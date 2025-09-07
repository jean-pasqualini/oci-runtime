Oui — c’est exactement ce qui peut t’arriver avec CLONE_NEWPID.

Pourquoi $$ ≠ /proc/self

Il y a trois “réalités” de PID en jeu :
•	PID vu par ta shell : $$ est le PID dans son namespace.
•	/proc : reflète le namespace PID du procfs monté à /proc.
•	PID hôte : ce que voit l’hôte pour ce même processus.

Si tu crées un nouveau PID namespace mais tu n’as pas remonté /proc après être entré dans ce namespace, alors /proc reste celui de l’hôte →
readlink /proc/self te donnera le PID hôte (ex: 1232), tandis que $$ (dans la shell du nouveau ns) montrera un PID “petit” (souvent 1, 2, 3…).
Certains outils (lib procps) exigent que “je me voie moi-même” dans /proc, sinon ils lèvent des erreurs du style “fatal library error, lookup self”.

Règle d’or
•	Monte (ou remonte) /proc dans le process déjà entré dans le PID namespace.
Autrement dit : NEWNS + NEWPID → (dans l’enfant) mount --make-rprivate / → umount ancien /proc → mount procfs /proc.

Shell (pour vérifier)

# Dans un nouveau mount ns + pid ns
unshare -mp bash

# IMPORTANT : /proc du namespace, pas celui de l'hôte
mount --make-rprivate /
umount -l /proc 2>/dev/null || true
mount -t proc proc /proc -o nosuid,nodev,noexec

# Compare
echo "$$"
readlink /proc/self
# Doivent correspondre (même PID attendu dans ce ns)

# Bonus : voir la table de correspondance des PID à travers les ns
grep NSpid /proc/self/status

Go (dans l’enfant déjà en NEWPID|NEWNS)

// Isoler la propagation
unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, "")

// Remplacer /proc (de l’hôte) par un procfs du ns courant
_ = unix.Unmount("/proc", 0) // pas -l si possible
_ = os.MkdirAll("/proc", 0555)
if err := unix.Mount("proc", "/proc", "proc",
uintptr(unix.MS_NOSUID|unix.MS_NODEV|unix.MS_NOEXEC), ""); err != nil {
panic(err)
}

Check-list quand tu vois “lookup self”
1.	findmnt -no FSTYPE /proc → doit être proc.
2.	readlink /proc/self vs echo $$ dans la même shell (pas celle du parent).
3.	Si tu utilises un “init” (PID 1) qui exec ensuite bash, c’est souvent l’init qui est 1 et bash sera 2, 3, etc.
4.	cat /proc/self/status | grep NSpid montre toutes les correspondances (hôte ↔ ns imbriqués).

Cause fréquente du décalage
•	Tu as monté /proc avant d’entrer dans NEWPID, ou dans le parent au lieu de l’enfant → donc /proc reste lié au namespace PID du parent (l’hôte).
•	Tu lis $$ dans une shell et /proc/self depuis une autre (ex: parent vs enfant).

Aligne ces deux points, et ps redevient heureux 😊.