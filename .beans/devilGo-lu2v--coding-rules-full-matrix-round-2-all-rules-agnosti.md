---
# devilGo-lu2v
title: 'coding-rules full matrix round 2: ALL rules (agnostic+Go+Go-test) x ALL 104 files'
status: in-progress
type: epic
priority: high
created_at: 2026-10-23T19:55:42Z
updated_at: 2026-10-23T21:20:15Z
---

User directive: find ALL rules (language-agnostic + Go + Go-test), list them, list all files, EVERY FILE x EVERY RULE. Method: (1) rule extraction from all skills (brainer, 3 parallel extractions running), (2) rule list consolidated into /tmp/devilgo-sweep/ALL-RULES.md, (3) per-package brainer batches (2 batches, ~9k lines each) judging every rule x every file, (4) fix batches via pi, (5) parent gates + residual ledger, (6) bean closes with PoW.

## Acceptance criteria
- [ ] ALL-RULES.md exists with every rule from every skill (M/J tagged)
- [ ] 104/104 files judged against every J-rule; M-rules verified mechanically
- [ ] Violations fixed via pi; gates green; CI success
- [ ] Residual ledger (legit + WHY) in bean PoW


## Progress log
- 2026-10-23 claimed: rule extraction (3 brainer extractions: code-style, testing, repo-gates) -> ALL-RULES.md (126 rules: 60 M / 66 J); file list 104 (units.json); 2 judgment batches dispatched (b0: 32 files, b1: 72 files).
