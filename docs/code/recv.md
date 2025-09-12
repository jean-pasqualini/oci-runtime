Très bonne question 👌 — ça touche au fonctionnement bas-niveau de recvfrom (et des sockets en général).

⸻

1. Comment Recvfrom fonctionne en Go (golang.org/x/sys/unix)

Quand tu appelles :

n, _, err := unix.Recvfrom(fd, buf, 0)

	•	fd = descripteur de socket (ici un socket Netlink).
	•	buf = ton slice cible, par ex. 8192 octets.
	•	Recvfrom demande au noyau : “donne-moi le prochain paquet reçu sur ce socket, et copie-le dans mon buffer de taille N”.
	•	La fonction renvoie :
	•	n = nombre réel d’octets copiés dans buf.
	•	éventuellement l’adresse source (ici, kernel).
	•	une erreur si problème.

👉 Tu dois ensuite ne lire que buf[:n].
C’est ça qui t’indique où t’arrêter.

⸻

2. Pourquoi il s’arrête ?

Le socket Netlink (comme UDP) est datagramme-oriented :
•	Chaque envoi (Sendto) côté userland correspond à un “message” côté noyau.
•	Le noyau regroupe la réponse (parfois plusieurs messages Netlink concaténés) et l’émet en un seul datagramme.
•	Recvfrom lit exactement un datagramme complet, jamais plus, jamais moins.

Donc deux cas :
•	Si ton buffer est assez grand → n = taille exacte du datagramme.
•	Si ton buffer est trop petit → n = taille du buffer, mais le reste est troncqué et tu perds la fin du datagramme.

⸻

3. Quand s’arrêter de parser à l’intérieur du buffer ?

Le noyau peut empaqueter plusieurs messages Netlink dans un seul datagramme (ton p := buf[:n]).
Chaque message a un champ h.Len qui dit combien d’octets il occupe (header + payload).
Donc ta boucle :

for len(p) >= unix.NLMSG_HDRLEN {
h := (*unix.NlMsghdr)(unsafe.Pointer(&p[0]))
...
adv := align4(int(h.Len))
p = p[adv:]
}

s’arrête quand :
•	len(p) < NLMSG_HDRLEN (plus assez de place pour un header complet),
•	ou adv dépasse len(p) (on a consommé tout le datagramme).

👉 Autrement dit : Recvfrom te donne une “boîte” avec N octets → tu parses message par message grâce à h.Len jusqu’à épuiser ces N octets.

⸻

4. Résumé
   •	Recvfrom sait s’arrêter car Netlink est basé sur des datagrammes : chaque appel lit exactement un paquet.
   •	La valeur n te dit combien d’octets sont valides dans ton buffer.
   •	À l’intérieur, tu dois toi-même boucler avec h.Len (+ alignement) pour traiter les différents messages Netlink concaténés.
   •	Quand il n’y a plus rien à lire, Recvfrom bloque (sauf si tu as mis le socket en non-bloquant ou avec un timeout).

⸻

👉 Veux-tu que je te montre la différence entre un socket flux (TCP) et un socket datagramme (UDP/Netlink) pour voir pourquoi Netlink est “auto-découpé” en paquets, alors qu’avec TCP tu dois gérer toi-même la notion de fin de message ?