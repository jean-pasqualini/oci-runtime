Great target! Here are practical ways to discover the syscalls a binary actually uses and turn that into a seccomp profile.

1) Quick & reliable: strace

Record everything the program (and its children) does under realistic workload.

# Run and produce a per-thread file: /tmp/trace.<pid>
strace -ff -o /tmp/trace -e trace=all -- your-binary args...

# Or get a summary table (counts, time)
strace -f -c -- your-binary args...

Get the unique syscall names (for a profile):

grep -ho '^[a-z0-9_]\+' /tmp/trace.* | sort -u > /tmp/syscalls.txt

Pros: simple, accurate for what actually ran.
Cons: misses syscalls not exercised by your test paths.

2) Low-overhead & system-wide: eBPF/bpftrace

Trace kernel tracepoints for syscalls; filter by process name or PID.

# By command name
sudo bpftrace -e 'tracepoint:syscalls:sys_enter_* /comm=="your-binary"/ { printf("%s\n", probe); }' \
| sed 's/tracepoint:syscalls:sys_enter_//' | sort -u

# By PID (start binary, note PID)
sudo bpftrace -e 'tracepoint:syscalls:sys_enter_* /pid==12345/ { printf("%s\n", probe); }'

You can also count frequency:

sudo bpftrace -e 'tracepoint:syscalls:sys_enter_* /comm=="your-binary"/ { @[probe] = count(); }'

Pros: very low overhead, good for containers/production-like runs.
Cons: requires bpftrace; needs root.

3) Auditing: auditd

Good when you want persistent logs (shows up in journalctl -k / ausearch).

sudo auditctl -D
sudo auditctl -a always,exit -F arch=b64 -S all -F exe=/full/path/your-binary -k learnseccomp
# run workload here
sudo ausearch -k learnseccomp | aureport -s

Pros: built-in auditing.
Cons: noisier to set up; parse reports.

4) Containers: tracing inside or from host
   •	Docker/Podman: run the container unconfined (or permissive) while tracing with bpftrace/strace from the host (filter by container PID namespace / comm).
   •	Tools like Tracee (Aqua Security) are great for container syscall capture via eBPF.

⸻

Turn the syscall set into an OCI seccomp profile

Minimal allow-list with deny-by-default:

{
"defaultAction": "SCMP_ACT_ERRNO",
"architectures": ["SCMP_ARCH_X86_64"],
"syscalls": [
{
"names": [
"read","write","exit","exit_group","futex","nanosleep",
"clock_gettime","brk","mmap","mprotect","munmap","close",
"rt_sigaction","rt_sigprocmask","rt_sigreturn","openat","statx","getrandom"
],
"action": "SCMP_ACT_ALLOW",
"args": [],
"comment": "your-binary observed syscalls",
"includes": {},
"excludes": {}
}
]
}

Apply (Docker example):

docker run --security-opt seccomp=profile.json your-image

“Learning mode”

While discovering, you can log instead of block:
•	With modern libseccomp, use SCMP_ACT_LOG as defaultAction (or for unknowns).
•	Then run workload, read dmesg / journalctl -k for SECCOMP logs, add missing syscalls, switch back to ERRNO.

⸻

Tips & caveats
•	Exercise all code paths (init, steady state, error handling, shutdown).
•	Syscalls vary by arch (x86_64 vs arm64) and libc (glibc vs musl). Generate per target.
•	Go binaries often use futex, nanosleep, clone, sched_*, epoll_*, openat, statx, mmap*, etc. Don’t guess—trace.
•	Consider allowing families as needed (e.g., epoll_wait, epoll_ctl).
•	For threads, add clone if you see failures.
•	Keep the profile as tight as possible, but add a small “safety” set for signals and memory mgmt to avoid brittle crashes.

⸻

If you tell me your target (host/container, arch) and I/O needs (network? files?), I can generate a starter seccomp JSON and a one-liner to auto-extract syscalls from strace/bpftrace into it.