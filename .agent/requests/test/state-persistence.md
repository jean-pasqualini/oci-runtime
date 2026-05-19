# Test Plan: state-persistence

## 1. Identification
- Plan ID:           TP-state-persistence
- Feature:           .agent/requests/feature/state-persistence.md
- Version / commit:  workflow-ai-dev @ b81f6e6
- Author:            Jean Pasqualini
- Date created:      2026-05-19
- Last executed:     never

## 2. Objective
Verify that container metadata is persisted to `<root>/<name>/state.json` at create time and that `domain.ContainerState.Status` is correctly derived on Load from the combination of `exec.fifo` presence and `/proc/<pid>` liveness. A passing run demonstrates that the three lifecycle statuses (`created`, `running`, `stopped`) resolve end-to-end and that the dead-code refusal path in `delete.go` ("container is running") is now reachable.

## 3. Scope
- In scope:
  - `state.json` written under `<root>/<name>/state.json` at create time
  - Persisted shape: `Name`, `Pid`, `Bundle`, `CreatedAt`; `Status` not persisted
  - `exec.fifo` removal at start time
  - Status derivation on Load:
    - `exec.fifo` present → `created`
    - `exec.fifo` absent + `/proc/<pid>` alive → `running`
    - `exec.fifo` absent + `/proc/<pid>` gone → `stopped`
  - `state.json` as the canonical existence gate (replaces former `exec.fifo` gate)
  - Delete refusal on running container (refusal path now reachable)
  - Delete success on `created` and `stopped` containers
- Out of scope:
  - Atomic write semantics (tmp+rename was intentionally dropped — see feature task 2 memory)
  - Concurrent create/delete races
  - Cleanup of namespaces / mounts / ipc artifacts beyond state dir
  - `list` / `kill` / `stop` handlers (not yet implemented)

## 4. References
- Feature request:   .agent/requests/feature/state-persistence.md
- Related code:
  - internal/domain/container_state.go (status constants + ContainerState fields, `json:"-"` on Status)
  - internal/app/ports.go (`ContainerStateManager.Save`)
  - internal/infrastructure/container/state.go (`Manager.Save`, `Manager.Load` w/ status derivation)
  - internal/app/create.go (`createHandler` persists state.json)
  - internal/app/start.go (`startHandler` removes `exec.fifo` after kick byte)
  - cmd/oci-runtime/main.go (shared `container.NewManager()` wiring)
- External specs:    OCI runtime-spec — lifecycle (created / running / stopped)
- Related test plan: .agent/requests/test/delete-container.md (TC-007 was the dead-code probe; now live)

## 5. Test Environment
- Build / launch:    `go run -tags medium ./cmd/oci-runtime <subcommand>`
- Runtime / OS:      Linux container with `/app` workspace mount (per CLAUDE.md)
- Required fixtures: `/app/bundle` is a valid OCI bundle; `/tmp/state` writable and empty
- Required tooling:  `cat`, `ls`, `test`, `jq` (optional, for state.json field inspection), `kill`

## 6. Test Data
- containerNames:  c1, c2, c3, c4, ghost
- root:            /tmp/state
- bundle:          /app/bundle

## 7. Entry Criteria
- Build is green: `go run -tags medium ./cmd/oci-runtime check` exits 0
- No residual `/tmp/state/{c1,c2,c3,c4,ghost}` directories before TC-001
- No leftover init processes from prior runs (`ps -ef | grep oci-runtime` empty besides this shell)

## 8. Exit Criteria
- All P0 and P1 cases: Pass
- No open Critical or High defects against state-persistence
- Execution Log filled for every case (Pass / Fail / Blocked / Skipped, never Not Run)

## 9. Test Cases

---

### TC-001 baseline check is healthy
- Priority:        P1
- Type:            Smoke
- Precondition:    Entry criteria met
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime check`
- Expected result: exit 0; check output reports no platform issues

### TC-002 state.json written at create time with expected shape
- Priority:        P0
- Type:            Functional
- Precondition:    TC-001 Pass; `/tmp/state/c1` does not exist
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime create --root /tmp/state --bundle /app/bundle c1`
  2. `ls /tmp/state/c1/state.json`
  3. `cat /tmp/state/c1/state.json`
- Expected result: step 1 exits 0; step 2 lists the file; step 3 prints JSON containing keys for `Name` (== c1), `Pid` (positive int), `Bundle` (== /app/bundle), `CreatedAt` (RFC3339 timestamp). No `Status` field present (or empty), per `json:"-"` tag.

### TC-003 exec.fifo present after create, absent after start
- Priority:        P0
- Type:            Functional
- Precondition:    TC-002 Pass; `c1` still in `created` state from TC-002
- Steps:
  1. `ls /tmp/state/c1/exec.fifo`  (must exist after create)
  2. `go run -tags medium ./cmd/oci-runtime start --root /tmp/state c1`
  3. `ls /tmp/state/c1/exec.fifo` (must be gone after start)
- Expected result: step 1 lists the fifo; step 2 exits 0; step 3 reports "No such file or directory".

### TC-004 delete on `created` container succeeds, state dir removed
- Priority:        P0
- Type:            Functional
- Precondition:    TC-001 Pass; `/tmp/state/c2` does not exist
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime create --root /tmp/state --bundle /app/bundle c2`
  2. Verify `/tmp/state/c2/state.json` exists and `/tmp/state/c2/exec.fifo` exists (Status would resolve to `created`)
  3. `go run -tags medium ./cmd/oci-runtime delete --root /tmp/state c2`
  4. `ls /tmp/state/c2`
- Expected result: step 3 exits 0; step 4 reports "No such file or directory".

### TC-005 delete on running container is refused with "container is running"
- Priority:        P0
- Type:            Negative
- Precondition:    TC-001 Pass; `/tmp/state/c3` does not exist; bundle long-running enough to remain alive during step 2
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime run --root /tmp/state --bundle /app/bundle c3` (do not wait — run in background or another shell)
  2. While the container init is alive (confirm `/proc/<pid>` exists, pid read from `/tmp/state/c3/state.json`), in another shell: `go run -tags medium ./cmd/oci-runtime delete --root /tmp/state c3`
  3. `ls /tmp/state/c3` — should still exist
- Expected result: step 2 exits non-zero; stderr contains "container is running"; `/tmp/state/c3` untouched.

### TC-006 delete on `stopped` container succeeds
- Priority:        P0
- Type:            Functional
- Precondition:    TC-001 Pass; `/tmp/state/c4` does not exist
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime run --root /tmp/state --bundle /app/bundle c4` (background)
  2. Read init pid from `/tmp/state/c4/state.json`
  3. `kill -9 <pid>` and wait for `/proc/<pid>` to disappear
  4. Verify `/tmp/state/c4/exec.fifo` is absent and `/proc/<pid>` is gone (Status would resolve to `stopped`)
  5. `go run -tags medium ./cmd/oci-runtime delete --root /tmp/state c4`
  6. `ls /tmp/state/c4`
- Expected result: step 5 exits 0; step 6 reports "No such file or directory".

### TC-007 delete on unknown container returns "container not found" via state.json gate
- Priority:        P1
- Type:            Negative
- Precondition:    `/tmp/state/ghost` does not exist
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime delete --root /tmp/state ghost`
- Expected result: non-zero exit; stderr contains "container not found"; error KV references `root`, `name`, `state_file` (not `exec_fifo` — gate moved per feature decision).

### TC-008 Status not persisted in state.json (json:"-" enforcement)
- Priority:        P1
- Type:            Functional
- Precondition:    TC-002 Pass (or rerun create on a fresh name)
- Steps:
  1. `cat /tmp/state/c1/state.json` (or equivalent for the most recently created container)
  2. Inspect for any `Status` / `status` key
- Expected result: no `Status` / `status` key in the JSON payload. If present, the test fails.

### TC-009 check still healthy after the full matrix
- Priority:        P2
- Type:            Regression
- Precondition:    TC-002..TC-007 executed
- Steps:
  1. `go run -tags medium ./cmd/oci-runtime check`
- Expected result: exit 0; output identical in shape to TC-001.

---

## 10. Risks & Known Gaps

- **Risk — TC-005 timing**: the delete in step 2 must fire before init exits naturally. Use a long-running bundle or insert a `sleep` between steps. If init exits first, mark TC-005 Blocked and retry.
- **Risk — TC-006 pid reuse**: between `kill -9 <pid>` and `delete`, the kernel could reuse the pid for an unrelated process, flipping the derived Status back to `running` and triggering a spurious refusal. Probability is low on a quiet test container; if it happens, mark Blocked and rerun.
- **Test gap — no `list` handler**: cannot enumerate persisted state without reading the filesystem directly. Each TC inspects `/tmp/state/<name>/state.json` by path.
- **Test gap — Save is not exercised on overwrite**: feature task 2 intentionally dropped the tmp+rename atomic-replace pattern since Save is only called once on a fresh dir today. Future overwriting handlers (stop/kill/update) will need a dedicated plan to cover crash-during-write.
- **Feature gap — CreatedAt timezone / precision not specified**: TC-002 only asserts a valid RFC3339 timestamp. If a stricter contract emerges (UTC only, ms precision, etc.), tighten the assertion.

## 11. Execution Log

- TC-001: baseline check is healthy
  Status: NOT_RUN
  Iteration: 0
  Notes:
- TC-002: state.json written at create time with expected shape
  Status: NOT_RUN
  Iteration: 0
  Notes:
- TC-003: exec.fifo present after create, absent after start
  Status: NOT_RUN
  Iteration: 0
  Notes:
- TC-004: delete on `created` container succeeds, state dir removed
  Status: NOT_RUN
  Iteration: 0
  Notes:
- TC-005: delete on running container is refused with "container is running"
  Status: NOT_RUN
  Iteration: 0
  Notes:
- TC-006: delete on `stopped` container succeeds
  Status: NOT_RUN
  Iteration: 0
  Notes:
- TC-007: delete on unknown container returns "container not found" via state.json gate
  Status: NOT_RUN
  Iteration: 0
  Notes:
- TC-008: Status not persisted in state.json (json:"-" enforcement)
  Status: NOT_RUN
  Iteration: 0
  Notes:
- TC-009: check still healthy after the full matrix
  Status: NOT_RUN
  Iteration: 0
  Notes:
