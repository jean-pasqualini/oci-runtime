Parfait — tu veux un **équivalent complet de runc**, pas un jouet. Voilà un plan d’architecture + une feuille de route très concrète pour recoder un **runtime OCI** compatible avec l’écosystème containerd/Docker, y compris rootless, cgroups v2, seccomp, AppArmor/SELinux, console TTY, checkpoint/restore, etc.

---

# Objectif fonctionnel (parité runc)

* **Compatibilité OCI Runtime Spec** (`create`, `start`, `state`, `kill`, `delete`, `run`, `exec`, `ps`, `list`, `pause`/`resume`, `events`, `update`, `features`).
* **Rootless** (user namespaces, cgroups rootless dégradé).
* **Cgroups v1 & v2** (ressources CPU/mémoire/IO, freezer, pids, device policy).
* **Sécurité** : seccomp-BPF, capabilities, **no\_new\_privs**, AppArmor, SELinux, masked/readonly paths, rlimits, ambient caps.
* **Nominal Linux** : UTS/PID/IPC/MNT/NET/USER namespaces, **pivot\_root**, mount propagation, tmpfs/proc/sys, mounts bind/ro, propagation privée, /dev management.
* **PTY/console** : attach, resize, `--console-socket`, stdio pass-through.
* **State mgmt** : `state.json`, répertoires sous `/run/<runtime>/<id>`, locks.
* **Hooks** : prestart, createRuntime, createContainer, startContainer, poststart, poststop.
* **Checkpoint/Restore** : CRIU (si présent).
* **Intégration** : compatible **containerd-shim** / Docker.

---

# Architecture recommandée (style Onion pragmatique)

```
/cmd/<name>/
  runtime/main.go            # CLI (urfave/cli v2), gestion des commandes et flags

/internal/
  oci/                       # Parsing/validation de la spec OCI + conversions
  runtime/                   # Orchestration haut niveau du cycle de vie
  container/
    init/                    # Code exécuté dans le process init (PID 1 du conteneur)
    proc/                    # Création du process, clone/unshare, sync pipes
    mount/                   # Préparation rootfs, pivot_root, /proc, /sys, /dev, mounts
    cgroup/                  # Appliquer limites (v1/v2), freezer, update
    ns/                      # Namespaces (setns/unshare/clone flags)
    sec/                     # seccomp, caps, no_new_privs, LSM (AppArmor/SELinux)
    idmap/                   # UID/GID map pour rootless, setgroups
    rlimits/                 # ulimit/rlimits
    console/                 # PTY, console-socket, resize
    hooks/                   # Orchestration des hooks OCI
    network/                 # (optionnel) ifaces veth, sysctl net; sinon via hooks
  platform/linux/            # Syscalls/Netlink/Prctl/capset abstraits et testables
  state/                     # state.json RW, layouts, verrous, gc
  events/                    # publication d’événements, métriques
  criu/                      # checkpoint/restore wrappers (si présent)
  rootless/                  # détections et chemins alternatifs (cgroups, mounts)
  log/                       # journaux structurés
  version/                   # ‘features’, versions, build info

/pkg/
  errorsx/                   # erreurs enrichies + causes
  fdx/                       # utilitaires de FD passing (SCM_RIGHTS), pipes
  fileutil/                  # chown_r, ensure_dirs, secure_open
```

**Dépendances vers l’intérieur** : `cmd` → `runtime` (orchestration) → `container/*` (composants) → `platform/linux` (syscalls).
Le **domaine** (concept “Container”, “ProcessSpec”) ne dépend pas des syscalls; l’**infra** Linux est concentrée dans `platform/linux` et les sous-packages spécialisés.

---

# Contrats & flux critiques

## 1) Cycle `create` → `start`

1. **CLI** (`cmd`) charge le bundle (`config.json`), valide la spec (`internal/oci`), réserve un **state dir** :
   `/run/<runtime>/<id>/` avec `state.json`, `init.pid`, `control.sock`, **sync FIFOs**.
2. **runtime.Manager.Create()** orchestre :

    * `mount.PrepareRootfs()` : mounts du spec (bind, tmpfs, ro, propagation privée), `proc`, `sysfs` (+ readonly si demandé), `devpts`/`mqueue` si nécessaire, `masked/readonly` paths.
    * `cgroup.Manager.Apply(pidParent?)` prépare le cgroup **parent** (freezer en `THAWED`).
    * `proc.SpawnInit()` fait le **double fork** / `clone3()` :

        * **Parent** (processus runtime) crée 2 pipes nommés/anon :

            * **INITPIPE** (communication init ↔ parent)
            * **SYNCPIPE** (barrière de synchro étroite)
            * **CONSOLE** (fd passé via UNIX socket si `--console-socket`)
        * **Child stage 1** : `clone3()` avec flags PID|UTS|IPC|NET|MNT|CGROUP|USER selon la spec.
        * **Child stage 2** (dans nouveaux NS) : sethostname, idmap (rootless), `chdir` rootfs, **pivot\_root**, remount `/` ro si demandé, `rlimits`, **prctl(PR\_SET\_NO\_NEW\_PRIVS)**, **capset()** (drop bounding set), **seccomp** (charge BPF), LSM label (AppArmor/SELinux).
        * Child signale **“ready”** via **SYNCPIPE**, puis `execve()` vers `process.args[0]` OU l’**init shim** minimal si `terminal=true`.
    * **hooks.createContainer/prestart** (dans l’ordre OCI) exécutés côté host (avec env/stdio contrôlés).
3. **start** : si `create` a démarré un init en pause (via `pause setns`/`ptrace stop`), `start` envoie **go** via **INITPIPE** (ou CRIU restore). Exécute `hooks.startContainer`.

## 2) `exec`

* Ouvre le **pid namespace** du container (via `/proc/<initpid>/ns/…`), prépare `setns()` dans les mêmes NS, mappe UID/GID si rootless, applique seccomp/caps/rlimits, attache au même cgroup (ou un souscgroup), gère PTY/stdio. Lancement d’un **processus fils** qui n’est pas PID 1.

## 3) `kill`, `delete`, `state`, `ps`, `update`

* `kill` : signale PID 1 (ou un process id) via `signal(2)`.
* `delete` : stoppe si nécessaire, démonte rootfs **en ordre inverse**, libère cgroups, nettoie state dir.
* `update` : réapplique limites cgroups live.
* `ps` : lit `/proc/<initpid>/…` dans le pidns du host (traduction d’IDs).
* `state` : retourne le `state.json` (OCI Runtime State), y compris `pid`, `status`, `bundle`, `annotations`.

---

# Mécanismes bas niveau (détails importants)

* **Subreaper** : le runtime **ne doit pas** devenir reaper des process du conteneur (c’est le rôle de l’init **dans** le conteneur). Utilise `PR_SET_CHILD_SUBREAPER` pour un **shim interne** uniquement si tu dois relayer des exits avant l’intégration containerd-shim ; en prod, le **shim** externe (containerd-shim) fait subreaper.
* **Sync pipes & FDs spéciaux (env libcontainer)** :

    * `_LIBCONTAINER_INITPIPE`, `_LIBCONTAINER_SYNCPIPE`, `_LIBCONTAINER_CONSOLE` (index FD via `ExtraFiles`) : permettent d’annoncer **ready**, de **bloquer tant que parent n’a pas fini** cgroups/mounts externes, et de câbler la console.
* **Ordre sécurisé** :

    1. Unshare/clone namespaces → 2) mounts/pivot → 3) idmap (userns)/setgroups → 4) rlimits → 5) no\_new\_privs → 6) caps drop → 7) seccomp load → 8) LSM label → 9) **execve()**.
       (Capsh/ambient caps à soigner pour conserver ce qu’il faut jusqu’au `execve`.)
* **Rootless** :

    * UserNS + idmap (deny `setgroups` si nécessaire), pas d’accès direct cgroup v1; en v2, cgroup **delegation** possible si parent l’autorise (systèmes modernes). Sinon, **no-op** cgroup avec warning.
* **/dev** :

    * `devtmpfs` n’est pas dispo dans le namespace ; crée via `mknod` limité ou bind-mount depuis `/.` préparé, puis applique **devices cgroup policy** (deny-all + allowlist).
* **Network** :

    * Minimal dans le runtime (config NET NS + loopback up). Les ifaces veth/bridge et IP peuvent être gérés par **hooks** (CNI) ou un sous-module optionnel `network/`.
* **Checkpoint/Restore (CRIU)** :

    * `checkpoint` : freezer cgroup → CRIU dump → pack metadata.
    * `restore` : prépare NS & mounts compatibles → CRIU restore → resynchronise `state.json`.

---

# Interfaces clés (extraits)

```go
// OCI Spec in → structures internes normalisées
type SpecAdapter interface {
    LoadBundle(path string) (*OCISpec, error)
    Validate(*OCISpec) error
}

// Gestion du cycle de vie
type Runtime interface {
    Create(id, bundle string, opts CreateOpts) error
    Start(id string) error
    Exec(id string, p ProcessSpec, io IO) (int, error)
    Kill(id string, sig syscall.Signal) error
    Delete(id string, force bool) error
    State(id string) (State, error)
    Update(id string, r UpdateResources) error
}

// Synchronisation init/parent (FDs passés via ExtraFiles)
type InitSync interface {
    SendReady() error
    WaitGo() error
    SendError(err error) error
}

// Abstractions Linux testables
type LinuxOps interface {
    Clone3(flags CloneFlags, attr *ProcAttr) (pid int, err error)
    Setns(fd int, nstype int) error
    Mount(source, target, fstype string, flags uintptr, data string) error
    PivotRoot(newroot, putold string) error
    Capset(sets CapSets) error
    SeccompLoad(*BPFProgram) error
    Prctl(option int, arg2, arg3, arg4, arg5 uintptr) (int, error)
}
```

---

# Layout du **state** & compatibilité shim

* `/run/<runtime>/<id>/state.json` (OCI Runtime State).
* `init.pid`, `exec.fifo`, `console.sock`, `log.json`.
* **Socket de contrôle** (facultatif) pour `events`, `ps`, `exec attach`.
* Conserver les **sémantiques** attendues par containerd-shim (retours d’erreurs, timings, `exit status`).

---

# Tests & conformité

1. **Unitaires** sur `platform/linux` avec **fakes** de syscalls.
2. **Tests d’intégration** en VM CI (kernel récent + cgroups v2) :

    * `create/run/exec/kill/delete` avec **BusyBox** & Alpine.
    * Montages ro/masked, rlimits, seccomp (bloquer `ptrace`), caps drop (test capability leakage).
    * Rootless : userns + idmap ; vérifier qu’on peut `run` sans privilèges.
3. **Conformance** : exécuter la suite **OCI runtime-tools**.
4. **Stress** :

    * Spawn concurrent (`100 run`), `exec` multiples, `kill -9` race, resize TTY.
    * `pause/resume` via freezer, `update` CPU/IO live.
5. **CRIU** : si détecté, tests gated.

**Matrice CI** :

* Kernels LTS (≥ 5.10, 5.15, 6.1), **cgroups v2 par défaut** + job v1.
* Distros : Debian/Ubuntu, Fedora (SELinux enforcing), openSUSE (AppArmor).
* Rootless vs rootfull.

---

# CLI (schéma de commandes)

```
<name> run [--bundle .] [--detach] [--pid-file ...] <id>
<name> create [--bundle .] <id>
<name> start <id>
<name> exec [--tty --console-socket /path] <id> <cmd> [args...]
<name> kill [--all] <id> <signal>
<name> delete [--force] <id>
<name> ps <id>
<name> state <id>
<name> update <id> --memory ... --cpus ...
<name> events <id>
<name> checkpoint <id> --image-path ...
<name> restore <id> --image-path ...
<name> features
```

---

# Sécurité & pièges réels

* **Ordre seccomp/caps** : charge le filtre **après** avoir configuré les mounts/rlimits, **juste avant** `execve()`, sinon tu te tires une balle dans le pied (le filtre bloquera des syscalls de setup).
* **Ambient capabilities** : ne les oublie pas si tu lances une cible qui re-exec (ex. `bash`) et qui fait `setuid` — tu perds tes caps si l’ambient set n’est pas géré.
* **/proc & /sys** : `nosuid,nodev,noexec` sur `proc`; `sysfs` souvent **ro**; monter `cgroupfs` finement en v1 seulement.
* **pivot\_root** : créer un `putold` **dans** le newroot, puis démonter proprement l’ancien `/`.
* **rootless cgroup** : détecter la délégation; sinon passer en mode “best effort” et documenter les limites (pas de hard limit mémoire, etc.).
* **TTY** : bien gérer `SIGWINCH` et la fermeture propre des FDs (leak FD = conteneur coincé).
* **Hooks** : n’exécute **jamais** les hooks dans le pidns du conteneur; borne les timeouts; contrôle env/path.

---

# Roadmap de dev (suggestion incrémentale)

1. **MVP kernel** : namespaces (sans net), mounts/pivot, exec, `state`, `delete`.
2. **TTY/console** + `exec`.
3. **Cgroups v2** (CPU, memory, pids) + `update`, `pause/resume`.
4. **Seccomp + caps + rlimits + masked/readonly**.
5. **Rootless** (idmap, setgroups lock, mode dégradé cgroups).
6. **Hooks** & `events`.
7. **CRIU** (optionnel).
8. **Conformance** & intégration containerd-shim.

---

# Exemples d’implémentation (snippets ciblés)

## clone + sync (parent)

```go
pid, err := linux.Clone3(linux.CloneFlags{
    NewMount: true, NewPid: true, NewUts: true, NewIpc: true,
    NewNet: wantsNet, NewCgroup: true, NewUser: rootless,
}, &linux.ProcAttr{ /* chroot & fds */ })
// parent: applique cgroups, envoie GO via SYNCPIPE quand prêt
```

## init (enfant, PID 1 conteneur)

```go
// 1) hostname, 2) mount rootfs & pivot_root, 3) idmap rootless, 4) rlimits
// 5) no_new_privs, 6) caps drop, 7) seccomp load, 8) LSM, 9) execve
if err := prctl.NoNewPrivs(); err != nil { /* ... */ }
if err := caps.ApplyBoundingDrop(spec); err != nil { /* ... */ }
if err := seccomp.Load(spec); err != nil { /* ... */ }
return unix.Exec(argv0, argv, envv)
```

---

# Pourquoi cette architecture fonctionne pour un **runc complet**

* Elle sépare clairement :

    * **Orchestration** (runtime)
    * **Contrats OCI** (oci/)
    * **Mécanique Linux** (platform/linux + container/\*)
* Elle permet de **tester** la logique sans le kernel (fakes), et de stresser les chemins réels en CI.
* Elle laisse la porte ouverte au **rootless**, à CRIU, aux LSMs, et aux évolutions (clone3, cgroup v2 only).

---

Si tu veux, je peux te pondre :

* un **squelette de repo** (arbo + `go.mod` + `Makefile` + workflow CI) prêt à compiler,
* et un **plan de tests** minimal (scripts + bundles BusyBox) que tu peux exécuter tout de suite en local.
