Tu parles du troisième argument de close_range(2) 👍

⸻

🔹 Signature

int close_range(unsigned int fd, unsigned int max_fd, unsigned int flags);

	•	fd = premier FD à fermer
	•	max_fd = dernier FD à fermer (souvent UINT_MAX → tous jusqu’au bout)
	•	flags = options pour changer le comportement

⸻

🔹 Flags possibles
•	0
→ fermeture classique (chaque FD du range est fermé).
•	CLOSE_RANGE_UNSHARE (0x1)
→ avant de fermer, le noyau fait un unshare(CLONE_FILES).
•	Sans ça, si ton process partage sa table de FDs avec d’autres (cas des threads créés avec clone(CLONE_FILES)), fermer les FDs les fermerait aussi dans les autres threads/process.
•	Avec UNSHARE, tu obtiens ta propre copie de la table, et tu peux fermer sans affecter les autres.
•	C’est plus sûr, mais un peu plus coûteux (copie de la table).
•	CLOSE_RANGE_CLOEXEC (0x2)
→ au lieu de fermer les FDs, le noyau leur applique FD_CLOEXEC.
•	Résultat : ils restent ouverts tant que tu es dans le process actuel,
•	mais seront automatiquement fermés au prochain execve.
•	C’est utile si tu veux éviter les fuites vers l’enfant, tout en gardant tes FDs dispo localement.

Tu peux combiner les deux : CLOSE_RANGE_UNSHARE | CLOSE_RANGE_CLOEXEC.

⸻

🔹 Exemple concret

Fermer tous les FDs ≥ 3 :

unix.CloseRange(3, ^uint(0), 0)

Marquer tous les FDs ≥ 3 comme CLOEXEC (mais pas les fermer maintenant) :

unix.CloseRange(3, ^uint(0), unix.CLOSE_RANGE_CLOEXEC)


⸻

👉 Donc le troisième argument (flags) contrôle si tu veux un close immédiat, un close seulement à l’exec, et si tu veux d’abord unshare ta table de FDs pour ne pas impacter d’autres threads.

⸻

Tu veux que je te montre ce que fait runc exactement (close ou cloexec) quand il fait son sweep final des FDs ?