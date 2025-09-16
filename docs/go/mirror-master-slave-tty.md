Short answer: you can’t “tap” a PTY master from two places at once. If you pass the master FD over a Unix socket and also read from it locally, both ends will race and steal bytes from each other.

What to do instead (pick one):
1.	Keep the master and act as a proxy (recommended)
Make your process the single reader/writer of the PTY master, and fan out output to multiple sinks (your stdout + whoever needs the stream). Do not pass the PTY FD; instead, forward bytes over your own Unix stream.

// sketch: one owner of PTY master; fan-out output, fan-in input
m, s, _ := pty.Open()
defer s.Close() // s goes to child’s stdio
// ... start init with s as stdin/stdout/stderr ...

conn, _ := net.DialUnix("unix", nil, &net.UnixAddr{Name: "/tmp/peer.sock", Net: "unix"})
defer conn.Close()

// OUTPUT: PTY -> (stdout, peer)
go func() {
mw := io.MultiWriter(os.Stdout, conn) // send to your console + peer
_, _ = io.Copy(mw, m)
}()

// INPUT from peer back into PTY (if you need interactive input)
go func() { _, _ = io.Copy(m, conn) }()

// also forward your own stdin if needed
go func() { _, _ = io.Copy(m, os.Stdin) }()

	2.	If you must send the PTY master FD
Send a dup of the master (so the receiver gets a valid FD) but don’t read locally. Ask the receiver to tee the stream back to you on a side channel.

// duplicate master and send the duplicate
dupFD, _ := unix.Dup(int(m.Fd()))
defer unix.Close(dupFD)
sendFDOverSocket(consoleSocket, dupFD)

// DO NOT read from m here; instead, receive mirrored bytes:
go func() { _, _ = io.Copy(os.Stdout, mirrorConn) }()

Why dup isn’t enough: duplicated masters refer to the same underlying PTY; reads on either FD consume the same byte stream. So you still need a single reader + explicit mirroring.
3.	Use a recording/mirroring layer
If you only need to display (not interact), the owner of the master can wrap its read with io.TeeReader and send one branch to you:

r := io.TeeReader(master, yourDisplayWriter)
_, _ = io.Copy(remoteWriter, r)

Key takeaways
•	Only one place should read from the PTY master. Mirror/forward bytes to others.
•	unix.Dup lets you send a different FD number, but doesn’t create an independent stream.
•	If you control both ends, build a tiny “console mux” that owns the PTY and fans out output (and optionally fans in input). That’s exactly what runc/containerd shims do.

---------------

Si tu n’as pas la main sur le shim mais tu veux quand même afficher la sortie, fais un pont PTY↔PTY (man-in-the-middle) que toi contrôles :

Idée :
•	ptyA : slaveA branché à init; tu lis masterA (unique lecteur).
•	ptyB : tu envoies masterB au shim (ce qu’il attend).
•	Tu recopies ce que tu lis sur masterA vers os.Stdout et vers slaveB.
→ le shim, en lisant son masterB, voit exactement le flux d’init.
•	(Optionnel) si le shim envoie de l’input, tu recopies masterB → masterA pour le renvoyer à init.

Aucun changement côté shim, et tu gardes l’affichage local.

package main

import (
"io"
"net"
"os"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

func main() {
// 1) PTY pour init
mA, sA, err := pty.Open()
if err != nil { panic(err) }
defer sA.Close()

	// Brancher sA (stdin/out/err) au process init, puis lancer init...
	//   cmd := exec.Command("/sbin/init", ...)
	//   cmd.Stdin, cmd.Stdout, cmd.Stderr = sA, sA, sA
	//   _ = cmd.Start()

	// 2) PTY pour le shim (on lui donnera masterB)
	mB, sB, err := pty.Open()
	if err != nil { panic(err) }
	defer sB.Close()

	// 3) Envoyer masterB au shim via le console socket existant
	ctrl, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: "/run/shim.ctrl", Net: "unix"})
	if err != nil { panic(err) }
	defer ctrl.Close()
	if err := sendFD(ctrl, int(mB.Fd())); err != nil { panic(err) }

	// 4) Pontage des flux
	// 4.a) Sortie init : masterA -> (stdout local) + slaveB (que lira le shim via masterB)
	go func() {
		mw := io.MultiWriter(os.Stdout, sB)
		_, _ = io.Copy(mw, mA)
	}()

	// 4.b) (Optionnel) Entrée depuis le shim : masterB -> masterA (redirigée vers init)
	// Si tu ne veux *pas* d'interaction, commente cette goroutine.
	go func() {
		_, _ = io.Copy(mA, mB)
	}()

	select {} // garde le process vivant
}

func sendFD(c *net.UnixConn, fd int) error {
oob := unix.UnixRights(fd)
_, _, err := c.WriteMsgUnix([]byte{0}, oob, nil)
return err
}

Pourquoi ça marche :
•	Un seul lecteur du flux réel d’init : masterA (chez toi).
•	Tu “tees” en espace utilisateur vers stdout et vers slaveB.
•	Le shim lit masterB et reçoit le même flux, sans que tu aies à modifier le shim.
•	Dupliquer le FD (dup) ou lire à deux sur le même master ne marche pas (perte d’octets) ; le pont PTY→PTY évite ça.