---
# devilGo-5wsb
title: 'Asset pipeline: MPQ, sprite decode, palettes, txtdata, audio'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - phase-1
created_at: 2026-10-15T16:37:32Z
updated_at: 2026-10-16T14:26:20Z
parent: devilGo-i56j
blocked_by:
    - devilGo-qna9
---

# Asset pipeline

## Goal
From retail MPQs + devilutionX assets to Go-loadable data: MPQ archive stack with priorities, CEL/CL2/CLX/PCX decode to paletted frames, palette+TRN, fonts, TSV tables, WAV/MP3 sound — everything rendering and gameplay consume.

## Research
- MPQ load order/priorities: Source/engine/assets.cpp:505-594,704-760; priorities Source/engine/assets.hpp:309-312 (devilutionx 9000, lang 9100, fonts 9200, main 1000, hellfire 8000, hfmonk 8100, hfmusic 8200, hfvoice 8500, hfbard 8110, hfbarb 8120, mods ≥10000).
- Compression needed: PKWARE DCL/zlib/bzip2 only — mpqfs mpq_stream.c:160-231; go-mpq supports exactly these read-side, Huffman not implemented (audio-only): https://pkg.go.dev/github.com/suprsokr/go-mpq
- CLX: 6-byte frame header, CL2 pixel RLE — Source/engine/clx_sprite.hpp:15-27 (+CL2.md link); loaders Source/engine/load_cel.cpp, load_clx.cpp, load_pcx.cpp; TRN Source/engine/trn.cpp; palettes/light tables Source/engine/light_tables.hpp:16-23, palette.cpp:175.
- txtdata format: assets/txtdata/Readme.md (TSV, BOM allowed); parser Source/data/file.hpp:37-110; loaders tables/itemdat.cpp:710, misdat.cpp:442, monstdat.cpp:409-472, objdat.cpp:257; playerdat.cpp:203 baseMeleeToHit; experience tables/playerdat.cpp:91.
- Sound: Source/engine/sound.cpp:281 Aulib init; default 22050 Source/options.cpp:55,520; music wav→mp3 swap :61-64; sfx table effects.cpp:190-226; Ebiten audio skill /home/blaze/.agents/skills/ebiten/audio/SKILL.md.
- Full study: docs/research/dg-requirements.md §6, Appendices D.5/J.2.

## Scope
internal/assets: mpq.go (stack+priority, go-mpq wrapper), embedded unpacked devilutionx assets (go:embed), clx/cl2/cel/pcx/trn/palette/font decoders, txtdata parser + generated tables, sfx/music via ebiten audio.Context(22050).

## Acceptance criteria
- [ ] Extraction smoke test: 100% files in spawn.mpq + diabdat.mpq + hellfire.mpq decode via go-mpq; failures listed in CI artifact (proves Huffman not needed)
- [ ] Priority stack: file present in devilutionx.mpq-equivalent overrides spawn.mpq (unit test)
- [ ] CLX decode: N fixture frames decode to identical index arrays vs C++ dump (golden files)
- [ ] PCX decode: ui_art background pixel-exact
- [ ] TRN applied to palette produces expected 256-entry table (fixture)
- [ ] All txtdata tables load with exact row counts (monstdat 113 data lines, spells, missiles, items, towners, quest_dialog, Experience 50 rows)
- [ ] txtdata edge cases pass (crlf/cr/lf_no_trail/empty/bom fixtures from test/fixtures/txtdata)
- [ ] SFX plays via ebiten audio at 22050 Hz; music track switch works; no per-frame allocation in audio path

## Skills for coder
/home/blaze/.agents/skills/ebiten/resources/SKILL.md — asset loading and embed patterns
/home/blaze/.agents/skills/ebiten/audio/SKILL.md — audio.Context, wav/mp3 decode
ego-dev — binary decoding discipline (little-endian, bounds checks)
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — golden-file tests

## Dependencies & blockers
blocked_by dg-epic-01 (module tree, harness). Blocks dg-epic-03 (render), dg-epic-04 (tables), dg-epic-08 (save writes MPQs).

## Open questions
- go-mpq listfile presence in retail archives (fallback: embedded listfile from devilutionX knowledge).

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:scope not amended for DECIDED #2 extract-at-build (runtime mpq.go stack framing violates "no runtime MPQ reading") + skills typo "ego-dev"


## IMPL NOTES (triage 2026-10-16)
1 root: Turns retail MPQs + txtdata into Go-loadable data. Wired blockedBy qna9, blocks qbiy/97mn/ftbu ✓. 8 children.
2 local: Cites verified: assets.hpp:309-312 priorities (1000/9000/9100/9200) ✓ + hf 8100/8110/8120/8200/8500 in assets.cpp:565-585 ✓; tables/monstdat.cpp:409-415 loader ✓; itemdat.cpp:710 ✓; misdat.cpp:442 ✓; data/file.hpp:37-50 BOM handling ✓; Experience.tsv = 51 lines (50 data rows + header) ✓ AC. Skills: "ego-dev" is a TYPO for go-dev (no such path).
3 online: pkg.go.dev/suprsokr/go-mpq claim (PKWARE DCL/zlib/bzip2 read-side, no Huffman) — verified during plan wave; AC itself (100% decode smoke test) is the live proof, so no pre-trust needed.
4 solution: (a) Scope line still says `internal/assets: mpq.go (stack+priority, go-mpq wrapper)` = runtime MPQ reading — contradicts DECIDED #2 (extract-at-build, per AGENTS.md "No runtime MPQ reading") and repo AGENTS.md §5 pipeline (extract.sh → convert → manifest). Reframe: build-time extractor producing loose assets + manifest; runtime consumes manifest only; runtime MPQ reader = deferred post-1.0 bean (decision says NOT cut). (b) Fix "ego-dev" → go-dev. ACs themselves are compatible (decode smoke test = build-time gate).
5 understanding: diablo/assets layer supply line; child vof6 ("MPQ archive stack over go-mpq") must inherit the build-time reframe (batch-03 to verify).
6 repro: `grep -n "Priority" Source/engine/assets.cpp | head`; `sed -n '309,312p' Source/engine/assets.hpp`; `grep -n "No runtime MPQ" AGENTS.md`; `ls /home/blaze/.agents/skills/go-dev 2>&1`
7 fix: `beans update devilGo-5wsb --body-replace-old "internal/assets: mpq.go (stack+priority, go-mpq wrapper), embedded unpacked devilutionx assets (go:embed), clx/cl2/cel/pcx/trn/palette/font decoders, txtdata parser + generated tables, sfx/music via ebiten audio.Context(22050)." --body-replace-new "Build-time extraction pipeline (per DECIDED #2 + AGENTS.md §5): extractor tool over go-mpq → assets/loose + generated manifest.json; runtime internal/assets loads loose assets + manifest via go:embed (NO runtime MPQ reading); clx/cl2/cel/pcx/trn/palette/font decoders, txtdata parser + generated tables, sfx/music via ebiten audio.Context(22050). Runtime MPQ reader = deferred post-1.0 feature bean."` and `beans update devilGo-5wsb --body-replace-old "ego-dev — binary decoding discipline" --body-replace-new "/home/blaze/.agents/skills/go-dev/SKILL.md — binary decoding discipline"`
8 validation: body shows extract-at-build scope + go-dev path; child vof6 reframed consistently (batch-03 check).
9 related: qna9 (blocker), qbiy+97mn+ftbu (blocked), children incl. vof6, n89r, zvwd, jx08, b782, ow6p, b8z9, l5fp.
VERDICT: NEEDS-WORK:scope not amended for DECIDED #2 extract-at-build (runtime mpq.go stack framing violates "no runtime MPQ reading") + skills typo "ego-dev"

Decision-alignment (triage): scope updated for extract-at-build — mpq.go runtime reader removed from scope; build-time extraction + go:embed loose assets + decoders remain.
