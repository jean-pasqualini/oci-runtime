Oui, si c’est vraiment du point-à-point strict en « requête → réponse » (un seul côté initie, une requête à la fois, pas de push asynchrone), tu peux faire sans boucle de lecture dédiée ni channels :
•	Client : Send(req) puis Recv() juste après, de manière synchrone.
•	Serveur : une boucle simple qui fait Recv(), appelle le handler, puis Send(resp). (La boucle est côté serveur parce que lui doit traiter plusieurs requêtes successives.)

Pas besoin de inFlight ni de goroutines tant que tu ne veux pas :
•	plusieurs requêtes en parallèle,
•	que les deux côtés puissent initier,
•	du streaming ou des notifications push.

Voici un squelette minimal corrigé (champs exportés pour JSON) :

type RpcRequest struct {
ID      string          `json:"id"`
OpName  string          `json:"op_name"`
Payload json.RawMessage `json:"payload"`
}

type RpcResponse struct {
ReplyTo string          `json:"reply_to"`
Payload json.RawMessage `json:"payload,omitempty"`
Error   string          `json:"error,omitempty"`
}

// Côté client : strict req/resp, pas de goroutine
func (r *RpcPipe) Op(ctx context.Context, name string, in any, out any) error {
id := "some-id" // mets un uuid si tu veux
b, err := json.Marshal(in)
if err != nil {
return err
}
req := RpcRequest{ID: id, OpName: name, Payload: b}
if err := r.ipc.Send(req); err != nil {
return err
}

	// bloc jusqu’à la réponse
	var resp RpcResponse
	if err := r.ipc.Recv(&resp); err != nil { // suppose Recv décode dans &resp
		return err
	}
	if resp.ReplyTo != id {
		return fmt.Errorf("mauvais id de corrélation: %s", resp.ReplyTo)
	}
	if resp.Error != "" {
		return errors.New(resp.Error)
	}
	if out != nil {
		return json.Unmarshal(resp.Payload, out)
	}
	return nil
}

Et le serveur (un thread suffit) :

func (r *RpcPipe) Serve() error {
for {
var req RpcRequest
if err := r.ipc.Recv(&req); err != nil {
return err // ou continue selon ta stratégie
}
h := r.handlers[req.OpName]
if h == nil {
_ = r.ipc.Send(RpcResponse{ReplyTo: req.ID, Error: "unknown op"})
continue
}
// Appel handler (synchrone)
val, err := h(context.Background(), req.Payload)
resp := RpcResponse{ReplyTo: req.ID}
if err != nil {
resp.Error = err.Error()
} else if val != nil {
resp.Payload, _ = json.Marshal(val)
}
if err := r.ipc.Send(resp); err != nil {
return err
}
}
}

Quand la boucle/goroutine devient nécessaire
•	Tu veux plusieurs appels simultanés côté client → utilise une map[id]chan et une goroutine de lecture.
•	Les deux côtés peuvent envoyer à tout moment (full-duplex).
•	Tu veux des timeouts/cancel fins → passe un context.Context et gère l’annulation pendant l’attente.

Sinon, pour ton cas « simple point à point », le modèle Send puis Recv (client) + boucle serveur suffit. Assure-toi juste d’exporter les champs (Payload, pas payload) pour JSON.