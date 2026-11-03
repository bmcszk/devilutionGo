---
# devilGo-mihf
title: txtdata loader accepts a lone CR where upstream returns BadRecordTerminator
status: draft
type: bug
priority: low
created_at: 2026-11-03T00:51:14Z
updated_at: 2026-11-03T00:51:14Z
---

Discovered during the asset verification wave (devilGo-5wsb, AC7).

## Evidence

- Upstream: Source/data/parser.cpp:4-27 — on '\r' it advances once and then requires '\n'; anything else returns Status::BadRecordTerminator, so a CR-only record separator is a hard error.
- Ours: internal/assets/txtdata.go normalizes CR and CRLF to LF, and TestLoadNormalizesLineEndings (internal/assets/txtdata_test.go:36) asserts that cr.tsv — CR-only records — loads identically to the CRLF export.
- The fixture README (internal/testdata/assets/txtdata/README.md) states the upstream rule, so docs and test contradict each other today.

## Impact

None for real inputs: the committed corpus and the retail txtdata are CRLF-only. It is a fidelity deviation, not a live failure, so it is recorded rather than patched mid-wave.

## Fix shape

Reject a CR that is not followed by LF (mirror HandleRecordTerminator), and turn cr.tsv into a rejection fixture. Needs a decision first: strict 1:1 vs documented leniency.
