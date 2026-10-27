# Pi delegation reliability research

Date: 2026-10-21

## Executive conclusion

The current delegation problem is primarily a **harness/orchestration failure**, not evidence that pi cannot perform the work. The dispatcher has several independent failure modes:

- the RPC client's timeout is an `agent_end` wait timeout, not a process deadline;
- the external runner can fail before launch because its output directory is created only by the child script, while the parent shell redirects into that directory;
- the scope guard is not deterministic: the live smoke failed with `comm: file 1 is not in sorted order`;
- the current client waits for top-level `agent_end`, although pi documents that `agent_end` can be followed by retries, compaction, steering, or follow-up work and recommends waiting for `agent_settled`;
- session identity is not verified after startup: pi can warn that a requested project session does not exist and create a new one;
- the worker is asked to review a whole file and run several gates in one model run, while the batch-level runner has no durable per-step checkpoint/heartbeat contract;
- the main worktree is contaminated, increasing attribution and safety risk.

The recommended architecture is a small durable queue manager around **isolated git worktrees**, with one leaf per worker, deterministic snapshots, an append-only event ledger, hard process-group cancellation, explicit heartbeats/progress, and parent-side verification. Keep pi as the first worker backend; add OpenCode v2 only behind the same backend interface after the pi path is reliable.

## Primary sources and verified implications

### Pi RPC documentation

- Pi RPC is a long-lived JSONL child process interface: https://pi.dev/docs/latest/rpc
- Each command accepts an ID and responses must be correlated by ID; stdout must be continuously consumed; clients must handle startup failure, unexpected exit, stderr diagnostics, cancellation, and their own deadlines.
- A successful `prompt` response means accepted/queued/handled, not completed.
- `agent_end` ends one low-level run. Retries, overflow recovery, compaction, steering, and follow-up work may continue; clients needing quiescence should wait for `agent_settled`.
- Pi's RPC command reference documents `steer`, `follow_up`, `abort`, `get_state`, `get_entries`, `get_session_stats`, `set_auto_retry`, `abort_retry`, and `bash`/`abort_bash`: https://pi.dev/docs/latest/rpc-commands
- Pi JSON events include tool execution start/update/end, turn start/end, queue updates, compaction, retry, and agent lifecycle events: https://pi.dev/docs/latest/json

**Required changes:** update the client to correlate request IDs, expose/handle `agent_settled`, `get_state`, retry/compaction events, and a heartbeat based on observed event timestamps. Do not treat a single `agent_end` or a zero exit code as completion.

### OpenCode v2 API and permissions

- OpenCode v2 exposes durable session prompt, wait, interrupt, inbox, diff, log, and shell operations in its HTTP API: https://opencode.ai/v2/docs/api
- OpenCode v2 supports session-scoped permissions with ordered allow/ask/deny rules and explicit resources: https://opencode.ai/v2/docs/permissions/
- OpenCode agents support separate primary/subagent profiles, permissions, models, and positive step limits: https://opencode.ai/v2/docs/agents/

**Required changes:** OpenCode can be a backend, but use its server API rather than treating `opencode run` as an opaque one-shot. Configure a dedicated worker agent with subagent denied, Beans/triage instructions disabled, explicit shell/edit/read permissions, and a finite `steps` limit. Use session interrupt/wait/log/diff endpoints. Still enforce filesystem scope in the parent because provider/tool permissions are not a substitute for postcondition checks.

### Git worktrees

Git officially supports linked worktrees with independent `HEAD` and index state, detached throwaway worktrees, porcelain listing, and safe removal: https://git-scm.com/docs/git-worktree

**Required change:** never run multiple coding workers in the contaminated main worktree. Create one detached worktree per leaf from an explicit baseline commit or immutable tree snapshot. Merge/reconcile only after verification. This removes the need to infer worker ownership from a dirty shared tree.

### Durable execution and heartbeats

Temporal's official documentation describes the relevant reliability primitives: timeouts, retry policies, heartbeats, checkpoint payloads, idempotency, and durable execution: https://docs.temporal.io/encyclopedia/failures-and-error-handling and https://docs.temporal.io/standalone-activity

**Required change:** model each leaf as a durable activity. Persist `attempt`, `started`, `last_event`, `last_heartbeat`, `checkpoint`, `process_group`, `workspace`, `input_fingerprint`, `output_fingerprint`, `verification`, and `next_action`. A retry must start from the last verified checkpoint, not from model memory. A heartbeat must distinguish provider wait from a wedged process; absence of both events and provider traffic triggers cancellation.

### Evaluation and agent design

Anthropic recommends simple composable workflows, ground truth from tool results, explicit stopping conditions, sandboxing, and evaluator/optimizer loops: https://www.anthropic.com/engineering/building-effective-agents
Anthropic's evaluation guidance distinguishes transcript/trajectory from final environment outcome and recommends deterministic graders, multiple trials, regression suites, and transcript review: https://www.anthropic.com/research/demystifying-evals-for-ai-agents

**Required change:** grade outcomes, not reports. For every leaf, the parent must run deterministic checks against the resulting workspace/diff and retain the raw transcript as diagnostic evidence. Build a small delegation eval suite from real failures: provider blip, empty exit-0, hung RPC, wrong cwd, output-redirection failure, scope violation, stale session, timeout during tool execution, and valid partial edit after kill.

### Observability

OpenTelemetry's model treats traces as correlated spans/events and logs as structured records with timestamps and trace/span context: https://opentelemetry.io/docs/concepts/signals/traces/ and https://opentelemetry.io/docs/specs/otel/logs/data-model/

**Required change:** the existing JSONL registry should become a structured event ledger with a single `run_id`, `attempt_id`, `bean`, `epic`, backend, model, cwd/worktree, contract hash, event sequence, timestamps, process IDs, provider request IDs when available, and artifact hashes. Never reconstruct lifecycle from shell notifications alone.

## Findings from the live failure

1. The parent launch used output redirection to `/tmp/dg-kefy-batch1-v2/driver.log` before that directory existed. The runner therefore failed before executing the script. The script's internal `mkdir -p` could not help.
2. The corrected smoke reached the scope guard, but `comm` failed because the snapshot inputs were not normalized under a fixed locale/order. A guard that cannot classify its own input is not a guard.
3. The current runner's `snapshot()` hashes only paths reported dirty by `git status`. That is useful for attribution, but the baseline must be captured once and then compared as a canonical sorted set under `LC_ALL=C`; all status parsing must handle rename quoting and untracked/generated files consistently.
4. The existing worktree contains 52 modified Go files, five Beans files, and previously an ELF build artifact. A shared dirty tree makes it impossible to safely infer which changes belong to a new leaf.
5. The first smoke log contained only startup narration and no final verdict. That is a failed artifact, regardless of process exit status.

## Target architecture

### Control plane

A single `delegation-runner` owns the queue and is the only writer to the ledger. It must:

1. validate repository path, branch, baseline commit/tree hash, bean/epic lineage, skill paths, contract hash, and clean isolated worktree;
2. create one detached worktree per leaf;
3. write `RUN_CREATED`, `ATTEMPT_STARTED`, `HEARTBEAT`, `EVENT`, `ATTEMPT_FINISHED`, `VERIFICATION_STARTED`, `VERIFICATION_FINISHED`, and `RUN_TERMINAL` events atomically to JSONL;
4. launch exactly one coding child per worktree, with a process group and a hard wall-clock deadline;
5. reconcile every restart from the ledger plus filesystem and git state;
6. never mark a leaf done from model text alone.

### Worker contract

Split each file into smaller bounded phases when possible:

- inspect and report plan (read-only);
- make the smallest edit;
- run the package-local checks;
- print one machine-readable verdict.

The worker must not create Beans, change status, commit, push, or touch files outside the declared worktree. For the current style sweep, a file review is a reasonable leaf, but the model should not be responsible for parent-wide gates.

### Verification contract

The parent runs, independently:

- canonical path/diff scope check;
- expected artifact/marker check;
- `gofmt -d` and package-local tests/lint;
- full repository gates only after a verified batch or milestone;
- contract-specific semantic checks;
- no untracked junk except explicitly allowed artifacts.

Verification records input and output tree fingerprints. A test result is invalid if the workspace changed after the test ran.

## Prioritized implementation plan

### P0 — required before another real batch

1. Fix launch wrapper: `mkdir -p "$ROOT/out"` in the parent before redirection, or redirect to an already-existing file path. Run `bash -n` and a launch-failure test.
2. Fix snapshot guard: use `LC_ALL=C`; write canonical `path<TAB>sha256` records; sort both operands inside the comparison; use Python set comparison instead of `comm` to eliminate locale and quoting ambiguity; add tests for modified, added, deleted, renamed, quoted, and unchanged paths.
3. Kill descendants by process group, not only the Python wrapper. Use `start_new_session=True` in the Python client or `setsid`, then TERM the process group and KILL after grace. Verify no descendant remains.
4. Add preflight checks that fail closed on missing cwd, list, contract, skill, output directory, wrong branch, or active worker lock.
5. Record launch metadata before spawning and reconcile registry status after every exit. Never claim a runner started without a PID and artifact directory.
6. Keep the current main tree frozen. Do not start batch work in it. Use a detached worktree from an explicitly named baseline.

### P1 — make runs observable and resumable

1. Add unique request IDs and handle `agent_settled`, `retry_*`, `compaction_*`, `tool_execution_*`, and unexpected EOF in the pi client.
2. Emit heartbeat events on every received frame and on provider-call-log advancement. Detect stalled process only from both signals being stale.
3. Add an append-only event ledger and a small `status` command that reports truth from PID, ledger, workspace fingerprint, last event, last provider traffic, and artifact existence.
4. Add attempt budgets and exponential backoff with jitter. Retry only classified transient failures; fork a new session after poisoned/degenerate output. Never reuse a session after a failed artifact gate.
5. Persist per-leaf checkpoint after every verified phase; resume from the leaf and phase, not from a monolithic batch prompt.

### P2 — improve quality and model routing

1. Create a dedicated pi worker profile with a short system contract and pass only task-relevant skills.
2. Add a read-only evaluator pass after the worker, using a separate model/session and independent evidence.
3. Maintain a delegation eval suite and run it on every launcher/client change and model-route change.
4. Add OpenCode v2 as an alternate backend through the same runner interface; use its HTTP session API, permissions, `steps`, `wait`, `interrupt`, and `diff`, not opaque `run` calls.
5. Measure per backend/model: completion rate, artifact-valid rate, scope-violation rate, timeout rate, median time, retries, provider errors, token/cost proxy, and human rework.

## Concrete design decision

Do **not** switch from pi to OpenCode merely because this run stalled. The observed failures are in the surrounding launcher and dirty-worktree strategy. First make the backend-neutral runner reliable with pi. Then A/B pi and OpenCode behind identical contracts, isolated worktrees, deterministic graders, and the same retry/time budgets. Select by measured verified outcome, not by final prose or a green wrapper exit code.

## References

- Pi RPC: https://pi.dev/docs/latest/rpc
- Pi RPC commands: https://pi.dev/docs/latest/rpc-commands
- Pi JSON events: https://pi.dev/docs/latest/json
- OpenCode v2 API: https://opencode.ai/v2/docs/api
- OpenCode v2 agents: https://opencode.ai/v2/docs/agents/
- OpenCode v2 permissions: https://opencode.ai/v2/docs/permissions/
- Git worktrees: https://git-scm.com/docs/git-worktree
- Temporal failures: https://docs.temporal.io/encyclopedia/failures-and-error-handling
- Temporal standalone activities: https://docs.temporal.io/standalone-activity
- Anthropic effective agents: https://www.anthropic.com/engineering/building-effective-agents
- Anthropic agent evals: https://www.anthropic.com/research/demystifying-evals-for-ai-agents
- OpenTelemetry traces: https://opentelemetry.io/docs/concepts/signals/traces/
- OpenTelemetry logs: https://opentelemetry.io/docs/specs/otel/logs/data-model/
