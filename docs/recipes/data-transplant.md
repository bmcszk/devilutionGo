# Data-transplant recipe (Phase-7 gamedata beans)

The proven loop for `internal/gamedata/<domain>` beans (spells, missiles,
monsters, objects, audio, text, quests, towners, levels, enums). Follow in
order; every step exists because skipping it cost a fix round.

## 1. Claim + measure the corpus FIRST

```bash
# row counts — awk, NOT wc -l (corpus files end without trailing newline;
# wc -l undercounts by exactly 1 and poisons every parity test):
# $DX points at the devilutionX checkout — OFFLINE PRODUCER STEP ONLY,
# never referenced by the Go tests themselves.
awk 'END{print NR}' $DX/assets/txtdata/<domain>/<table>.tsv
# column shapes — numeric-sounding columns are often enum-NAME strings:
head -1 <table>.tsv | tr '\t' '\n' | cat -n      # column list
awk -F"\t" 'NR>1{print $COL}' <table>.tsv | sort -u   # real values
# distinct-value counts for registry/enum ACs:
... | sort -u | wc -l
```

Read the upstream loader (sed -n the cited range) before typing the struct:
column semantics, empty-value handling, and any custom parse function
(e.g. treasure `Uniq(NAME)` / `None` decoding) live there, not in the TSV.

## 2. RED: tests against COMMITTED Go-owned fixture bytes

The live-corpus read in step 1 is PRODUCER-side only: it happens once, offline,
before the worker contract is written. The committed test path is:

- Copy the measured rows (or the probe's canonical parsed dump) into
  `test/fixtures/**` and COMMIT the bytes + a manifest (producer commit +
  sha256 per file). Tests read the committed fixture; a missing/corrupt
  fixture FAILS the test — never `t.Skip` on an absent upstream checkout
  (`DG_DEVILUTIONX`-style gating makes CI parity vacuously green).
- Row-count parity tests assert committed fixture counts (fail on wrong
  count). Include a **positive hf-absence drift guard** for tables upstream
  has no hf override for.
- Golden rows copied verbatim as Go string constants pin column order and
  CRLF handling; row-count assertions belong in tests, not the report.
- Enum/registry completeness: assert distinct-value counts AND that every
  data-referenced name resolves to a registered implementation (and no
  extra entries).

## 3. GREEN: strict parser + data structures

- Strict TSV split: header-driven, ragged row = error; tolerate the hf
  quirk of ONE trailing empty column (strip it), empty numerics = 0.
- Keep a Raw() accessor when byte-parity assertions need the source line.
- Enum tables copied verbatim from the .hpp in declaration order; count
  constants asserted in tests.
- Function-pointer tables (AiProc, addFn/processFn) become name-keyed
  registries; null entries (FireMan) stay null and resolve-by-name still
  finds them.

## 4. hf delta shape: compare BY ID SET

Positional diffs lie when hf inserts mid-file. Measure: shared ids
(content-equal?), appended ids (count + no base dupes), reordered ids
(content identical, position changed). Encode the measured shape in tests.

## 5. AC discrepancies: re-measure, encode, record

Bean bodies carry planner estimates that disk measurement corrects
(1122→1080 class rows, "10 modified"→reorder artifact, 10→9 movies,
51→50 named registry entries). When measured truth differs: tests assert
measured values, PoW gets a numbered decision entry citing the disk
evidence. Never silently ship either the bean number or a silent fix.

## 6. Close

Full gates (`gofmt -l`, test ./internal/..., vet, golangci-lint), PoW with
decisions + Residual, `beans update -s completed`, commit (split unrelated
diffs), push, CI to terminal. Out-of-scope AC items (need Player types,
RNG probe, Q_* enums) get a recorded move to the consumer bean, never a
silent drop and never stubs pretending coverage.
