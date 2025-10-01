Bien vu — c’est curieux mais pas impossible. Si ton `exec.CommandContext(ctx, "/proc/self/exe", ...)` **passe** alors que `/proc` est monté `noexec`, il y a quelques explications plausibles. Je te donne les plus probables et la marche à suivre pour trancher (rapide, sans casser ta prod) 🕵️‍♂️

### 4 explications possibles

1. **Exécution par descripteur (FD) — bypass du `noexec`)**
   Le binaire appelé peut **ouvrir** `/proc/self/exe` et appeler `fexecve(fd,…)` ou `execveat(fd,"",...,AT_EMPTY_PATH)`. Dans ce cas on n’utilise *pas* la résolution par chemin et `noexec` n’est pas appliqué.

2. **Remount furtif en `exec` avant l’`execve`**
   Quelque code (ton runtime, un wrapper, ou dockerd) a fait `mount -o remount,exec /proc` avant d’appeler `execve`. Si le remount s’effectue juste avant l’appel, tu peux avoir l’impression que tout “marche” même si `/proc` était `noexec` auparavant.

3. **Le binaire est lancé via son vrai chemin résolu par Go**
   Même si tu écris `"/proc/self/exe"`, la lib ou le process pourrait résoudre le chemin réel (via `os.Executable()` ou lecture du lien) et appeler ensuite `execve("/real/path",…)`. Dans ce cas `noexec` sur `/proc` n’intervient pas.

4. **Comportement kernel/optimisation spécifique**
   Il existe des cas limites où le kernel, si le fichier est déjà mappé/exécuté par le processus, autorise certaines opérations — mais ce n’est pas le comportement général attendu et dépend de détails d’implémentation. C’est la moins probable.

---

### Comment trancher rapidement (3 commandes utiles)

1. **Tracer les appels d’`exec` du processus parent** (strace) — voit si c’est `execve`, `fexecve`, ou `execveat` :

```bash
# remplace <PID_PARENT> par le PID du processus qui lance l'exec
sudo strace -f -e trace=execve,execveat -p <PID_PARENT>
```

Tu verras par exemple :

* `execve("/proc/self/exe", ...) = 0` → exec par chemin
* `execveat(3, "", ..., AT_EMPTY_PATH) = 0`  → exec par FD
* `fexecve(...)` (rare via strace, mais voir execveat is the common case)

2. **Tracer open/openat juste avant exec** (prouve qu’un FD est ouvert) :

```bash
sudo strace -f -e trace=open,openat,execve,execveat -p <PID_PARENT>
```

Si tu vois `open("/proc/self/exe", O_RDONLY) = 3` puis `execveat(3, "", ...)` → c’est fexecve-style.

3. **Vérifier remounts** — surveille les remounts dans dmesg/journal pendant l’opération :

```bash
sudo journalctl -kf | grep -i remount
# ou juste relancer la commande et garder un terminal journalctl -kf ouvert
```

---

### Si tu veux être robuste — options à envisager

* Si tu veux **que ça marche partout** indépendamment des montages, remplace l’`exec` par une **exécution par FD** (Option B) — exemple Go minimal avec `golang.org/x/sys/unix` :

```go
f, _ := os.Open("/proc/self/exe")
defer f.Close()
fd := int(f.Fd())
argv := []string{"/proc/self/exe", "arg1", "arg2"}
env := os.Environ()
unix.Fexecve(fd, argv, env)
```

* Ou mieux : exécuter via le **chemin réel** (`os.Executable()`) pour éviter `/proc` complètement.

---

Si tu veux, je peux :

* te fournir le **snippet Go** complet pour `fexecve` prêt à coller,
* ou t’aider à analyser la sortie de `strace` si tu la colles ici — comme ça on voit exactement quelle piste se vérifie chez toi. 🚀


--- autre

readlink -f /proc/self/exe              # affiche la cible réelle
stat -c '%m' $(readlink -f /proc/self/exe)  # affiche le point de montage (ou utilise findmnt)
# ou, plus complet :
findmnt -T $(readlink -f /proc/self/exe)   # montre options du montage (dont noexec si présent)
