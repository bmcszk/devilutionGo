---
# devilGo-gmtn
title: 'De-transliteration round: idiomatic Go structure, features stay 1:1 (user mandate 2026-10-23)'
status: todo
type: epic
priority: high
created_at: 2026-10-23T16:10:34Z
updated_at: 2026-10-23T16:11:37Z
---

devilGo-zk5z  todo   task  [high]                                                            
De-transliteration round: idiomatic Go structure, features stay 1:1 (user mandate 2026-10-23)
──────────────────────────────────────────────────                                           
                                                                                             

  Brainer audit (devilGo-cppaudit-1, _glm_first thinking-high, report         
  /tmp/devilgo-sweep/report-cppaudit.md) found 9 C++-transliteration          
  structures                                                                  
  in 8 files. Port mandate: FEATURES 1:1, CODE idiomatic Go. Fixes via pi     
  dispatch per finding cluster; comments/citations NOT in scope (kc11 closed  
  those). Findings ranked worst-first:                                        
                                                                              
  1. animation/animationinfo_test.go:10-60 — step/tape gtest harness (F/T/R + 
  run interpreter) -> per-scenario table subtests calling SetNewAnimation+    
  ProcessAnimation inline (file's own second half shows the pattern)          
  2. animation/animationinfo.go:44-47 — trailing-underscore member names (C++ 
  private idiom) -> drop underscores                                          
  3. animation/animationinfo.go:99-103 — 7-positional-param func mirrors C++  
  signature (ignored sink, dummy 0) -> named Config struct or split call      
  shapes                                                                      
  4. combat/combat_test.go:14,69,119 — fixture type + expect() oracle = hand- 
  copied player.cpp arithmetic -> literal want values per table case; keep tap
  for RNG seam                                                                
  5. missiles/collision_test.go:63-78,147-172 — rollArgs mirrors C++ arg pack -
  > named want fields on case struct                                          
  6. game/world/tileprops_test.go:180,199,226-253 — opaque [4]int tuples +    
  gtest-fixture-named tests -> struct{from,to Point}, keep citations          
  7. missiles/process_ii_test.go:28-29 — [][4]int tuple spy -> named call     
  structs                                                                     
  8. missiles/process_test.go:58-59 — same tuple spy -> same fix              
  9. combat/combat_fuzz_test.go:16-20 — triple-nested C++ param loop, rand.New
  not testing.F -> t.Run table or real F.Fuzz                                 
  EXCLUDED as parity-by-design (not style): rotation tapes, codec/diablosha   
  goldens, random goldens, missilelist order, process tapes (recordings still 
  flagged), 5 nolint-cited ports.                                             
                                                                              
  ## Skills for coder                                                         
                                                                              
  • /home/blaze/.agents/skills/coding-rules/SKILL.md — read fully BEFORE      
  acting                                                                      
  • /home/blaze/.agents/skills/ponytail/SKILL.md — read fully BEFORE acting   
  • /home/blaze/.agents/skills/go-dev/SKILL.md — read fully BEFORE acting     
                                                                              
  ## Acceptance criteria                                                      
                                                                              
  [ ] All 9 findings fixed, each verified by parent re-read of the diff       
  [ ] Test semantics preserved: same scenarios, same assertions strength      
  (rotation/parity tapes untouched)                                           
  [ ] Gates per cluster: gofmt/vet/golangci-lint 0/test ok; full suite before 
  push                                                                        
  [ ] Final: go test ./... green, push, CI success                            
                                                                              
  ## Triage                                                                   
                                                                              
  Brainer audit (pi-as-brainer lane per brainer-delegation HARD rule): session
  devilGo-cppaudit-1, --model _glm_first --thinking high --tools read,grep,   
  find,ls --ns,                                                               
  READ-ONLY, resumed once for the impl scan. Child read all 45 test files +   
  impl scan, returned 9 findings with file:line + idiomatic alternative +     
  EXCLUDE list; parent verified exhibit finding against the live file         
  (step/tape harness at animationinfo_test.go:10-60 confirmed) before         
  promoting                                                                   
  to todo. User mandate verbatim: rewrite cpp app to go with ALL features 1 to
  1 but NOT the code 1 to 1.                                                  
                                                                              
  ## Libraries                                                                
                                                                              
  • none new (test/impl restructuring only; no dependency additions)          



## Skills for coder
- /home/blaze/.agents/skills/coding-rules/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/ponytail/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/go-dev/SKILL.md — read fully BEFORE acting
- /home/blaze/.hermes/skills/testing/port-parity-testing/SKILL.md — read fully BEFORE acting (idiom boundary section governs)

## Triage
Delegated brainer audit (session devilGo-cppaudit-1, _glm_first thinking-high, READ-ONLY) swept all 45 test files + impl; 9 findings, ranked, with EXCLUDE list. Parent verified the flagged exhibit against the live file before promoting. Full findings embedded in the body above; split into 3 cluster tasks.


## Acceptance criteria
- [ ] All 9 audit findings fixed; each diff parent-re-read
- [ ] No transliteration pattern remains (brainer spot-check re-run on touched files)
- [ ] Full gates: gofmt/vet/golangci-lint 0, go test ./... green, CI success
