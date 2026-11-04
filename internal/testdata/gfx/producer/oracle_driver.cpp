// Offline fixture producer: port-parity oracle for devilutionX CLX converters.
// Build: g++ -std=c++2b -O1 -I<devilutionX>/Source $(pkg-config --cflags sdl2) driver.cpp -o oracle $(pkg-config --libs sdl2)
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <string_view>
#include <vector>
#include <fstream>
#include <iterator>
#include <memory>

namespace devilution {
[[noreturn]] void app_fatal(std::string_view s) { fprintf(stderr, "APP_FATAL: %.*s\n", (int)s.size(), s.data()); exit(3); }
[[noreturn]] void DisplayFatalErrorAndExit(std::string_view t, std::string_view b) { fprintf(stderr, "FATAL %.*s: %.*s\n", (int)t.size(), t.data(), (int)b.size(), b.data()); exit(3); }
[[noreturn]] void assert_fail(int l, const char* f, const char* e) { fprintf(stderr, "ASSERT %s:%d %s\n", f, l, e); exit(3); }
}

#define UNPACKED_MPQS
#define private public
#define protected public
#include "utils/cel_to_clx.cpp"
#include "utils/cl2_to_clx.cpp"
#include "utils/surface_to_clx.cpp"
#undef private
#undef protected

static std::vector<uint8_t> readf(const char* p) {
  std::ifstream f(p, std::ios::binary);
  return std::vector<uint8_t>((std::istreambuf_iterator<char>(f)), std::istreambuf_iterator<char>());
}

// widthOrWidths: one value for all frames, or one per frame.
static std::vector<uint16_t> parseWidths(int argc, char** argv, int from) {
  std::vector<uint16_t> widths;
  for (int i = from; i < argc; ++i) widths.push_back((uint16_t)atoi(argv[i]));
  return widths;
}

// surface: raw 8-bit pixels (row-major, pitch == width), args are
// width numFrames (transparentColor | -).
static std::vector<uint8_t> surfaceToClx(const std::vector<uint8_t>& data, int width, int numFrames,
    bool hasTransparent, uint8_t transparentColor) {
  const int height = (int)(data.size() / (size_t)width);
  SDL_Surface* raw = SDL_CreateRGBSurfaceFrom(const_cast<uint8_t*>(data.data()), width, height, 8, width,
      0, 0, 0, 0);
  if (raw == nullptr) { fprintf(stderr, "surface create failed\n"); exit(3); }
  devilution::Surface surface(raw);
  devilution::OwnedClxSpriteList owned = devilution::SurfaceToClx(surface, numFrames,
      hasTransparent ? std::optional<uint8_t>(transparentColor) : std::nullopt);
  devilution::ClxSpriteList list(owned);
  const size_t size = list.dataSize();
  std::vector<uint8_t> out(list.data(), list.data() + size);
  SDL_FreeSurface(raw);
  return out;
}

int main(int argc, char** argv) {
  if (argc < 4) { fprintf(stderr, "usage: %s cel|cl2|surface in out [widths...|width numFrames [transparentColor|-]]\n", argv[0]); return 2; }
  const std::string mode = argv[1];
  auto data = readf(argv[2]);
  if (data.empty()) { fprintf(stderr, "empty input\n"); return 3; }
  std::vector<uint8_t> out;

  if (mode == "surface") {
    if (argc < 6) { fprintf(stderr, "surface needs width numFrames\n"); return 2; }
    const bool hasTransparent = argc > 6 && strcmp(argv[6], "-") != 0;
    out = surfaceToClx(data, atoi(argv[4]), atoi(argv[5]), hasTransparent,
        hasTransparent ? (uint8_t)atoi(argv[6]) : 0);
  } else if (mode == "cel") {
    auto widths = parseWidths(argc, argv, 4);
    devilution::PointerOrValue<uint16_t> w(widths.size() == 1 ? widths[0] : 0);
    devilution::PointerOrValue<uint16_t> wp = widths.size() == 1 ? devilution::PointerOrValue<uint16_t>(widths[0]) : devilution::PointerOrValue<uint16_t>(widths.data());
    auto r = devilution::CelToClx(data.data(), data.size(), wp);
    size_t n = r.dataSize();
    out.assign(r.data_.get(), r.data_.get() + n);
  } else if (mode == "cl2") {
    auto widths = parseWidths(argc, argv, 4);
    devilution::PointerOrValue<uint16_t> wp = widths.size() == 1 ? devilution::PointerOrValue<uint16_t>(widths[0]) : devilution::PointerOrValue<uint16_t>(widths.data());
    devilution::Cl2ToClx(data.data(), data.size(), wp, out);
  } else {
    fprintf(stderr, "unknown mode\n"); return 2;
  }

  FILE* f = fopen(argv[3], "wb");
  fwrite(out.data(), 1, out.size(), f);
  fclose(f);
  printf("%zu bytes\n", out.size());
  return 0;
}
