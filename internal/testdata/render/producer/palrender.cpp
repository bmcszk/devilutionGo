// Offline fixture producer for the dungeon tile micro renderer
// (devilGo-gk5e / TestRenderTileGolden). Compiled BY HAND, never run by
// `go test` or CI, against /tmp/dvlx-pin (pinned upstream ca91d778f with
// Source/engine/render/dun_render.cpp's per-pixel-lighting option read
// replaced by a constant false - identical to the benchmark's InitOnce pin,
// dun_render_benchmark.cpp:65).
//
// Algorithm code is included verbatim (dun_render.cpp, palette_blending.cpp,
// palette_kd_tree.cpp, str_cat.cpp). Every frame fed to the renderer is
// fabricated here: unique, mechanically decodable patterns per data layout, so
// the dumps pin the upstream DISPATCH AND GEOMETRY (which rows, which source
// bytes, which blit), not any retail asset. The palette is a fabricated grey
// ramp and the blend lookup is upstream's own GenerateBlendedLookupTable over
// it; the light tables are fabricated pointer-tables (content (i + k*13) &
// 0xFF) whose pointer membership drives the FullyLit/FullyDark/PartiallyLit
// dispatch exactly as the benchmark's tables do.
//
// Usage: palrender <outdir>
#include <array>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <filesystem>
#include <fstream>
#include <string>
#include <vector>

#include <SDL.h>

#include "engine/render/dun_render.hpp"
#include "engine/render/dun_render.cpp"
#include "utils/palette_blending.cpp"
#include "utils/palette_kd_tree.cpp"
#include "utils/str_cat.cpp"

namespace devilution {
[[noreturn]] void app_fatal(std::string_view) { std::abort(); }

// verbatim: engine/render/light_render.cpp:499-511 (the ctor only; the rest of
// the TU pulls in the game systems).
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

using namespace devilution;

constexpr int SurfaceW = 640;
constexpr int SurfaceH = 480;
// Upstream placement: RenderTileFrame's origin is the tile's bottom-center
// pixel? The dump records the exact region each variant wrote, so placement
// conventions are pinned by the region rows in the dump itself.
constexpr int TileX = 320;
constexpr int TileY = 240;

// Dump region: generous diamond margin.
constexpr int DumpX0 = TileX - 64;
constexpr int DumpX1 = TileX + 63;
constexpr int DumpY0 = TileY - 96;
constexpr int DumpY1 = TileY + 16;

constexpr uint8_t Background = 0xAA; // never used by a fabricated frame

SDLSurfaceUniquePtr SdlSurface;

std::array<std::array<uint8_t, 256>, NumLightingLevels> tables;

void BuildLightTables()
{
	for (size_t k = 0; k < tables.size(); ++k) {
		for (int i = 0; i < 256; ++i) {
			tables[k][i] = uint8_t((i + static_cast<int>(k) * 13) & 0xFF);
		}
	}
}

// ---- fabricated frames (RE-ENCODED /trtnn layouts) -------------------------

std::array<uint8_t, 1024> squareFrame;
std::vector<uint8_t> transparentSquareStream;
std::array<uint8_t, 512> leftTriangleFrame;
std::array<uint8_t, 512> rightTriangleFrame;
std::array<uint8_t, 784> leftTrapezoidFrame;
std::array<uint8_t, 784> rightTrapezoidFrame;

void BuildSquare()
{
	for (unsigned i = 0; i < squareFrame.size(); ++i) {
		squareFrame[i] = uint8_t((i + 1) & 0xFF);
	}
}

// 32 rows bottom-up; row r (0 = bottom): alternate solid runs of 3 and
// transparent runs of 4, values = 1 + ((r * 7 + x) & 0x7F). Every row ends
// with a solid run, so the RLE walk lands exactly on the row edge.
void BuildTransparentSquare()
{
	transparentSquareStream.clear();
	for (unsigned r = 0; r < 32; ++r) {
		int x = 0;
		while (x < 32) {
			const int solid = 3;
			transparentSquareStream.push_back(uint8_t(solid));
			for (int j = 0; j < solid && x + j < 32; ++j) {
				transparentSquareStream.push_back(uint8_t(1 + ((r * 7 + x + j) & 0x7F)));
			}
			x += solid;
			int trans = std::min(4, 32 - x);
			transparentSquareStream.push_back(uint8_t(-int8_t(trans)));
			x += trans;
		}
	}
	transparentSquareStream.resize(transparentSquareStream.size());
}

// Triangle base (the lower 16 rows, widths 2..32), then upper 15 rows
// (widths 30..2), every byte = 1 + ((row * 4 + px) & 0x7F).
template <size_t N>
void FillTriangle(std::array<uint8_t, N> &frame, int // left/right flavour
)
{
	size_t out = 0;
	int width = 2;
	for (int row = 0; row < 16; ++row, width += 2) {
		for (int j = 0; j < width; ++j) {
			frame[out++] = uint8_t(1 + ((row * 4 + j) & 0x7F));
		}
	}
	width = 30;
	for (int row = 0; row < 15; ++row, width -= 2) {
		for (int j = 0; j < width; ++j) {
			frame[out++] = uint8_t(1 + ((row * 4 + j) & 0x7F));
		}
	}
}

void BuildTrapezoids()
{
	for (unsigned i = 0; i < 272; ++i) {
		leftTrapezoidFrame[i] = leftTriangleFrame[i];
		rightTrapezoidFrame[i] = rightTriangleFrame[i];
	}
	for (int r = 0; r < 16; ++r) {
		for (int x = 0; x < 32; ++x) {
			uint8_t v = uint8_t(1 + ((r * 6 + x) & 0x7F));
			leftTrapezoidFrame[272 + static_cast<unsigned>(r * 32 + x)] = v;
			rightTrapezoidFrame[272 + static_cast<unsigned>(r * 32 + x)] = v;
		}
	}
}

// RLE text: "v:count" runs with singleton values bare, lines joined by \n.
void DumpRegion(std::ostream &file)
{
	auto *pixels = static_cast<uint8_t *>(SdlSurface->pixels);
	for (int y = DumpY0; y <= DumpY1; ++y) {
		int runValue = -1;
		int runCount = 0;
		bool first = true;
		auto flush = [&]() {
			if (runCount == 0) {
				return;
			}
			if (!first) {
				file << ' ';
			}
			first = false;
			if (runCount == 1) {
				file << int(runValue);
			} else {
				file << runValue << ':' << runCount;
			}
		};
		for (int x = DumpX0; x <= DumpX1; ++x) {
			const auto v = pixels[y * SurfaceW + x];
			if (runCount > 0 && v == runValue) {
				++runCount;
				continue;
			}
			flush();
			runValue = v;
			runCount = 1;
		}
		flush();
		file << '\n';
	}
}

int main(int argc, char **argv)
{
	const char *outDir = argc > 1 ? argv[1] : "/tmp/dumper/out";
	std::filesystem::create_directories(outDir);

	SDL_Color palette[256];
	for (int i = 0; i < 256; ++i) {
		palette[i] = SDL_Color { Uint8(i), Uint8(i), Uint8(i), SDL_ALPHA_OPAQUE };
	}
	GenerateBlendedLookupTable(palette, -1, -1);
	BuildLightTables();

	SdlSurface.reset(SDL_CreateRGBSurfaceWithFormat(0, SurfaceW, SurfaceH, 8, SDL_PIXELFORMAT_INDEX8));
	if (SdlSurface == nullptr) {
		std::fprintf(stderr, "SDL Error: %s\n", SDL_GetError());
		return 1;
	}
	devilution::Surface out(SdlSurface.get());

	BuildSquare();
	BuildTransparentSquare();
	FillTriangle(leftTriangleFrame, 0);
	FillTriangle(rightTriangleFrame, 1);
	BuildTrapezoids();

	devilution::Lightmap lightmap(nullptr, 1, {}, 1, tables, tables[0].data(), tables.back().data());

	struct Case {
		const uint8_t *src;
		TileType tile;
		int height;
		std::string tileName;
	};
	std::vector<std::pair<Case, MaskType>> cases = {
		{ { squareFrame.data(), TileType::Square, DunFrameHeight, "Square" }, MaskType::Solid },
		{ { squareFrame.data(), TileType::Square, DunFrameHeight, "Square" }, MaskType::Transparent },
		{ { transparentSquareStream.data(), TileType::TransparentSquare, DunFrameHeight, "TransparentSquare" }, MaskType::Solid },
		{ { transparentSquareStream.data(), TileType::TransparentSquare, DunFrameHeight, "TransparentSquare" }, MaskType::Transparent },
		{ { transparentSquareStream.data(), TileType::TransparentSquare, DunFrameHeight, "TransparentSquare" }, MaskType::Left },
		{ { transparentSquareStream.data(), TileType::TransparentSquare, DunFrameHeight, "TransparentSquare" }, MaskType::Right },
		{ { leftTriangleFrame.data(), TileType::LeftTriangle, DunFrameTriangleHeight, "LeftTriangle" }, MaskType::Solid },
		{ { leftTriangleFrame.data(), TileType::LeftTriangle, DunFrameTriangleHeight, "LeftTriangle" }, MaskType::Transparent },
		{ { rightTriangleFrame.data(), TileType::RightTriangle, DunFrameTriangleHeight, "RightTriangle" }, MaskType::Solid },
		{ { rightTriangleFrame.data(), TileType::RightTriangle, DunFrameTriangleHeight, "RightTriangle" }, MaskType::Transparent },
		{ { leftTrapezoidFrame.data(), TileType::LeftTrapezoid, DunFrameHeight, "LeftTrapezoid" }, MaskType::Solid },
		{ { leftTrapezoidFrame.data(), TileType::LeftTrapezoid, DunFrameHeight, "LeftTrapezoid" }, MaskType::Transparent },
		{ { leftTrapezoidFrame.data(), TileType::LeftTrapezoid, DunFrameHeight, "LeftTrapezoid" }, MaskType::Left },
		{ { rightTrapezoidFrame.data(), TileType::RightTrapezoid, DunFrameHeight, "RightTrapezoid" }, MaskType::Solid },
		{ { rightTrapezoidFrame.data(), TileType::RightTrapezoid, DunFrameHeight, "RightTrapezoid" }, MaskType::Transparent },
		{ { rightTrapezoidFrame.data(), TileType::RightTrapezoid, DunFrameHeight, "RightTrapezoid" }, MaskType::Right },
	};

	std::ofstream file(std::string(outDir) + "/rtf_goldens.tsv");
	file << "# devilGo-gk5e TestRenderTileGolden goldens, produced by the offline\n";
	file << "# C++ dumper (palrender.cpp) against /tmp/dvlx-pin @ ca91d778f.\n";
	file << "# region x [" << DumpX0 << ", " << DumpX1 << "] y [" << DumpY0 << ", " << DumpY1 << "], background " << int(Background) << "\n";
	(void)out;

	for (const auto &[tile, mask] : cases) {
		const char *maskNames[] = { "Solid", "Transparent", "Right", "Left" };
		const char *lightNames[] = { "FullyLit", "PartiallyLit", "FullyDark" };
		const int lightLevels[] = { 0, 5, NumLightingLevels - 1 };
		for (int li = 0; li < 3; ++li) {
			const int lightLevel = lightLevels[li];
			memset(SdlSurface->pixels, Background, SurfaceW * SurfaceH);
			RenderTileFrame(out, lightmap, Point { TileX, TileY }, tile.tile, tile.src, tile.height, mask, tables[size_t(lightLevel)].data());
			file << "---- " << tile.tileName << " " << maskNames[int(mask)] << " " << lightNames[li] << "\n";
			DumpRegion(file);
		}
	}

	// black tile (world_draw_black_tile), one dump.
	memset(SdlSurface->pixels, Background, SurfaceW * SurfaceH);
	world_draw_black_tile(out, TileX, TileY);
	file << "---- BlackTile\n";
	DumpRegion(file);

	std::printf("wrote %s/rtf_goldens.tsv\n", outDir);
	return 0;
}
