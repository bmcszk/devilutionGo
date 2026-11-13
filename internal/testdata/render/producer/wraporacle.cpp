// Offline fixture producer for the wrap-layout parity golden. Never run by
// `go test` or CI: it runs the PINNED upstream wrap/measure algorithms
// (checkout /tmp/dvlx-pin-ca91 @ ca91d778f, ADR-002 canonical) over the exact
// corpus internal/render/text_test.go pins, and dumps one TSV row per
// (string, line):
//
//	stringIdx \t lineIdx \t lineWidth \t startX \t lineText
//
// lineWidth is upstream GetLineWidth over the wrapped line; startX is the
// left-aligned cursor offset (line width + spacing), matching lineStartX with
// AlignLeft; lineText uses literal \n escapes so the TSV stays one row per
// line. Glyph widths are decoded at runtime from the committed 12-05 CLX hex
// golden (same file internal/assets decodes), so no widths are hand-written
// here. Route: verbatim copies of WordWrapString + GetLineWidth + their
// direct helpers (IsBreakableWhitespace, IsBreakAllowed, UTF-8 decode) —
// #include-ing the whole text_render.cpp would drag in the full font-loading
// and surface stack, an unbounded dependency surface for an offline probe.
//
// Build and run (once, offline):
//
//	g++ -std=c++20 -O2 wraporacle.cpp -o /tmp/wraporacle
//	/tmp/wraporacle internal/testdata/assets/fonts/12-05.clx.hex
//	    > internal/testdata/assets/fonts/wrap_golden.tsv
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <string>
#include <string_view>
#include <vector>

namespace wraporacle {

// verbatim from ca91d778f Source/engine/render/text_render.cpp:46
constexpr char32_t ZWSP = U'\u200B'; // Zero-width space

// Upstream's Utf8DecodeError sentinel (utils/sheen_bidi.hpp
// SBCodepointInvalid); DecodeFirstUtf8CodePoint returns it on malformed input
// and WordWrap breaks on it (text_render.cpp:772-774).
constexpr char32_t Utf8DecodeError = 0xFFFFFFFF;

// DecodeFirstUtf8CodePoint: utils/utf8.cpp:9-14 behaviour for valid UTF-8.
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

// verbatim from ca91d778f Source/engine/render/text_render.cpp:631-659, with
// the Utf8CodePoints range spelled as an explicit DecodeFirstUtf8CodePoint
// walk (identical codepoint sequence) and CurrentFont.load replaced by the
// decoded sheet lookup — the only substitution in this probe.
int GetLineWidth(const int *sheetWidths, std::string_view text, int spacing, int *charactersInLine)
{
	int lineWidth = 0;
	uint32_t codepoints = 0;
	std::string_view remaining = text;
	while (!remaining.empty()) {
		std::size_t len = 0;
		char32_t next = DecodeFirstUtf8CodePoint(remaining, &len);
		remaining.remove_prefix(len);
		if (next == ZWSP)
			continue;

		if (next == U'\n')
			break;

		const uint8_t frame = next & 0xFF;
		lineWidth += sheetWidths[frame] + spacing;
		++codepoints;
	}
	if (charactersInLine != nullptr)
		*charactersInLine = codepoints;

	return lineWidth != 0 ? (lineWidth - spacing) : 0;
}

// verbatim from ca91d778f Source/engine/render/text_render.cpp:1012-1015
bool IsBreakableWhitespace(char32_t c)
{
	return c == U' ' || c == U'　' || c == ZWSP;
}

// verbatim from ca91d778f Source/engine/render/text_render.cpp:286-289:
// IsFullWidthPunct is false for every corpus codepoint, so the helper reduces
// to false; kept for structural fidelity with the upstream call site.
bool IsBreakAllowed(char32_t /*codepoint*/, char32_t /*nextCodepoint*/)
{
	return false;
}

// verbatim from ca91d778f Source/engine/render/text_render.cpp:741-821, with
// the CurrentFont surface replaced by the sheet-width lookup (same
// substitution as GetLineWidth above).
std::string WordWrapString(const int *sheetWidths, std::string_view text, unsigned width, int spacing)
{
	std::string output;
	if (text.empty() || text[0] == '\0')
		return output;

	output.reserve(text.size());
	const char *begin = text.data();
	const char *processedEnd = text.data();
	std::string_view::size_type lastBreakablePos = std::string_view::npos;
	std::size_t lastBreakableLen = 0;
	unsigned lineWidth = 0;

	char32_t codepoint = U'\0'; // the current codepoint
	char32_t nextCodepoint;     // the next codepoint
	std::size_t nextCodepointLen;
	std::string_view remaining = text;
	nextCodepoint = DecodeFirstUtf8CodePoint(remaining, &nextCodepointLen);
	do {
		codepoint = nextCodepoint;
		const std::size_t codepointLen = nextCodepointLen;
		if (codepoint == Utf8DecodeError)
			break;
		remaining.remove_prefix(codepointLen);
		nextCodepoint = !remaining.empty() ? DecodeFirstUtf8CodePoint(remaining, &nextCodepointLen) : U'\0';

		if (codepoint == U'\n') { // Existing line break, scan next line
			lastBreakablePos = std::string_view::npos;
			lineWidth = 0;
			output.append(processedEnd, remaining.data());
			processedEnd = remaining.data();
			continue;
		}

		if (codepoint != ZWSP) {
			const uint8_t frame = codepoint & 0xFF;
			lineWidth += sheetWidths[frame] + spacing;
		}

		if (IsBreakableWhitespace(codepoint)) {
			lastBreakablePos = remaining.data() - begin - codepointLen;
			lastBreakableLen = codepointLen;
			continue;
		}

		if (lineWidth - spacing <= width) {
			if (IsBreakAllowed(codepoint, nextCodepoint)) {
				lastBreakablePos = remaining.data() - begin;
				lastBreakableLen = 0;
			}

			continue; // String is still within the limit, continue to the next symbol
		}

		if (lastBreakablePos == std::string_view::npos) { // Single word longer than width
			lastBreakablePos = remaining.data() - begin - codepointLen;
			lastBreakableLen = 0;
		}

		// Break line and continue to next line
		const char *end = &text[lastBreakablePos];
		output.append(processedEnd, end);
		output += '\n';

		// Restart from the beginning of the new line.
		remaining = text.substr(lastBreakablePos + lastBreakableLen);
		processedEnd = remaining.data();
		lastBreakablePos = std::string_view::npos;
		lineWidth = 0;
		nextCodepoint = !remaining.empty() ? DecodeFirstUtf8CodePoint(remaining, &nextCodepointLen) : U'\0';
	} while (!remaining.empty() && remaining[0] != '\0');
	output.append(processedEnd, remaining.data());
	return output;
}

// ---------------------------------------------------------------------------
// 12-05 CLX sheet decoding (the widths upstream reads through
// currentFont.glyph(frame).width(), decoded from the same committed hex
// golden internal/assets/font.go parses).
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

// SheetWidths decodes the 256 CLX frames and returns the width of each glyph;
// a zero-width frame is legal (blank glyph, internal/testdata/assets/fonts/
// README.md). Frame header: pixel-data offset, width, height, all LE uint16
// (clx_sprite.hpp:15-27, the same parse internal/assets/font.go:139-160 does).
std::vector<int> SheetWidths(const std::vector<uint8_t> &data)
{
	if (data.size() < 4) {
		std::fprintf(stderr, "sheet too short\n");
		std::exit(1);
	}
	uint32_t numFrames = ReadLE32(data.data());
	if (numFrames != 256) {
		std::fprintf(stderr, "sheet holds %u frames, want 256\n", numFrames);
		std::exit(1);
	}
	std::vector<uint32_t> offsets(numFrames + 1);
	for (uint32_t i = 0; i <= numFrames; ++i)
		offsets[i] = ReadLE32(data.data() + 4 + 4 * i);
	if (offsets[numFrames] != data.size()) {
		std::fprintf(stderr, "last offset %u != size %zu\n", offsets[numFrames], data.size());
		std::exit(1);
	}

	std::vector<int> widths(numFrames);
	for (uint32_t i = 0; i < numFrames; ++i) {
		const uint8_t *frame = data.data() + offsets[i];
		size_t frameLen = offsets[i + 1] - offsets[i];
		if (frameLen < 6) {
			std::fprintf(stderr, "frame %u too short\n", i);
			std::exit(1);
		}
		uint16_t headerSize = ReadLE16(frame);
		widths[i] = ReadLE16(frame + 2);
		uint16_t height = ReadLE16(frame + 4);
		(void)height;
		if (headerSize < 6 || headerSize > frameLen) {
			std::fprintf(stderr, "frame %u bad header %u\n", i, headerSize);
			std::exit(1);
		}
	}
	return widths;
}

// ---------------------------------------------------------------------------
// Corpus: the exact strings TestWrapLayout + TestFontLineWidthParity pin in
// internal/render/text_test.go (spacing = 1). startX is the left-aligned
// cursor offset after the line (line width + spacing), the value Go's
// Result.LastCursorX carries for AlignLeft.
// ---------------------------------------------------------------------------

struct Case {
	const char *name;
	const char *text; // UTF-8
	int width;        // wrap rect width; -1 = LineWidth-only case
};

const Case kCases[] = {
	// LineWidth-parity corpus (TestFontLineWidthParity): measured via wrap at
	// a width that never splits so the golden still carries the full row set.
	{"lw-empty", "", -1},
	{"lw-single", "\xC3\x90", -1},
	{"lw-two", "\xC3\x90\xC3\x9E", -1},
	{"lw-three", "\xC3\x90\xC3\x95\xC3\x9E", -1},
	{"lw-five", "\xC3\x90\xC3\x9E\xC3\x92\xC3\x98\xC3\xA9", -1},
	{"lw-ten", "\xC3\x90\xC3\x91\xC3\x92\xC3\x93\xC3\x94\xC3\x95\xC3\x96\xC3\x97\xC3\x98\xC3\x99", -1},
	{"lw-narrow-repeat", "\xC3\x95\xC3\x9F\xC3\x99", -1},
	{"lw-mixed-ascii-blank", "ab\xC3\x90", -1},
	{"lw-wide-tail", "\xC3\x93\xC3\x9E\xC3\xAA", -1},
	{"lw-zwsp-solo", "\xE2\x80\x8B", -1},
	{"lw-zwsp-between", "\xE2\x80\x8B\xC3\x90", -1},
	{"lw-zwsp-run", "\xE2\x80\x8B\xE2\x80\x8B", -1},
	{"lw-zwsp-word", "\xC3\x90\xE2\x80\x8B\xC3\x9E", -1},
	{"lw-ascii-then-zwsp", "a\xE2\x80\x8B", -1},
	{"lw-newline-solo", "\n", -1},
	{"lw-newline-first", "\n\xC3\x90", -1},
	{"lw-newline-last", "\xC3\x90\n", -1},
	{"lw-newline-middle", "\xC3\x90\n\xC3\x9E", -1},
	{"lw-zwsp-then-newline", "\xE2\x80\x8B\n", -1},
	{"lw-crlf", "\xC3\x90\r\n", -1},
	// Wrap-layout corpus (TestWrapLayout).
	{"fits-exactly", "\xC3\x90\xC3\x9E", 19},
	{"one-under", "\xC3\x90\xC3\x9E", 20},
	{"two-words-fit", "\xC3\x90 \xC3\x9E", 200},
	{"two-words-split", "\xC3\x90 \xC3\x9E", 10},
	{"line-one-ends-at-space", "\xC3\x90\xC3\x93 \xC3\x9E", 20},
	{"second-word-wider", "\xC3\x95 \xC3\x9E", 12},
	{"third-word-forces-split", "\xC3\x90 \xC3\x95 \xC3\x9E", 14},
	{"newline-two-lines", "\xC3\x90\n\xC3\x9E", 200},
	{"newline-then-split", "\xC3\x90\n\xC3\x90 \xC3\x9E", 10},
	{"hard-split", "\xC3\x90\xC3\x91\xC3\x92\xC3\x93", 10},
	{"hard-split-exact", "\xC3\x90\xC3\x91\xC3\x92", 24},
};

std::string Escape(std::string_view s)
{
	std::string out;
	for (char c : s) {
		switch (c) {
		case '\n': out += "\\n"; break;
		case '\t': out += "\\t"; break;
		case '\\': out += "\\\\"; break;
		default: out += c;
		}
	}
	return out;
}

} // namespace wraporacle

int main(int argc, char **argv)
{
	if (argc != 2) {
		std::fprintf(stderr, "usage: wraporacle <12-05.clx.hex>\n");
		return 1;
	}
	std::vector<uint8_t> sheet = wraporacle::ReadHexFile(argv[1]);
	std::vector<int> widths = wraporacle::SheetWidths(sheet);
	constexpr int spacing = 1;

	int idx = 0;
	for (const wraporacle::Case &c : wraporacle::kCases) {
		std::string text = c.text;
		if (c.width < 0) {
			// LineWidth-only corpus row: lineWidth is the whole string's
			// GetLineWidth (stops at '\n'), one line, empty lineText.
			int lw = wraporacle::GetLineWidth(widths.data(), text, spacing, nullptr);
			int startX = lw != 0 ? lw + spacing : 0;
			std::printf("%d\t0\t%d\t%d\t\n", idx, lw, startX);
		} else {
			std::string wrapped = wraporacle::WordWrapString(widths.data(), text, c.width, spacing);
			std::size_t lineIdx = 0;
			std::size_t begin = 0;
			while (true) {
				std::size_t nl = wrapped.find('\n', begin);
				std::string_view line = std::string_view(wrapped).substr(
				    begin, nl == std::string::npos ? std::string::npos : nl - begin);
				int lw = wraporacle::GetLineWidth(widths.data(), line, spacing, nullptr);
				int startX = lw != 0 ? lw + spacing : 0;
				std::printf("%d\t%zu\t%d\t%d\t%s\n", idx, lineIdx, lw, startX,
				    wraporacle::Escape(line).c_str());
				if (nl == std::string::npos)
					break;
				begin = nl + 1;
				++lineIdx;
			}
		}
		++idx;
	}
	return 0;
}
