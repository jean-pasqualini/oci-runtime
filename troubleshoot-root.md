readelf -l /app/root/bin/bash | grep 'Requesting program interpreter'
ldd /app/root/bin/bash
go doc syscall.SysProcAttr | grep -A5 Chroot

----

Yep — that’s expected. cp -a preserves symlinks, it doesn’t copy their targets.
So if you copy libtinfo.so.6 (a symlink) with cp -a, you’ll get the symlink inside your rootfs; unless you also copy the real file it points to (e.g. libtinfo.so.6.5), the link is broken → /bin/bash can’t load and you see no such file or directory.

How to fix (copy symlink and its target)

Use a tiny loop that:
1.	copies the path exactly as ldd prints it (preserve symlink), and
2.	if that path is a symlink, also copy its resolved target.

ROOT=/app/root

# Ensure base dirs exist
mkdir -p "$ROOT"

# Copy bash itself
mkdir -p "$ROOT/bin"
cp -a /bin/bash "$ROOT/bin/"

# Copy loader + libs from ldd (handles both "=> /path" and bare "/lib/...")
ldd /bin/bash \
| awk '/=> \//{print $3} /^\/lib\//{print $1}' \
| while read -r p; do
# copy the symlink or file as-is
mkdir -p "$ROOT$(dirname "$p")"
cp -a "$p" "$ROOT$p"

# if it's a symlink, also copy its ultimate target
if [ -L "$p" ]; then
t="$(readlink -f "$p")"
mkdir -p "$ROOT$(dirname "$t")"
cp -a "$t" "$ROOT$t"
fi
done

# (On many distros the loader is itself a symlink; ensure real file exists too)
interp="$(readelf -l /bin/bash | awk -F': ' '/interpreter/{print $2}')"
mkdir -p "$ROOT$(dirname "$interp")"
cp -a "$interp" "$ROOT$interp"
if [ -L "$interp" ]; then
it="$(readlink -f "$interp")"
mkdir -p "$ROOT$(dirname "$it")"
cp -a "$it" "$ROOT$it"
fi

Now you should have, for ARM64 for example:

/app/root/bin/bash
/app/root/lib/ld-linux-aarch64.so.1 -> aarch64-linux-gnu/ld-linux-aarch64.so.1
/app/root/lib/aarch64-linux-gnu/ld-linux-aarch64.so.1        # real file
/app/root/lib/aarch64-linux-gnu/libc.so.6  [and its real file if symlink]
/app/root/lib/aarch64-linux-gnu/libtinfo.so.6 -> libtinfo.so.6.5
/app/root/lib/aarch64-linux-gnu/libtinfo.so.6.5              # real file

Alternative approaches
•	Dereference symlinks: cp -L copies the target file instead of the link (you’ll lose the symlink name). That can work, but some loaders expect the SONAME symlink to exist, so I prefer copying both.
•	rsync: rsync -a preserves links; rsync -aL dereferences them. Same trade-off as cp -a vs cp -L.

Don’t forget in your Go code

After Chroot, set the child’s working dir inside the chroot:

cmd.SysProcAttr = &syscall.SysProcAttr{ Chroot: "/app/root" }
cmd.Dir = "/" // make sure cwd exists inside the new root

If it still fails, run this litmus test from the host:

chroot /app/root /lib/ld-linux-aarch64.so.1 /bin/bash -lc 'echo OK'

If that prints OK, your rootfs is now correct.
