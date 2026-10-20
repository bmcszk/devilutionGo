---
# devilGo-w67z
title: Diablo-codec + Diablo-SHA1 ports
status: completed
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - crypto
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-17T13:57:40Z
parent: devilGo-6xu2
---

devilGo-w67z  in-progress   task  [normal]  plan:task, phase-2, crypto
Diablo-codec + Diablo-SHA1 ports                                      
──────────────────────────────────────────────────                    
parent: devilGo-6xu2                                                  
──────────────────────────────────────────────────                    
                                                                      

  # Save crypto                                                               
                                                                              
  ## Goal                                                                     
                                                                              
  codec_encode/decode + custom SHA1 exact.                                    
                                                                              
  ## Research                                                                 
                                                                              
  codec.cpp:92-160; sha.cpp:8-25 (custom circular shift); tests codec_test.cpp.
                                                                              
  ## Scope                                                                    
                                                                              
  internal/diablo/save/{codec,diablosha}.go.                                  
                                                                              
  ## Acceptance criteria                                                      
                                                                              
  [ ] codec_test port green incl. wrong-password failure path                 
  [ ] Block size 32 + signature layout asserted                               
                                                                              
  ## Libraries                                                                
                                                                              
  • stdlib only (no third-party deps at this stage)                           
                                                                              
  ## Skills for coder                                                         
                                                                              
  /home/blaze/.agents/skills/go-dev/SKILL.md — crypto without crypto pkg      
  (custom algorithm)                                                          
                                                                              
  ## Dependencies & blockers                                                  
                                                                              
  None.                                                                       
                                                                              
  ## Open questions                                                           
                                                                              
  None.                                                                       
                                                                              
  ## Fidelity & TDD gates (program mandate devilGo-i56j — binding)            
                                                                              
  [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit —   
  per domain) before implementation; PoW includes the failing→passing run.    
  [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-
  compat roundtrip, or row-exact data table (pick per domain; no              
  approximations).                                                            
  [ ] Any deviation from devilutionX behavior requires a recorded decision    
  entry in this bean + requirements doc — otherwise it is a defect.           
                                                                              
  TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK           
                                                                              
  ## IMPL NOTES (triage 2026-10-16)                                           
                                                                              
  1 root: Persistence epic leaf — codec_encode/decode + custom SHA1; no       
  blockers (matches tree devilGo-w67z unblocked; correct: crypto has no deps).
  2 local: codec.cpp:92 codec_decode verified; codec.cpp:15-21                
  CodecSignature{checksum,error,lastChunkSize} + SignatureSize 8 +            
  BlockSizeBytes layout verified (AC "block size 32 + signature layout"       
  assertable); sha.cpp:8-25 custom arithmetic-shift SHA1 comment +            
  SHA1CircularShift verified; test/codec_test.cpp exists. 3 online: n/a. 4    
  solution: none — cleanest bean in batch: C++ test suite portable as golden  
  vectors first (TDD-first trivially satisfiable), wrong-password path is in  
  codec_test. 5 understanding: persistence-layer crypto leaf feeding oqho/xwtt
  (hero pack encryption via pfile passwords, pfile.cpp:60-64 verified). 6     
  repro: beans show devilGo-w67z; sed -n '92,96p' .../Source/codec.cpp; sed -n
  '8,25p' .../Source/sha.cpp; ls                                              
  ~/github/bmcszk/devilutionX/test/codec_test.cpp. 7 fix: none. 8 validation: 
  n/a. 9 related: oqho, xwtt, 6xu2. VERDICT: TRIAGED-OK                       
                                                                              
  LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-    
  layer                                                                       
  (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in        
  milestone devilGo-i56j. Re-point Scope paths at claim time.


## Proof of Work (2026-10-17)

TDD: tests written first (`codec_test.go`, `diablosha_test.go`), confirmed RED (`no non-test Go files in .../internal/diablo/save`), then implemented to green. 14 tests pass (`go test ./internal/diablo/save/ -count=1 -v`: TestSHA1CircularShiftSignExtensionQuirk, TestSHA1CalculateGoldenFromFreshContext, TestSHA1CalculatePasswordBlockGolden, TestCodecInitKeyGolden, TestCodecGetEncodedLen, TestCodecEncodeGoldenBytes, TestCodecEncodeDecodeRoundtrip, TestCodecSignatureLayout, TestCodecDecodeWrongPassword, TestCodecDecodeTamperedSignature, TestCodecDecodeMalformedSizes, TestCodecEncodeRejectsInvalidSize64, TestCodecDecodeEmptyEncodeIsFailure, TestCodecDifferentialAgainstCppDump).

Citations (read this session, ~/github/bmcszk/devilutionX):
- sha.cpp:16-24 SHA1CircularShift sign-extension quirk: `if ((word & (1 << 31)) != 0) return (0xFFFFFFFF << bits) | (word >> (32 - bits));` — ported verbatim with uint32 arithmetic.
- sha.cpp:26-90 SHA1ProcessMessageBlock: 80-word expansion `w[i]=w[i-16]^w[i-14]^w[i-8]^w[i-3]`, four 20-round loops (0x5A827999/0x6ED9EBA1/0x8F1BBCDC/0xCA62C1D6), state += round vars.
- sha.h:10-11 `BlockSize = 16` (words), `SHA1HashSize = 5`; sha.h:13 IV 67452301 EFCDAB89 98BADCFE 10325476 C3D2E1F0. NOTE: bean AC said "block size 32" — upstream truth is 16 words = 64 bytes; asserted upstream values (test/codec_test.cpp: 50→72, 128→136 implies 64-byte blocks + 8-byte signature). Contract's "BlockSize=17?" probe also resolves to 16 words.
- codec.cpp:19-53 CodecInitKey (password repeat → SHA1 → xor with 16 constant mask words codec.cpp:40-42 → SHA1).
- codec.cpp:92-130 codec_decode: `size <= SignatureSize → 0`; `size -= 8`; `size % BlockSize != 0 → 0` (upstream compares BYTES against BlockSize=16 WORDS — a real save always passes since real lengths are 64-byte multiples; ported verbatim); per block: load→xor(sha)→SHA1Calculate→store; sig error byte (codec.cpp:117-120) and checksum (codec.cpp:123-128) failure paths return 0.
- codec.cpp:132-138 codec_get_encoded_len: pad to 64 + 8. codec.cpp:140-173 codec_encode: chunk≤64, pad zero, xor AFTER SHA1Calculate (opposite order vs decode), signature `[checksum LE32][error=0][lastChunkSize][0][0]`.
- codec.cpp:66-70 ByteSwapBlock uses Swap32LE = SDL_Swap32LE (endian_swap.hpp:31) = **no-op on LE hosts**. Port is host-independent LE load/store (encoding/binary), byte-identical to upstream-on-x86.
- test/codec_test.cpp ported: codec_get_encoded_len(50)=72, (128)=136 (TestCodecGetEncodedLen).

Parity evidence: verbatim upstream sha.cpp + codec.cpp compiled on x86-64 (g++), goldens dumped (/tmp/dg-w67z-ref): diablo-SHA1("abc" block) = 45A07558 D5B60157 A31F2F7B 0BCB00D4 B7E9798B (vs real SHA1 abc = A9993E36 4706816A BA3E2571 7850C26C 9CD0D89D — quirk proven); CodecInitKey("xrgyrkj1") = 2BF46D62 9423E44E 97CBBED2 6C4F1519 23135736 (pfile.cpp:62 password, plus szqnlsk1/adslhfb1/lshbkfg1/password/"" all asserted); full 72-byte encode of 50 pattern bytes byte-exact (TestCodecEncodeGoldenBytes); 40 random (size 0-199, 8/16-char passwords) C++-encoded dumps decode exactly in Go (TestCodecDifferentialAgainstCppDump).

Known UB (documented divergence, not ported): codec.cpp:28-31 restarts the 4-byte password window on '\0' but re-reads past the terminator when strlen%4≠0 (C++ OOB read; Go zero-pads deterministically). All devilutionX passwords are 8 chars (pfile.cpp:60-63) → unreachable in production. Second: upstream `size%BlockSize(16)` decode gate admits non-64-byte lengths, whose trailing partial block reads OOB in C++ (codec.cpp:101); Go bounds-check panics — also unreachable for real saves (codec_get_encoded_len only emits 64-byte multiples).

Gates: `go test ./internal/... -count=1` all ok (assets, demo, diablo/save, engine, game) • `go vet ./...` clean • `golangci-lint run ./...` → 0 issues • `gofmt -l internal/` clean. No third-party deps (stdlib `encoding/binary`, `encoding/hex` in tests only). Untouched: .beans (status line only, pre-set by dispatch), .github/, justfile, docs/, cmd/, internal/{game,engine,assets,demo}/.

Residual: none blocking; two upstream-UB divergences documented above.


## Proof of Work
Worker (pi _glm_flash_first) completed FULLY this time (report + gates done before exit; lane healthy ~20min).
Independent re-gate by orchestrator:
- 14 tests PASS incl. codec roundtrip, wrong-password checksum-mismatch failure path (codec.cpp:123-128), block/signature layout asserts, SHA1 vectors
- Upstream fidelity: SHA1CircularShift sign-extension quirk ported exactly (sha.cpp:16-24); codec per-block swap/xor/calculate chain (codec.cpp:92-130); signature layout (codec.cpp:16-17)
- Worker documented 2 upstream-UB spots NOT ported (password restart past terminator when strlen%4≠0 — unreachable for real 8-char passwords; decode size%16 gate admitting OOB lengths — real saves emit 64-byte multiples): deviations recorded per program mandate
- go vet clean; golangci-lint 0 issues; stdlib only
