Dans un conteneur
•	Si tu veux des terminaux interactifs (docker run -it …, runc --console-socket), ton conteneur doit avoir :
•	un /dev/pts monté (-t devpts)
•	un /dev/ptmx qui pointe vers ce devpts (le “multiplexeur”).
•	Sinon, les programmes qui veulent allouer un TTY échouent.

👉 C’est pour ça que dans ton test précédent, tty disait not a tty : /dev/pts n’était pas monté → pas de device node visible, même si tes FDs existaient déjà.

----

```
userland
   |
   | isatty(fd)
   v
libc (glibc/musl)
   |
   | → ioctl(fd, TCGETS, &termios)
   v
kernel
   |
   | file descriptor -> inode -> major/minor
   v
driver associé (ex: tty, pty, mem, null...)
   |
   +-- si le driver implémente termios (tty_ioctl) → OK → isatty = true
   +-- sinon → renvoie ENOTTY → isatty = false
```

------

Bonne remarque 👌 tu mélanges deux morceaux du puzzle : FD déjà ouvert et résolution du nom du TTY.

⸻

🔹 1. Si le FD est déjà ouvert

Ton process a par ex. fd=0 qui pointe vers un PTY slave.
•	Quand il fait isatty(0) → la libc fait un ioctl(TCGETS) directement sur le fd.
•	Le noyau suit :
FD table → struct file → inode → major/minor → driver PTY.
•	Le driver répond “OK, je sais gérer TCGETS” → donc isatty = true.

👉 À ce stade, pas besoin que /dev/pts/0 existe dans le namespace du conteneur.
Le FD est un handle déjà valide → tout marche.

⸻

🔹 2. Pourquoi tty râle alors ?

Parce que la commande tty ne fait pas que isatty.
•	Si isatty(0) dit false → elle affiche not a tty.
•	Si isatty(0) dit true → elle appelle ttyname(0).

Et ttyname() fait quoi ?
•	Il regarde le major/minor derrière le FD.
•	Puis il cherche un chemin dans /dev qui correspond à ce major/minor (par ex. /dev/pts/0).
•	Si le device node n’existe pas dans ton mount namespace (ton conteneur n’a pas monté /dev/pts), la recherche échoue → ttyname renvoie une erreur → la commande tty retombe sur not a tty.

👉 Donc : le FD marche, mais tty ne sait pas le nommer dans /dev → d’où le message trompeur.

⸻

🔹 3. Résumé
•	FD ouvert → parle directement au driver → isatty() OK.
•	Nom symbolique → ttyname() doit trouver un fichier spécial dans /dev qui a le même major/minor.
•	Si ton conteneur n’a pas /dev/pts, la résolution échoue, même si techniquement le FD est un vrai TTY.

⸻

✅ Donc la clé est :
•	isatty se fiche de /dev, il ne regarde que le driver derrière le FD.
•	ttyname dépend de /dev → si le device node n’existe pas dans ton conteneur, ça casse.

⸻

👉 Tu veux que je t’écrive un petit bout de code C/Go qui montre la différence entre isatty(fd) (qui passe) et ttyname(fd) (qui échoue) dans un conteneur sans /dev/pts ?