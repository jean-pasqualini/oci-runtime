Yes 🙂 — mais ça dépend du niveau que tu veux :

⸻

1. Standard library (sans dépendances)

Go embarque déjà deux mécanismes RPC simples :
•	net/rpc (Go standard depuis longtemps)
•	Permet de déclarer des méthodes sur un objet (exportées)
•	Gère le request/response automatiquement
•	Fonctionne au-dessus de n’importe quel io.ReadWriteCloser (donc tu peux l’utiliser avec deux pipes, pas seulement TCP).

Exemple rapide :

// service.go
type Arith int

type Args struct{ A, B int }
type Reply struct{ C int }

func (t *Arith) Multiply(args *Args, reply *Reply) error {
reply.C = args.A * args.B
return nil
}

// serveur
server := rpc.NewServer()
_ = server.Register(new(Arith))
server.ServeConn(pipe) // pipe = ton io.ReadWriteCloser

// client
client := rpc.NewClient(pipe)
var reply Reply
err := client.Call("Arith.Multiply", &Args{A: 2, B: 3}, &reply)
fmt.Println(reply.C) // 6

👉 Avantages : déjà inclus, pas besoin de réinventer rpcPipe.
👉 Inconvénients : un peu rigide (pas de JSON par défaut, c’est du gob), mais tu peux écrire un codec JSON custom.

⸻

2. Libs plus modernes (avec JSON)

Si tu veux directement un protocole JSON request/response, il y a quelques options externes :
•	go-JsonRpc2 (implémentation JSON-RPC 2.0 utilisée par VSCode, LSP, etc.)
•	Gestion standardisée des requêtes/réponses, erreurs, notifications.
•	Fonctionne sur pipes (io.ReadWriteCloser).
•	gorilla/rpc (moins actif, mais simple).
•	jrpc2 (très propre, orienté JSON-RPC 2.0).

⸻

3. Cas Unix spécifique

Si tu veux juste des pipes nommés et des appels RPC JSON simples, sourcegraph/jsonrpc2 est un bon compromis :
tu gardes ton transport (os.Pipe, net.Conn, etc.), et la lib s’occupe du protocole requête/réponse.

⸻

✅ Résumé
•	Si tu veux rester 100% stdlib → utilise net/rpc (tu peux brancher dessus tes pipes).
•	Si tu veux un vrai protocole JSON request/response → regarde jsonrpc2 (Sourcegraph ou Creachadair).
•	Si ton besoin est très basique → ton rpcPipe maison suffit (mais il va vite ressembler à JSON-RPC).

⸻

👉 Tu veux que je te montre un exemple minimal avec net/rpc au-dessus d’un io.Pipe() pour que tu voies comment remplacer ton rpcPipe actuel sans lib externe ?