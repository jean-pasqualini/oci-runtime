Parfait 👍 tu veux savoir si le format transporte ou non les informations de type (au sens “type Go” ou “type schéma”). Voici le tableau enrichi :

⸻

🔎 Tableau comparatif des encodages

Format	Stdlib ?	Interop Langages	Lisible	Compact	Schéma strict	Transporte le type ?	Usage typique
JSON	✔️	✔️ universel	✔️ oui	❌ verbeux	❌ (souple, dynamique)	❌ Non → on doit savoir à l’avance dans quoi unmarshal	API web, debug
Gob	✔️	❌ Go-only	❌ non	✔️ bon	implicite Go	✔️ Oui → envoie une définition de type Go puis les valeurs	IPC Go↔Go, net/rpc
Protobuf	❌	✔️ multi-langage	❌ non	✔️ excellent	✔️ .proto obligatoire	✔️ Oui → via le schéma .proto, pas dans le message lui-même	gRPC, microservices
MessagePack	❌	✔️ multi-langage	❌ non	✔️ bon	❌ (comme JSON)	❌ Non → juste des valeurs structurées, pas d’info de type fort	Réseau rapide JSON-like
CBOR	❌	✔️ multi-langage	❌ non	✔️ bon	optionnel via tags	⚠️ Partiel → peut transporter des “tags” (ex: date, bignum), mais pas un vrai type struct	IoT, WebAuthn, standards IETF
FlatBuffers	❌	✔️ multi-langage	❌ non	✔️ excellent	✔️ .fbs obligatoire	✔️ Oui → défini par le schéma, structure figée	Jeux, perf critique
Cap’n Proto	❌	✔️ multi-langage	❌ non	✔️ excellent	✔️ .capnp obligatoire	✔️ Oui → défini par le schéma, sérialisation zero-copy	Perf extrême, systèmes embarqués


⸻

✅ Résumé clair
•	JSON / MessagePack → ne transportent pas le type → tu dois savoir dans quoi déserialiser (flexibles mais faibles sur la sûreté).
•	Gob → transporte le type Go (auto-descriptif, mais Go-only).
•	Protobuf / FlatBuffers / Cap’n Proto → n’embarquent pas le type dans chaque message, mais reposent sur un schéma externe (le type est connu grâce au .proto / .fbs / .capnp).
•	CBOR → un entre-deux : format binaire JSON, avec la possibilité d’annoter des valeurs (tags), mais pas de définition de struct complexe comme gob.

⸻

👉 Veux-tu que je te fasse un exemple avec gob où tu encodes deux structs différents et on voit que gob envoie la définition de type la première fois (preuve que le type est bien transporté) ?