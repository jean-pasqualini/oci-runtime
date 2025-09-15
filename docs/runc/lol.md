
✅ Excellente question encore une fois, et la réponse est subtile mais essentielle.

⸻

❓ Tu demandes :

Si je fais runc create sans --tty (donc avec "terminal": false),
est-ce que je perds la sortie (stdout/stderr) du conteneur ?

⸻

✅ Réponse courte :

Non, tu ne perds pas la sortie.
🎯 Mais tu dois la récupérer toi-même.

⸻

🧠 Explication détaillée :

Quand tu fais :

runc create --bundle ./bundle mycontainer

Et que dans ton config.json tu as :

"process": {
"terminal": false"
}

Alors :
•	runc crée 3 pipes anonymes :
•	pour stdin
•	pour stdout
•	pour stderr

Mais…

🛑 runc ne redirige pas ces pipes vers ton terminal
🛑 Et il ne garde pas les extrémités côté hôte
🧠 Parce qu’il termine après le create : c’est à toi de t’en occuper

⸻

📋 Donc, que deviennent stdout / stderr ?

Par défaut :
•	Ils existent dans le conteneur
•	Mais aucun processus ne les lit côté hôte
•	Résultat : 🕳️ la sortie part dans le vide (pipe non connecté)

⸻

✅ Tu as 3 options pour ne pas perdre la sortie :

⸻

🔧 Option 1 : Utiliser runc run

runc run --bundle ./bundle mycontainer > out.log 2> err.log

✅ Ici, runc intercepte les pipes et les redirige vers ton shell, tu ne perds rien.

⸻

🔧 Option 2 : Gérer les FDs toi-même après create

Tu fais :

runc create --bundle ./bundle mycontainer

Puis tu écris un programme (en Go, C, Python…) qui :
•	Ouvre /proc/[pid]/fd/1 et /proc/[pid]/fd/2 (stdout/stderr du conteneur)
•	Ou : passe ses propres pipes à runc via l’interface OCI (process.stdin, stdout, stderr via FD passing, mais runc ne le supporte pas facilement sans orchestrateur)

📦 Autrement dit : tu deviens l’orchestrateur.

⸻

🔧 Option 3 : Attacher plus tard avec runc exec

Tu peux utiliser :

runc exec mycontainer some-command

Et là tu peux utiliser --tty ou rediriger la sortie vers ton terminal.

Mais pour le processus principal du conteneur, si tu ne captures pas les pipes lors de create, c’est trop tard.

⸻

🧪 Exemple d’effet :

runc create --bundle ./bundle mycontainer
runc start mycontainer

	•	Tu vois rien du tout dans ton terminal
	•	Parce que les FDs sont connectés à des pipes non lus

⸻

✅ TL;DR

Cas	Sortie perdue ?	Pourquoi ?
runc run sans --tty	❌ Non	FDs connectés à ton shell
runc create sans --tty	⚠️ Oui, si tu ne lis pas les pipes toi-même	FDs non capturés
runc create avec --tty + --console-socket	✅ Tu peux recevoir un PTY et tout rediriger


⸻

Tu veux que je te montre comment écrire un petit programme en Go qui capture les pipes stdout/stderr d’un conteneur runc create ?