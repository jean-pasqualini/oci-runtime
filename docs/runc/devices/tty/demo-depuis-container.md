parfait — voilà un exercice complet, pas-à-pas, pensé pour un shell dans un conteneur lancé avec --privileged. Tu auras tout : préparation de /dev, création d’une paire PTY master/slave, test en deux terminaux, et vérifs isatty/ttyname.

⸻

0) Pré-requis rapides
   •	Tu es dans le conteneur, avec --privileged.
   •	Tu as deux shells vers ce même conteneur (ex. un docker exec -it <id> sh pour ouvrir le 2ᵉ).
   •	Idéalement : bash, mount, python3. (Si pas python3, je te donne une alternative plus bas.)

⸻

1) Préparer /dev dans le conteneur

Monte devpts et assure le lien ptmx → devpts :

# dans le conteneur, shell A
set -e

# /dev/pts (devpts)
mkdir -p /dev/pts
mountpoint -q /dev/pts || mount -t devpts devpts /dev/pts -o mode=620,ptmxmode=666,gid=5

# /dev/ptmx doit pointer vers l'instance locale de devpts
ln -sf /dev/pts/ptmx /dev/ptmx

# devices /dev de base si manquants
[ -e /dev/null ]    || (mknod /dev/null    c 1 3  && chmod 666 /dev/null)
[ -e /dev/zero ]    || (mknod /dev/zero    c 1 5  && chmod 666 /dev/zero)
[ -e /dev/full ]    || (mknod /dev/full    c 1 7  && chmod 666 /dev/full)
[ -e /dev/random ]  || (mknod /dev/random  c 1 8  && chmod 666 /dev/random)
[ -e /dev/urandom ] || (mknod /dev/urandom c 1 9  && chmod 666 /dev/urandom)
[ -e /dev/tty ]     || (mknod /dev/tty     c 5 0  && chmod 666 /dev/tty)

# vérifs
mount | grep devpts || true
ls -l /dev/ptmx /dev/pts | sed -n '1,3p'


⸻

2) Créer le master (ouvre /dev/ptmx) et retrouver le slave

Toujours dans le shell A :

# ouvrir le master (fd 3)
exec 5<> /dev/ptmx

# retrouver le /dev/pts/N associé (unlock + query number)
python3 - <<'PY'
import fcntl, termios, struct, os
TIOCSPTLCK = 0x40045431
TIOCGPTN = 0x80045430 # From asm-generic/ioctls.h
fd = 5
# déverrouiller le slave
fcntl.ioctl(fd, TIOCSPTLCK, struct.pack('i', 0))
# obtenir le numéro N du slave
n = struct.unpack('i', fcntl.ioctl(fd, TIOCGPTN, struct.pack('i', 0)))[0]
print(f"/dev/pts/{n}")
PY

Note le chemin affiché, par ex. /dev/pts/7. C’est le slave.

⸻

3) Tester l’échange master ↔ slave (deux shells)

Shell B (esclave)

# remplace N par ton numéro
cat > /dev/pts/N
# (reste bloqué en lecture; ce que tu tapes ici peut partir vers A si tu lis côté master)

Shell A (maître)
•	Écrire vers l’esclave (ça s’affiche dans B) :

echo "Hello depuis master" >&5

	•	Lire ce que tape B (depuis le master) :

# lis 12 octets sur le master
head -c 12 <&5 | hexdump -C

(ou dd bs=1 count=12 <&3 si hexdump manque)

⸻

4) Vérifier isatty vs ttyname

Dans le shell A (ou B), teste les cas :

# 4.1 stdin sur un TTY (le shell interactif)
python3 - <<'PY'
import sys
print("stdin.isatty() =", sys.stdin.isatty())
PY

# 4.2 sur le device slave concret
python3 - <<'PY'
import os, sys
path = sys.argv[1] if len(sys.argv)>1 else "/dev/tty"
fd = os.open(path, os.O_RDONLY)
print(path, "isatty? ->", os.isatty(fd))
os.close(fd)
PY /dev/pts/N

# 4.3 la commande 'tty' doit maintenant afficher /dev/pts/N
tty || true

Si plus tôt tu n’avais pas monté devpts, tty disait “not a tty”. Maintenant qu’il est monté, tty peut résoudre le nom (ttyname) et affiche /dev/pts/N.

⸻

5) (Optionnel) Tester le job control

Toujours en conteneur, avec un TTY dispo :

sleep 1000 &      # lance en background
jobs
stty tostop       # optionnel : suspendre les jobs bg qui écrivent au TTY

	•	Si tu fais cat > /dev/tty & (bg) et qu’il tente de lire le TTY, il recevra SIGTTIN (arrêté).
	•	Avec stty tostop, un job bg qui écrit vers le TTY recevra SIGTTOU.

⸻

6) (Optionnel) Redimensionnement (SIGWINCH / TIOCSWINSZ)

Depuis A, simule un resize de l’esclave sur le master :

python3 - <<'PY'
import fcntl, struct, sys
fd = 3
rows, cols = 50, 120
winsz = struct.pack("HHHH", rows, cols, 0, 0)
TIOCSWINSZ = 0x5414
fcntl.ioctl(fd, TIOCSWINSZ, winsz)
print("resize envoyé:", rows, "x", cols)
PY

Applis côté slave (vim, less) reçoivent SIGWINCH.

⸻

7) Nettoyage (si tu as tout monté toi-même)

# ferme le master
exec 3>&-

# démonter devpts si tu l'as monté dans cet exercice
umount /dev/pts || true


⸻

⚠️ Alternative si tu n’as pas python3
•	Pour créer une paire test facilement : socat -d -d pty,raw,echo=0 pty,raw,echo=0
•	Il t’affiche deux slaves (/dev/pts/X et /dev/pts/Y).
•	Utilise deux shells : cat > /dev/pts/X d’un côté, echo "..." > /dev/pts/Y de l’autre.

⸻

TL;DR
1.	Monte devpts et pointe /dev/ptmx -> /dev/pts/ptmx.
2.	exec 3<> /dev/ptmx (master), puis via ioctl récupère /dev/pts/N (slave).
3.	Dans un 2ᵉ shell, écris/lis sur /dev/pts/N ; dans le 1ᵉʳ, utilise >&3 et <&3.
4.	tty fonctionne (nommage) une fois devpts monté ; isatty fonctionne via le FD de toute façon.

Si tu veux, je te fournis un script bash unique qui fait les étapes 1→4 et affiche directement le chemin du slave à utiliser dans le second shell.