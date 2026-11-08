// Offline fixture producer for the palette-LUT / light-variant / blit parity
// goldens. Never run by `go test` or CI: it compiles the PINNED upstream TUs
// (/tmp/dvlx-pinned @ ca91d778) and writes the reference bytes to ./raw.
//
// Algorithm code is included verbatim (palette_blending.cpp,
// palette_kd_tree.cpp, blit_impl.hpp). The only hand-written block is the
// verbatim copy of Lightmap's ctor (engine/render/light_render.cpp:499-511),
// because the blit flavours need the object but the rest of light_render.cpp
// pulls in the dungeon tile rasteriser.
#include <algorithm>
#include <array>
#include <cstdint>
#include <cstdio>
#include <cstring>
#include <fstream>
#include <iterator>
#include <string>
#include <vector>

#include <SDL.h>

#include "engine/palette.h"
#include "engine/render/blit_impl.hpp"
#include "utils/palette_kd_tree.hpp"

namespace devilution {
// engine/palette.cpp:39 storage for the palette header's extern.
std::array<SDL_Color, 256> logical_palette;

// verbatim: engine/render/light_render.cpp:499-511
Lightmap::Lightmap(const uint8_t *outBuffer, uint16_t outPitch,
    std::span<const uint8_t> lightmapBuffer, uint16_t lightmapPitch,
    std::span<const std::array<uint8_t, LightTableSize>, NumLightingLevels> lightTables,
    const uint8_t *fullyLitLightTable, const uint8_t *fullyDarkLightTable)
    : outBuffer(outBuffer)
    , outPitch(outPitch)
    , lightmapBuffer(lightmapBuffer)
    , lightmapPitch(lightmapPitch)
    , lightTables(lightTables)
    , fullyLitLightTable_(fullyLitLightTable)
    , fullyDarkLightTable_(fullyDarkLightTable)
{
}
} // namespace devilution

#include "utils/palette_blending.cpp"
#include "utils/palette_kd_tree.cpp"

using namespace devilution;

namespace {

void WriteFile(const std::string &path, const std::vector<uint8_t> &data)
{
	std::ofstream f(path, std::ios::binary);
	f.write(reinterpret_cast<const char *>(data.data()), static_cast<std::streamsize>(data.size()));
}

std::string HexOf(const std::vector<uint8_t> &v)
{
	static const char *digits = "0123456789abcdef";
	std::string s;
	s.reserve(v.size() * 2);
	for (uint8_t b : v) {
		s.push_back(digits[b >> 4]);
		s.push_back(digits[b & 0xF]);
	}
	return s;
}

std::string Sha256Of(const std::vector<uint8_t> &v)
{
	const std::string path = "/tmp/oracle/.hash.tmp";
	WriteFile(path, v);
	const std::string cmd = "sha256sum " + path;
	FILE *p = popen(cmd.c_str(), "r");
	char buf[128] = {};
	if (p == nullptr) return "popen-failed";
	(void)fgets(buf, sizeof(buf), p);
	pclose(p);
	return std::string(buf, 64);
}

std::vector<uint8_t> ReadFile(const std::string &path)
{
	std::ifstream f(path, std::ios::binary);
	return std::vector<uint8_t>((std::istreambuf_iterator<char>(f)), std::istreambuf_iterator<char>());
}

void LoadPaletteBytes(const std::vector<uint8_t> &raw, SDL_Color pal[256])
{
	for (int i = 0; i < 256; ++i) {
		pal[i].r = raw[i * 3];
		pal[i].g = raw[i * 3 + 1];
		pal[i].b = raw[i * 3 + 2];
		pal[i].a = 255;
	}
}

std::array<std::array<uint8_t, 256>, 16> ReadLightTables(const std::string &path)
{
	// The committed golden is the 16x256 grid flattened to 256 rows of 256
	// decimals (internal/light/tables_test.go decodeTSV).
	std::array<uint8_t, 4096> flat {};
	size_t n = 0;
	std::ifstream f(path);
	std::string line;
	while (std::getline(f, line)) {
		if (line.empty() || line[0] == '#') continue;
		for (size_t i = 0; i < line.size();) {
			while (i < line.size() && (line[i] == ' ' || line[i] == '\t')) ++i;
			int value = 0;
			bool any = false;
			while (i < line.size() && line[i] >= '0' && line[i] <= '9') {
				value = value * 10 + (line[i] - '0');
				++i;
				any = true;
			}
			if (any) flat[n++] = static_cast<uint8_t>(value);
		}
	}
	if (n != flat.size()) { std::fprintf(stderr, "light tables: %zu values, want %zu\n", n, flat.size()); std::exit(1); }
	std::array<std::array<uint8_t, 256>, 16> tables {};
	for (size_t i = 0; i < flat.size(); ++i) tables[i / 256][i % 256] = flat[i];
	return tables;
}

// G4 variant mapping: variant[shade][index] = palette[lightTables[shade][index]].
std::vector<uint8_t> VariantBytes(const SDL_Color pal[256], const std::array<std::array<uint8_t, 256>, 16> &tables)
{
	std::vector<uint8_t> out;
	out.reserve(16 * 256 * 3);
	for (int shade = 0; shade < 16; ++shade) {
		for (int idx = 0; idx < 256; ++idx) {
			const SDL_Color &c = pal[tables[shade][idx]];
			out.push_back(c.r);
			out.push_back(c.g);
			out.push_back(c.b);
		}
	}
	return out;
}

std::vector<uint8_t> PaletteBytes(const SDL_Color pal[256])
{
	std::vector<uint8_t> out(768);
	for (int i = 0; i < 256; ++i) {
		out[i * 3] = pal[i].r;
		out[i * 3 + 1] = pal[i].g;
		out[i * 3 + 2] = pal[i].b;
	}
	return out;
}

std::vector<uint8_t> LutBytes()
{
	return std::vector<uint8_t>(&paletteTransparencyLookup[0][0], &paletteTransparencyLookup[0][0] + 65536);
}

// engine/palette.cpp:52-64 verbatim rotate shape (both axes, then rows).
void CycleColors(int from, int to)
{
	std::rotate(logical_palette.begin() + from, logical_palette.begin() + from + 1, logical_palette.begin() + to + 1);
	for (auto &palette : paletteTransparencyLookup) {
		std::rotate(std::begin(palette) + from, std::begin(palette) + from + 1, std::begin(palette) + to + 1);
	}
	std::rotate(&paletteTransparencyLookup[from][0], &paletteTransparencyLookup[from + 1][0], &paletteTransparencyLookup[to + 1][0]);
}

// engine/palette.cpp:72-84 verbatim rotate shape.
void CycleColorsReverse(int from, int to)
{
	std::rotate(logical_palette.begin() + from, logical_palette.begin() + to, logical_palette.begin() + to + 1);
	for (auto &palette : paletteTransparencyLookup) {
		std::rotate(std::begin(palette) + from, std::begin(palette) + to, std::begin(palette) + to + 1);
	}
	std::rotate(&paletteTransparencyLookup[from][0], &paletteTransparencyLookup[to][0], &paletteTransparencyLookup[to + 1][0]);
}

struct BlitCase {
	std::string name;
	std::vector<uint8_t> src;
	bool hasSrc;
	int fill;
	bool useLight;
	bool useMap;
	bool blend;
};

} // namespace

int main(int argc, char **argv)
{
	if (argc < 4) {
		std::fprintf(stderr, "usage: oracle <outdir> <lighttables.tsv> <pal-dir> [pal:skipFrom:skipTo ...]\n");
		return 2;
	}
	const std::string out = argv[1];
	const std::string palDir = argv[3];
	const auto tables = ReadLightTables(argv[2]);

	// ---- 1. blend LUT per palette + skip range ----
	for (int i = 4; i < argc; ++i) {
		const std::string spec = argv[i];
		const size_t c1 = spec.find(':');
		const size_t c2 = spec.find(':', c1 + 1);
		const std::string name = spec.substr(0, c1);
		const int skipFrom = std::stoi(spec.substr(c1 + 1, c2 - c1 - 1));
		const int skipTo = std::stoi(spec.substr(c2 + 1));
		SDL_Color pal[256];
		LoadPaletteBytes(ReadFile(palDir + "/" + name + ".pal"), pal);
		GenerateBlendedLookupTable(pal, skipFrom, skipTo);
		WriteFile(out + "/lut_" + name + ".bin", LutBytes());
	}
	// LUT for the same palettes with both skip values out of range: the fill.
	// (skipFrom/skipTo of -1/-1 is GenerateBlendedLookupTable's default.)

	// ---- 2. blit flavour trace (>=3 light shades) ----
	SDL_Color blitPal[256];
	LoadPaletteBytes(ReadFile(palDir + "/diablo.pal"), blitPal);
	GenerateBlendedLookupTable(blitPal);
	WriteFile(out + "/variants_diablo.bin", VariantBytes(blitPal, tables));
	{
		// G4 composition: variant_trn[shade][i] = palette[tables[shade][trn[i]]].
		const std::vector<uint8_t> raw = ReadFile(palDir + "/pause.trn");
		std::array<uint8_t, 256> trn {};
		std::memcpy(trn.data(), raw.data(), 256);
		std::vector<uint8_t> out2;
		out2.reserve(16 * 256 * 3);
		for (int shade = 0; shade < 16; ++shade)
			for (int idx = 0; idx < 256; ++idx) {
				const SDL_Color &c = blitPal[tables[shade][trn[idx]]];
				out2.push_back(c.r);
				out2.push_back(c.g);
				out2.push_back(c.b);
			}
		WriteFile(out + "/variants_diablo_pause.bin", out2);
	}

	{
		const std::vector<uint8_t> lightmap = {0, 1, 3, 5, 8, 12, 14, 15, 0, 3, 12, 15};
		const std::vector<uint8_t> dstIn = {9, 40, 199, 0, 255, 7, 128, 31, 200, 4, 90, 61};
		const std::vector<uint8_t> src = {199, 199, 199, 7, 7, 7, 0, 255, 12, 34, 200, 3};
		std::array<uint8_t, 256> trn {};
		{
			const std::vector<uint8_t> raw = ReadFile(palDir + "/pause.trn");
			std::memcpy(trn.data(), raw.data(), 256);
		}
		const int fillColor = 37;
		// The lightmap resolves a destination pointer back to its row, so it must
		// be built over the very buffer the flavour writes into.
		std::array<uint8_t, 12> dstBuf {};
		std::memcpy(dstBuf.data(), dstIn.data(), dstIn.size());
		Lightmap lm(dstBuf.data(), static_cast<uint16_t>(dstIn.size()), lightmap,
		    static_cast<uint16_t>(lightmap.size()),
		    std::span<const std::array<uint8_t, 256>, 16>(tables), nullptr, nullptr);

		std::vector<BlitCase> cases;
		cases.push_back({"direct", src, true, -1, false, false, false});
		cases.push_back({"withmap", src, true, -1, false, true, false});
		cases.push_back({"blended", src, true, -1, false, false, true});
		cases.push_back({"blended_withmap", src, true, -1, false, true, true});

		std::ofstream trace(out + "/blit_trace.tsv");
		trace << "# Blit flavour trace over the committed cathedral light tables\n";
		trace << "# (internal/testdata/gamedata/lighttables/cathedral_tables.tsv). One row\n";
		trace << "# per upstream blit flavour of engine/render/blit_impl.hpp:20-190.\n";
		trace << "# Columns: flavour, length, src, dst, light, fillColor, want.\n";
		trace << "# src/dst/light/want are hex byte strings; '-' means unused.\n";
		trace << "# flavour\tlength\tsrc\tdst\tlight\tfillColor\twant\n";
		auto emit = [&](const BlitCase &c, const std::vector<uint8_t> &want) {
			trace << c.name << '\t' << dstIn.size() << '\t'
			      << (c.hasSrc ? HexOf(c.src) : std::string("-")) << '\	'
			      << HexOf(dstIn) << '\t'
			      << (c.useLight ? HexOf(lightmap) : std::string("-")) << '\t'
			      << c.fill << '\t' << HexOf(want) << '\n';
		};

		const unsigned len = static_cast<unsigned>(dstIn.size());
		for (const BlitCase &c : cases) {
			std::vector<uint8_t> d(dstIn);
			if (c.blend && c.useMap) {
				if (c.hasSrc) BlitBlendedWithMap { trn.data() }(len, d.data(), c.src.data());
				else BlitBlendedWithMap { trn.data() }(len, static_cast<uint8_t>(c.fill), d.data());
			} else if (c.blend) {
				if (c.hasSrc) BlitBlended {}(len, d.data(), c.src.data());
				else BlitBlended {}(len, static_cast<uint8_t>(c.fill), d.data());
			} else if (c.useMap) {
				if (c.hasSrc) BlitWithMap { trn.data() }(len, d.data(), c.src.data());
				else BlitWithMap { trn.data() }(len, static_cast<uint8_t>(c.fill), d.data());
			} else {
				if (c.hasSrc) BlitDirect {}(len, d.data(), c.src.data());
				else BlitDirect {}(len, static_cast<uint8_t>(c.fill), d.data());
			}
			emit(c, d);
		}
		BlitCase lmCase {"lightmap", src, true, -1, true, false, false};
		{
			std::memcpy(dstBuf.data(), dstIn.data(), dstIn.size());
			BlitWithLightmap { lm }(len, dstBuf.data(), src.data());
			emit(lmCase, std::vector<uint8_t>(dstBuf.begin(), dstBuf.end()));
		}
		{
			std::memcpy(dstBuf.data(), dstIn.data(), dstIn.size());
			BlitBlendedWithLightmap { lm }(len, dstBuf.data(), src.data());
			lmCase.name = "blended_lightmap";
			emit(lmCase, std::vector<uint8_t>(dstBuf.begin(), dstBuf.end()));
		}
	}

	// ---- 3. cycle after-state (palette + variants + LUT) ----
	{
		LoadPaletteBytes(ReadFile(palDir + "/diablo.pal"), logical_palette.data());
		GenerateBlendedLookupTable(logical_palette.data());
		std::ofstream steps(out + "/cycle_steps.tsv");
		steps << "# Cycle after-state. One line per step.\n";
		steps << "# step, palette bytes, variant sha256, LUT sha256. The variant rows are\n";
		steps << "# re-derived from the cycled palette (G4 mapping: palette[table[i]]).\n";
		steps << "# step\tpalette\tvariants_sha256\tlut_sha256\n";
		auto capture = [&](const std::string &name) {
			const std::vector<uint8_t> pal = PaletteBytes(logical_palette.data());
			const std::vector<uint8_t> variants = VariantBytes(logical_palette.data(), tables);
			steps << name << '\t' << HexOf(pal) << '\t' << Sha256Of(variants) << '\t' << Sha256Of(LutBytes()) << '\n';
		};
		capture("init");
		CycleColors(1, 31);
		capture("cycle 1..31");
		CycleColorsReverse(16, 31);
		capture("cycle 16..31 reverse");
		CycleColors(1, 8);
		capture("cycle 1..8");
	}
	return 0;
}
