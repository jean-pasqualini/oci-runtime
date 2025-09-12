Bonne question 👌 — quand tu envoies un message RTM_NEWADDR ou que tu listes les adresses avec netlink (RTM_GETADDR), tu peux rencontrer tout un tas d’attributs IFA_*.

Ces attributs sont définis dans <linux/if_addr.h>.

⸻

🔹 Les principaux IFA_* types

Attribut (unix.IFA_*)	Usage
IFA_UNSPEC	Non utilisé (valeur par défaut = 0).
IFA_ADDRESS	Adresse de l’interface (IPv6 : c’est l’adresse, IPv4 : peer/broadcast, sauf si c’est du unicast → même que IFA_LOCAL).
IFA_LOCAL	Adresse locale de l’interface (IPv4 : ton IP source).
IFA_LABEL	Nom de l’alias (ex : eth0:1). Stocké comme string.
IFA_BROADCAST	Adresse broadcast IPv4 (ex : 192.168.1.255).
IFA_ANYCAST	Adresse anycast (rare).
IFA_CACHEINFO	Informations de cache (struct ifa_cacheinfo → validité, préférences, timestamps).
IFA_MULTICAST	Adresses multicast associées.
IFA_FLAGS	Flags spécifiques à l’adresse (cf. IFA_F_*, genre IFA_F_PERMANENT, IFA_F_NOPREFIXROUTE).
IFA_RT_PRIORITY	Priorité de routage pour cette adresse (rarement utilisé).
IFA_TARGET_NETNSID	ID de network namespace cible (utile pour opérations cross-netns).


⸻

🔹 Exemple d’utilisation
•	ip addr add 192.168.1.10/24 broadcast 192.168.1.255 dev eth0
→ ajoute IFA_LOCAL=192.168.1.10, IFA_ADDRESS=192.168.1.10, IFA_BROADCAST=192.168.1.255.
•	ip addr add 10.0.0.1 peer 10.0.0.2 dev ppp0
→ ajoute IFA_LOCAL=10.0.0.1, IFA_ADDRESS=10.0.0.2.
•	ip -6 addr add 2001:db8::1/64 dev eth0
→ ajoute uniquement IFA_ADDRESS=2001:db8::1.

⸻

🔹 Côté kernel

Quand tu fais un dump (RTM_GETADDR), le kernel peut te renvoyer plusieurs de ces attributs.
Exemple sur une IPv6 :

inet6 fe80::42:acff:fe11:2/64 scope link
valid_lft forever preferred_lft forever

Netlink retourne :
•	IFA_ADDRESS = fe80::42:acff:fe11:2
•	IFA_FLAGS = IFA_F_PERMANENT
•	IFA_CACHEINFO = {valid_lft=INFINITY, preferred_lft=INFINITY}
•	IFA_SCOPE = RT_SCOPE_LINK

⸻

✅ Résumé

Les IFA_* les plus courants que tu manipuleras dans ton runtime sont :
•	IFA_LOCAL (IPv4)
•	IFA_ADDRESS (IPv4 + IPv6)
•	IFA_BROADCAST (IPv4)
•	IFA_FLAGS (comportement de l’adresse)
•	IFA_CACHEINFO (durée de validité, préféré/déprécié)

Les autres (ANYCAST, MULTICAST, TARGET_NETNSID) sont plus rares et spécifiques.

⸻

👉 Veux-tu que je te fasse un tableau d’exemples “commande ip addr → attributs IFA_ utilisés”* pour que tu visualises mieux la correspondance ?