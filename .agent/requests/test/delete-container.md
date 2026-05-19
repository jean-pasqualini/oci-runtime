# Test Plan: delete-container

## 1. Identification
- Plan ID:           TP-delete-container
- Feature:           .agent/requests/feature/delete-container.md
- Version / commit:  workflow-ai-dev @ 878db72
- Author:            Jean Pasqualini
- Date created:      2026-05-18
- Last executed:     2026-05-19

## 2. Objective
Verify that the `delete` subcommand removes a previously created or stopped container's on-disk state under `--root`, refuses to delete a running container, and surfaces clear errors for unknown containers. A passing run demonstrates the delete-handler + `container.Manager` integration is safe to ship for stopped/created containers.

## 3. Scope
- In scope:
  - `delete` subcommand wiring (cmd -> action -> handler -> Manager)
  - State-dir removal under `<root>/<name>`
  - "container not found" error path
  - Running-container refusal path (probe only — see §10)
  - No regression on `check` after delete runs
- Out of scope:
  - Cleanup of namespaces / mounts / ipc artifacts (not yet implemented — see §10)
  - Concurrent delete behaviour
  - Bundle teardown / image GC

## 4. References
- Feature request:   .agent/requests/feature/delete-container.md
- Related code:
  - cmd/oci-runtime/cmd.go (subcommand wiring)
  - internal/app/delete.go (handler, refusal check)
  - internal/app/ports.go (`ContainerStateManager` port)
  - internal/infrastructure/container/state.go (`Manager` impl)
- External specs:    OCI runtime-spec — lifecycle / delete operation

## 5. Test Environment
- Build / launch:    `go run -tags medium ./cmd/oci-runtime <subcommand>`
- Runtime / OS:      Linux container with `/app` workspace mount (per CLAUDE.md)
- Required fixtures: `/app/bundle` is a valid OCI bundle; `/tmp/state` writable and empty
- Required tooling:  `mount`, `lsns`, `ip` (iproute2), `ls`, `test`

## 6. Test Data
- containerNames:  c1, c2, c3, ghost
- root:            /tmp/state
- bundle:          /app/bundle

## 7. Entry Criteria
- Build is green: `go run -tags medium ./cmd/oci-runtime check` exits 0
- No residual `/tmp/state/{c1,c2,c3,ghost}` directories before TC-001
- `mount | grep -E 'c1|c2|c3'` empty; `lsns | grep -E 'c1|c2|c3'` empty

## 8. Exit Criteria
- TC-001..TC-006: Pass
- TC-007: documented outcome (Pass or Fail with reference to feature gap — see §10)
- No open Critical or High defects against delete-container
- Execution Log filled for every TC; none left Not Run

## 9. Test Cases

---

### TC-001 baseline check is healthy
- Priority:        P1
- Type:            Smoke
- Precondition:    Entry criteria met
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime check`
- Expected result: exit 0; check output reports no platform issues

### TC-002 delete removes state dir after a run completes
- Priority:        P0
- Type:            Functional
- Precondition:    TC-001 Pass; `/tmp/state/c1` does not exist
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime run --root /tmp/state --bundle /app/bundle c1`
  2. Wait for the container process to exit
  3. Verify `/tmp/state/c1` still exists (state left behind by run)
  4. `go run -tags medium ./cmd/oci-runtime delete --root /tmp/state c1`
  5. `ls /tmp/state/c1`
- Expected result: step 4 exits 0; step 5 reports "No such file or directory"

### TC-003 delete removes state dir after create (no start)
- Priority:        P0
- Type:            Functional
- Precondition:    TC-001 Pass; `/tmp/state/c2` does not exist
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime create --root /tmp/state --bundle /app/bundle c2`
  2. Verify `/tmp/state/c2/exec.fifo` exists
  3. `go run -tags medium ./cmd/oci-runtime delete --root /tmp/state c2`
  4. `ls /tmp/state/c2`
- Expected result: step 3 exits 0; step 4 reports "No such file or directory"

### TC-004 delete on unknown container returns "container not found"
- Priority:        P1
- Type:            Negative
- Precondition:    `/tmp/state/ghost` does not exist
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime delete --root /tmp/state ghost`
- Expected result: non-zero exit; stderr contains "container not found"; error has KV with `root`, `name`, `exec_fifo`

### TC-005 no leftover infrastructure resources after delete
- Priority:        P1
- Type:            Integration
- Precondition:    TC-002 Pass; TC-003 Pass
- Steps:
  1. `mount | grep -E 'c1|c2'`
  2. `lsns | grep -E 'c1|c2'`
  3. `ip netns list | grep -E 'c1|c2'` (or `true` if `ip` unavailable)
- Expected result: all three commands produce no matching lines

### TC-006 check still healthy after delete runs
- Priority:        P2
- Type:            Regression
- Precondition:    TC-002..TC-005 executed
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime check`
- Expected result: exit 0; output identical in shape to TC-001

### TC-007 delete-while-running refusal (known feature gap)
- Priority:        P1
- Type:            Negative
- Precondition:    `/tmp/state/c3` does not exist
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime run --root /tmp/state --bundle /app/bundle c3` (do not wait)
  2. While the container is still running, in another shell: `go run -tags medium ./cmd/oci-runtime delete --root /tmp/state c3`
- Expected result: non-zero exit; stderr contains "container is running"; `/tmp/state/c3` untouched

---

## 10. Risks & Known Gaps

- **Feature gap — refusal-path unreachable**: `container.Manager.Load` always returns `Status == ""` because no component persists status yet. The refusal check at `internal/app/delete.go:38-43` is correct but dead code in practice. Closing this requires a status persister (out of scope of feature `delete-container`; see feature memory task 4). TC-007 documents the gap; a Fail here is expected today and does **not** block exit criteria.
- **Feature gap — infra cleanup**: feature description mentions cleaning up "namespaces, mounts, ipc artifacts". `delete` today only removes the state dir. TC-005 surfaces leftovers as findings, not test failures — record observations and file follow-ups rather than blocking.
- **Test gap — environment tooling**: `ip`, `lsns`, or `mount` may be missing inside minimal containers. TC-005 should be marked Blocked (not Fail) in that case, and the env upgraded before re-running.
- **Risk — port 9 timing**: TC-007 depends on the container still running when step 2 fires. Use a long-running bundle or insert a short wait; if the container exits before step 2, mark Blocked and retry.

## 11. Execution Log

- TC-001: baseline check is healthy
  Status: PASSED
  Iteration: 2
  Notes: Re-run 2026-05-19 (post state-persistence task 2). check exit 0; CAP_SYS_ADMIN effective. Output shape identical to iteration 1.
- TC-002: delete removes state dir after a run completes
  Status: PASSED
  Iteration: 1
  Notes: run c1 ok; /tmp/state/c1 left after run; delete exit 0; dir removed.
- TC-003: delete removes state dir after create (no start)
  Status: PASSED
  Iteration: 1
  Notes: create c2 ok; exec.fifo present; delete exit 0; dir removed.
- TC-004: delete on unknown container returns "container not found"
  Status: PASSED
  Iteration: 1
  Notes: exit 1; stderr "container not found"; KV root=/tmp/state name=ghost exec_fifo=/tmp/state/ghost/exec.fifo.
- TC-005: no leftover infrastructure resources after delete
  Status: PASSED
  Iteration: 1
  Notes: mount/lsns/ip netns list — no matching lines for c1|c2.
- TC-006: check still healthy after delete runs
  Status: PASSED
  Iteration: 1
  Notes: exit 0; same shape as TC-001.
- TC-007: delete-while-running refusal (known feature gap)
  Status: FAILED (documented per §10)
  Iteration: 3
  Notes: Re-run 2026-05-19 @ b81f6e6. Container c3 started (init pid 117 running /bin/sh -l via run subcommand); delete issued while running → exit 0, log "container deleted"; /tmp/state/c3 removed. Same outcome as iterations 1 & 2 — Manager.Load returns Status=="" so refusal check at internal/app/delete.go:38-43 remains dead code. Not blocking exit criteria.