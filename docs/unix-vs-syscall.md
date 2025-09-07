Bonne nuance 👌

🔹 syscall.Mount
•	Package standard syscall (dans la stdlib Go).
•	Direct binding bas-niveau des appels système.
•	⚠️ Deprecated depuis Go 1.4+ : la doc indique clairement que syscall n’évolue plus.
•	Fonctionne encore, mais pas portable et ne reçoit pas de correctifs.

🔹 unix.Mount
•	Dans golang.org/x/sys/unix.
•	C’est la librairie recommandée par l’équipe Go pour tous les appels système Unix/Linux.
•	Maintenue activement (ajout de constantes récentes, support kernels récents, etc.).
•	Même API (presque identique) mais plus complète et à jour.

⸻

Exemple avec unix.Mount

import "golang.org/x/sys/unix"

err := unix.Mount("proc", "/proc", "proc",
uintptr(unix.MS_NOSUID|unix.MS_NODEV|unix.MS_NOEXEC), "")
if err != nil {
panic(err)
}


⸻

⚖️ Comparatif rapide

Aspect	syscall.Mount (stdlib)	unix.Mount (x/sys/unix)
Maintien	❌ obsolète	✅ maintenu activement
Constantes récentes	❌ souvent manquantes	✅ toujours à jour
Recommandation Go	❌ déconseillé	✅ recommandé


⸻

✅ Donc : utilise toujours unix.Mount (dans golang.org/x/sys/unix) pour un code conteneur/runtime moderne.

⸻

Veux-tu que je te montre un diff exact entre syscall.Mount et unix.Mount (leurs signatures, imports, etc.) pour voir que c’est quasi interchangeable ?