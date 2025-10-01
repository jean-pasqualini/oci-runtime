Parfait — voici exactement comment lancer `strace` sur **le binaire Go que tu démarres**, en suivant aussi tous les enfants (fork/clone), et comment repérer vite si l’exec passe par chemin (`execve("/proc/self/exe"...)`) ou par FD (`execveat(..., AT_EMPTY_PATH)` / `execveat(3, "", ..., AT_EMPTY_PATH)`).

Je te donne 1) commande générique à lancer **dans le conteneur**, 2) variante si tu veux tracer `go run ...`, 3) comment lire les logs.

> Remarque : `strace` a besoin de privilèges pour attacher/trace (sudo) et change un peu le timing — c’est normal.

---

### 1) Tracer un binaire déjà compilé (suivi des enfants, logs séparés)

Exécute dans ton conteneur (ou sur l’hôte si tu lances le binaire depuis l’hôte) :

```bash
# lance le binaire sous strace, suit enfants (-f), crée un fichier par processus (-ff),
# capture chaînes longues (-s 200), ajoute timestamps (-ttt) et trace les appels utiles
sudo strace -ff -f -o /tmp/strace.%p.log -s 200 -ttt \
  -e trace=execve,execveat,open,openat,mount,umount2,clone,fork,vfork \
  /app/oci-runtime --root /tmp/state run --bundle /app/bundle cid
```

* `/tmp/strace.%p.log` → fichiers `/tmp/strace.<pid>.log` (un pour chaque processus).
* `trace=` : on suit `execve/execveat` (pour voir exécution par chemin vs par FD), `open/openat` (pour voir si `/proc/self/exe` est ouvert), `mount/umount2` (remounts), et `clone/fork` (suivi enfants).

---

### 2) Tracer `go run` directement

Si tu veux tracer exactement ce que fait `make run` (qui appelle `go run ...`), tu peux lancer :

```bash
sudo strace -ff -f -o /tmp/strace.%p.log -s 200 -ttt \
  -e trace=execve,execveat,open,openat,mount,umount2,clone,fork,vfork \
  go run -tags medium ./cmd/oci-runtime --root /tmp/state run --bundle /app/bundle cid
```

(attention : `go run` compile puis exécute un binaire temporaire — strace tracera le processus d’exécution résultant aussi).

---

### 3) Dans ton cas “make run”

Si tu veux garder `make run`, lance `make` sous `strace` :

```bash
sudo strace -ff -f -o /tmp/strace.%p.log -s 200 -ttt \
  -e trace=execve,execveat,open,openat,mount,umount2,clone,fork,vfork \
  make run
```

---

### 4) Rechercher rapidement les événements intéressants

Après exécution, regarde les logs :

* Voir si `/proc/self/exe` est **ouvert** :

```bash
grep '/proc/self/exe' /tmp/strace.*.log
```

* Trouver les `execve` par chemin :

```bash
grep 'execve("' /tmp/strace.*.log | head
# ex. execve("/proc/self/exe", ["..."], 0x...) = -1 EPERM
```

* Trouver les `execveat`/exec par FD (bypass noexec):

```bash
grep 'execveat(' /tmp/strace.*.log || grep 'AT_EMPTY_PATH' /tmp/strace.*.log
```

* Chercher remounts de /proc :

```bash
grep -i 'remount' /tmp/strace.*.log
grep -i 'mount.*proc' /tmp/strace.*.log
```

* Regarder les erreurs EPERM pour exec :

```bash
grep -E 'execve.*EPERM|execveat.*EPERM' -n /tmp/strace.*.log
```

---

### 5) Conseils pratiques

* Si tu veux surveiller **en temps réel** pendant que tu fais la manip :

  ```bash
  tail -F /tmp/strace.*.log
  ```
* Pour ne tracer **que** exec/open/mount (moins verbeux), adapte `-e trace=...`.
* Si `strace` crée beaucoup de fichiers, nettoie après : `rm /tmp/strace.*.log`.
* Si tu es dans Docker et n’as pas `strace` dans l’image, installe-le (ex: `apt update && apt install -y strace`) ou exécute le binaire sur l’hôte en pointant vers le binaire du conteneur (moins pratique).

---

Donne-moi la ou les lignes clés que tu trouves (par ex. `open("/proc/self/exe", O_RDONLY) = 3` suivi de `execveat(3, "", ... AT_EMPTY_PATH)` ou `execve("/proc/self/exe", ...) = -1 EPERM`) — je t’expliquerai exactement ce que ça signifie et la modification minimale à appliquer.
