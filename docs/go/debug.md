Tu as plusieurs façons de “débugger” un programme Go lancé en ligne de commande. L’outil standard est Delve (dlv), le débogueur officiel Go.

⸻

1. Installer Delve

go install github.com/go-delve/delve/cmd/dlv@latest

Le binaire sera dans $(go env GOPATH)/bin/dlv (pense à l’ajouter dans ton $PATH).

⸻

2. Lancer ton binaire en debug

a) Lancer le programme directement avec dlv

dlv debug ./cmd/oci-runtime -- arg1 arg2 ...

	•	debug = compile ton code en mode debug et l’exécute.
	•	-- sépare les options de dlv de celles de ton programme (arg1 arg2).

⸻

b) Attacher Delve à un binaire déjà compilé

go build -gcflags="all=-N -l" -o mybin ./cmd/oci-runtime
dlv exec ./mybin -- arg1 arg2

	•	-gcflags="all=-N -l" = désactive les optimisations (indispensable pour debug).
	•	dlv exec = exécute un binaire déjà construit.

⸻

c) Attacher à un process déjà en cours

Si tu lances ton binaire à la main (hors dlv) :

ps aux | grep oci-runtime   # récupérer le PID
dlv attach <PID>


⸻

3. Commandes utiles dans dlv

Une fois dans la console dlv :
•	break main.main → pose un breakpoint
•	break file.go:42 → breakpoint sur une ligne
•	continue (ou c) → continue l’exécution
•	next (ou n) → exécute la ligne suivante (step over)
•	step (ou s) → entre dans la fonction appelée
•	print varName (ou p) → affiche une variable
•	locals → montre les variables locales
•	bt → backtrace (pile d’appels)
•	quit → sortir du débogueur

⸻

4. Exemple concret avec ton cas

Si ton binaire est oci-runtime et tu veux débugger run-shim :

cd oci-runtime
dlv debug ./cmd/oci-runtime -- run-shim --config=config.json


⸻

👉 Question pour être plus précis :
Veux-tu déboguer dans un terminal interactif avec dlv, ou intégrer le debug dans un IDE/éditeur (GoLand, VS Code, Vim, etc.) ?