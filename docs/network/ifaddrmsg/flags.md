Parfait ⚡ — on parle bien des flags de la struct ifaddrmsg :

struct ifaddrmsg {
__u8    ifa_family;
__u8    ifa_prefixlen;
__u8    ifa_flags;   /* <-- ici */
__u8    ifa_scope;
__u32   ifa_index;
};

Dans Go → unix.IfAddrmsg.Flags.
Ces IFA_F_* flags indiquent des propriétés particulières de l’adresse IP.

⸻

🎯 Principaux flags IFA_F_* (définis dans <linux/if_addr.h>)

Flag	Valeur	Signification
IFA_F_SECONDARY / IFA_F_HOMEADDRESS	0x01	Adresse “secondaire” sur une interface (IPv4 alias).
IFA_F_TEMPORARY	0x01 (IPv6)	Adresse temporaire (privacy extension IPv6, RFC 4941).
IFA_F_NODAD	0x02	Ne pas faire de Duplicate Address Detection (IPv6).
IFA_F_OPTIMISTIC	0x04	Adresse IPv6 optimiste (utilisable avant fin DAD).
IFA_F_DADFAILED	0x08	DAD a échoué (IPv6).
IFA_F_HOMEADDRESS	0x10	Adresse home (IPv6 Mobile IP).
IFA_F_DEPRECATED	0x20	Adresse encore configurée mais à éviter (sera supprimée bientôt).
IFA_F_TENTATIVE	0x40	Adresse encore en phase DAD (pas encore confirmée unique).
IFA_F_PERMANENT	0x80	Adresse permanente (non supprimée automatiquement).
IFA_F_MANAGETEMPADDR	0x100	Kernel gère la génération d’adresses temporaires.
IFA_F_NOPREFIXROUTE	0x200	Ne pas installer automatiquement de route de préfixe.
IFA_F_MCAUTOJOIN	0x400	Joindre automatiquement les groupes multicast liés.
IFA_F_STABLE_PRIVACY	0x800	Adresse générée de manière stable (RFC 7217).


⸻

🛠 Usage concret
•	Pour ajouter une IP “normale” : tu laisses Flags = 0.
•	Pour IPv6 privacy (adresse temporaire) → Flags = IFA_F_TEMPORARY.
•	Si tu veux empêcher la route de préfixe auto (cas particulier conteneur/networking) → Flags |= IFA_F_NOPREFIXROUTE.
•	Si tu veux qu’une IPv6 soit permanente (toujours valide) → Flags |= IFA_F_PERMANENT.

⸻

Exemple en Go

ifam := unix.IfAddrmsg{
Family:    unix.AF_INET6,
Prefixlen: 64,
Flags:     unix.IFA_F_NOPREFIXROUTE | unix.IFA_F_PERMANENT,
Scope:     unix.RT_SCOPE_UNIVERSE,
Index:     uint32(ifi.Index),
}

👉 Ici : on ajoute une IPv6 permanente sans route de préfixe automatique.

⸻

💡 Dans un runtime de conteneur, le flag IFA_F_NOPREFIXROUTE est assez courant, parce que tu ne veux pas que le kernel installe des routes implicites dans le netns qui pourraient interférer avec celles que tu configures explicitement.

⸻

Veux-tu que je te fasse un tableau “runtime container” avec :
•	quels flags sont pertinents,
•	lesquels tu peux ignorer,
•	et lesquels éviter absolument ?