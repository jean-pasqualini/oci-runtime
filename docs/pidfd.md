# pidfd

## What is pidfd?

`pidfd` (process file descriptor) is a Linux kernel feature that provides a stable, race-free handle to a process. Instead of identifying a process by its PID (which can be recycled), a `pidfd` is a file descriptor that refers to a specific process instance.

Introduced progressively starting with Linux 5.1 and expanded in later kernels (5.2, 5.3, 5.4).

## The problem it solves

Traditional PIDs suffer from a race condition:

1. Process A obtains PID `1234`.
2. Process `1234` exits.
3. The kernel recycles PID `1234` to a new, unrelated process.
4. Process A sends a signal to PID `1234` — and unintentionally signals the wrong process.

A `pidfd` is bound to a specific process instance. Once that process exits, the `pidfd` continues to refer to the same (now-terminated) process — it is never recycled.

## How to obtain a pidfd

### 1. `pidfd_open(2)`

Open a pidfd for an existing process by PID:

```c
int pidfd = syscall(SYS_pidfd_open, pid, 0);
```

### 2. `clone(2)` / `clone3(2)` with `CLONE_PIDFD`

Receive a pidfd at the moment of process creation — eliminates any race window:

```c
struct clone_args args = {
    .flags     = CLONE_PIDFD,
    .pidfd     = (uint64_t)(uintptr_t)&pidfd,
    .exit_signal = SIGCHLD,
};
pid_t pid = syscall(SYS_clone3, &args, sizeof(args));
```

### 3. `/proc/<pid>` directory fd

Opening `/proc/<pid>` with `O_DIRECTORY` yields a pidfd-compatible fd in recent kernels.

## What you can do with a pidfd

### Send signals — `pidfd_send_signal(2)`

Race-free replacement for `kill(2)`:

```c
syscall(SYS_pidfd_send_signal, pidfd, SIGTERM, NULL, 0);
```

### Wait for process exit — `poll(2)` / `epoll(7)`

A pidfd becomes readable when the referenced process terminates. This lets you wait for a process the same way you wait for I/O — fully integratable into an event loop:

```c
struct pollfd pfd = { .fd = pidfd, .events = POLLIN };
poll(&pfd, 1, -1);  // returns when process exits
```

This is a major win: previously, waiting for a child required `SIGCHLD` handling or blocking `waitpid()`.

### Reap status — `waitid(2)` with `P_PIDFD`

```c
siginfo_t info;
waitid(P_PIDFD, pidfd, &info, WEXITED);
```

### Duplicate file descriptors — `pidfd_getfd(2)`

Steal an fd out of another process (requires `PTRACE_MODE_ATTACH_REALCREDS`):

```c
int remote_fd = syscall(SYS_pidfd_getfd, pidfd, target_fd, 0);
```

## Relevance to an OCI runtime

A container runtime spends most of its life managing the lifecycle of a single child process (the container init). `pidfd` is well-suited for this:

| Concern | Without pidfd | With pidfd |
|---|---|---|
| Wait for container exit | Blocking `waitpid` or `SIGCHLD` handler | `poll`/`epoll` on the fd — fits the event loop naturally |
| Send signals (`stop`, `kill`) | `kill(pid, sig)` — PID race possible | `pidfd_send_signal` — safe |
| Persist process identity across runtime restarts | Re-read PID from disk, hope it's the same | pidfd can be passed via SCM_RIGHTS or stored in a holder process |
| Integrate with a supervisor loop | Mix of signal-driven + poll-driven code | Uniform fd-driven model |

For this runtime, using `clone3(CLONE_PIDFD)` when spawning container init gives:

- A guaranteed-correct handle to the container process from the very first instruction.
- Race-free `stop` / `kill` / `delete` commands, even if the PID is recycled by the kernel.
- Clean integration of "container exited" into any future `epoll`-based supervisor.

## Caveats

- **Kernel version**: `pidfd_open` requires ≥ 5.3; `CLONE_PIDFD` requires ≥ 5.2; `pidfd_getfd` requires ≥ 5.6.
- **No `pidfd_open` for arbitrary processes pre-5.3** — only your own children via `CLONE_PIDFD`.
- **Not inherited across `exec`** of unrelated processes — pidfds are FDs and follow normal FD semantics (close-on-exec by default with `pidfd_open`).
- **Cross-PID-namespace**: a pidfd obtained in one PID namespace still works in another (it identifies the process, not its visible PID).

## References

- `man 2 pidfd_open`
- `man 2 pidfd_send_signal`
- `man 2 pidfd_getfd`
- `man 2 clone3`
- LWN: "The pidfd API" — https://lwn.net/Articles/794707/
