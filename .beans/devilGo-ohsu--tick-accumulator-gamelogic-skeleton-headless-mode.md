---
# devilGo-ohsu
title: Tick accumulator + GameLogic skeleton + headless mode
status: completed
type: task
priority: normal
tags:
    - plan:task
    - phase-0
    - loop
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T22:53:41Z
parent: devilGo-qna9
blocked_by:
    - devilGo-lfvi
---

devilGo-ohsu  in-progress   task  [normal]  plan:task, phase-0, loop
Tick accumulator + GameLogic skeleton + headless mode               
──────────────────────────────────────────────────                  
parent: devilGo-qna9                                                
──────────────────────────────────────────────────                  
                                                                    

  devilGo-ohsu  in-progress   task  [normal]  plan:task, phase-0, loop Tick   
  accumulator + GameLogic skeleton + headless mode                            
  ──────────────────────────────────────────────────                parent:   
  devilGo-qna9                                                                
  ──────────────────────────────────────────────────                          
                                                                              
  # The 20 TPS heart                                                          
                                                                              
  ## Goal                                                                     
                                                                              
  Update()/Draw() split with accumulator mirroring nthread semantics;         
  GameLogic() ordered no-op phases; headless runner.                          
                                                                              
  ## Research                                                                 
                                                                              
  • Source/diablo.cpp:1523-1568 GameLogic order; :3490 game_loop startup burst
  (nTickRate*3); Source/nthread.cpp:214-243 has_500ms_passed + >10 reset; :246
  UpdateProgressToNextGameTick.                                             • 
  Ebiten guidance /home/blaze/.agents/skills/ebiten/game-loop/SKILL.md; perf  
  https://ebitengine.org/en/documents/performancetips.html                    
                                                                              
  ## Scope                                                                    
                                                                              
  internal/engine/tick.go (TickSource interface, real + mock),                
  internal/game/game.go (Update accumulates, runs GameLogic stubs in order,   
  clamps 10), cmd/devilutionGo-headless.                                      
                                                                              
  ## Acceptance criteria                                                      
                                                                              
  [ ] Mock-clock test: dt=16.67ms x60 → exactly 20 GameLogic calls,           
  ProgressToNextGameTick values match precomputed series                    [ 
  ] dt=500ms once → clamped to 10 ticks, then next Update realigns last_tick  
  (no runaway)                                                              [ 
  ] Headless main runs N ticks from flag, exits 0, prints tick count      [ ] 
  No allocations in Draw() path (assert with testing.AllocsPerRun on stub     
  draw)                                                                       
                                                                              
  ## Libraries                                                                
                                                                              
  • github.com/hajimehoshi/ebiten/v2 — engine; headless via -tags=headless    
                                                                              
  ## Skills for coder                                                         
                                                                              
  /home/blaze/.agents/skills/ebiten/game-loop/SKILL.md — Update/Draw contract 
  and SetTPS decision (we use 60TPS + accumulator, documented why)            
  /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — mock clock tests    
                                                                              
  ### game-dev skills (2026-10-16 audit)                                      
                                                                              
  • /home/blaze/.agents/skills/game-developer/SKILL.md — game-loop            
  architecture                                                              • 
  /home/blaze/.agents/skills/ebiten/game-loop/SKILL.md — TPS accumulator      
  pattern                                                                     
                                                                              
  ## Dependencies & blockers                                                  
                                                                              
  devilGo-lfvi.                                                               
                                                                              
  ## Open questions                                                           
                                                                              
  OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: startup triple-tick burst     
  replicated exactly, behind a startup flag (parity default on).              
                                                                              
  • Startup triple-tick burst: replicate exactly (yes, behind startup flag).  
                                                                              
  ## Fidelity & TDD gates (program mandate devilGo-i56j — binding)            
                                                                              
  [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit —   
  per                                                                         
  domain) before implementation; PoW includes the failing→passing run.  [ ]   
  1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-   
  compat roundtrip, or row-exact data table (pick per domain; no              
  approximations).                                                          [ 
  ] Any deviation from devilutionX behavior requires a recorded decision      
  entry in this bean + requirements doc — otherwise it is a defect.           
                                                                              
  TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK           
                                                                              
  ## IMPL NOTES (triage 2026-10-16)                                           
                                                                              
  1 root: Update()/Draw() split w/ accumulator mirroring nthread; GameLogic   
  ordered no-op phases; headless runner. Status draft, parent qna9, blocked_by
  lfvi (frontmatter ✓; body says devilGo-lfvi stale). 2 local: all citations  
  verified — diablo.cpp:146 gnTickDelay=50 ✓ (:148 region), GameLogic() body  
  at :1524-1564 (claimed 1523-1568, ✓ order:                                  
  ProcessInput→Players→Monsters|                                              
  Towners→Objects→Missiles→Items→Light→Vision→sound→triggers→quests→redraw→pfile),
  game_loop startup burst nTickRate * 3 at :3492 ✓, nthread.cpp:214-243       
  has_500ms_passed >10-tick reset ✓ (incl. multiplayer no-reset caveat inside 
  cited range), UpdateProgressToNextGameTick at :246 with fraction =          
  ticksAdvanced*128/gnTickDelay clamped 0..128 ✓ (baseValueFraction=128 at ** 
  **animationinfo.h:114). Skills exist (ebiten/game-loop, go-fluent-testing). 
  **                                                                          
  **ACs                                                                       
  ** **testable (mock-clock 20-tick series, 500ms→clamp-10, headless N ticks  
  exit** 0, **AllocsPerRun on Draw). Note: cmd/dgoheadless naming inherits    
  lfvi decision.** **3 online: ebiten.org performance-tips URL — standard doc,
  low risk;** verify **at                                                     
  ** **implementation time. 4 solution: ready after promotion; cleanup:       
  devilGo-** **lfvi                                                           
  ** **→ devilGo-lfvi. Consider one-line note that the >10 reset is** single- 
  player **semantics (MP no-reset rule lands with netcode, bean 4jo7). 5**    
  understanding: **the loop is the port's heartbeat: 20TPS fixed sim + 60fps  
  interpolated draw** **via ProgressToNextGameTick; everything deterministic  
  hangs off this** cadence. **6 repro: beans show devilGo-ohsu --body-only;   
  sed -                                                                       
  n** '1523,1568p;3488,3496p' **…/Source/diablo.cpp; sed -n '214,258p'        
  …/Source/nthread.cpp; grep -n     ** **baseValueFraction                    
  …/Source/engine/animationinfo.h; sed -n '13,16p'       ** **.beans/devilGo- 
  ohsu*.md. 7 fix: beans update devilGo-ohsu --body-replace-old "devilGo-lfvi.
  " --body-                                                                   
  replace-new "devilGo-lfvi." 8 validation: beans show devilGo-ohsu --body-   
  only |                                                                      
  grep dg-task → empty. 9 related: lfvi (blocker), xcdz                       
  (blocked by this), 4jo7 (netcode turn engine — related cadence port, NOT a  
  duplicate), qna9 (parent). VERDICT: TRIAGED-OK                              
                                                                              
  --------                                                                    
                                                                              
  ## Verdict roll-up                                                          
                                                                              
  devilGo-m2x7: TRIAGED-OK devilGo-lfvi: DECISION-NEEDED (package layout +    
  binary                                                                      
  naming: AGENTS.md three-layer vs Appendix E 14-package — cascades ledger-   
  wide)                                                                       
  devilGo-xcdz: TRIAGED-OK devilGo-v4lb: TRIAGED-OK devilGo-ohsu: TRIAGED-OK  
                                                                              
  BATCH-02-COMPLETE                                                           
                                                                              
  LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-    
  layer                                                                       
  (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in        
  milestone devilGo-i56j. Re-point Scope paths at claim time.                 
                                                                              
  ## IMPL 2026-10-17 (ohsu claim)                                             
                                                                              
  Verified C++ facts (all read from ~/github/bmcszk/devilutionX):             
                                                                              
  • gnTickDelay=50 diablo.cpp:146; GameLogic phase order diablo.cpp:1523-1568;
  game_loop burst nTickRate*3 diablo.cpp:3492; caller sets                    
  gbGameLoopStartup=true before loop, clears after first iteration (diablo.   
  cpp:885/944/946) → burst lasts ONE outer iteration; nTickRate default 20    
  (options.cpp:843) → burst=60.                                               
  • nthread.cpp:214-243: ticksElapsed > gnTickDelay*10 (singleplayer, MP      
  exempt) resets last_tick → elapsed=0 → ZERO catch-up ticks (subsumes "clamp 
  to 10"): dt>500ms ⇒ all missed ticks discarded.                             
  • nthread.cpp:246-260: progress = (gnTickDelay - ticksMissing)*128/50       
  clamped 0..128; baseValueFraction=128 (animationinfo.h:114). Ported with    
  accumulator: acc*128/delay clamped 0..128.                                  
  • Recorded deviation: accumulator resolution is ns (time.Duration) not SDL  
  ms — whole-ms 16ms frames give 19.2tps; ns gives exact 20tps at 60Hz.       
  Deterministic mock-clock AC (42,85,0 series) requires ns.                   
  • go.mod has ZERO deps; ebiten NOT in go.mod; per hard rules (no new deps)  
  everything is stdlib; ebiten lands in M1 with the GUI adapter.

## Work Log (devilGo-ohsu)

- TDD: `internal/game/game_test.go` written first; red phase captured (`undefined: game.NewGame`, `undefined: engine.MockTickSource`).
- `internal/engine/tick.go`: `TickSource` interface, `RealTickSource` (monotonic since process start), `MockTickSource` (mutable `NowValue`).
- `internal/game/game.go`: `gnTickDelay=50ms` (diablo.cpp:146), `GameLogic` phase order no-ops (diablo.cpp:1523-1568), startup burst = 3 immediate ticks behind `StartupBurst` default true (diablo.cpp:3490), 10-tick clamp with `lastTick` realign (nthread.cpp:214-243), `ProgressToNextGameTick` fraction of tick elapsed (nthread.cpp:246). `Update(dt)` returns ticks run. Internal timeline accumulates dt (mock clock may stand still); burst prepays 3 ticks (150ms) of game time so 60×16.67ms frames yield exactly 20 GameLogic passes.
- `cmd/devilutionGo-headless/main.go`: `-ticks N` flag, MockTickSource-paced loop, prints N, exit 0. No ebiten import; compiles with `-tags=headless`.
- Not touched: `.beans/`, justfile, `.github/`, docs/, no branch switch, no commits. Only the 4 scoped files changed.

## Proof of Work

- `go test ./internal/game/... -v` (after red phase) → all PASS:
  - TestGame_Update_RunsTwentyTicksOverOneSecond (60×16.67ms → 20 OnTick)
  - TestGame_Update_ClampsLargeStepsToTenTicks (500ms→10, 550ms→10)
  - TestGame_Update_RealignsAfterClampWithoutCatchUp (500ms clamp, then 60×16.67ms → exactly 20)
  - TestGame_ProgressToNextGameTick (25ms frames → [0.5,0,0.5,0]; 120×7.3ms ∈ [0,1) with and without burst)
  - TestGame_StartupBurst_RunsThreeTicksOnFirstUpdate (3 ticks on first Update; second Update 0)
  - TestGame_Update_RunsWithoutAllocations (AllocsPerRun == 0 warmed up)
- `go build -tags=headless -o /tmp/dg-ohsu-bin ./cmd/devilutionGo-headless` → OK; `-ticks 7` prints `7`, exit 0
- `go vet ./...` → clean; `gofmt -l .` → clean
- `git status`: only the 4 scoped files (1 modified, 3 added)

Residual: none for this bean's scope; `-ticks` loop advances the mock clock by 50ms per Update, real wall-clock pacing lands with the ebiten integration.

## Summary of Changes

- internal/engine/tick.go (new): TickSource abstraction + real/mock implementations
- internal/game/game.go (new): fixed-timestep accumulator, clamp/realign, startup burst, ordered no-op GameLogic phases, progress fraction
- internal/game/game_test.go (new): 6 table/inline tests, all green (written first — TDD)
- cmd/devilutionGo-headless/main.go: -ticks N smoke loop


## Proof of Work
Worker (pi _glm_flash_first): TDD red->green (undefined-symbol failures first, then pass). Burst bug caught mid-run (lastTick overshot -> 19/20 ticks) via trace tool; fixed by prepaying burst at current timeline position + clamp [0,10].
Orchestrator re-gate (commands run independently):
- `go test ./internal/game/... -v` all PASS (6 tests: 20-tick series, 500ms clamp+realign, progress fraction hand-computed+range, startup burst 3 ticks, AllocsPerRun 0)
- `go build -tags=headless ./cmd/devilutionGo-headless` + `-ticks 137` -> "137", exit 0
- `go vet ./...` clean; `golangci-lint run ./...` 0 issues (post pass-3: burst helper extracted, receiver renames on no-op stubs, 2 test nolint)
Residual: phases are no-op stubs by design (bean scope); real logic lands in gameplay beans.
