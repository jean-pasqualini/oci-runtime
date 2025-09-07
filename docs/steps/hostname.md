When you’re writing your own little init inside a UTS namespace, you can set the hostname directly with the sethostname(2) system call. In Go, the clean way is to use golang.org/x/sys/unix.

⸻

🔹 Steps in your init
1.	Make sure the process was started with CLONE_NEWUTS, so it has its own UTS namespace for hostname/domainname.

cmd.SysProcAttr = &syscall.SysProcAttr{
Cloneflags: syscall.CLONE_NEWUTS,
}


	2.	In your init code, call unix.Sethostname with the desired name:

package main

import (
"fmt"
"golang.org/x/sys/unix"
)

func main() {
// this is your init
if err := unix.Sethostname([]byte("my-container")); err != nil {
panic(err)
}
fmt.Println("hostname set to my-container")
select {} // keep init running
}


	3.	If you spawn a shell in that namespace, hostname or uname -n will show my-container, independent of the host.

⸻

🔹 Low-level alternative

You can also use syscall.Sethostname (in the stdlib), but it’s deprecated in favor of x/sys/unix. The system call behind it is:

int sethostname(const char *name, size_t len);


⸻

🔹 Things to note
•	Needs CAP_SYS_ADMIN or CAP_SYS_NICE in that UTS namespace (normally, if you created the UTS ns yourself, you have it).
•	If you don’t unshare UTS (CLONE_NEWUTS), you’ll just rename the host’s hostname → careful.
•	Maximum length = 64 bytes (constant __NEW_UTS_LEN in kernel headers).

⸻

✅ TL;DR: in your init, just call:

unix.Sethostname([]byte("container1"))

as long as you created the init process inside CLONE_NEWUTS.

⸻

Want me to show you how to extend your init Go handler so it mounts /proc and also sets the hostname, like a mini runc?

----


Great instinct 🙂, the naming is confusing.

⸻

🔹 What “UTS” really is
•	UTS namespace in Linux isolates:
•	the hostname (uname -n, gethostname(2), sethostname(2)),
•	the domainname (uname -d, setdomainname(2)).
•	The name “UTS” comes from the old struct utsname in UNIX (uname(2) syscall), not from “time”.

So:
•	CLONE_NEWUTS → gives the process its own view of system identity (hostname/domain).
•	That’s why containers can each have their own hostname without touching the host.

⸻

🔹 What about time?

Time isolation uses:
•	CLONE_NEWTIME (Linux ≥ 5.6) → new time namespace.
•	Lets you have per-namespace offsets for CLOCK_MONOTONIC and CLOCK_BOOTTIME.
•	That’s unrelated to UTS.

⸻

🔹 Quick check

unshare -u hostname container1   # -u = UTS ns
unshare -T ...                   # -T = time ns (newer kernels)


⸻

✅ So, to change the hostname for your init: you need a UTS namespace, not the time ns.

👉 Want me to give you a quick side-by-side cheatsheet of the main CLONE_NEW* flags (PID, UTS, NET, NS, TIME, …) so you have them all clear?