Yes 👌 ! Bash est un très bon exemple, parce qu’il combine les 3 cas (stdin, stdout, stderr) pour décider s’il doit être interactif ou pas.

⸻

🔹 Comment Bash décide s’il est interactif

Quand bash démarre, il teste :
1.	FD 0 (stdin) :
•	est-ce que isatty(0) est vrai ?
•	si oui, il peut lire directement un utilisateur → bon signe pour mode interactif.
2.	FD 1 (stdout) et parfois FD 2 (stderr) :
•	est-ce que isatty(1) ou isatty(2) est vrai ?
•	s’il n’y a pas de tty en sortie → pas très logique d’être interactif.
3.	Option -i :
•	force l’interactivité, même si stdin n’est pas un tty.
•	ex: bash -i < script.sh → reste interactif.

⸻

🔹 Exemples concrets

1. Terminal interactif

$ bash
bash$ echo $-
himBH

	•	$- contient i → bash est interactif.
	•	Pourquoi ? Parce que isatty(0) et isatty(1) sont vrais.

⸻

2. Redirection d’entrée

$ echo "echo coucou" | bash
coucou

	•	Ici stdin n’est pas un tty (isatty(0) == false), donc bash n’est pas interactif.
	•	$- ne contient pas i.

⸻

3. -i pour forcer

$ echo "echo coucou" | bash -i
coucou
bash$   # tu restes dans un shell interactif

	•	Même si stdin est un pipe, -i force le mode interactif.

⸻

4. Sortie redirigée

$ bash > out.txt

	•	Ici stdout n’est pas un tty (isatty(1) == false).
	•	Bash démarre quand même, mais désactive certaines fonctions interactives (par ex. le job control, prompts spéciaux).

⸻

🔹 Résumé
•	bash regarde surtout isatty(stdin) pour décider “interactif ou pas”.
•	Il regarde aussi stdout/stderr pour activer/désactiver certains comportements (prompt, job control, messages d’erreur colorés).
•	Tu peux forcer avec -i.

⸻

👉 Veux-tu que je t’écrive un mini programme Go qui reproduit exactement la logique de bash : tester isatty(0/1/2) et afficher “interactive shell” ou “non-interactive shell” selon les cas ?