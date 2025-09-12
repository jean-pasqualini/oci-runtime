Parfait, allons pas à pas 👇

Le code :

var buf bytes.Buffer
if err := binary.Write(&buf, binary.LittleEndian, h); err != nil {
return nil, err
}
return buf.Bytes(), nil


⸻

1. var buf bytes.Buffer
   •	bytes.Buffer est un type standard de Go (bytes package).
   •	C’est un buffer croissant en mémoire, qui implémente io.Writer et io.Reader.
   •	Ici on l’utilise comme un tampon de sortie pour recevoir les octets encodés.

⸻

2. binary.Write(&buf, binary.LittleEndian, h)
   •	binary.Write = fonction du package encoding/binary qui sérialise une valeur en binaire et l’écrit dans un io.Writer.

Arguments :
1.	&buf → la destination (implémente io.Writer).
2.	binary.LittleEndian → choix de l’endianness pour écrire les champs numériques.
3.	h → la valeur à écrire, ici une struct (nlmsghdr, ifinfomsg, etc.).

Que fait binary.Write ?
•	Il parcourt champ par champ la valeur de h.
•	Pour chaque champ (ex: uint16, uint32), il écrit les octets correspondants dans l’ordre demandé (ici little-endian).
•	Il n’ajoute pas de padding Go (contrairement à un cast unsafe.Pointer).
•	Résultat : un flux d’octets qui colle à la struct C, tant que :
•	tu as défini la struct avec les bons types (uint32, uint16…),
•	dans le même ordre que la struct C.

Exemple :

type nlmsghdr struct {
Len   uint32 // 4 octets
Type  uint16 // 2 octets
Flags uint16 // 2 octets
Seq   uint32 // 4 octets
Pid   uint32 // 4 octets
}

→ binary.Write écrit exactement :
Len(4B) | Type(2B) | Flags(2B) | Seq(4B) | Pid(4B) = 16 octets.
Pas de trous, pas de padding.

⸻

3. if err != nil
   •	binary.Write peut renvoyer une erreur si :
   •	h contient un type qu’il ne sait pas encoder (ex: string, slice sans taille fixe).
   •	ou si l’écriture dans le Writer échoue (rare avec bytes.Buffer).
   •	Ici on gère l’erreur normalement.

⸻

4. return buf.Bytes(), nil
   •	buf.Bytes() renvoie le contenu actuel du buffer comme []byte.
   •	Donc on récupère directement l’encodage binaire de la struct.
   •	C’est ce slice qu’on peut ensuite envoyer via Netlink (Sendto).

⸻

Exemple concret

h := nlmsghdr{
Len:   20,
Type:  16,
Flags: 5,
Seq:   1,
Pid:   0,
}

var buf bytes.Buffer
_ = binary.Write(&buf, binary.LittleEndian, h)
fmt.Printf("% x\n", buf.Bytes())

Sortie (exemple) :

14 00 00 00 10 00 05 00 01 00 00 00 00 00 00 00

Décomposition :
•	14 00 00 00 = Len = 20
•	10 00       = Type = 16
•	05 00       = Flags = 5
•	01 00 00 00 = Seq = 1
•	00 00 00 00 = Pid = 0

⸻

Pourquoi c’est pratique
•	Tu n’as pas besoin de gérer les offsets et PutUint* manuellement.
•	Pas de unsafe.
•	Le code reste lisible et proche de la struct C d’origine.

⸻

👉 Tu veux que je te montre la comparaison directe entre :
•	la version binary.Write (comme ci-dessus),
•	et la version unsafe.Pointer (cast struct → []byte),

pour bien voir la différence dans les octets produits quand il y a du padding ?