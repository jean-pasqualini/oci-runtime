Awesome—let’s make you a **custom AppArmor profile** that’s basically “docker-default, but allow `mount --make-rprivate` (and friends)”.

Two key facts first:

* Docker’s **`docker-default`** profile is **generated in tmpfs at runtime** and loaded with `apparmor_parser` (not a file you can edit in-place). You provide your **own** profile and select it with `--security-opt apparmor=<name>`. ([Docker Documentation][1])
* AppArmor **can allow just the mount-propagation flags** (`private`, `rprivate`, etc.). The mount rule syntax supports these flags directly (e.g., `options=rprivate`). ([manpages.debian.org][2])

Below is a minimal, production-friendly profile that mirrors the spirit of `docker-default` (denies sensitive `/sys` writes, no raw network, etc.) but **allows only propagation changes** (`private` / `rprivate`) anywhere under `/`. It **does not** allow arbitrary mounts.

---

# 1) Create the profile

Save this as `/etc/apparmor.d/containers/docker-allow-rprivate`:

```
#include <tunables/global>

profile docker-allow-rprivate flags=(attach_disconnected,mediate_deleted) {
  # Base sensible defaults
  # (Same include pattern used in Docker docs examples)
  # https://docs.docker.com/engine/security/apparmor/
  #include <abstractions/base>

  # Networking: allow inet/{tcp,udp,icmp}; deny raw/packet like docker-default style hardening
  network inet tcp,
  network inet udp,
  network inet icmp,
  deny network raw,
  deny network packet,

  # Files: read/write/exec mediation
  file,

  # --- Mount mediation ---
  # Allow changing mount propagation to private (non-recursive and recursive) anywhere.
  # This covers: `mount --make-private <path>` and `mount --make-rprivate <path>`
  # For your case on `/` specifically, these two rules are sufficient; the /** variants
  # allow you to apply it to subtrees too.
  mount options=private  -> /,
  mount options=rprivate -> /,
  mount options=private  -> /**,
  mount options=rprivate -> /**,

  # You can also allow other propagation toggles if needed:
  # mount options=slave   -> /**,
  # mount options=rslave  -> /**,
  # mount options=shared  -> /**,
  # mount options=rshared -> /**,

  # Allow unmounts (helpful in some setups; remove if you don’t need it)
  umount,

  # DO NOT add a blanket `deny mount,` here — deny rules override allow rules.

  # Harden /sys like docker-default examples
  deny /sys/[^f]*/** wklx,
  deny /sys/f[^s]*/** wklx,
  deny /sys/fs/[^c]*/** wklx,
  deny /sys/fs/c[^g]*/** wklx,
  deny /sys/fs/cg[^r]*/** wklx,
  deny /sys/firmware/** rwklx,
  deny /sys/kernel/security/** rwklx,
}
```

Why this works: AppArmor’s mount rules understand **propagation flags** (`private`, `rprivate`, `shared`, `slave`, etc.), so we can allow *only* those. The rule tail `-> <path>` is the mountpoint filter; we allow `/` and any subtree (`/**`). ([manpages.debian.org][2])

---

# 2) Load the profile

```bash
sudo apparmor_parser -r /etc/apparmor.d/containers/docker-allow-rprivate
# verify
sudo aa-status | grep docker-allow-rprivate || true
```

---

# 3) Run your container with this profile

You still need the usual kernel gates open for `mount(2)`:

* `CAP_SYS_ADMIN` (grant via Docker),
* seccomp must allow `mount` (either unconfined or a profile that allows it).

Example test:

```bash
docker run --rm -it \
  --cap-add SYS_ADMIN \
  --security-opt seccomp=unconfined \
  --security-opt apparmor=docker-allow-rprivate \
  alpine sh -c 'mount --make-rprivate / && echo OK'
```

If you want to **log** these events too, keep your seccomp profile/action as `SCMP_ACT_LOG` for `mount(2)`; you’ll then see entries in `/var/log/audit/audit.log` or `journalctl -k`. ([manpages.debian.org][2])

---

## Notes & gotchas

* **Don’t combine** this profile with a blanket `deny mount,` (from some examples): deny overrides allow. Keep only the **specific allows** you need. ([manpages.debian.org][2])
* The example reuses the same “deny `/sys/**` writes” pattern you see in Docker docs’ profile snippets so you keep most of the `docker-default` hardening, while opening *only* propagation toggles. ([Docker Documentation][1])
* Docker’s stock `docker-default` is **not meant to be edited**; Docker generates it and loads it at runtime. Use `--security-opt apparmor=<your-profile>` to pick your custom profile. ([Docker Documentation][1])
* This doesn’t grant capabilities by itself; Docker controls those. Keep `--cap-add SYS_ADMIN` for the actual `mount` syscall, and make sure you’re not in **rootless** mode for host-affecting mounts.
* If you need to allow **only on `/`** and nowhere else, drop the `/**` lines and keep just:

  ```
  mount options=private  -> /,
  mount options=rprivate -> /,
  ```

Want me to tailor a profile for a **specific path** (e.g., `/my/runtime/root`) or to also allow `rslave`/`rshared` safely?

[1]: https://docs.docker.com/engine/security/apparmor/?utm_source=chatgpt.com "AppArmor security profiles for Docker"
[2]: https://manpages.debian.org/testing/apparmor/apparmor.d.5.en.html "apparmor.d(5) — apparmor — Debian testing — Debian Manpages"
