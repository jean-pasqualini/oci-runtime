technical_name: stop-container
description: Allow the user to stop a previously started container by sending a termination signal to its init process.

tasks:
- [ ] Add `stop` subcommand wiring in cmd/oci-runtime/main.go and cmd.go (flags: --root, --bundle; arg: container name)
- [ ] Create stop handler in internal/app/ (load state, validate "running", signal init pid, update state)
- [ ] Read container state from --root/<container-name> via existing state loader
- [ ] Validate container status is "running" before signalling (refuse otherwise)
- [ ] Send SIGTERM to container init PID via infrastructure/linux/ (add helper if missing)
- [ ] Support optional --timeout flag; escalate to SIGKILL after timeout
- [ ] Update container state file to "stopped" after process exits
- [ ] Manual test using run subcommand then stop subcommand

Memory:
(fill one line per completed task describing what was actually done)

how to test:
1. Run `run --root /tmp/state --bundle /app/bundle mycontainer` in background
2. Run `stop --root /tmp/state --bundle /app/bundle mycontainer`
3. Verify init process gone (no such pid)
4. Verify state file at /tmp/state/mycontainer reports status "stopped"
5. Edge case: stop on non-existent container -> clear error
6. Edge case: stop on already stopped container -> clear error
