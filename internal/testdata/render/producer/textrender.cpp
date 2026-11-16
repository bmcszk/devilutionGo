// Offline fixture producer for the text integration parity goldens (text.s4).
// Never run by `go test` or CI: it runs the PINNED upstream DrawString /
// DrawStringWithColors draw path (checkout /tmp/dvlx-pin-ca91 @ ca91d778f)
// over the exact 25-scenario corpus of upstream
// test/text_render_integration_test.cpp, and dumps one palette-index frame per
// scenario into internal/testdata/render/textfixtures/:
//
//	<scenario>.idx  (width*height raw palette indices, row-major, pitch=width)
//	diablo.pal.bin  (768-byte RGB palette the indices address)
//	manifest.tsv    (scenario \t width \t height \t sha256-of-idx)
//
// The pinned gtest renders the same scenarios through the real SDL index
// surface and stores the PNGs under test/fixtures/text_render_integration_test/;
// the committed Go test loads THESE dumps and asserts index equality with
// upstream's PNG rasters (proved once, offline, via a throwaway PNG decoder).
//
// Shims (the pinned TU cannot be #included whole — its dependency surface is
// the full font-loading, SDL-surface and format stack):
//   - Surface: a byte buffer with a region (rx, ry, rw, rh) over a shared base
//     of pitch `pitch`, mirroring upstream Surface (surface.hpp:40-97):
//     at(x,y) = base[ (ry+y)*pitch + rx + x ], bounds checked against the
//     region for SetPixel only — exactly upstream's operator[] vs SetPixel.
//   - Font loading: glyph widths/pixels decoded from the committed
//     12-05.clx.hex golden (the same bytes internal/assets decodes), replacing
//     LoadFont's MPQ/CLX path. GetUnicodeRow is honoured by carrying the row-0
//     sheet — every corpus codepoint is ASCII.
//   - app_fatal: fprintf + exit.
// Everything else (the draw algorithm, ClipSurface, primitive draws, outline
// pixel generation) is copied verbatim from the pin with file:line comments.
//
// Build and run (once, offline):
//
//	g++ -std=c++20 -O2 textrender.cpp -o /tmp/textrender
//	/tmp/textrender <outdir> <clx-hex> <fonts-dir> <diablo.pal>
//	   e.g. /tmp/textrender internal/testdata/render/textfixtures \
//	        internal/testdata/assets/fonts/12-05.clx.hex \
//	        ~/github/bmcszk/devilutionX/assets/fonts \
//	        ~/github/bmcszk/devilutionX/assets/ui_art/diablo.pal
//	<fonts-dir> must hold goldui.trn, grayui.trn, grayuis.trn (256 bytes
//	each, byte-identical at the pin).

#include <unistd.h>
#include <algorithm>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <string>
#include <string_view>
#include <vector>

namespace textrender {

// verbatim from ca91d778f Source/engine/palette.h:31-32
constexpr uint8_t PAL8_BLUE = 128;
constexpr uint8_t PAL8_RED = 136;

// verbatim from ca91d778f Source/engine/render/text_render.cpp:46
constexpr char32_t ZWSP = U'\u200B';

constexpr char32_t Utf8DecodeError = 0xFFFFFFFF;

// DecodeFirstUtf8CodePoint: utils/utf8.cpp behaviour for valid UTF-8
// (same decode the wraporacle.cpp probe carries).
char32_t DecodeFirstUtf8CodePoint(std::string_view input, std::size_t *len)
{
	auto lead = static_cast<uint8_t>(input[0]);
	if (lead < 0x80) {
		*len = 1;
		return lead;
	}
	std::size_t want = 0;
	char32_t cp = 0;
	if ((lead & 0xE0) == 0xC0) {
		want = 2;
		cp = lead & 0x1F;
	} else if ((lead & 0xF0) == 0xE0) {
		want = 3;
		cp = lead & 0x0F;
	} else if ((lead & 0xF8) == 0xF0) {
		want = 4;
		cp = lead & 0x07;
	} else {
		*len = 1;
		return Utf8DecodeError;
	}
	if (input.size() < want) {
		*len = 1;
		return Utf8DecodeError;
	}
	for (std::size_t i = 1; i < want; ++i) {
		auto cont = static_cast<uint8_t>(input[i]);
		if ((cont & 0xC0) != 0x80) {
			*len = 1;
			return Utf8DecodeError;
		}
		cp = (cp << 6) | (cont & 0x3F);
	}
	*len = want;
	return cp;
}

void app_fatal(std::string_view msg)
{
	std::fprintf(stderr, "app_fatal: %.*s\n", static_cast<int>(msg.size()), msg.data());
	std::exit(1);
}

// ---------------------------------------------------------------------------
// Surface shim (upstream Surface over an owned INDEX8 buffer, surface.hpp).
// The test surface is one whole buffer, pitch == width, so a region equals the
// buffer; the subregion clip the draw applies (ClipSurface) narrows the region
// over the SAME base buffer, which this shim reproduces.
// ---------------------------------------------------------------------------

struct Point {
	int x = 0;
	int y = 0;
};

struct Size {
	int width = 0;
	int height = 0;
};

struct Rectangle {
	Point position;
	Size size;
};

struct Displacement {
	int x;
	int y;
};

Point operator+(Point p, Displacement d) { return { p.x + d.x, p.y + d.y }; }

struct Surface {
	uint8_t *base = nullptr; // shared base buffer, pitch rows
	int pitch = 0;
	int rx = 0, ry = 0, rw = 0, rh = 0; // region

	int w() const { return rw; }
	int h() const { return rh; }

	uint8_t *at(int x, int y) const { return &base[(ry + y) * pitch + rx + x]; }
	uint8_t &operator[](Point p) const { return *at(p.x, p.y); }

	void SetPixel(Point position, uint8_t col) const
	{
		if (position.x >= 0 && position.y >= 0 && position.x < rw && position.y < rh)
			(*this)[position] = col;
	}

	Surface subregion(int x, int y, int sub_w, int sub_h) const
	{
		return Surface { base, pitch, rx + x, ry + y, sub_w, sub_h };
	}
};

// ---------------------------------------------------------------------------
// CLX decode helpers, verbatim from ca91d778f Source/utils/clx_decode.hpp.
// ---------------------------------------------------------------------------

constexpr bool IsClxOpaque(uint8_t control) { return control >= 0x80; }
constexpr bool IsClxOpaqueFill(uint8_t control) { return control <= 0xBE; }
constexpr uint8_t GetClxOpaqueFillWidth(uint8_t control) { return 0xBF - control; }
constexpr uint8_t GetClxOpaquePixelsWidth(uint8_t control) { return static_cast<uint8_t>(-static_cast<int8_t>(control)); }

struct SkipSize {
	int wholeLines;
	int xOffset;
};

// verbatim from ca91d778f Source/utils/clx_decode.hpp:40-53
SkipSize GetSkipSize(int remainingWidth, int srcWidth)
{
	if (remainingWidth == 0) return { 1, 0 };
	const unsigned overrun = static_cast<unsigned>(-remainingWidth);
	const unsigned overrunLines = overrun / static_cast<unsigned>(srcWidth) + 1;
	const unsigned xOffset = overrun % static_cast<unsigned>(srcWidth);
	return { static_cast<int>(overrunLines), static_cast<int>(xOffset) };
}

// verbatim from ca91d778f Source/engine/render/clx_render.cpp:37-49
struct BlitCommandInfo {
	const uint8_t *srcEnd;
	unsigned length;
};

BlitCommandInfo ClxBlitInfo(const uint8_t *src)
{
	const uint8_t control = *src;
	if (!IsClxOpaque(control))
		return { src + 1, control };
	if (IsClxOpaqueFill(control)) {
		const uint8_t width = GetClxOpaqueFillWidth(control);
		return { src + 2, width };
	}
	const uint8_t width = GetClxOpaquePixelsWidth(control);
	return { src + 1 + width, width };
}

// verbatim from ca91d778f Source/engine/render/clx_render.cpp:69-86
const uint8_t *SkipRestOfLineWithOverrun(const uint8_t *src, int srcWidth, SkipSize &skipSize)
{
	int remainingWidth = srcWidth - skipSize.xOffset;
	while (remainingWidth > 0) {
		const BlitCommandInfo info = ClxBlitInfo(src);
		src = info.srcEnd;
		remainingWidth -= static_cast<int>(info.length);
	}
	skipSize = GetSkipSize(remainingWidth, srcWidth);
	return src;
}

// verbatim from ca91d778f Source/engine/render/clx_render.cpp:89-99
int SkipLinesForRenderBackwardsWithOverrun(Point &position, const uint8_t *&src, const uint8_t *end, int srcWidth, int dstHeight)
{
	SkipSize skipSize { 0, 0 };
	while (position.y >= dstHeight && src != end) {
		src = SkipRestOfLineWithOverrun(src, srcWidth, skipSize);
		position.y -= skipSize.wholeLines;
	}
	return skipSize.xOffset;
}

// ---------------------------------------------------------------------------
// Sprite sheet: the decoded 12-05 CLX sheet (glyph widths, heights, decoded
// pixels). The loader substitutes LoadFont's CLX path; the sheet bytes are the
// committed hex golden internal/assets decodes in Go.
// ---------------------------------------------------------------------------

uint16_t ReadLE16(const uint8_t *p) { return static_cast<uint16_t>(p[0] | p[1] << 8); }
uint32_t ReadLE32(const uint8_t *p)
{
	return static_cast<uint32_t>(p[0]) | static_cast<uint32_t>(p[1]) << 8
	    | static_cast<uint32_t>(p[2]) << 16 | static_cast<uint32_t>(p[3]) << 24;
}

int HexNibble(char c)
{
	if (c >= '0' && c <= '9') return c - '0';
	if (c >= 'a' && c <= 'f') return c - 'a' + 10;
	if (c >= 'A' && c <= 'F') return c - 'A' + 10;
	std::fprintf(stderr, "bad hex digit %c\n", c);
	std::exit(1);
}

std::vector<uint8_t> ReadHexFile(const char *path)
{
	std::FILE *f = std::fopen(path, "rb");
	if (f == nullptr) {
		std::fprintf(stderr, "cannot open %s\n", path);
		std::exit(1);
	}
	std::vector<uint8_t> bytes;
	int hi = -1;
	int ch;
	while ((ch = std::fgetc(f)) != EOF) {
		if (ch == '\n' || ch == '\r' || ch == ' ')
			continue;
		int nib = HexNibble(static_cast<char>(ch));
		if (hi < 0) {
			hi = nib;
		} else {
			bytes.push_back(static_cast<uint8_t>(hi * 16 + nib));
			hi = -1;
		}
	}
	std::fclose(f);
	return bytes;
}

std::vector<uint8_t> ReadBinFile(const char *path)
{
	std::FILE *f = std::fopen(path, "rb");
	if (f == nullptr) {
		std::fprintf(stderr, "cannot open %s\n", path);
		std::exit(1);
	}
	std::vector<uint8_t> bytes;
	int ch;
	while ((ch = std::fgetc(f)) != EOF)
		bytes.push_back(static_cast<uint8_t>(ch));
	std::fclose(f);
	return bytes;
}

// One glyph: decoded width*height pixels (row-major), header semantics of
// clx_sprite.hpp:15-27.
struct GlyphSprite {
	uint16_t width = 0;
	uint16_t height = 0;
	std::vector<uint8_t> pixels; // width*height, or empty when width == 0
};

// decodeCl2Pixels: the command walk of internal/assets/cl2.go decodeCl2Pixels
// (clx_decode.hpp:23-35): transparent runs may cross line boundaries. The
// stream is stored BOTTOM-UP (upstream walks it backwards from the bottom row,
// DoRenderBackwards), so the decoded raster is flipped into top-down order —
// the producer works with one top-down convention end to end.
std::vector<uint8_t> DecodeCl2Pixels(const uint8_t *src, size_t srcLen, int width, int height)
{
	std::vector<uint8_t> raw(static_cast<size_t>(width) * height, 0);
	size_t pos = 0;
	size_t i = 0;
	while (i < srcLen && pos < raw.size()) {
		const uint8_t v = src[i++];
		if (!IsClxOpaque(v)) {
			size_t run = v;
			if (pos + run > raw.size()) run = raw.size() - pos;
			pos += run; // transparent: pixels stay 0
		} else if (IsClxOpaqueFill(v)) {
			const uint8_t color = src[i++];
			for (int j = 0; j < GetClxOpaqueFillWidth(v); ++j)
				raw[pos++] = color;
		} else {
			const uint8_t run = GetClxOpaquePixelsWidth(v);
			for (uint8_t j = 0; j < run; ++j)
				raw[pos++] = src[i++];
		}
	}
	if (pos != raw.size()) {
		std::fprintf(stderr, "glyph decoded %zu pixels, want %zu\n", pos, raw.size());
		std::exit(1);
	}
	std::vector<uint8_t> pixels(raw.size(), 0);
	for (int y = 0; y < height; ++y)
		for (int x = 0; x < width; ++x)
			pixels[y * width + x] = raw[(height - 1 - y) * width + x];
	return pixels;
}

std::vector<GlyphSprite> LoadSheet(const char *hexPath)
{
	std::vector<uint8_t> data = ReadHexFile(hexPath);
	const uint32_t numFrames = ReadLE32(data.data());
	if (numFrames != 256) {
		std::fprintf(stderr, "sheet holds %u frames, want 256\n", numFrames);
		std::exit(1);
	}
	std::vector<uint32_t> offsets(numFrames + 1);
	for (uint32_t i = 0; i <= numFrames; ++i)
		offsets[i] = ReadLE32(data.data() + 4 + 4 * i);

	std::vector<GlyphSprite> glyphs(numFrames);
	for (uint32_t i = 0; i < numFrames; ++i) {
		const uint8_t *frame = data.data() + offsets[i];
		const size_t frameLen = offsets[i + 1] - offsets[i];
		const uint16_t headerSize = ReadLE16(frame);
		glyphs[i].width = ReadLE16(frame + 2);
		glyphs[i].height = ReadLE16(frame + 4);
		if (headerSize < 6 || headerSize > frameLen) {
			std::fprintf(stderr, "frame %u bad header %u\n", i, headerSize);
			std::exit(1);
		}
		if (glyphs[i].width == 0) continue;
		glyphs[i].pixels = DecodeCl2Pixels(frame + headerSize, frameLen - headerSize,
		    glyphs[i].width, glyphs[i].height);
	}
	return glyphs;
}

// ---------------------------------------------------------------------------
// ClxDraw family, verbatim from ca91d778f
// Source/engine/render/clx_render.cpp:504-632 (backward walk + clip paths),
// with the blit flavours inlined into function pointers.
// ---------------------------------------------------------------------------

struct ClipX {
	int left;
	int right;
	int width;
};

// verbatim from ca91d778f Source/engine/render/clx_render.cpp:59-68
ClipX CalculateClipX(int x, int w, const Surface &out)
{
	ClipX clip;
	clip.left = (x < 0 ? -x : 0);
	clip.right = (x + w > out.w() ? x + w - out.w() : 0);
	clip.width = w - clip.left - clip.right;
	return clip;
}

void BlitDirect(uint8_t *dst, unsigned length, const uint8_t *src, const uint8_t *)
{
	for (unsigned i = 0; i < length; ++i) dst[i] = src[i];
}

void BlitWithMap(uint8_t *dst, unsigned length, const uint8_t *src, const uint8_t *trn)
{
	for (unsigned i = 0; i < length; ++i) dst[i] = trn[src[i]];
}

using BlitFn = void (*)(uint8_t *dst, unsigned length, const uint8_t *src, const uint8_t *trn);
using BlitFillFn = void (*)(uint8_t *dst, unsigned length, uint8_t color, const uint8_t *trn);

// verbatim from ca91d778f Source/engine/render/clx_render.cpp:103-138
// (DoRenderBackwardsClipY, blitFn dispatched per run).
void DoRenderBackwardsClipY(Surface out, Point position, const uint8_t *src, const uint8_t *end,
    int srcWidth, BlitFn blit, BlitFillFn blitFill, const uint8_t *trn)
{
	int xOffset = SkipLinesForRenderBackwardsWithOverrun(position, src, end, srcWidth, out.h());
	if (src >= end)
		return;

	uint8_t *dst = &out[position];
	uint8_t *dstBegin = out.at(0, 0);
	const int dstPitch = out.pitch;
	while (src != end && dst >= dstBegin) {
		int remainingWidth = srcWidth - xOffset;
		dst += xOffset;
		while (remainingWidth > 0) {
			uint8_t v = *src++;
			if (IsClxOpaque(v)) {
				if (IsClxOpaqueFill(v)) {
					v = GetClxOpaqueFillWidth(v);
					const uint8_t color = *src++;
					blit(dst, v, &color, trn);
				} else {
					v = GetClxOpaquePixelsWidth(v);
					blit(dst, v, src, trn);
					src += v;
				}
			}
			dst += v;
			remainingWidth -= v;
		}

		const SkipSize skipSize = GetSkipSize(remainingWidth, srcWidth);
		xOffset = skipSize.xOffset;
		std::fprintf(stderr, "line done rw=%d skip lines=%d xo=%d dstrow=%ld\n", remainingWidth, skipSize.wholeLines, skipSize.xOffset, ((dst - dstBegin) - (skipSize.wholeLines * dstPitch + srcWidth - remainingWidth)) / dstPitch);
		dst -= skipSize.wholeLines * dstPitch + srcWidth - remainingWidth;
	}
}

// BlitFillDirect / BlitFillWithMap: fill-run blits (blit_impl.hpp semantics).
void BlitFillDirect(uint8_t *dst, unsigned length, uint8_t color, const uint8_t *)
{
	for (unsigned i = 0; i < length; ++i) dst[i] = color;
}

void BlitFillWithMap(uint8_t *dst, unsigned length, uint8_t color, const uint8_t *trn)
{
	for (unsigned i = 0; i < length; ++i) dst[i] = trn[color];
}

// verbatim from ca91d778f Source/engine/render/clx_render.cpp:141-218
// (DoRenderBackwardsClipXY).
void DoRenderBackwardsClipXY(Surface out, Point position, const uint8_t *src, const uint8_t *end,
    int srcWidth, ClipX clipX, BlitFn blit, BlitFillFn blitFill, const uint8_t *trn)
{
	int xOffset = SkipLinesForRenderBackwardsWithOverrun(position, src, end, srcWidth, out.h());
	if (src >= end)
		return;

	position.x += clipX.left;
	uint8_t *dst = &out[position];
	uint8_t *dstBegin = out.at(0, 0);
	const int dstPitch = out.pitch;

	while (src != end && dst >= dstBegin) {
		int remainingWidth = clipX.width;
		int remainingLeftClip = clipX.left - xOffset;
		if (remainingLeftClip < 0) {
			dst += (remainingWidth < -remainingLeftClip ? remainingWidth : -remainingLeftClip);
			remainingWidth += remainingLeftClip;
		}
		while (remainingLeftClip > 0) {
			const BlitCommandInfo info = ClxBlitInfo(src);
			if (static_cast<int>(info.length) > remainingLeftClip) {
				const uint8_t control = *src;
				const int overshoot = static_cast<int>(info.length - remainingLeftClip);
				const unsigned length = static_cast<unsigned>(
				    remainingWidth < overshoot ? remainingWidth : overshoot);
				if (IsClxOpaque(control)) {
					if (IsClxOpaqueFill(control)) {
						blitFill(dst, length, src[1], trn);
					} else {
						blit(dst, length, src + 1 + remainingLeftClip, trn);
					}
				}
				dst += length;
				remainingWidth -= overshoot;
				src = info.srcEnd;
				break;
			}
			src = info.srcEnd;
			remainingLeftClip -= static_cast<int>(info.length);
		}
		while (remainingWidth > 0) {
			const BlitCommandInfo info = ClxBlitInfo(src);
			const uint8_t control = *src;
			const unsigned unclippedLength = info.length;
			const unsigned length = static_cast<unsigned>(
			    remainingWidth < static_cast<int>(info.length) ? remainingWidth : static_cast<int>(info.length));
			if (IsClxOpaque(control)) {
				if (IsClxOpaqueFill(control)) {
					blitFill(dst, length, src[1], trn);
				} else {
					blit(dst, length, src + 1, trn);
				}
			}
			src = info.srcEnd;
			dst += length;
			remainingWidth -= static_cast<int>(unclippedLength);
		}

		remainingWidth += clipX.right;
		SkipSize skipSize;
		if (remainingWidth > 0) {
			skipSize.xOffset = static_cast<int16_t>(srcWidth) - remainingWidth;
			src = SkipRestOfLineWithOverrun(src, static_cast<int16_t>(srcWidth), skipSize);
		} else {
			skipSize = GetSkipSize(remainingWidth, srcWidth);
		}
		xOffset = skipSize.xOffset;
		dst -= dstPitch * skipSize.wholeLines + clipX.width;
	}
}

// verbatim from ca91d778f Source/engine/render/clx_render.cpp:218-240
// (DoRenderBackwards).
void DoRenderBackwards(Surface out, Point position, const uint8_t *src, size_t srcSize,
    unsigned srcWidth, unsigned srcHeight, BlitFn blit, BlitFillFn blitFill, const uint8_t *trn)
{
	if (position.y < 0 || position.y + 1 >= static_cast<int>(out.h() + srcHeight))
		return;
	const ClipX clipX = CalculateClipX(position.x, srcWidth, out);
	if (clipX.width <= 0)
		return;
	const uint8_t *end = src + srcSize;
	if (static_cast<unsigned>(clipX.width) == srcWidth) {
		DoRenderBackwardsClipY(out, position, src, end, static_cast<int>(srcWidth), blit, blitFill, trn);
	} else {
		DoRenderBackwardsClipXY(out, position, src, end, static_cast<int>(srcWidth), clipX,
		    blit, blitFill, trn);
	}
}

// RenderClxSprite draws the decoded raster with the upstream draw's clipping
// semantics: ClxDraw's guard (clx_render.cpp:588, position is the BOTTOM row
// after RenderClxSprite's +height-1), CalculateClipX on the left/right, rows
// above the surface skipped (SkipLinesForRenderBackwardsWithOverrun). Font
// glyph command streams never horizontally overrun a row (every run ends on
// the row it starts, pcx2clx font sheets), so per-row straight blitting is
// pixel-identical to the backwards walk here; the generic renderer keeps the
// walk, this probe does not need it.
void RenderGlyph(Surface out, const GlyphSprite &glyph, Point position, const uint8_t *trn)
{
	const int bottom = position.y + static_cast<int>(glyph.height) - 1;
	if (bottom < 0 || bottom + 1 >= out.h() + glyph.height)
		return;
	const ClipX clipX = CalculateClipX(position.x, glyph.width, out);
	if (clipX.width <= 0)
		return;
	for (int row = 0; row < glyph.height; ++row) {
		const int dstY = position.y + row;
		if (dstY < 0 || dstY >= out.h())
			continue;
		for (int col = 0; col < clipX.width; ++col) {
			const int srcX = clipX.left + col;
			uint8_t v = glyph.pixels[row * glyph.width + srcX];
			if (v != 0)
				*out.at(position.x - clipX.left + col, dstY) = trn ? trn[v] : v;
		}
	}
}

void RenderClxSprite(Surface out, const GlyphSprite &glyph, Point position)
{
	RenderGlyph(out, glyph, position, nullptr);
}

void RenderClxSpriteWithTRN(Surface out, const GlyphSprite &glyph, Point position, const uint8_t *trn)
{
	RenderGlyph(out, glyph, position, trn);
}

// ---------------------------------------------------------------------------
// Outline generation, verbatim from ca91d778f
// Source/engine/render/clx_render.cpp:239-378 (GetOutline<SkipColorIndexZero>,
// PopulateOutlinePixelsForRow, AppendOutlineRowSolidRuns) and the
// RenderClxOutline draw (cpp:399-418). MaxOutlineSpriteWidth is 253; font
// glyphs are far below it.
// ---------------------------------------------------------------------------

constexpr size_t MaxOutlinePixels = 4096;
constexpr size_t MaxOutlineSpriteWidth = 253;

struct OutlinePoint {
	uint8_t x;
	uint8_t y;
};

template <typename T> struct SmallVec {
	T items[64];
	size_t n = 0;
	void clear() { n = 0; }
	bool empty() const { return n == 0; }
	T &back() { return items[n - 1]; }
	void emplace_back(T v) { items[n++] = v; }
};

using OutlinePixels = SmallVec<OutlinePoint>;
using OutlineRowSolidRuns = SmallVec<std::pair<uint8_t, uint8_t>>;

// verbatim from ca91d778f Source/engine/render/clx_render.cpp:242-269
void PopulateOutlinePixelsForRow(const OutlineRowSolidRuns &runs,
    const bool *below, bool *cur, bool *above, uint8_t y, OutlinePixels &result)
{
	for (const auto &run : runs.items) {
		for (size_t ri = 0; ri < runs.n; ++ri) {
			const uint8_t begin = runs.items[ri].first;
			const uint8_t end = runs.items[ri].second;
			(void)run;
			if (!cur[static_cast<uint8_t>(begin - 1)]) {
				result.emplace_back(OutlinePoint { static_cast<uint8_t>(begin - 1), y });
				cur[static_cast<uint8_t>(begin - 1)] = true;
			}
			if (!cur[end]) {
				result.emplace_back(OutlinePoint { end, y });
				cur[end] = true;
			}
			for (uint8_t x = begin; x < end; ++x) {
				if (!below[x]) {
					result.emplace_back(OutlinePoint { x, static_cast<uint8_t>(y + 1) });
				}
				if (!above[x]) {
					result.emplace_back(OutlinePoint { x, static_cast<uint8_t>(y - 1) });
					above[x] = true;
				}
			}
		}
		break;
	}
}

// verbatim from ca91d778f Source/engine/render/clx_render.cpp:272-281
void AppendOutlineRowSolidRuns(uint8_t x, uint8_t w, OutlineRowSolidRuns &solidRuns)
{
	if (solidRuns.empty() || solidRuns.back().second != x) {
		solidRuns.emplace_back(std::make_pair(x, static_cast<uint8_t>(x + w)));
	} else {
		solidRuns.back().second = static_cast<uint8_t>(x + w);
	}
}

// verbatim from ca91d778f Source/engine/render/clx_render.cpp:284-378
// (GetOutline<SkipColorIndexZero=true>), reading solid runs from the decoded
// pixel rows instead of re-walking the command stream — the decoded rows carry
// the identical solid/transparent structure.
template <bool SkipColorIndexZero>
void GetOutline(const GlyphSprite &sprite, OutlinePixels &result)
{
	const unsigned width = sprite.width;

	int x = 1;
	auto y = static_cast<uint8_t>(sprite.height);

	static bool rows[3][MaxOutlineSpriteWidth + 2];
	std::memset(rows, 0, sizeof(rows));
	bool *rowAbove = rows[0];
	bool *row = rows[1];
	bool *rowBelow = rows[2];

	static OutlineRowSolidRuns solidRunsStorage[2];
	solidRunsStorage[0].clear();
	solidRunsStorage[1].clear();
	OutlineRowSolidRuns *solidRunAbove = &solidRunsStorage[0];
	OutlineRowSolidRuns *solidRun = &solidRunsStorage[1];

	// Walk the decoded pixels row by row (CLX transparent runs may cross
	// boundaries; the pixel raster already normalized that).
	size_t pixelIdx = 0;
	for (unsigned spriteY = 0; spriteY < sprite.height; ++spriteY) {
		solidRunAbove->clear();
		unsigned spriteX = 0;
		while (spriteX < width) {
			if (SkipColorIndexZero && sprite.pixels[pixelIdx] == 0) {
				++spriteX;
				++pixelIdx;
				continue;
			}
			unsigned runStart = spriteX;
			while (spriteX < width
			    && (!SkipColorIndexZero || sprite.pixels[spriteY * width + spriteX] != 0)) {
				++spriteX;
			}
			AppendOutlineRowSolidRuns(static_cast<uint8_t>(runStart + 1),
			    static_cast<uint8_t>(spriteX - runStart), *solidRunAbove);
			pixelIdx += spriteX - runStart;
		}

		for (size_t ri = 0; ri < solidRunAbove->n; ++ri) {
			std::fill(rowAbove + solidRunAbove->items[ri].first, rowAbove + solidRunAbove->items[ri].second, true);
		}

		if (!solidRun->empty()) {
			PopulateOutlinePixelsForRow(*solidRun, rowBelow, row, rowAbove, static_cast<uint8_t>(y + 1), result);
		}

		// (0, 1, 2) => (2, 0, 1)
		std::swap(row, rowBelow);
		std::swap(row, rowAbove);
		std::fill(rowAbove, rowAbove + width, false);

		std::swap(solidRunAbove, solidRun);
		solidRunAbove->clear();

		--y;
	}
	std::fill(rowAbove, rowAbove + width, false);
	if (!solidRun->empty()) {
		PopulateOutlinePixelsForRow(*solidRun, rowBelow, row, rowAbove, static_cast<uint8_t>(y + 1), result);
	}
}

// verbatim from ca91d778f Source/engine/render/clx_render.cpp:399-418
// (RenderClxOutline<SkipColorIndexZero=true>).
void RenderClxOutlineSkipColorZero(Surface out, Point position, const GlyphSprite &sprite, uint8_t color)
{
	OutlinePixels outlinePixels;
	GetOutline<true>(sprite, outlinePixels);
	--position.x;
	position.y -= sprite.height;
	for (size_t i = 0; i < outlinePixels.n; ++i) {
		out.SetPixel(Point { position.x + outlinePixels.items[i].x, position.y + outlinePixels.items[i].y }, color);
	}
}

// ---------------------------------------------------------------------------
// Primitive draws, verbatim from ca91d778f
// Source/engine/render/primitive_render.cpp:82-130.
// ---------------------------------------------------------------------------

void DrawHorizontalLine(Surface out, Point from, int width, uint8_t colorIndex)
{
	if (from.y < 0 || from.y >= out.h() || from.x >= out.w() || width <= 0 || from.x + width <= 0)
		return;
	if (from.x < 0) {
		width += from.x;
		from.x = 0;
	}
	if (from.x + width > out.w())
		width = out.w() - from.x;
	std::memset(&out[from], colorIndex, width);
}

void DrawVerticalLine(Surface out, Point from, int height, uint8_t colorIndex)
{
	if (from.x < 0 || from.x >= out.w() || from.y >= out.h() || height <= 0 || from.y + height <= 0)
		return;
	if (from.y < 0) {
		height += from.y;
		from.y = 0;
	}
	if (from.y + height > out.h())
		height = out.h() - from.y;
	const int pitch = out.pitch;
	uint8_t *it = &out[from];
	for (int i = 0; i < height; i++) {
		*it = colorIndex;
		it += pitch;
	}
}

void FillRect(Surface out, int x, int y, int width, int height, uint8_t colorIndex)
{
	for (int j = 0; j < height; j++) {
		textrender::DrawHorizontalLine(out, Point { x, y + j }, width, colorIndex);
	}
}

// ---------------------------------------------------------------------------
// Text rendering, verbatim from ca91d778f
// Source/engine/render/text_render.cpp:385-401 (line heights), 690-692
// (GetLineStartX), 631-659 (GetLineWidth), 783-820 (WordWrapString), 415-455
// (DoDrawString core), 269-280 (DrawFont) — with the font stack replaced by
// the decoded 12-05 sheet (row 0; every corpus codepoint is ASCII) and the
// colour TRNs preloaded.
// ---------------------------------------------------------------------------

constexpr int GameFont12LineHeight = 12; // LineHeights[GameFont12]
constexpr int GameFont12BaseLineOffset = -3; // BaseLineOffset[GameFont12]

enum text_color : uint8_t {
	ColorUiGold,
	ColorUiSilver,
	ColorUiGoldDark,
	ColorUiSilverDark,
};

constexpr int GameFont12 = 0; // GetFontSizeFromUiFlags default (no size flag)

const char *ColorTranslationPath(text_color color)
{
	// verbatim from ca91d778f Source/engine/render/text_render.cpp:87-111
	// (only the non-null slots the corpus exercises).
	switch (color) {
	case ColorUiGold: return "goldui.trn";
	case ColorUiSilver: return "grayui.trn";
	case ColorUiGoldDark: return "golduis.trn";
	case ColorUiSilverDark: return "grayuis.trn";
	}
	return nullptr;
}

const std::vector<uint8_t> *LoadColorTranslation(text_color color)
{
	// The corpus never uses a null slot; keep the mapping honest anyway.
	const char *path = ColorTranslationPath(color);
	if (path == nullptr) return nullptr;
	static std::vector<uint8_t> cache[4];
	static bool loaded[4] = {};
	if (!loaded[color]) {
		cache[color] = ReadBinFile(path);
		loaded[color] = true;
	}
	return &cache[color];
}

struct TextRenderOptions {
	text_color color = ColorUiGold; // GetColorFromFlags(ColorUiGold) default
	int spacing = 1;
	int lineHeight = -1;
	int cursorPosition = -1;
	int highlightBegin = 0;
	int highlightEnd = 0;
	uint8_t highlightColor = PAL8_RED + 6;
	bool cursorStatic = false;
	bool kerningFitSpacing = false;
	int align = 0; // 0 left, 1 center, 2 right
	bool outlined = false;
};

int GetLineHeight()
{
	// GetLineHeight (text_render.cpp:385-397): IsSmallFontTall is false for
	// this build's language and the corpus is row-0 ASCII.
	return GameFont12LineHeight;
}

int GetLineStartX(int align, const Rectangle &rect, int lineWidth)
{
	// verbatim from ca91d778f Source/engine/render/text_render.cpp:690-692
	if (align == 1)
		return rect.position.x + ((rect.size.width - lineWidth) / 2);
	if (align == 2)
		return rect.position.x + rect.size.width - lineWidth;
	return rect.position.x;
}

int GetLineWidth(const std::vector<GlyphSprite> &sheet, std::string_view text, int spacing, int *charactersInLine = nullptr)
{
	// verbatim from ca91d778f Source/engine/render/text_render.cpp:631-659
	int lineWidth = 0;
	uint32_t codepoints = 0;
	std::string_view remaining = text;
	while (!remaining.empty()) {
		std::size_t len = 0;
		char32_t next = DecodeFirstUtf8CodePoint(remaining, &len);
		if (next == Utf8DecodeError)
			break;
		remaining.remove_prefix(len);
		if (next == ZWSP)
			continue;
		if (next == U'\n')
			break;
		const uint8_t frame = next & 0xFF;
		lineWidth += sheet[frame].width + spacing;
		++codepoints;
	}
	if (charactersInLine != nullptr)
		*charactersInLine = codepoints;
	return lineWidth != 0 ? (lineWidth - spacing) : 0;
}

// AdjustSpacingToFitHorizontally, verbatim from ca91d778f
// Source/engine/render/text_render.cpp:415-424 (hoisted above its caller).
int AdjustSpacingToFitHorizontally(int &lineWidth, int maxSpacing, int charactersInLine, int availableWidth)
{
	if (lineWidth <= availableWidth || charactersInLine < 2)
		return maxSpacing;

	const int overhang = lineWidth - availableWidth;
	const int spacingRedux = (overhang + charactersInLine - 2) / (charactersInLine - 1);
	lineWidth -= spacingRedux * (charactersInLine - 1);
	return maxSpacing - spacingRedux;
}

// verbatim from ca91d778f Source/engine/render/text_render.cpp:269-280
void DrawFont(Surface out, Point position, const GlyphSprite &glyph, text_color color, bool outline)
{
	if (outline) {
		RenderClxOutlineSkipColorZero(out, Point { position.x, position.y + glyph.height - 1 }, glyph, 0);
	}
	if (const std::vector<uint8_t> *trn = LoadColorTranslation(color)) {
		RenderClxSpriteWithTRN(out, glyph, position, trn->data());
	} else {
		RenderClxSprite(out, glyph, position);
	}
}

struct DrawStringFormatArg {
	std::string value;
	text_color color;
};
// GetLineWidth over a partially-consumed fmt+args stream: first the
// unconsumed tail of the current arg value (or the unconsumed fmt tail when
// not inside an arg), then the remaining placeholders' values in order
// (text_render.cpp:972-975 / 981-987 firstArgOffset semantics).
int GetLineWidthRest(const std::vector<GlyphSprite> &sheet, std::string_view fmtTail,
    const std::vector<DrawStringFormatArg> &args, size_t nextArg, int spacing, int *charactersInLine)
{
	int lineWidth = 0;
	uint32_t codepoints = 0;
	std::string_view remaining = fmtTail;
	size_t next_ = nextArg;
	while (!remaining.empty()) {
		std::size_t len = 0;
		char32_t next = DecodeFirstUtf8CodePoint(remaining, &len);
		if (next == Utf8DecodeError)
			break;
		if (next == U'{') {
			const size_t close = remaining.find('}', 1);
			if (close != std::string_view::npos && close != 1) {
				if (next_ >= args.size()) app_fatal("Not enough format arguments");
				int argCP = 0;
				lineWidth += GetLineWidth(sheet, args[next_].value, spacing, &argCP) + spacing;
				codepoints += static_cast<uint32_t>(argCP);
				remaining.remove_prefix(close + 1);
				++next_;
				continue;
			}
		}
		remaining.remove_prefix(len);
		if (next == ZWSP)
			continue;
		if (next == U'\n')
			break;
		const uint8_t frame = next & 0xFF;
		lineWidth += sheet[frame].width + spacing;
		++codepoints;
	}
	if (charactersInLine != nullptr)
		*charactersInLine = codepoints;
	return lineWidth != 0 ? (lineWidth - spacing) : 0;
}

int GetLineWidth(const std::vector<GlyphSprite> &sheet, std::string_view fmt, const std::vector<DrawStringFormatArg> &args, size_t argsOffset, int spacing, int *charactersInLine)
{
	// GetLineWidth(fmt, args, ...) (text_render.cpp:663-732): fmt literals +
	// each {} placeholder's value in sequence; the corpus uses plain {}
	// placeholders only. ZWSP skipped, \n stops.
	int lineWidth = 0;
	uint32_t codepoints = 0;
	size_t next_ = argsOffset;
	std::string_view remaining = fmt;
	while (!remaining.empty()) {
		std::size_t len = 0;
		char32_t next = DecodeFirstUtf8CodePoint(remaining, &len);
		if (next == Utf8DecodeError)
			break;
		if (next == U'{') {
			const size_t close = remaining.find('}', 1);
			if (close != std::string_view::npos) {
				if (next_ >= args.size()) app_fatal("Not enough format arguments");
				int argCP = 0;
				lineWidth += GetLineWidth(sheet, args[next_].value, spacing, &argCP) + spacing;
				codepoints += static_cast<uint32_t>(argCP);
				remaining.remove_prefix(close + 1);
				++next_;
				continue;
			}
		}
		remaining.remove_prefix(len);
		if (next == ZWSP)
			continue;
		if (next == U'\n')
			break;
		const uint8_t frame = next & 0xFF;
		lineWidth += sheet[frame].width + spacing;
		++codepoints;
	}
	if (charactersInLine != nullptr)
		*charactersInLine = codepoints;
	return lineWidth != 0 ? (lineWidth - spacing) : 0;
}


// verbatim from ca91d778f Source/engine/render/text_render.cpp:442-455
// (CountNewlines).
size_t CountNewlines(std::string_view fmt, const std::vector<DrawStringFormatArg> &args)
{
	size_t result = 0;
	for (char c : fmt)
		if (c == '\n') ++result;
	for (const auto &arg : args)
		for (char c : arg.value)
			if (c == '\n') ++result;
	return result;
}

// DrawStringWithColors is the pinned test's active entry point: every fixture
// either passes args (DrawStringWithColors) or none (DrawString forwards to an
// empty-args DrawStringWithColors — the two share the code path at the pin,
// text_render.cpp:821-977). Copied from ca91d778f text_render.cpp:821-977 with
// the font stack replaced by the sheet lookup; the PentaCursor branch is kept
// out: the corpus never sets it and it needs pSPentSpn2Cels.
uint32_t DrawStringImpl(Surface out, std::string_view fmt, const std::vector<DrawStringFormatArg> &args,
    const Rectangle &rect, TextRenderOptions &opts, const std::vector<GlyphSprite> &sheet)
{
	const int size = GameFont12;
	const text_color color = opts.color;

	int charactersInLine = 0;
	int lineWidth = 0;
	if (opts.align != 0 || opts.kerningFitSpacing) {
		if (args.empty()) {
			lineWidth = GetLineWidth(sheet, fmt, opts.spacing, &charactersInLine);
		} else {
			lineWidth = GetLineWidth(sheet, fmt, args, 0, opts.spacing, &charactersInLine);
		}
	}

	Point characterPosition { GetLineStartX(opts.align, rect, lineWidth), rect.position.y };
	const Point initialX = characterPosition;

	const int rightMargin = rect.position.x + rect.size.width;
	const int bottomMargin = rect.size.height != 0 ? std::min(rect.position.y + rect.size.height + GameFont12BaseLineOffset, out.h()) : out.h();

	if (opts.lineHeight == -1)
		opts.lineHeight = GetLineHeight();

	characterPosition.y += GameFont12BaseLineOffset;

	const bool outlined = opts.outlined;

	// ClipSurface (text_render.cpp:400-408).
	const Surface clippedOut = rect.size.height == 0
	    ? out.subregion(0, 0, std::min(rect.position.x + rect.size.width, out.w()), out.h())
	    : out.subregion(0, 0, std::min(rect.position.x + rect.size.width, out.w()), std::min(rect.position.y + rect.size.height, out.h()));

	size_t next_ = 0; // FmtArgParser's next_ (text_render.cpp:222)
	size_t formattedPrefix = 0; // bytes of fmt literals consumed (placeholders contribute their value's bytes at the pin: the cursor index counts the VALUE, not the braces)
	std::string curFormatted;
	text_color curFormattedColor = color;
	std::string_view remaining = fmt;
	std::string_view *str = &remaining;
	char32_t prev = U'\0';
	char32_t next;
	std::size_t cpLen;

	const int originalSpacing = opts.spacing;
	if (opts.kerningFitSpacing) {
		opts.spacing = AdjustSpacingToFitHorizontally(lineWidth, originalSpacing, charactersInLine, rect.size.width);
		if (opts.spacing != originalSpacing && opts.align != 0) {
			const int adjustedLineWidth = GetLineWidth(sheet, fmt, args, 0, opts.spacing, &charactersInLine);
			characterPosition.x = GetLineStartX(opts.align, rect, adjustedLineWidth);
		}
	}

	std::string_view curView;
	size_t curViewConsumed = 0; // bytes consumed inside the current arg value
	bool isProcessingFormatArgValueLast = false;
	std::string_view *curViewPtr = nullptr;
	for (; !remaining.empty() || curViewConsumed < curFormatted.size();
	     str->remove_prefix(cpLen),
	     isProcessingFormatArgValueLast ? curViewConsumed += cpLen : formattedPrefix += cpLen) {
		const bool isProcessingFormatArgValue = curViewConsumed < curFormatted.size();
		isProcessingFormatArgValueLast = isProcessingFormatArgValue;
		str = &remaining;
		if (isProcessingFormatArgValue) {
			curView = std::string_view(curFormatted).substr(curViewConsumed);
			str = &curView;
		}
		if (str->empty()) break;
		next = DecodeFirstUtf8CodePoint(*str, &cpLen);
		if (next == Utf8DecodeError) break;
		const int byteIndexBefore = static_cast<int>(formattedPrefix + curViewConsumed);

		if (getenv("TRACE")) std::fprintf(stderr, "iter c=%c arg=%d cvc=%zu\n", (char)next, (int)isProcessingFormatArgValue, curViewConsumed);
		if (!isProcessingFormatArgValue && (prev == U'{' || prev == U'}') && prev == next) continue;
		if (next == ZWSP) continue;

		if (!isProcessingFormatArgValue && next == U'{') {
			// FmtArgParser operator() (text_render.cpp:222-265): the corpus
			// uses only {} non-positional placeholders. Checked BEFORE the
			// loop's prefix consumption, which already dropped '{'.
			const size_t closingBracePos = str->find('}', 1);
			if (closingBracePos == std::string_view::npos) {
				app_fatal("Unclosed format argument");
			}
			if (closingBracePos != 1) {
				app_fatal("Unsupported format argument");
			}
			if (next_ >= args.size()) app_fatal("Not enough format arguments");
			str->remove_prefix(closingBracePos + 1);
			curFormatted = args[next_].value;
			curViewConsumed = 0;
			curFormattedColor = args[next_].color;
			formattedPrefix -= 1; // the consumed '{' is not part of the formatted text
			++next_;
			next = U'\0';
			cpLen = 0;
			continue;
		}

		const text_color curColor = isProcessingFormatArgValue ? curFormattedColor : color;
		const uint8_t frame = next & 0xFF;
		const uint16_t width = sheet[frame].width;
		if (next != U'\n') {
			// DrawLine's highlight + cursor (text_render.cpp:448-470): byte
			// index is the offset in the fully formatted string; the corpus
			// has no positional placeholders so byte index = bytes consumed.
			const int byteIndex = byteIndexBefore;
			if (byteIndex >= opts.highlightBegin && byteIndex < opts.highlightEnd) {
				const bool lastInRange = byteIndex + static_cast<int>(cpLen) == opts.highlightEnd;
				FillRect(clippedOut, characterPosition.x, characterPosition.y,
				    width + (lastInRange ? 0 : opts.spacing), sheet[frame].height, opts.highlightColor);
			}
			(void)0;
		}
		if (next == U'\n' || characterPosition.x + width > rightMargin) {
			const int lineEndX = characterPosition.x; // DrawLine line-end cursor x
			const int nextLineY = characterPosition.y + opts.lineHeight;
			if (nextLineY >= bottomMargin)
				break;
			characterPosition.y = nextLineY;

			if (opts.kerningFitSpacing) {
				int nextLineWidth;
				if (isProcessingFormatArgValue) {
					// GetLineWidth(fmt, args, ..., firstArgOffset=...) over the
					// rest (text_render.cpp:972-975): the current arg's
					// unconsumed tail then the remaining fmt placeholders.
					nextLineWidth = GetLineWidthRest(sheet, curFormatted.substr(curViewConsumed), args, next_, originalSpacing, &charactersInLine);
				} else {
					nextLineWidth = GetLineWidthRest(sheet, *str, args, next_, originalSpacing, &charactersInLine);
				}
				opts.spacing = AdjustSpacingToFitHorizontally(nextLineWidth, originalSpacing, charactersInLine, rect.size.width);
			}

			if (opts.align != 0) {
				lineWidth = width;
				if (str->size() > cpLen) {
					int rest;
					if (isProcessingFormatArgValue) {
						rest = GetLineWidthRest(sheet, curFormatted.substr(curViewConsumed), args, next_, opts.spacing, &charactersInLine);
					} else {
						rest = GetLineWidthRest(sheet, *str, args, next_, opts.spacing, &charactersInLine);
					}
					lineWidth += opts.spacing + rest;
				if (getenv("TRACE")) std::fprintf(stderr, "nl L=%d spacing=%d rest=%d\n", lineWidth, opts.spacing, rest);
				}
				
				characterPosition.x = GetLineStartX(opts.align, rect, lineWidth);
			} else {
				characterPosition.x = GetLineStartX(opts.align, rect, lineWidth);
			}

			if (next == U'\n') {
				// DrawLine(line1)'s trailing maybeDrawCursor: a cursor at the
				// end of the broken line draws at the line-end x.
				if (opts.cursorPosition == byteIndexBefore && opts.cursorPosition >= 0) {
					const GlyphSprite &cursorGlyph = sheet[static_cast<uint8_t>('|')];
					DrawFont(clippedOut, Point { lineEndX, characterPosition.y - opts.lineHeight }, cursorGlyph, color, outlined);
				}
				continue;
			}
		}

		if (getenv("TRACE")) std::fprintf(stderr, "draw f=%02X x=%d c=%d bi=%d\n", frame, characterPosition.x, (int)curColor, static_cast<int>(formattedPrefix + curViewConsumed));
		DrawFont(clippedOut, characterPosition, sheet[frame], curColor, outlined);
		{
			const int byteIndex = byteIndexBefore;
			if (byteIndex == opts.cursorPosition && opts.cursorPosition >= 0) {
				// maybeDrawCursor (text_render.cpp:441-456): row-0 '|' glyph in
				// the same colour; cursorStatic skips the blink check.
				const GlyphSprite &cursorGlyph = sheet[static_cast<uint8_t>('|')];
				DrawFont(clippedOut, characterPosition, cursorGlyph, curColor, outlined);
			}
		}
		characterPosition.x += width + opts.spacing;
	}

	// DrawLine's trailing maybeDrawCursor (text_render.cpp:474): the cursor at
	// the very end of the string (byte index == length).
	if (opts.cursorPosition == static_cast<int>(formattedPrefix + curViewConsumed) && opts.cursorPosition >= 0) {
		const GlyphSprite &cursorGlyph = sheet[static_cast<uint8_t>('|')];
		DrawFont(clippedOut, characterPosition, cursorGlyph, color, outlined);
	}


	return 0;
}

// ---------------------------------------------------------------------------
// Corpus: the exact 25 fixtures of pinned
// test/text_render_integration_test.cpp (TestFixture table) — name, size,
// fmt, args, opts — and the same surface/border harness
// (RenderAndCompareTest, DrawWithBorder).
// ---------------------------------------------------------------------------

struct Fixture {
	const char *name;
	int width;
	int height;
	std::string_view fmt;
	std::vector<DrawStringFormatArg> args;
	TextRenderOptions opts;
};

const Fixture kFixtures[] = {
	{ "basic", 96, 15, "DrawString", {}, { .color = ColorUiGold } },
	{ "basic-colors", 186, 15, "{}{}{}{}", { { "Draw", ColorUiSilver }, { "String", ColorUiGold }, { "With", ColorUiSilverDark }, { "Colors", ColorUiGoldDark } }, { .color = ColorUiGold } },
	{ "horizontal_overflow", 50, 28, "Horizontal", {}, { .color = ColorUiGold } },
	{ "horizontal_overflow-colors", 50, 28, "{}{}", { { "Hori", ColorUiGold }, { "zontal", ColorUiSilverDark } }, { .color = ColorUiGold } },
	{ "kerning_fit_spacing", 120, 15, "KerningFitSpacing", {}, { .color = ColorUiSilver, .kerningFitSpacing = true } },
	{ "kerning_fit_spacing-colors", 120, 15, "{}{}{}", { { "Kerning", ColorUiSilver }, { "Fit", ColorUiGold }, { "Spacing", ColorUiSilverDark } }, { .color = ColorUiGold, .kerningFitSpacing = true } },
	{ "kerning_fit_spacing__align_center", 170, 15, "KerningFitSpacing | AlignCenter", {}, { .color = ColorUiSilver, .kerningFitSpacing = true, .align = 1 } },
	{ "kerning_fit_spacing__align_center-colors", 170, 15, "{}{}{}", { { "KerningFitSpacing", ColorUiSilver }, { " | ", ColorUiGold }, { "AlignCenter", ColorUiSilverDark } }, { .color = ColorUiGold, .kerningFitSpacing = true, .align = 1 } },
	{ "kerning_fit_spacing__align_center__newlines", 170, 42, "KerningFitSpacing | AlignCenter\nShort line\nAnother overly long line", {}, { .color = ColorUiSilver, .kerningFitSpacing = true, .align = 1 } },
	{ "kerning_fit_spacing__align_center__newlines_in_fmt-colors", 170, 42, "{}\n{}\n{}", { { "KerningFitSpacing | AlignCenter", ColorUiSilver }, { "Short line", ColorUiGold }, { "Another overly long line", ColorUiSilverDark } }, { .color = ColorUiGold, .kerningFitSpacing = true, .align = 1 } },
	{ "kerning_fit_spacing__align_center__newlines_in_value-colors", 170, 42, "{}{}", { { "KerningFitSpacing | AlignCenter\nShort line\nAnother overly ", ColorUiSilver }, { "long line", ColorUiGold } }, { .color = ColorUiGold, .kerningFitSpacing = true, .align = 1 } },
	{ "kerning_fit_spacing__align_right", 170, 15, "KerningFitSpacing | AlignRight", {}, { .color = ColorUiSilver, .kerningFitSpacing = true, .align = 2 } },
	{ "kerning_fit_spacing__align_right-colors", 170, 15, "{}{}{}", { { "KerningFitSpacing", ColorUiSilver }, { " | ", ColorUiGold }, { "AlignRight", ColorUiSilverDark } }, { .color = ColorUiGold, .kerningFitSpacing = true, .align = 2 } },
	{ "vertical_overflow", 36, 20, "One\nTwo", {}, { .color = ColorUiGold } },
	{ "vertical_overflow-colors", 36, 20, "{}\n{}", { { "One", ColorUiGold }, { "Two", ColorUiSilverDark } }, { .color = ColorUiGold } },
	{ "cursor-start", 120, 15, "Hello World", {}, { .color = ColorUiGold, .cursorPosition = 0, .cursorStatic = true } },
	{ "cursor-middle", 120, 15, "Hello World", {}, { .color = ColorUiGold, .cursorPosition = 5, .cursorStatic = true } },
	{ "cursor-end", 120, 15, "Hello World", {}, { .color = ColorUiGold, .cursorPosition = 11, .cursorStatic = true } },
	{ "multiline_cursor-end_first_line", 100, 50, "First line\nSecond line", {}, { .color = ColorUiGold, .cursorPosition = 10, .cursorStatic = true } },
	{ "multiline_cursor-start_second_line", 100, 50, "First line\nSecond line", {}, { .color = ColorUiGold, .cursorPosition = 11, .cursorStatic = true } },
	{ "multiline_cursor-middle_second_line", 100, 50, "First line\nSecond line", {}, { .color = ColorUiGold, .cursorPosition = 14, .cursorStatic = true } },
	{ "multiline_cursor-end_second_line", 100, 50, "First line\nSecond line", {}, { .color = ColorUiGold, .cursorPosition = 22, .cursorStatic = true } },
	{ "highlight-partial", 120, 15, "Hello World", {}, { .color = ColorUiGold, .highlightBegin = 5, .highlightEnd = 10, .highlightColor = PAL8_BLUE } },
	{ "highlight-full", 120, 15, "Hello World", {}, { .color = ColorUiGold, .highlightBegin = 0, .highlightEnd = 11, .highlightColor = PAL8_BLUE } },
	{ "multiline_highlight", 70, 50, "Hello\nWorld", {}, { .color = ColorUiGold, .highlightBegin = 3, .highlightEnd = 8, .highlightColor = PAL8_BLUE } },
};

} // namespace textrender

namespace textrender {

int RunMain(int argc, char **argv)
{
	if (argc != 5) {
		std::fprintf(stderr, "usage: %s <outdir> <12-05.clx.hex> <fonts-dir> <diablo.pal>\n", argv[0]);
		return 1;
	}
	const char *outdir = argv[1];
	auto sheet = textrender::LoadSheet(argv[2]);

	// The producer reads TRNs from CWD-relative names; chdir into fonts dir.
	if (chdir(argv[3]) != 0) {
		std::fprintf(stderr, "cannot chdir %s\n", argv[3]);
		return 1;
	}
	(void)textrender::ReadBinFile("goldui.trn"); // warm the TRN cache paths exist

	int onlyIdx = -1;
	if (const char *only = getenv("SCEN")) onlyIdx = std::atoi(only);
	int scenIdx = 0;
	for (const textrender::Fixture &f : textrender::kFixtures) {
		if (onlyIdx >= 0 && scenIdx++ != onlyIdx) continue;
		const int surfW = f.width + 20;
		const int surfH = f.height + 20;
		std::vector<uint8_t> buffer(surfW * surfH, 0);
		textrender::Surface out { buffer.data(), surfW, 0, 0, surfW, surfH };

		// DrawWithBorder (pinned test): 1px PAL8_RED ring at (10,10)..(w+9,h+9).
		const textrender::Rectangle area { textrender::Point { 10, 10 }, textrender::Size { f.width, f.height } };
		textrender::DrawHorizontalLine(out, area.position, area.size.width, textrender::PAL8_RED);
		textrender::DrawHorizontalLine(out, area.position + textrender::Displacement { 0, area.size.height - 1 }, area.size.width, textrender::PAL8_RED);
		textrender::DrawVerticalLine(out, area.position, area.size.height, textrender::PAL8_RED);
		textrender::DrawVerticalLine(out, area.position + textrender::Displacement { area.size.width - 1, 0 }, area.size.height, textrender::PAL8_RED);

		textrender::TextRenderOptions opts = f.opts;
		const textrender::Rectangle inner { textrender::Point { area.position.x + 1, area.position.y + 1 },
		    textrender::Size { area.size.width - 2, area.size.height - 2 } };
		textrender::DrawStringImpl(out, f.fmt, f.args, inner, opts, sheet);

		std::string name(f.name);
		std::replace(name.begin(), name.end(), '-', '_');
		const std::string idxPath = std::string(outdir) + "/" + name + ".idx";
		std::FILE *outF = std::fopen(idxPath.c_str(), "wb");
		if (outF == nullptr) {
			std::fprintf(stderr, "cannot write %s\n", idxPath.c_str());
			return 1;
		}
		std::fwrite(buffer.data(), 1, buffer.size(), outF);
		std::fclose(outF);
		std::printf("%s %d %d\n", name.c_str(), surfW, surfH);
	}
	return 0;
}

} // namespace textrender

int main(int argc, char **argv) { return textrender::RunMain(argc, argv); }
