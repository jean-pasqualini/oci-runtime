Parfait — voici une implémentation simple et idiomatique avec github.com/creack/pty pour avoir un terminal interactif (TTY) façon crun --tty :
•	crée un PTY,
•	met ton terminal en raw mode,
•	relaye stdin ↔ PTY ↔ stdout/stderr,
•	gère le resize (SIGWINCH).

Dépendances

go get github.com/creack/pty
go get golang.org/x/term

Code (Linux)

//go:build linux

package app

import (
"context"
"io"
"os"
"os/exec"
"os/signal"
"time"

	"github.com/creack/pty"
	"golang.org/x/term"
	"syscall"
)

// RunAttachPTY lance argv[0] (ex: "/bin/bash") attaché à un PTY interactif.
func RunAttachPTY(ctx context.Context, argv []string) error {
if len(argv) == 0 {
argv = []string{"/bin/bash"} // défaut : shell interactif
}

	// Prépare la commande
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)

	// Démarre la commande avec un PTY (le PTY slave devient son stdio)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = ptmx.Close() }() // fermer le master ferme aussi le slave côté enfant

	// Mettre le terminal local en mode RAW et restaurer à la fin
	var restore func()
	if term.IsTerminal(int(os.Stdin.Fd())) {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err == nil {
			restore = func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }
			defer restore()
		}
	}

	// Synchroniser la taille initiale du PTY avec ton terminal
	_ = pty.InheritSize(os.Stdin, ptmx)

	// Gérer les resize (SIGWINCH)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	defer signal.Stop(ch)
	go func() {
		for range ch {
			_ = pty.InheritSize(os.Stdin, ptmx)
		}
	}()
	// resize initial (encore) au cas où
	_ = pty.InheritSize(os.Stdin, ptmx)

	// Relay I/O bidirectionnel
	// stdin -> pty
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	// pty -> stdout
	copyDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
		close(copyDone)
	}()

	// Attendre fin du process OU annulation du contexte
	waitErr := make(chan error, 1)
	go func() { waitErr <- cmd.Wait() }()

	select {
	case err := <-waitErr:
		// laisser finir la copie de sortie un court instant
		select {
		case <-copyDone:
		case <-time.After(200 * time.Millisecond):
		}
		return err
	case <-ctx.Done():
		// Contexte annulé : tenter un stop propre
		_ = cmd.Process.Signal(syscall.SIGTERM)
		select {
		case err := <-waitErr:
			return err
		case <-time.After(2 * time.Second):
			_ = cmd.Process.Kill()
			return ctx.Err()
		}
	}
}

Utilisation

ctx := context.Background()
if err := RunAttachPTY(ctx, []string{"/bin/bash", "-l"}); err != nil {
// handle err
}

Notes
•	Cette version est parfaite pour prototyper un mode attach.
•	pty.Start(cmd) règle pour toi les bonnes options TTY et connecte le slave au process enfant.
•	On met le terminal hôte en raw pour un vrai comportement interactif (pas de double echo, touches directes).
•	Le resize suit ton terminal (splits, redimensionnements).

Quand tu voudras aller plus loin (namespaces, mount /proc, etc.), tu pourras réutiliser ce squelette et ajouter cmd.SysProcAttr + ta logique avant cmd.Start().