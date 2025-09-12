Très bonne question ⚡ — le champ Scope de la struct ifaddrmsg (champ ifa_scope en C, Scope dans unix.IfAddrmsg en Go) indique la portée de validité d’une adresse.

C’est défini par des constantes RT_SCOPE_* dans <linux/rtnetlink.h>.

⸻

🌍 Les valeurs principales de Scope

Constante (unix.*)	Valeur	Signification
RT_SCOPE_UNIVERSE	0	Adresse globale (par défaut). Routable en dehors du host.
RT_SCOPE_SITE	200	Adresse de portée “site” (obsolete, peu utilisé).
RT_SCOPE_LINK	253	Adresse de portée link-local (valide uniquement sur ce lien, ex: fe80::/64 en IPv6).
RT_SCOPE_HOST	254	Adresse de portée host (boucle locale, ex: 127.0.0.1, ::1).
RT_SCOPE_NOWHERE	255	Pas de portée, adresse invalide.


⸻

🛠 Utilisation typique
•	IPv4 / IPv6 globales (192.168.1.10/24, 2001:db8::5/64) →
Scope = RT_SCOPE_UNIVERSE (par défaut).
•	Adresses link-local (fe80::1/64 en IPv6, ou 169.254.0.0/16 en IPv4) →
Scope = RT_SCOPE_LINK.
→ Important : le kernel les ajoute souvent automatiquement, mais si tu veux le faire à la main, il faut mettre LINK.
•	Loopback (127.0.0.1/8, ::1) →
Scope = RT_SCOPE_HOST.
•	Cas particuliers :
•	SITE (200) = adresse scope site (abandonné, pas utilisé aujourd’hui).
•	NOWHERE (255) = utilisé dans certains dumps de netlink pour indiquer “pas de scope”.

⸻

Exemple Go

ifam := unix.IfAddrmsg{
Family:    unix.AF_INET6,
Prefixlen: 64,
Flags:     unix.IFA_F_PERMANENT,
Scope:     unix.RT_SCOPE_LINK, // fe80::/64
Index:     uint32(ifi.Index),
}

Ici → on ajoute une IPv6 link-local sur une interface.

⸻

👉 Dans ton cas (runtime conteneur) :
•	la plupart de tes adresses seront RT_SCOPE_UNIVERSE.
•	tu peux tomber sur RT_SCOPE_LINK si tu manipules des adresses link-local (nécessaires pour le Neighbor Discovery IPv6 et certaines opérations ARP/NDP).
•	RT_SCOPE_HOST tu l’auras surtout sur lo dans chaque netns.

⸻

Veux-tu que je t’écrive un schéma concret : “type d’adresse → Scope à utiliser” (genre tableau IPv4/IPv6 globales, loopback, link-local) ?