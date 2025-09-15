Top 👍 tu peux tester la mécanique master/slave PTY directement depuis ton shell avec quelques commandes simples.

⸻

🔹 1. Vérifier ton montage devpts

mount | grep devpts

Tu devrais voir quelque chose comme :

devpts on /dev/pts type devpts (rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=666)

Et :

ls -l /dev/ptmx
ls -l /dev/pts

	•	/dev/ptmx = le multiplexeur.
	•	/dev/pts/ = où apparaissent les slaves.

⸻

🔹 2. Créer une paire PTY avec openvt ou unshare

Si tu es root, tu peux manipuler /dev/ptmx directement :

# ouvrir un master
exec 3<> /dev/ptmx
# voir quel slave est associé
ls -l /proc/$$/fd/3

Tu verras un lien comme :

/proc/12345/fd/3 -> /dev/pts/5

👉 Ici, fd 3 = master, /dev/pts/5 = slave.

⸻

🔹 3. Jouer avec master/slave

Ouvre deux terminaux et fais :

Terminal A (master côté)

exec 3<> /dev/ptmx
ls -l /proc/$$/fd/3
# disons que c’est /dev/pts/7

Terminal B (slave côté)

cat > /dev/pts/7

(tout ce que tu tapes ici sera lu côté master)

⸻

Retour dans Terminal A

echo "Hello depuis master" >&3

→ Ça apparaîtra dans Terminal B.

Et si tu tapes dans Terminal B (où tourne le cat), ça apparaîtra en lecture sur le FD master de Terminal A :

dd bs=1 count=5 <&3

(va lire 5 octets écrits par l’autre côté)

⸻

🔹 4. Test rapide avec socat

Encore plus simple :

socat -d -d pty,raw,echo=0 pty,raw,echo=0

→ socat crée une paire PTY et te montre les deux slaves /dev/pts/X et /dev/pts/Y.
•	Tout ce que tu tapes dans l’un ressort dans l’autre.

⸻

🔹 Résumé
•	Ouvre /dev/ptmx → tu obtiens un master FD et le noyau crée un slave /dev/pts/N.
•	Lis/écris sur l’un → ça ressort sur l’autre.
•	Tu peux tester ça avec exec + /dev/ptmx ou avec socat pty,pty.

⸻

👉 Tu veux que je t’écrive un petit script bash qui fait tout ça automatiquement et montre les échanges master/slave en direct ?