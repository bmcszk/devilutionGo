# Upstream gtest port recipe (entity/behavior beans)

For beans whose acceptance criterion reads "`<name>_test.cpp` port passes"
(AnimationInfo, tile properties, pathfinding, enum tables). The upstream
gtest IS the RED suite — port it scenario-for-scenario before writing any
implementation. No invented parity tests needed; do not trim expectations.

## 1. Read the whole gtest + the code under test first

- sed the full test file AND the `.h`/`.cpp` under test. Extract: the
  fixture runner (e.g. `RunAnimationTest`), every scenario table, and the
  exact constant semantics (fixed point: `baseValueFraction = 128`).
- Pull helper tables verbatim: `PathDirections[9]` from `GetPathDirection`,
  the direction→displacement map (`displacement.hpp fromDirection`), enums
  with explicit slots (`QUEST_HIVE_TEASE1 = 7`).

## 2. RED: port the gtest as Go table tests

- One Go test per C++ `TEST()`. A compact step DSL (a `step` struct with
  kind + args, plus one-letter constructors `F()`/`T()`/`R()`) keeps
  hundreds of fixture lines readable and diffable against the C++.
- Assert EVERY expectation, including mid-sequence ones (a GameTick asserts
  frame AND tick-counter, not just the final frame). Assertion density is
  what catches refactor slips later.
- Keep upstream comments that explain WHY ("distribution must not place
  skipped frames after the hit frame") — they become the Go comments.

## 3. GREEN: port the .cpp verbatim

- Field-for-field struct, same integer widths (`int8_t` stays `int8`), same
  fixed-point math and truncation order — Go integer division truncates
  like C++, so no float substitutions.
- Globals the code reads (`ProgressToNextGameTick`) become an injected
  sink struct: tests script them, the engine loop owns the real one later.
- Keep semantics identical through repo lint:
  - flag-parameter ban → split exported variants (`ProcessAnimation` /
    `ProcessAnimationReverse`) over one internal body, never a bool param;
  - nestif → extract the deep branch into a helper (verbatim, named);
  - switch needs a `default` clause; 120-col lines → wrap signatures;
  - test functions trip the cognitive/cyclomatic lint too — extract
    run-model and assert-order helpers, never trim the scenario count.

## 4. Refactor trap the tables catch

Extracting "counter advance + threshold" into a helper and then advancing
the frame unconditionally compiles, looks right, and silently advances
frames on delayed animations (`ticksPerFrame > 1`) — the ported fixture
tables fail on the first delayed scenario. Rule: after ANY refactor of
ported logic, run the full scenario set before touching anything else; the
tables are the parity oracle.

## 5. Multi-AC beans: slice it, one pi worker contract per slice

Land each independently-green slice (ported file + its ported tests) as
one commit with a `## Progress log` entry in the bean body; the bean stays
in-progress until the remaining ACs (C++ probes, struct size tables) close
for real. Place the package per the three-layer DDD rules
(`internal/diablo/animation`, `internal/world`, ...). Each slice = one
pi worker contract (GIT JAIL, RED-first, inlined upstream file:line,
OUT-OF-SCOPE list naming the later slices, exact gates, `</dev/null`,
2100 s wrapper budget — a 4-file TDD slice with full gates overran 900 s
and killed a healthy run mid-flight); the orchestrator re-gates, reads the
full diff, commits `Refs: <bean-id>`, pushes, and watches CI per slice.

Sizing anchor: ONE ported file + its tests ≈ one contract (10–25 min).
Two ports + tests + a refactor fold overran a 2100 s wrapper and died
mid-edit with an inconsistent tree (test file complete, implementation
unbuildable on missing struct fields). After any kill, inventory the tree
against the ORIGINAL contract's FULL deliverable list — implementation
files AND test files — before writing the completion contract: a
"make it compile + tests pass" finish pass silently drops the test files
the dead worker never wrote, and a green gate over incomplete coverage
reads as done. Completion contracts finish named existing artifacts;
they never re-scope silently.

Slice close-out: when a slice's port needs entity structs the entity beans
own (Player/Monster/Item fields), the pipeline takes an explicit-inputs
Ref struct + an injected `func(int) int` RNG now, and the struct wiring is
recorded in the PoW as the consuming bean's job — the roll ORDER and
fixed-point math are this bean's deliverable and are fully testable
without the structs. Name every deferred C++ branch in the PoW residual
(item-flag effects, monster-state checks, steal-life) so no branch is
silently dropped.

## 6. Container-semantics ACs: randomized differential testing

When an AC demands C++ container semantics ("iteration order equals
std::list for N random add/del sequences"): port the container (append =
emplace_back at the tail; remove_if = in-place filter keeping survivor
order relative to each other), then write three tests — a hand-traced
golden sequence, a seeded randomized differential test vs a plain-slice
reference model stepped in lockstep with the container, and the
flagged-row edge case (delete-flagged elements stay in place and iterable
until the delete sweep; appends land after them). Fixed seed, never
time-based, so failures reproduce. The op count (e.g. 1000 sequences ×
200 ops) is the parity evidence — if the test function trips the
complexity lint, extract helpers, do not reduce ops.

## 7. Roll-pipeline ACs: scripted-tape parity

When the AC pins an RNG-consuming pipeline ("exact roll order", "1/64
fixed point"): model every `GenerateRnd`/`RandomIntBetween` call as a
consumed value from a scripted tape (`func(int) int` closing over a slice,
cycling — C++ rolls repeat under constant streams), and write the expected
values in an INDEPENDENT trace function transcribed line-by-line from the
C++ body, never by calling the code under test — expected values derived
by running the Go port (or by adjusting the tape until green) are circular
and assert nothing; every expectation must trace to upstream source or a
probe dump. `RandomIntBetween(min,max)`
= `GenerateRnd(max-min+1)+min` — keep the `+min`; dropping it in the trace
only produces a mismatch that costs a fix round. Feed BOTH sides the same
tape in lockstep: scripted tapes for the matrix (classes × monster
classes × edge values incl. clamp boundaries 5/95/94/95) plus a seeded
randomized harness (1000 seeds × random fixtures) to catch order drift the
curated tapes miss. Damage parity asserts the 1/64 word exactly
(`want << 6`) and that the word is a multiple of 64 — no float anywhere.
Split the pipeline's helpers by the repo lint from the start (the
monster-class switch needs per-class helpers + `default`; bool flags become
split functions or move to the caller), or the first lint round reopens
everything.
When the AC says "N counted at runtime, never hardcoded" (corpus row count
== registry size): load the COMMITTED fixture file in the test
(`os.ReadFile` + the existing parser, fail when the fixture is absent —
parity inputs are committed Go-owned bytes; an env-var-gated skip on an
upstream checkout makes CI parity vacuously green).
Arithmetic constants like `1<<16` and `3<<16+49152` in assertions grep as
numbers — a grep for the reported count is NOT evidence of a hardcode;
read the assertion sites instead.

## 8. Behavior-table port (missile/monster/sfx registries): scope by the TSV's fn column

Modern upstream splits these tables: the enum-to-fields mapping lives in a
TSV corpus (`misdat.tsv`: addFn/processFn columns), parsed at load into a
vector indexed by enum ordinal. Port order: (1) count the corpus rows and
group ids by their `processFn`/`addFn` cell (`csv.DictReader` over the TSV
gives the exact per-fn id counts in one pass — the table C++ no longer
lists them); (2) register EVERY id from the corpus with stub Add/Process so
the count-parity gate is green from slice 1 (stub = logged no-op + upstream
cite, never panic — the registry must be exhaustive before any behavior
lands); (3) port behavior in archetype clusters sharing a core primitive
(projectiles share UpdateMissilePos/UpdateMissileVelocity/PutMissile;
explosions share duration-countdown + hit-window), one cluster per contract,
each cluster's tests table-driven with hand-computed fixed-point
trajectories and scripted tapes; (4) ids whose fn cell is EMPTY are
animation-only — assert that class separately instead of inventing behavior.
Entity interactions inside a Process body (monster/player hit calls) go
through the existing injected-callback seam — the entity damage functions
belong to the entity beans, not the missile contract.
