// dvg affix-pool probe — dumps DevilutionX's own SelectAffix pool for a fixed,
// printed set of (table, type, minlvl, maxlvl, onlygood, goe, excludeCharges)
// tuples. Data in (the committed item_prefixes/item_suffixes TSVs), data out
// (pool size, eligible count, expanded row list), no interpretation.
//
// Build: see docs/pow/asset-verification-wave.md "Slice 2 — closers".
#include <cstdint>
#include <cstdio>
#include <cstring>
#include <expected>
#include <fstream>
#include <optional>
#include <sstream>
#include <string>
#include <string_view>
#include <vector>

#include "tables/itemdat.h"

namespace devilution {

int32_t GenerateRnd(int32_t v);
// SelectAffix lives in items.cpp's anonymous namespace, so it has internal
// linkage; the object is linked with --globalize-symbol (see the build notes in
// docs/pow/asset-verification-wave.md) and bound here through its mangled name.
std::optional<const PLStruct *> SelectAffix(const std::vector<PLStruct> &affixList,
    AffixItemType type, int minlvl, int maxlvl, bool onlygood, goodorevil goe,
    bool excludeChargesForStaffs)
    asm("_ZN10devilution12_GLOBAL__N_1L11SelectAffixERKSt6vectorINS_8PLStructESaIS2_EENS_13AffixItemTypeEiibNS_10goodorevilEb");

namespace {
int32_t gRndArg = -1;
int32_t gRndScript = 0;
} // namespace

// items.cpp:1090 was redirected here by -include rnd_override.h.
int32_t probeGenerateRnd(int32_t v)
{
	gRndArg = v;
	return gRndScript;
}

} // namespace devilution

using namespace devilution;

namespace {

std::vector<std::string> SplitTabs(const std::string &line)
{
	std::vector<std::string> out;
	std::string cur;
	for (char c : line) {
		if (c == '\t') {
			out.push_back(cur);
			cur.clear();
		} else {
			cur.push_back(c);
		}
	}
	out.push_back(cur);
	return out;
}

// Copied from itemdat.cpp:547-565 (anonymous namespace there, so not linkable):
// the only member the selection logic reads is CHARGES (items.cpp:1082).
AffixItemType ParseType(std::string_view v)
{
	if (v == "Misc") return AffixItemType::Misc;
	if (v == "Bow") return AffixItemType::Bow;
	if (v == "Staff") return AffixItemType::Staff;
	if (v == "Weapon") return AffixItemType::Weapon;
	if (v == "Shield") return AffixItemType::Shield;
	if (v == "Armor") return AffixItemType::Armor;
	return AffixItemType::None;
}

goodorevil ParseAlignment(std::string_view v)
{
	if (v == "Evil") return GOE_EVIL;
	if (v == "Good") return GOE_GOOD;
	return GOE_ANY;
}

item_effect_type ParseEffect(std::string_view v)
{
	if (v == "CHARGES") return IPL_CHARGES;
	return IPL_INVALID;
}

bool LoadAffixes(const char *path, std::vector<PLStruct> &out)
{
	std::ifstream in(path);
	if (!in) {
		std::fprintf(stderr, "cannot open %s\n", path);
		return false;
	}
	std::string line;
	if (!std::getline(in, line)) return false; // header
	while (std::getline(in, line)) {
		if (line.empty()) continue;
		std::vector<std::string> f = SplitTabs(line);
		if (f.size() < 12) {
			std::fprintf(stderr, "short row in %s: %s\n", path, line.c_str());
			return false;
		}
		PLStruct pl {};
		pl.PLName = f[0];
		pl.power.type = ParseEffect(f[1]);
		pl.power.param1 = std::atoi(f[2].c_str());
		pl.power.param2 = std::atoi(f[3].c_str());
		pl.PLMinLvl = static_cast<int8_t>(std::atoi(f[4].c_str()));
		AffixItemType types = AffixItemType::None;
		std::string cur;
		std::string list = f[5];
		list.push_back(',');
		for (char c : list) {
			if (c == ',') {
				if (!cur.empty() && cur != "-") {
					types = types | ParseType(cur);
				}
				cur.clear();
			} else {
				cur.push_back(c);
			}
		}
		pl.PLIType = types;
		pl.PLGOE = ParseAlignment(f[6]);
		pl.PLChance = static_cast<uint8_t>(std::atoi(f[7].c_str()));
		pl.PLOk = f[8] == "true";
		pl.minVal = std::atoi(f[9].c_str());
		pl.maxVal = std::atoi(f[10].c_str());
		pl.multVal = std::atoi(f[11].c_str());
		out.push_back(pl);
	}
	return true;
}

uint64_t Fnv1a(const std::string &s)
{
	uint64_t h = 1469598103934665603ULL;
	for (unsigned char c : s) {
		h ^= c;
		h *= 1099511628211ULL;
	}
	return h;
}

const char *TypeName(AffixItemType t)
{
	switch (t) {
	case AffixItemType::Misc: return "Misc";
	case AffixItemType::Bow: return "Bow";
	case AffixItemType::Staff: return "Staff";
	case AffixItemType::Weapon: return "Weapon";
	case AffixItemType::Shield: return "Shield";
	case AffixItemType::Armor: return "Armor";
	case AffixItemType::None: return "None";
	}
	return "?";
}

struct Case {
	const char *table;
	AffixItemType type;
	int minlvl, maxlvl;
	bool onlygood;
	goodorevil goe;
	bool excludeCharges;
};

// Deterministic case generator: a plain LCG so the fixture is reproducible with
// any standard library (no <random> implementation drift).
struct CaseGen {
	uint64_t s = 0x5150ULL;
	uint32_t Next()
	{
		s = s * 6364136223846793005ULL + 1442695040888963407ULL;
		return static_cast<uint32_t>(s >> 33);
	}
};

const AffixItemType kTypes[] = {AffixItemType::None, AffixItemType::Misc, AffixItemType::Bow,
	AffixItemType::Staff, AffixItemType::Weapon, AffixItemType::Shield, AffixItemType::Armor};

void EmitCase(std::FILE *out, const Case &c, const std::vector<PLStruct> &table, bool detail)
{
	gRndScript = 0;
	gRndArg = -1;
	(void)SelectAffix(table, c.type, c.minlvl, c.maxlvl, c.onlygood, c.goe, c.excludeCharges);
	int32_t poolSize = gRndArg < 0 ? 0 : gRndArg; // pool size == the value GenerateRnd was handed
	int32_t eligible = 0;
	std::string rows;
	if (poolSize > 0) {
		std::string prev;
		for (int32_t i = 0; i < poolSize; ++i) {
			gRndScript = i;
			gRndArg = -1;
			auto picked = SelectAffix(table, c.type, c.minlvl, c.maxlvl, c.onlygood, c.goe, c.excludeCharges);
			if (!picked) continue;
			const std::string &name = (*picked)->PLName;
			if (name != prev) {
				++eligible;
				prev = name;
			}
			rows += name;
			rows.push_back('\n');
			if (detail) {
				std::fprintf(out, "DETAIL\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%s\n", c.table,
				    TypeName(c.type), c.minlvl, c.maxlvl, c.onlygood ? 1 : 0,
				    static_cast<int>(c.goe), c.excludeCharges ? 1 : 0, i, name.c_str());
			}
		}
	}
	std::fprintf(out, "%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%016llx\n", c.table, TypeName(c.type),
	    c.minlvl, c.maxlvl, c.onlygood ? 1 : 0, static_cast<int>(c.goe), c.excludeCharges ? 1 : 0,
	    eligible, poolSize, static_cast<unsigned long long>(Fnv1a(rows)));
}

} // namespace

int main(int argc, char **argv)
{
	if (argc < 6) {
		std::fprintf(stderr, "usage: %s <prefixes.tsv> <suffixes.tsv> <variant> <samples> <out.tsv>\n", argv[0]);
		return 2;
	}
	const std::string variant = argv[3];
	const int sampled = std::atoi(argv[4]);
	std::vector<PLStruct> prefixes;
	std::vector<PLStruct> suffixes;
	if (!LoadAffixes(argv[1], prefixes)) return 1;
	if (!LoadAffixes(argv[2], suffixes)) return 1;

	std::FILE *out = std::fopen(argv[5], "w");
	if (out == nullptr) {
		std::perror("fopen");
		return 1;
	}
	const auto baseCount = static_cast<int>(prefixes.size());
	std::fprintf(out, "table\ttype\tminlvl\tmaxlvl\tonlygood\tgoe\texcludeCharges\teligible\tpoolSize\trowsHash64\n");

	// Worked examples from V1.6 (detail rows for both).
	const std::string prefixTable = variant + "prefixes";
	const std::string suffixTable = variant + "suffixes";
	EmitCase(out, Case {prefixTable.c_str(), AffixItemType::Weapon, 1, 6, false, GOE_ANY, false}, prefixes, true);
	EmitCase(out, Case {suffixTable.c_str(), AffixItemType::Weapon, 1, 6, false, GOE_EVIL, false}, suffixes, true);

	CaseGen gen;
	for (int i = 0; i < sampled; ++i) {
		int minlvl = static_cast<int>(gen.Next() % 31);
		int maxlvl = minlvl + static_cast<int>(gen.Next() % 61);
		Case c {prefixTable.c_str(), kTypes[gen.Next() % 7], minlvl, maxlvl, (gen.Next() % 2) == 0,
			static_cast<goodorevil>(gen.Next() % 3), (gen.Next() % 2) == 0};
		if (i % 2 == 1) {
			c.table = suffixTable.c_str();
			EmitCase(out, c, suffixes, false);
		} else {
			EmitCase(out, c, prefixes, false);
		}
	}
	std::fclose(out);
	std::fprintf(stderr, "prefixes=%d suffixes=%d cases=%d\n", baseCount, static_cast<int>(suffixes.size()), sampled + 2);
	return 0;
}
