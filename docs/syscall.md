Top — décortiquons cet appel et corrigeons-le au passage.

1) L’API noyau visée

prctl côté Linux a la signature C :

int prctl(int option, unsigned long arg2, unsigned long arg3,
unsigned long arg4, unsigned long arg5);

Pour PR_SET_NAME, seul arg2 est utilisé (pointeur vers un buffer char[16] NUL-terminé). Les autres args doivent être 0.

2) La forme Go correcte

Comme prctl prend 5 paramètres, en Go il faut utiliser Syscall6 (ou RawSyscall6) afin de passer explicitement les zéros restants — c’est le plus sûr.

package main

import (
"log"
"syscall"
"unsafe"
)

const PR_SET_NAME = 15 // linux/prctl.h

func setComm(name string) error {
// Buffer de 16 octets (15 + NUL). Le noyau tronque si >16.
b := make([]byte, 16)
copy(b, name) // le reste est déjà zéro-initialisé

	_, _, errno := syscall.RawSyscall6(
		syscall.SYS_PRCTL,
		uintptr(PR_SET_NAME),
		uintptr(unsafe.Pointer(&b[0])),
		0, 0, 0, 0, // <- args 3..6 à zéro
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func main() {
if err := setComm("banana"); err != nil {
log.Fatal(err)
}
select {} // laissez tourner pour inspecter /proc/$$/comm
}

Variante “non-raw” : remplace RawSyscall6 par Syscall6.
Recommandation moderne : utilisez plutôt golang.org/x/sys/unix → unix.Prctl(unix.PR_SET_NAME, uintptr(unsafe.Pointer(&b[0])), 0, 0, 0).

3) Pourquoi éviter Syscall (3 args) ici

Votre ligne :

syscall.Syscall(syscall.SYS_PRCTL, syscall.PR_SET_NAME, uintptr(unsafe.Pointer(&b[0])), 0)

n’envoie que 3 registres au noyau. prctl en attend 5.
Selon l’archi et l’implémentation du stub, les registres « manquants » peuvent ne pas être garantis à 0 → comportement non fiable. D’où l’usage de Syscall6/RawSyscall6 pour forcer arg3..arg5 à 0.

4) Détails importants à connaître
   •	Limite 16 octets : le noyau stocke le nom dans task_struct->comm (16). Il tronque silencieusement si plus long.
   •	NUL-terminé : fournissez un buffer zéro-rempli (comme ci-dessus) ou ajoutez un 0 final.
   •	Thread vs “process” : PR_SET_NAME renomme le thread courant.
   •	Dans un binaire Go, une goroutine peut migrer entre threads.
   •	Si vous devez renommer le thread principal (ex. avant toute création de threads), faites :

runtime.LockOSThread()
defer runtime.UnlockOSThread()
setComm("banana")


	•	/proc/<pid>/comm reflète le nom du thread principal.

	•	Privilèges : pas besoin d’être root pour PR_SET_NAME.
	•	Retour : succès → r1=0, err==0. Échec → err contient l’errno (ex. EINVAL si option invalide).

5) Résumé
   •	Utilisez Syscall6/RawSyscall6 (ou unix.Prctl) pour prctl.
   •	Passez PR_SET_NAME, un buffer 16 octets zéro-rempli, et zéro pour les autres args.
   •	Verrouillez le thread si vous voulez cibler le thread principal en Go.

-----


Exactement 👌

⸻

🔹 prctl
•	Abréviation de process control.
•	C’est un appel système Linux (pas standard POSIX, donc spécifique Linux).
•	Permet de configurer ou interroger certains attributs de processus ou threads qui n’ont pas leur place ailleurs.

Signature C :

int prctl(int option, unsigned long arg2,
unsigned long arg3, unsigned long arg4,
unsigned long arg5);

	•	option → l’action à effectuer (ex : PR_SET_NAME)
	•	arg2…arg5 → paramètres dont l’usage dépend de l’option

⸻

🔹 Exemple d’options prctl
•	PR_SET_NAME (15) → change le nom du thread (champ comm, max 16 octets)
•	PR_GET_NAME (16) → récupère le nom du thread courant
•	PR_SET_NO_NEW_PRIVS (38) → active le flag no_new_privs (empêche d’obtenir plus de privilèges via execve/setuid)
•	PR_SET_SECCOMP (22) → active un filtre seccomp (sécurité, sandboxing syscalls)
•	PR_CAPBSET_DROP → enlève une capability de la bounding set
•	PR_GET_DUMPABLE / PR_SET_DUMPABLE → contrôle si le process peut être core dumpé

⸻

🔹 Exemple concret en Go

Changer le nom courant :

import (
"golang.org/x/sys/unix"
"unsafe"
)

func main() {
name := []byte("banana\x00") // NUL-terminé
unix.Prctl(unix.PR_SET_NAME, uintptr(unsafe.Pointer(&name[0])), 0, 0, 0)
}

→ Modifie /proc/self/comm → banana
(ps -o comm= -p $$ montrera le nouveau nom).

⸻

👉 Donc :
•	prctl = syscall Linux “process control”.
•	PR_SET_NAME = une option parmi d’autres, qui dit à prctl “change mon nom (comm)”.

Veux-tu que je te fasse un mini-cheatsheet des options prctl les plus utilisées avec leur effet ?

----


Parfait 👌, allons-y pas à pas :

L’appel système Linux prctl a cette signature C :

int prctl(int option,
unsigned long arg2,
unsigned long arg3,
unsigned long arg4,
unsigned long arg5);

	•	option (int) : le code de commande → dit au noyau quelle action faire (PR_SET_NAME, PR_SET_SECCOMP, …).
	•	arg2 … arg5 : 4 registres supplémentaires (tous en unsigned long) → leur signification dépend complètement de option.
	•	Si une option n’utilise pas un argument → il doit être passé à 0.

⸻

🔹 Exemple avec PR_SET_NAME (15)

prctl(PR_SET_NAME, (unsigned long) name);

	•	option = 15 (PR_SET_NAME)
	•	arg2 = pointeur vers un buffer char[16] NUL-terminé
	•	arg3..arg5 = ignorés → passer 0

👉 Résultat : change le champ comm du thread courant.

⸻

🔹 Exemple avec PR_GET_NAME (16)

prctl(PR_GET_NAME, (unsigned long) buf);

	•	arg2 = pointeur vers un buffer de 16 octets où écrire le nom
	•	arg3..arg5 = 0

⸻

🔹 Exemple avec PR_SET_NO_NEW_PRIVS (38)

prctl(PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0);

	•	arg2 = 1 → active le flag no_new_privs
	•	arg3..arg5 = 0

⸻

🔹 Exemple avec PR_SET_SECCOMP (22)

prctl(PR_SET_SECCOMP, SECCOMP_MODE_FILTER, prog);

	•	arg2 = mode (ex. SECCOMP_MODE_FILTER)
	•	arg3 = pointeur vers un programme BPF (filtre des syscalls)
	•	arg4..arg5 = 0

⸻

🔹 Exemple avec PR_CAPBSET_DROP

prctl(PR_CAPBSET_DROP, CAP_SYS_ADMIN);

	•	arg2 = numéro de la capability à retirer
	•	arg3..arg5 = 0

⸻

🧩 Résumé général
•	option = quel contrôle process tu veux (SET_NAME, SECCOMP, CAPBSET_DROP, …).
•	arg2 = paramètre principal (souvent pointeur ou valeur).
•	arg3 = paramètre secondaire (selon l’option).
•	arg4 = rarement utilisé, dépend de l’option.
•	arg5 = idem.

Dans beaucoup de cas, seul arg2 est utilisé → les autres doivent être 0.

⸻

👉 Veux-tu que je te fasse un tableau résumé des options prctl les plus courantes avec ce que signifient arg2..arg5 pour chacune ?