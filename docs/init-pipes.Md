Très bonne question 👌 — et la réponse tient au fait que les canaux ne servent pas aux mêmes interlocuteurs.

⸻

🔹 INITPIPE & SYNCPIP = communication parent/enfant immédiats
•	Qui parle ?
runc create (parent) ↔ runc init (enfant qu’il vient de fork/exec).
•	Durée de vie ?
Seulement pendant l’appel runc create.
•	Comment faire ?
➝ Un pipe/sockpair anonyme (ExtraFiles) suffit, car le parent et l’enfant partagent déjà le même espace de FDs.
•	Pas besoin de nom dans le FS :
•	personne d’autre n’a besoin d’y accéder,
•	c’est un one-shot,
•	c’est détruit automatiquement quand les deux extrémités se ferment.

⸻

🔹 FIFOFD = communication entre deux processus distincts et non reliés
•	Qui parle ?
runc init (qui reste bloqué) ↔ plus tard un nouveau processus runc start (pas l’ancien parent, qui est mort).
•	Durée de vie ?
Entre runc create et runc start. Peut durer plusieurs secondes/minutes/heures.
•	Pourquoi une FIFO nommée ?
•	Parce que ces deux processus n’ont pas de relation parent/enfant → impossible de se passer un pipe anonyme via ExtraFiles.
•	Ils ont besoin d’un point de rendez-vous persistant dans le FS (/run/runc/<id>/exec.fifo).
•	C’est ce “fichier spécial” qui permet au runc start de retrouver l’init et de le débloquer.

⸻

🔹 En résumé
•	INITPIPE & SYNCPIPE : éphémères, privés, internes au create. Un simple pipe anonyme suffit.
•	FIFOFD : doit survivre au create et être réutilisable par un autre runc plus tard (start). Donc → nommé dans le FS.

⸻

👉 Tu veux que je te fasse un schéma avec 3 temps (create → pause → start) pour montrer quels pipes existent encore et pourquoi FIFO doit rester dans le FS ?