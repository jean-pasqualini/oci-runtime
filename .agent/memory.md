### Memory

#### Container state persistence (2026-05-18)
- Create handler (`internal/app/create.go`) does **not** write any `state.json`. It only:
  - creates dir `<MetadataRoot>/<Name>/` (mode 0777)
  - writes `<MetadataRoot>/<Name>/exec.fifo` (named pipe, mode 0622)
  - optionally writes the init PID to `cmd.PidFile` (path provided by caller, not necessarily inside the state dir)
- `domain.ContainerState` (`internal/domain/container_state.go`) is `{Name, Status string}` — status field exists but is never persisted by any handler today.
- Implication for `ContainerStateLoader` (port in `internal/app/ports.go`): must **infer** state from filesystem artifacts (dir presence, exec.fifo presence, pid file + `/proc/<pid>` liveness). Not a simple JSON unmarshal.
- Status vocabulary partially defined: `domain.StatusRunning = "running"` exists in `internal/domain/container_state.go` (added by delete-container Task 3). Other statuses (created, stopped) not yet introduced.
- `delete` handler (`internal/app/delete.go`) refuses when `state.Status == domain.StatusRunning`, but the check is **dead code today** — no handler writes `Status`, so the loader will always return an empty string. Refusal becomes live only when a future feature persists status.

#### Handler inventory wired in `cmd/oci-runtime/main.go`
Create, Start, Delete, Init, Check. No Stop, no Run handler currently wired through `Actions`.
