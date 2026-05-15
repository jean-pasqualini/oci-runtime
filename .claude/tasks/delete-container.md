technical_name: delete-container
description: Add a `delete` subcommand to remove a previously created/stopped container, cleaning up its on-disk state under --root and any leftover infrastructure resources (namespaces, mounts, ipc artifacts).

tasks:
- [x] Wire the `delete` subcommand in `cmd/oci-runtime/` (register in `cmd.go`, hook into `main.go`); accept `--root` and a container name argument, mirroring `create`/`start`/`run` flag conventions.
- [ ] Add a delete handler under `internal/app/` that loads the container state from `--root/<name>` and orchestrates teardown via the infrastructure layer.
- [ ] Refuse deletion when the container is in a running state (read state, return a clear error); only proceed when stopped/created.
- [ ] Release linux resources via `infrastructure/linux/` — unmount the rootfs, drop namespaces, remove cgroup/proc artifacts left by create/start.
- [ ] Tear down any ipc transport endpoints under `infrastructure/transport/ipc/` that were established for this container.
- [ ] Remove the container's state directory `--root/<name>` last, after infrastructure teardown succeeds.
- [ ] Apply the existing middleware chain in `internal/app/mw/` (logging, error mapping) to the new handler for consistency.
- [ ] Manual verification: create a container, delete it, confirm state dir and resources are gone; then try deleting a running container and confirm the refusal path.

Memory:
- Task 1 (2026-05-15): Registered `delete` subcommand in `cmd/oci-runtime/cmd.go` with `<name>` arg and required `--root` flag, mirroring `create` flag style (local required flag rather than relying on the global one). Action is a stub returning `cli.Exit("delete not implemented yet", 2)`; wiring through the `Actions` struct + `main.go` is deferred to task 2, when `app.DeleteCmd` / `NewDeleteHandler` exist. Build verified via `go run -tags medium ./cmd/oci-runtime check`.

how to test:
1. `run` subcommand with `--root /tmp/state --bundle /app/bundle <name>`, stop it, then invoke the new `delete --root /tmp/state <name>` and check `/tmp/state/<name>` is removed and no stale mounts/namespaces remain.
2. `create` a container without starting it, `delete` it, verify clean teardown.
3. Start a container, attempt `delete` while running — expect a non-zero exit with a clear "container is running" error and no state changes.
4. Use the `check` subcommand before/after to confirm the platform state is still healthy.
