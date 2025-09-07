Yes 👍 — there are cheat-sheet style references that focus on the common / portable syscalls that appear on both amd64 (x86_64) and arm64 (aarch64).

⸻

🔹 Big picture
	•	Most syscalls are shared across architectures (open, read, write, mmap, clone, etc.).
	•	Differences:
	•	Some legacy x86 syscalls don’t exist on arm64 (e.g. socketcall, ipc).
	•	Newer syscalls like openat2, clone3 appear first on modern kernels and archs.
	•	Numbering differs, but names are mostly consistent.

⸻

🔹 Common “core” syscalls (present on amd64 & arm64)

🗂 File I/O
	•	openat, close, read, write, pread64, pwrite64, readv, writev
	•	statx, newfstatat, fstat, lseek, getdents64, readlinkat
	•	unlinkat, renameat2, symlinkat, linkat, mkdirat, rmdir
	•	chmod, fchmod, fchmodat, chown, fchown, fchownat, lchown
	•	utimensat

🧵 Process / Threads
	•	clone, clone3 (newer)
	•	fork, vfork (rarely used in Go, more in shells)
	•	execve, execveat
	•	wait4, waitid
	•	setpgid, getpgid, setsid, getsid
	•	exit, exit_group

🧠 Memory mgmt
	•	mmap, munmap, mprotect, mremap, brk
	•	madvise

🔔 Signals / Synchronization
	•	rt_sigaction, rt_sigprocmask, rt_sigreturn, sigaltstack
	•	futex, nanosleep, restart_syscall

⏱ Time & randomness
	•	clock_gettime, clock_nanosleep
	•	getrandom

📡 Networking
	•	socket, bind, connect, listen, accept4
	•	getsockopt, setsockopt, recvfrom, sendto, sendmsg, recvmsg
	•	shutdown

📑 Metadata / system info
	•	uname, getpid, getppid, gettid
	•	prlimit64
	•	prctl

🖥 Event loop / polling
	•	epoll_create1, epoll_ctl, epoll_pwait (or epoll_wait)
	•	eventfd2
	•	timerfd_create, timerfd_settime, timerfd_gettime
	•	signalfd4
	•	poll, ppoll, select, pselect6

🔒 Locking
	•	flock, fcntl (and fcntl64)

🗄 Mount / namespace / container stuff
	•	mount, umount2, pivot_root
	•	unshare, setns
	•	sethostname, chroot

⸻

🔹 Syscalls you’ll see on x86_64 but not arm64
	•	socketcall, ipc (obsolete multiplexors)
	•	stat, lstat, fstat64 (replaced by statx/newfstatat)
	•	arch_prctl (x86-specific CPU state setup)
	•	iopl, ioperm (x86 port I/O)

⸻

🔹 Syscalls you’ll see on arm64 but newer on x86
	•	clone3 (added first for modern kernels, also supported on recent x86_64)
	•	openat2 (new, used by glibc & runc checks)
	•	faccessat2

⸻

📎 Resources
	•	Docker’s default seccomp profile — curated allow list, portable across x86_64/arm64.
	•	man 2 syscall + man 2 <name> for details.
	•	ausyscall --dump to see what your kernel/arch supports.

⸻

👉 Do you want me to build you a “portable seccomp baseline profile” (arm64 + amd64) that includes only this common set, so you don’t need to maintain two separate JSONs?