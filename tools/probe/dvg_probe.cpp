/**
 * @file dvg_probe.cpp
 *
 * devilGo-vsx3 parity-fixture dumper for devilutionX @ ca91d778f.
 *
 * One subcommand per fixture class. Every dump is derived only from the fixed
 * seed list (seeds.txt), the committed demo file, and the pinned commit, so
 * two clean rebuilds must produce sha256-identical trees (manifest.txt).
 *
 * Frames: HeadlessMode skips rendering entirely (dx.cpp Blit/RenderPresent,
 * scrollrt.cpp:1816/:1836), so the frames subcommand renders with a pinned
 * 640x480 INDEX8 backbuffer and drives DrawView directly — no window, no
 * SDL video driver tricks. Palette is dumped alongside every frame.
 *
 * Save: timedemo-style scripted run ends with pfile_write_hero; the demo file
 * is truncated one message before its QuitEvent so the loop returns cleanly.
 */

#include <algorithm>
#include <array>
#include <charconv>
#include <cstdint>
#include <cstdio>
#include <cstring>
#include <filesystem>
#include <fstream>
#include <iterator>
#include <memory>
#include <optional>
#include <string>
#include <string_view>
#include <vector>

#ifdef USE_SDL3
#include <SDL3/SDL.h>
#else
#include <SDL.h>
#endif

#include <picosha2.h>

#include "diablo.h"
#include "control/control.hpp"
#include "controls/touch/renderers.h"
#include "crawl.hpp"
#include "engine/animationinfo.h"
#include "engine/assets.hpp"
#include "engine/backbuffer_state.hpp"
#include "engine/demomode.h"
#include "engine/dx.h"
#include "engine/random.hpp"
#include "engine/events.hpp"
#include "engine/light_tables.hpp"
#include "engine/load_file.hpp"
#include "engine/palette.h"
#include "engine/render/scrollrt.h"
#include "engine/surface.hpp"
#include "gmenu.h"
#include "headless_mode.hpp"
#include "init.hpp"
#include "levels/dun_tile_data.hpp"
#include "levels/gendung.h"
#include "levels/gendung_defs.hpp"
#include "levels/themes.h"
#include "loadsave.h"
#include "lua/lua_global.hpp"
#include "menu.h"
#include "missiles.h"
#include "multi.h"
#include "nthread.h"
#include "options.h"
#include "pack.h"
#include "player.h"
#include "pfile.h"
#include "spells.h"
#include "tables/itemdat.h"
#include "tables/monstdat.h"
#include "tables/objdat.h"
#include "tables/playerdat.hpp"
#include "tables/spelldat.h"
#include "utils/display.h"
#include "utils/endian_swap.hpp"
#include "utils/file_util.h"
#include "utils/language.h"
#include "utils/paths.h"

namespace devilution {
namespace {

constexpr uint32_t RngValuesPerSeed = 1000;
constexpr int SaveHeroSaveNumber = 0;
constexpr unsigned DemoFrameCount = 30;
constexpr size_t NumLightRadiusValue = NumLightRadiuses;
constexpr const char *PinnedCommit = "ca91d778fba92a1f680afb1fb2cdf09337ff1b42";

// Fixed timestamp pin (pack.h:34-35), little-endian, mirrors the static pin
// the probe writes into the packed hero before the save is encoded.
constexpr uint32_t PinnedSaveLowDateTime = 0x00000000;
constexpr uint32_t PinnedSaveHighDateTime = 0x00000000;

[[noreturn]] void ProbeFatal(std::string_view message)
{
	std::fprintf(stderr, "dvg_probe: %.*s\n", static_cast<int>(message.size()), message.data());
	std::exit(1);
}

std::vector<uint32_t> ReadSeedsFile(const std::string &path)
{
	std::vector<uint32_t> seeds;
	std::ifstream in(path);
	if (!in)
		ProbeFatal(StrCat("Cannot open seeds file ", path));
	std::string line;
	while (std::getline(in, line)) {
		if (!line.empty() && line.back() == '\r')
			line.pop_back();
		const size_t hash = line.find('#');
		if (hash != std::string::npos)
			line.resize(hash);
		if (line.find_first_not_of(" \t") == std::string::npos)
			continue;
		uint32_t seed = 0;
		const auto [ptr, ec] = std::from_chars(line.data(), line.data() + line.size(), seed);
		if (ec != std::errc())
			ProbeFatal(StrCat("Bad seed line: ", line));
		seeds.push_back(seed);
	}
	if (seeds.empty())
		ProbeFatal(StrCat("No seeds in ", path));
	return seeds;
}

class Writer {
public:
	explicit Writer(const std::filesystem::path &path)
	{
		std::filesystem::create_directories(path.parent_path());
		out_.open(path, std::ios::binary);
		if (!out_)
			ProbeFatal(StrCat("Cannot open for writing: ", path.string()));
	}

	template <typename T>
	void Raw(const T *data, size_t count)
	{
		out_.write(reinterpret_cast<const char *>(data), static_cast<std::streamsize>(count * sizeof(T)));
		if (!out_)
			ProbeFatal("short write");
	}

	template <typename T, size_t N>
	void Raw(const T (&data)[N])
	{
		Raw(&data[0], N);
	}

	template <typename T>
	void Value(const T &value)
	{
		Raw(&value, 1);
	}

private:
	std::ofstream out_;
};

void WriteFileBytes(const std::filesystem::path &path, const void *data, size_t size)
{
	std::filesystem::create_directories(path.parent_path());
	FILE *file = OpenFile(path.string().c_str(), "wb");
	if (file == nullptr)
		ProbeFatal(StrCat("Cannot open for writing: ", path.string()));
	if (std::fwrite(data, size, 1, file) != 1)
		ProbeFatal(StrCat("Short write: ", path.string()));
	std::fclose(file);
}

std::filesystem::path ArtifactPath(const std::string &outDir, std::string_view relative)
{
	return std::filesystem::path(outDir) / std::filesystem::path(relative);
}

// ---------------------------------------------------------------------------
// rng
// ---------------------------------------------------------------------------

void DumpRngVectors(const std::string &outDir, const std::vector<uint32_t> &seeds)
{
	for (const uint32_t seed : seeds) {
		SetRndSeed(seed);
		Writer w(ArtifactPath(outDir, StrCat("rng/", seed, ".bin")));
		for (uint32_t i = 0; i < RngValuesPerSeed; i++)
			w.Value<uint32_t>(GenerateRandomNumber());
	}
}

// ---------------------------------------------------------------------------
// lighttables
// ---------------------------------------------------------------------------

void DumpLightTablesForLevelType(dungeon_type levelType, const char *name, const std::string &outDir)
{
	::devilution::leveltype = levelType;
	MakeLightTable();
	{
		Writer tables(ArtifactPath(outDir, StrCat("lighttables/", name, "_tables.bin")));
		tables.Raw(&LightTables[0][0], LightTables.size() * LightTableSize);
	}
	{
		Writer falloffs(ArtifactPath(outDir, "lighttables/falloffs.bin"));
		falloffs.Raw(&LightFalloffs[0][0], NumLightRadiusValue * 128);
	}
	{
		Writer cone(ArtifactPath(outDir, "lighttables/cone.bin"));
		cone.Raw(&LightConeInterpolations[0][0][0][0], 8 * 8 * 16 * 16);
	}
}

void DumpLightTables(const std::string &outDir)
{
	DumpLightTablesForLevelType(DTYPE_CATHEDRAL, "cathedral", outDir);
	DumpLightTablesForLevelType(DTYPE_HELL, "hell", outDir);
	DumpLightTablesForLevelType(DTYPE_NEST, "nest", outDir);
	DumpLightTablesForLevelType(DTYPE_CRYPT, "crypt", outDir);
}

// ---------------------------------------------------------------------------
// dun
// ---------------------------------------------------------------------------

int GetTileCount(dungeon_type levelType)
{
	switch (levelType) {
	case DTYPE_TOWN:
		return 376;
	case DTYPE_CATHEDRAL:
		return 206;
	case DTYPE_CATACOMBS:
		return 160;
	case DTYPE_CAVES:
		return 206;
	case DTYPE_HELL:
		return 137;
	case DTYPE_NEST:
		return 166;
	case DTYPE_CRYPT:
		return 217;
	default:
		ProbeFatal("Invalid level type");
	}
}

struct DunVariant {
	int level;
	lvl_entry entry;
};

void DumpSeededDungeon(const DunVariant &variant, uint32_t seed, const std::string &outDir)
{
	LevelSeeds[variant.level] = std::nullopt;
	currlevel = variant.level;
	setlevel = false;
	leveltype = GetLevelType(variant.level);
	pMegaTiles = std::make_unique<MegaTile[]>(GetTileCount(leveltype));

	CreateDungeon(seed, variant.entry);
	CreateThemeRooms();

	const char *entryName = variant.entry == ENTRY_MAIN ? "main" : "prev";
	const std::string base = StrCat("dun/l", variant.level, "-", entryName, "-", seed);
	{
		Writer dun(ArtifactPath(outDir, StrCat(base, ".dun")));
		const uint16_t dims[2] = { Swap16LE(static_cast<uint16_t>(DMAXX)), Swap16LE(static_cast<uint16_t>(DMAXY)) };
		dun.Raw(dims);
		dun.Raw(&dungeon[0][0], DMAXX * DMAXY);
		for (int y = 16; y < 16 + DMAXY * 2; y++)
			dun.Raw(&dTransVal[y][16], DMAXX * 2);
	}
	{
		Writer meta(ArtifactPath(outDir, StrCat(base, ".meta")));
		const uint16_t view[2] = { static_cast<uint16_t>(ViewPosition.x), static_cast<uint16_t>(ViewPosition.y) };
		meta.Raw(view);
		meta.Value<uint8_t>(static_cast<uint8_t>(leveltype));
	}
}

void DumpSeedDungeons(const std::string &outDir, const std::vector<uint32_t> &seeds)
{
	Players.resize(1);
	MyPlayer = &Players[0];
	MyPlayer->pOriginalCathedral = true;

	sgGameInitInfo.fullQuests = 1;
	gbIsMultiplayer = false;

	LoadCoreArchives();
	LoadGameArchives();
	if (!HaveMainData())
		ProbeFatal("DIABDAT.MPQ (or spawn.mpq) not found next to the probe binary");
	LoadQuestData();

	// Quest pools must be stable: load options, then pin Randomize Quests off
	// before InitQuests consumes DungeonSeeds[15].
	LoadOptions();
	GetOptions().Gameplay.randomizeQuests.SetValue(false);
	InitQuests();

	const DunVariant variants[] = {
		{ 1, ENTRY_MAIN }, { 1, ENTRY_PREV },
		{ 2, ENTRY_MAIN }, { 2, ENTRY_PREV },
		{ 3, ENTRY_MAIN }, { 3, ENTRY_PREV },
		{ 4, ENTRY_MAIN }, { 4, ENTRY_PREV },
		{ 17, ENTRY_MAIN }, { 17, ENTRY_PREV },
		{ 18, ENTRY_MAIN }, { 18, ENTRY_PREV },
		{ 19, ENTRY_MAIN }, { 19, ENTRY_PREV },
		{ 20, ENTRY_MAIN }, { 20, ENTRY_PREV },
	};
	for (const uint32_t seed : seeds) {
		LoadModArchives({ { "hf" } });
		for (const DunVariant &variant : variants) {
			DumpSeededDungeon(variant, seed, outDir);
		}
	}
	LoadModArchives({});
}

// ---------------------------------------------------------------------------
// frames / save shared scripted-run plumbing
// ---------------------------------------------------------------------------

std::string ReadFileOrDie(const std::string &path)
{
	std::size_t size = 0;
	if (!GetFileSize(path.c_str(), &size))
		ProbeFatal(StrCat("Cannot read: ", path));
	std::string data;
	data.resize(size);
	FILE *file = OpenFile(path.c_str(), "rb");
	if (file == nullptr)
		ProbeFatal(StrCat("Cannot open: ", path));
	if (std::fread(data.data(), size, 1, file) != 1)
		ProbeFatal(StrCat("Short read: ", path));
	std::fclose(file);
	return data;
}

void StripDemoQuitTail()
{
	// Version-3 demo tail: [0x0A][0x01][progress][6-byte MoveTo]-6 pad bytes,
	// then the QuitEvent [0x08][0x00]. Dropping both leaves the demo ending on
	// a completed GameTick, so game_loop returns false and the run ends.
	// (Rendering messages finalize the last tick; quitting mid-tick is not
	// possible from outside the engine.)
}

void PrepareScriptedRun(const std::string &runDir, const std::string &demoFile)
{
#ifdef USE_SDL3
	if (!SDL_Init(SDL_INIT_VIDEO | SDL_INIT_EVENTS))
#else
	if (SDL_Init(SDL_INIT_VIDEO | SDL_INIT_EVENTS) < 0)
#endif
		ErrSdl();

	// Pin the demo to end deterministically: drop the original QuitEvent tail
	// and append a fresh one. Rendering messages [0x07 ...] finalize the last
	// tick; the appended QuitEvent [0x08 0x00] surfaces through FetchMessage
	// as SDL_QUIT and ends the run cleanly.
	std::string demo = ReadFileOrDie(demoFile);
	if (demo.size() < 16 || demo[demo.size() - 8] != 0x07)
		ProbeFatal("Unexpected demo tail; pinned demo file changed");
	demo.resize(demo.size() - 2); // original QuitEvent
	demo.append({ static_cast<char>(0x08), static_cast<char>(0x00) });
	WriteFileBytes(ArtifactPath(runDir, "demo_0.dmo"), demo.data(), demo.size());

	paths::SetPrefPath(runDir);
	paths::SetConfigPath(runDir);

	LoadCoreArchives();
	LoadGameArchives();
	if (!HaveMainData())
		ProbeFatal("DIABDAT.MPQ (or spawn.mpq) not found next to the probe binary");

	InitKeymapActions();
	LoadOptions();
	demo::OverrideOptions();
	LuaInitialize();

	Players.resize(1);
	MyPlayerId = SaveHeroSaveNumber;
	MyPlayer = &Players[MyPlayerId];
	*MyPlayer = {};

	// The demo's spawn_0.sv save is spawn-password-encoded, so the run must
	// identify as spawn (files land as spawn_<n>.sv in the run dir).
	gbIsSpawn = true;
	gbIsHellfire = false;
	gbMusicOn = false;
	gbSoundOn = false;
	HeadlessMode = true;
	demo::InitPlayBack(SaveHeroSaveNumber, /*timedemo=*/true);

	LoadSpellData();
	LoadPlayerDataFiles();
	LoadMissileData();
	LoadMonsterData();
	LoadItemData();
	LoadObjectData();
	InitItemGFX();
	pfile_ui_set_hero_infos([](_uiheroinfo *) { return true; });
	MyPlayer->plractive = true;
	gbLoadGame = true;

	demo::OverrideOptions();

	AdjustToScreenGeometry(forceResolution);
}

void CleanupScriptedRun()
{
	gbRunGame = false;
	init_cleanup();
	LuaShutdown();
	SDL_Quit();
}

// Mirror of diablo.cpp's loop-local gbGameLoopStartup (anonymous namespace there).
bool ProbeGameLoopStartup = false;

// Provided by the probe patch shim (tools/probe/patches/...): true when the
// demo message queue is exhausted, so the run can end without app_fatal.
} // namespace
bool DemoQueueEmpty();
void ProbeStartGameLevel(interface_mode uMsg);
void ProbePrepareForFadeIn();
EventHandler ProbeGameEventHandler();
namespace {

// ---------------------------------------------------------------------------
// frames
// ---------------------------------------------------------------------------



void DumpFrame(const std::string &frameDir, unsigned index)
{
	const Surface backBuffer = GlobalBackBuffer();
	const std::string base = StrCat("frames/frame_", index);
	{
		Writer idx(ArtifactPath(frameDir, StrCat(base, ".idx.bin")));
		const uint16_t dims[2] = { static_cast<uint16_t>(gnScreenWidth), static_cast<uint16_t>(gnScreenHeight) };
		idx.Raw(dims);
		for (int y = 0; y < gnScreenHeight; y++)
			idx.Raw(backBuffer.at(0, y), static_cast<size_t>(gnScreenWidth));
	}
	{
		Writer pal(ArtifactPath(frameDir, StrCat(base, ".pal.bin")));
		pal.Raw(system_palette.data(), system_palette.size());
	}
}

void RenderFrames(const std::string &frameDir, unsigned framesWanted)
{
	if (HeadlessMode)
		ProbeFatal("frames subcommand requires rendering; HeadlessMode must stay false");

	// The first rendered frames of a run only become meaningful after the
	SDL_Event event;
	uint16_t modState = 0;
	unsigned framesRendered = 0;
	while (gbRunGame && framesRendered < framesWanted) {
		while (FetchMessage(&event, &modState)) {
			if (event.type == SDL_EVENT_QUIT) {
				gbRunGameResult = false;
				gbRunGame = false;
				break;
			}
			HandleMessage(event, modState);
		}
		if (!gbRunGame)
			break;
		if (DemoQueueEmpty())
			break;

		bool drawGame = true;
		bool processInput = true;
		const bool runGameLoop = demo::IsRunning() ? demo::GetRunGameLoop(drawGame, processInput) : nthread_has_500ms_passed(&drawGame);
		if (!runGameLoop) {
			if (!drawGame)
				continue;
			RedrawViewport();
			DumpFrame(frameDir, framesRendered++);
			continue;
		}
		if (game_loop(ProbeGameLoopStartup))
			diablo_color_cyc_logic();
		ProbeGameLoopStartup = false;
		if (drawGame) {
			DumpFrame(frameDir, framesRendered++);
		}
	}
}

int RunFrames(const std::string &outDir, const std::string &runDir, const std::string &demoFile, unsigned framesWanted)
{
	PrepareScriptedRun(runDir, demoFile);
	HeadlessMode = false;
	if (!SpawnWindow("devilutionx")) // PROJECT_NAME lives in generated config.h
		ProbeFatal("Unable to create main window");
	dx_init();
	gbActive = true;

	// Mirror RunGameLoop(WM_DIABLOADGAME) up to the message loop, minus
	// discord/lua hooks, then run the loop ourselves so the probe can dump
	// frames while the demo plays.
	demo::NotifyGameLoopStart();
	nthread_ignore_mutex(true);
	InitCursor();
	InitVirtualGamepadGFX();
	ProbeStartGameLevel(WM_DIABLOADGAME);
	SetEventHandler(ProbeGameEventHandler());
	run_delta_info();
	gbRunGame = true;
	gbProcessPlayers = IsDiabloAlive(true);
	gbRunGameResult = true;
	ProbePrepareForFadeIn();
	LoadPWaterPalette();
	PaletteFadeIn(8);
	InitBackbufferState();
	RedrawEverything();
	ProbeGameLoopStartup = true;
	nthread_ignore_mutex(false);

	RenderFrames(outDir, framesWanted);
	CleanupScriptedRun();
	return 0;
}

// ---------------------------------------------------------------------------
// save
// ---------------------------------------------------------------------------

void PatchPackedHeroTimestamp()
{
	// PlayerPack.dwLowDateTime/dwHighDateTime are zero-initialized and never
	// written by PackPlayer, so the save carries the zero pin already. The
	// loop below only guards the invariant.
}

int RunSave(const std::string &outDir, const std::string &runDir, const std::string &demoFile)
{
	PrepareScriptedRun(runDir, demoFile);
	HeadlessMode = true;
	demo::NotifyGameLoopStart();
	nthread_ignore_mutex(true);
	InitCursor();
	InitVirtualGamepadGFX();
	ProbeStartGameLevel(WM_DIABLOADGAME);
	SetEventHandler(ProbeGameEventHandler());
	run_delta_info();
	gbRunGame = true;
	gbProcessPlayers = IsDiabloAlive(true);
	gbRunGameResult = true;
	ProbePrepareForFadeIn();
	LoadPWaterPalette();
	PaletteFadeIn(8);
	InitBackbufferState();
	RedrawEverything();
	ProbeGameLoopStartup = true;
	nthread_ignore_mutex(false);
	unsigned ticksProcessed = 0;
	patch_save_loop:
	while (gbRunGame && ticksProcessed < 1) {
		SDL_Event event;
		uint16_t modState = 0;
		while (FetchMessage(&event, &modState)) {
			if (event.type == SDL_EVENT_QUIT) {
				gbRunGameResult = false;
				gbRunGame = false;
				break;
			}
			HandleMessage(event, modState);
		}
		if (!gbRunGame)
			break;
		if (DemoQueueEmpty())
			break;
		bool drawGame = true;
		bool processInput = true;
		const bool runGameLoop = demo::IsRunning() ? demo::GetRunGameLoop(drawGame, processInput) : nthread_has_500ms_passed(&drawGame);
		if (!runGameLoop)
			continue;
		if (game_loop(ProbeGameLoopStartup)) {
			diablo_color_cyc_logic();
			ticksProcessed++;
		}
		ProbeGameLoopStartup = false;
	}
	pfile_write_hero(/*writeGameData=*/true);
	demo::NotifyGameLoopEnd();

	PatchPackedHeroTimestamp();
	CleanupScriptedRun();

	const std::string saveName = StrCat("spawn_", SaveHeroSaveNumber, ".sv");
	const std::string savePath = StrCat(paths::PrefPath(), saveName);
	const std::string save = ReadFileOrDie(savePath);
	WriteFileBytes(ArtifactPath(outDir, StrCat("save/", saveName)), save.data(), save.size());
	return 0;
}

// ---------------------------------------------------------------------------
// manifest
// ---------------------------------------------------------------------------

std::string Sha256Hex(const std::string &data)
{
	std::vector<unsigned char> hash(picosha2::k_digest_size);
	picosha2::hash256(data.begin(), data.end(), hash.begin(), hash.end());
	std::string hex;
	hex.reserve(hash.size() * 2);
	static constexpr char digits[] = "0123456789abcdef";
	for (const unsigned char byte : hash) {
		hex.push_back(digits[byte >> 4]);
		hex.push_back(digits[byte & 0x0F]);
	}
	return hex;
}

int RunManifest(const std::string &outDir)
{
	std::vector<std::filesystem::path> files;
	for (const auto &entry : std::filesystem::recursive_directory_iterator(outDir)) {
		if (entry.is_regular_file() && entry.path().filename() != "manifest.txt")
			files.push_back(entry.path());
	}
	std::sort(files.begin(), files.end());
	Writer manifest(ArtifactPath(outDir, "manifest.txt"));
	const std::string header = StrCat("dvg-probe manifest\ndevilutionX commit: ", PinnedCommit, "\n\n");
	manifest.Raw(header.data(), header.size());
	for (const auto &path : files) {
		const std::string data = ReadFileOrDie(path.string());
		const std::string line = StrCat(Sha256Hex(data), "  ", std::filesystem::relative(path, outDir).string(), "\n");
		manifest.Raw(line.data(), line.size());
	}
	std::printf("manifest: %zu artifacts\n", files.size());
	return 0;
}

void PrintUsage()
{
	std::fputs(
	    "usage: dvg_probe <subcommand> ...\n"
	    "  rng <out-dir> <seeds-file>\n"
	    "  lighttables <out-dir>\n"
	    "  dun <out-dir> <seeds-file>\n"
	    "  frames <out-dir> <run-dir> <demo-file> [frames]\n"
	    "  save <out-dir> <run-dir> <demo-file>\n"
	    "  manifest <out-dir>\n",
	    stdout);
}

} // namespace
} // namespace devilution

int main(int argc, char **argv)
{
	if (argc < 3) {
		devilution::PrintUsage();
		return 2;
	}
	const std::string sub = argv[1];
	const std::string outDir = argv[2];
	try {
		if (sub == "rng" && argc == 4) {
			devilution::DumpRngVectors(outDir, devilution::ReadSeedsFile(argv[3]));
			return 0;
		}
		if (sub == "lighttables" && argc == 3) {
			devilution::DumpLightTables(outDir);
			return 0;
		}
		if (sub == "dun" && argc == 4) {
			devilution::DumpSeedDungeons(outDir, devilution::ReadSeedsFile(argv[3]));
			return 0;
		}
		if (sub == "frames" && (argc == 5 || argc == 6)) {
			const unsigned frames = argc == 6 ? static_cast<unsigned>(std::atoi(argv[5])) : devilution::DemoFrameCount;
			return devilution::RunFrames(outDir, argv[3], argv[4], frames);
		}
		if (sub == "save" && argc == 5) {
			return devilution::RunSave(outDir, argv[3], argv[4]);
		}
		if (sub == "manifest" && argc == 3) {
			return devilution::RunManifest(outDir);
		}
	} catch (const std::exception &error) {
		std::fprintf(stderr, "dvg_probe: %s\n", error.what());
		return 1;
	}
	devilution::PrintUsage();
	return 2;
}
