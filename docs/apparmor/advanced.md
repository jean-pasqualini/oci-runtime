Yes — ça peut *quand même* marcher avec `/proc` en **noexec** dans au moins ces cas :

1. **Exécution par descripteur (FD), pas par chemin**
   Si ton runtime n’appelle pas `execve("/proc/self/exe", ...)` mais **ouvre d’abord** ce fichier puis lance **`fexecve(fd, ...)`** (ou `execveat(fd, "", ..., AT_EMPTY_PATH)`), le noyau n’applique **pas** le check `noexec` du montage de `/proc` (le contrôle `noexec` s’applique au **chemin**).
   → Résultat : `/proc` peut rester `noexec`, et ça marche quand même.

2. **Remount furtif de `/proc`**
   Un composant (ton binaire ou le runtime) peut faire `mount -o remount,exec /proc` juste avant l’exec. Vu que tu as `CAP_SYS_ADMIN`, c’est autorisé si AppArmor/seccomp ne l’interdisent pas.
   → Selon le timing, tu vois encore `noexec` quand tu vérifies, mais au moment du `exec` c’était `exec`.

3. **Pas vraiment /proc/self/exe**
   Certains chemins de fallback finissent par exécuter le **vrai binaire** via son chemin réel (p. ex. `/app/oci-runtime`) et pas via `/proc/self/exe`.

---

### Comment savoir lequel se passe chez toi (rapide)

Dans le conteneur, trace juste le `exec` du process qui échoue/réussit :

```bash
# 1) Trouve le PID du process cible
pidof oci-runtime   # ou autre

# 2) Trace ses exec
strace -f -e execve,execveat -p <PID>
```

Regarde ce qui apparaît :

* `execve("/proc/self/exe", ...) = -1 EPERM` → bloqué par noexec (classique).
* `fexecve(<fd>, ...)` ou `execveat(<fd>, "", ..., AT_EMPTY_PATH)` → **bypass noexec** (exécution par FD).
* `execve("/app/oci-runtime", ...)` → pas via `/proc`.

Autre test simple :

```bash
# Par chemin (doit échouer si /proc noexec)
execve /proc/self/exe 2>/dev/null || echo "exec by path: FAIL"

# Par FD (devrait marcher même si noexec)
python3 - <<'PY'
import os, sys
fd = os.open("/proc/self/exe", os.O_RDONLY)
os.execveat(fd, "", [b"self"], os.environ, os.AT_EMPTY_PATH)
PY
```

---

### TL;DR

* **noexec** bloque **exec par chemin** (ex. `execve("/proc/self/exe")`).
* **fexecve / execveat(AT_EMPTY_PATH)** exécutent par **FD** → ne subissent pas `noexec` du montage `/proc`.
* Donc oui, tu peux voir `/proc` en *noexec* et que “ça marche” quand même : ton code (ou le runtime) **n’exécute pas par chemin**, mais par **descripteur**.
