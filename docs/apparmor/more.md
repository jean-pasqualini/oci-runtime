Good question — that `--security-opt apparmor=xxx` flag in Docker trips up a lot of folks 👀

It **only takes a profile name**, not a direct path to a file.

Here’s the deal:

* AppArmor works by loading profiles into the kernel ahead of time (usually under `/etc/apparmor.d/`).
* Docker can’t just consume a raw profile file; it has to reference a **profile that the kernel already knows about**.
* So when you run:

  ```bash
  docker run --security-opt apparmor=my-profile ...
  ```

  Docker tells the kernel “enforce the `my-profile` AppArmor profile on this container,” and the kernel looks up that profile by name.

If you’ve written a custom profile, you’ll need to:

1. Put it in `/etc/apparmor.d/` (or wherever your distro expects).
2. Load it with `apparmor_parser -r /etc/apparmor.d/my-profile`.
3. Then you can launch your container with `--security-opt apparmor=my-profile`.

So yeah, no direct file path magic — just profile names that are already registered.

👉 Do you want me to walk you through creating a minimal custom AppArmor profile for a container, so you can test it quickly?

----

Ah, yep — you spotted a key difference between how Docker handles **AppArmor** vs **seccomp** 🎯

Here’s why:

* **AppArmor** is a **Linux Security Module (LSM)**. Profiles live in the kernel’s AppArmor subsystem. You don’t feed it a file at container start — you load the profile into the kernel ahead of time, and then refer to it *by name*. That’s how AppArmor is designed.

* **seccomp**, on the other hand, is **not an LSM**. It’s a syscall filter mechanism built into the kernel, and Docker itself is responsible for installing the filter rules when a container starts.

    * Because of that, Docker can just take a JSON file at runtime (via `--security-opt seccomp=/path/to/profile.json`) and apply it directly. No kernel “profile registry” needed.

👉 So:

* AppArmor = **kernel-managed profiles**, referenced by name.
* seccomp = **user-space JSON policy files**, parsed/applied by Docker at container start.

That’s why the interface feels asymmetric.

Would you like me to give you a little side-by-side example of “hello world” style AppArmor and seccomp profiles so you can see the contrast?


----

En court-circuit 😅 : **le flag `noexec` du montage `/proc` gagne toujours**.
Même si ton profil AppArmor “autorise l’exec”, **le noyau refuse l’exécution depuis un système de fichiers monté `noexec`**. AppArmor (LSM) ne peut *jamais* “donner” un droit qu’un niveau plus bas (VFS/mount options) a retiré.

### Ordre (simplifié) des contrôles à l’`exec`

1. Bits Unix + options de montage (ex. `noexec`)
2. LSM/AppArmor (règles `ix/px/ux`, chemins, etc.)
3. Seccomp (filtrage de `execve/execveat`), etc.

Donc si `/proc` est `noexec`, l’`exec` de `/proc/self/exe` fait **EPERM** avant même qu’AppArmor ne puisse aider. C’est pour ça que tu ne vois pas de `DENIED` AppArmor.

### Comment faire si tu DOIS lancer via `/proc/self/exe`

* **Remonter `/proc` avec `exec`** (nécessite `CAP_SYS_ADMIN` et que seccomp + AppArmor l’autorisent) :

  ```bash
  # dans le conteneur
  mount -o remount,exec /proc
  ```

  Côté profil AppArmor du conteneur, il faut des règles qui laissent passer le remount (selon ta version d’AppArmor) et l’exec via /proc :

  ```apparmor
  mount,
  umount,
  # si supporté par ta version :
  # mount fstype=proc options=(remount,exec) -> /proc,

  /proc/self/exe mr,
  /proc/self/exe ix,
  /proc/[0-9]*/exe mr,
  /proc/[0-9]*/exe ix,
  ```

  Et côté seccomp, autoriser `mount`, `umount2`, `execve`, `execveat`.

### Alternatives (souvent mieux)

* **Ne pas utiliser `/proc/self/exe`** : exécute ton binaire via son **vrai chemin** (`/app/oci-runtime`) → pas d’impact du `noexec` de `/proc`.
* **Copier** le binaire vers un FS **exécutable** (ex. `tmpfs` : `/tmp`, `/dev/shm`) et l’exécuter de là.
* **`--security-opt apparmor=unconfined`** : ça *peut* marcher car ton process peut alors remonter `/proc` en `exec`. Mais c’est large et pas idéal en prod.

👉 Moralité : **AppArmor peut restreindre plus**, mais **ne peut pas assouplir** un `noexec`. Si tu veux absolument l’exec via `/proc`, il faut lever `noexec` (remount) *et* laisser passer côté AppArmor/seccomp.
