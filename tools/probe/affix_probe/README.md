# affix_probe — SelectAffix pool dumper

Dumps DevilutionX's own `SelectAffix` pool (`Source/items.cpp:1063-1093`) for a
fixed, printed case list. The committed output is
`internal/testdata/gamedata/items/affix_pool_probe.tsv`, asserted row-for-row by
`TestEligibleAffixesMatchesCppProbe`.

The devilutionX tree is used read-only: the probe compiles two of its
translation units with the include flags of an existing build directory and
links them with the probe's own `main`. No CMake target, patch or file inside
that tree is touched.

## Why the linker tricks

* `SelectAffix` sits in an anonymous namespace, so it has internal linkage. The
  object is linked after `objcopy --globalize-symbol=<mangled name>`, and
  `affix_probe.cpp` binds to it through that mangled name in an `asm` label.
* `items.cpp`'s `GenerateRnd` call site is redirected to the probe's recorder by
  `-include rnd_override.h`, which lets the probe read the pool size the C++
  hands to `GenerateRnd` (and script the index) without editing upstream.

## Build (devilutionX pinned checkout with an existing build dir)

```sh
DVLX=~/github/bmcszk/devilutionX      # read-only
BUILD=$DVLX/build-v2                 # any dir with compile_commands.json
OUT=/tmp/dg-slice2-probe && mkdir -p "$OUT"

# 1. the compile flags the build used for items.cpp, minus compiler/LTO/PIC
python3 - "$BUILD/compile_commands.json" > "$OUT/flags.txt" <<'PY'
import json, shlex, sys
cc = json.load(open(sys.argv[1]))
cmd = [x for x in cc if x["file"].endswith("Source/items.cpp")][0]["command"].split()
drop = {"-c", "-flto=auto", "-fno-fat-lto-objects", "-fipa-pta",
        "-fdevirtualize-at-ltrans", "-fPIC"}
out, i = [], 0
while i < len(cmd):
    t = cmd[i]
    if t in ("-o", "-MT", "-MF"):
        i += 2
        continue
    if t.startswith(("-o", "-MT", "-MF")) or t.endswith((".cpp", ".c")) or t in drop:
        i += 1
        continue
    if t.endswith("c++"):
        i += 1
        continue
    out.append(t)
    i += 1
print(" ".join(out))
PY

# 2. upstream objects (items.cpp carries the RNG redirection)
FLAGS=$(cat "$OUT/flags.txt")
/usr/sbin/c++ $FLAGS -std=c++23 -ffunction-sections -fdata-sections -fno-pie \
    -include "$PWD/tools/probe/affix_probe/rnd_override.h" \
    -c "$DVLX/Source/items.cpp" -o "$OUT/items.o"
/usr/sbin/c++ $FLAGS -std=c++23 -ffunction-sections -fdata-sections -fno-pie \
    -c "$DVLX/Source/engine/random.cpp" -o "$OUT/random.o"

# 3. the probe itself
/usr/sbin/c++ $FLAGS -std=c++23 -ffunction-sections -fdata-sections -fno-pie \
    -c tools/probe/affix_probe/affix_probe.cpp -o "$OUT/probe.o"

# 4. promote the anonymous-namespace SelectAffix, then link
objcopy --globalize-symbol='_ZN10devilution12_GLOBAL__N_1L11SelectAffixERKSt6vectorINS_8PLStructESaIS2_EENS_13AffixItemTypeEiibNS_10goodorevilEb' \
    "$OUT/items.o" "$OUT/items_g.o"
/usr/sbin/c++ -no-pie -Wl,--gc-sections -Wl,--unresolved-symbols=ignore-all \
    "$OUT/probe.o" "$OUT/items_g.o" "$OUT/random.o" -o "$OUT/affixprobe"

# 5. dump: base tables, then Hellfire, concatenated into the committed fixture
R=internal/testdata/gamedata/items
"$OUT/affixprobe" $R/item_prefixes.tsv $R/item_suffixes.tsv "" 1000 "$OUT/base.tsv"
"$OUT/affixprobe" $R/hf-item_prefixes.tsv $R/hf-item_suffixes.tsv hf- 300 "$OUT/hf.tsv"
{ cat "$OUT/base.tsv"; tail -n +2 "$OUT/hf.tsv"; } > $R/affix_pool_probe.tsv
```

`-Wl,--unresolved-symbols=ignore-all` is safe here: the retained sections
(`SelectAffix` and the probe) only call `GenerateRnd` and inline helpers, and
`--gc-sections` drops everything else.
