Yes 👍 tu as tout à fait raison : avec github.com/creack/pty tu as un wrapper bien pratique autour de openpty(3) et ioctl(TIOCSCTTY), ce qui rend le code très lisible.

⸻

🔹 Et runc, il fait quoi ?

runc n’utilise pas creack/pty.
Il utilise directement les appels système exposés par golang.org/x/sys/unix et ses propres helpers dans libcontainer/console.

En gros :
•	Le code qui gère la console est dans libcontainer/console.
•	Il appelle directement :
•	unix.Openpt()
•	unix.Grantpt()
•	unix.Unlockpt()
•	unix.IoctlSetInt(fd, unix.TIOCSCTTY, 0)
•	gestion de setsid() / dup des FDs → stdin/stdout/stderr
•	Puis il envoie le master FD via le socket --console-socket avec sendmsg + SCM_RIGHTS.

Donc là où creack/pty encapsule tout ça dans un pty.Open(), runc réimplémente en natif pour avoir plus de contrôle et éviter une dépendance externe.

⸻

🔹 Exemple côté runc

Dans libcontainer/console/console_linux.go (simplifié) :

func New() (*Console, error) {
fd, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY, 0)
// -> grantpt, unlockpt
// -> ptsname pour trouver le slave
// -> open du slave
}

Puis dans libcontainer/console_linux.go :

func (c *linuxConsole) setup() error {
// setsid()
// ioctl(TIOCSCTTY)
// dup2(slave → 0,1,2)
}

Et côté envoi FD (runc --console-socket) :

// unix.UnixRights(fd) + unix.Sendmsg pour transmettre le master


⸻

🔹 Donc en résumé
•	creack/pty : pratique, tout-en-un (open + dup + ioctl).
•	runc : utilise x/sys/unix + code maison dans libcontainer/console, pour maîtriser les détails et éviter dépendances.

⸻

👉 Tu veux que je t’écrive un mini-exemple en Go qui reproduit le flux exact de runc (openpt + grantpt + unlockpt + ioctl + sendmsg) mais sans creack/pty ? Ça peut te servir si tu veux coller au comportement interne de runc.