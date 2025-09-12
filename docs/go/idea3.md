Top ! La solution propre (sans génériques) est de stocker un handler “générique” avec la signature uniforme :

type RpcHandler func(ctx context.Context, payload any) (any, error)

…et de fournir un adaptateur qui accepte tes fonctions fortement typées (func(ctx, FirstReq) FirstRes, etc.), les valide avec reflect et les wrap en RpcHandler.
Ainsi, tu écris tes handlers “beaux et typés”, et l’infra les appelle via une passerelle unique.

⸻

Ce qu’on vise

Tu écris ça :

func First(ctx context.Context, req FirstReq) FirstRes { /* ... */ }
func Second(ctx context.Context, req *SecondReq) (*SecondRes, error) { /* ... */ }

// enregistrement
_ = pipe.RegisterFunc("First", First)
_ = pipe.RegisterFunc("Second", Second)

Et ton infra garde un map[string]RpcHandler derrière.

⸻

Implémentation

1) Types & registre

type RpcHandler func(ctx context.Context, payload any) (any, error)

type rpcPipe struct {
handlers map[string]RpcHandler
ipc      app.Ipc
}

func newRpcPipe(ipc app.Ipc) *rpcPipe {
return &rpcPipe{
ipc:      ipc,
handlers: make(map[string]RpcHandler),
}
}

2) Adapter une fonction typée → RpcHandler

On accepte des variations pratiques :
•	arg req par valeur ou par pointeur
•	retour res par valeur ou par pointeur
•	avec ou sans error en 2ᵉ valeur de retour

func (r *rpcPipe) RegisterFunc(name string, fn any) error {
h, err := makeAdapter(fn)
if err != nil {
return err
}
r.handlers[name] = h
return nil
}

func makeAdapter(fn any) (RpcHandler, error) {
v := reflect.ValueOf(fn)
t := v.Type()
// Doit être une fonction
if t.Kind() != reflect.Func {
return nil, fmt.Errorf("handler must be a function")
}
// Signature attendue: func(context.Context, Req) (Res[, error])
if t.NumIn() != 2 {
return nil, fmt.Errorf("handler must have 2 params (context.Context, Req)")
}
if !t.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
return nil, fmt.Errorf("first param must be context.Context")
}
if t.NumOut() < 1 || t.NumOut() > 2 {
return nil, fmt.Errorf("handler must return (Res) or (Res, error)")
}
var withErr bool
if t.NumOut() == 2 {
if !t.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
return nil, fmt.Errorf("second return must be error")
}
withErr = true
}

    reqType := t.In(1)  // T or *T
    resType := t.Out(0) // U or *U

    return func(ctx context.Context, payload any) (any, error) {
        pv := reflect.ValueOf(payload)

        // Adapter payload au type attendu par le handler (T vs *T)
        // - si handler attend T et on a *T → deref
        // - si handler attend *T et on a T  → prendre l'adresse
        if !pv.IsValid() {
            return nil, fmt.Errorf("nil payload")
        }
        if reqType.Kind() == reflect.Ptr && pv.Kind() != reflect.Ptr {
            // T -> *T
            if pv.CanAddr() {
                pv = pv.Addr()
            } else {
                // recrée une *T et copie
                addr := reflect.New(pv.Type())
                addr.Elem().Set(pv)
                pv = addr
            }
        } else if reqType.Kind() != reflect.Ptr && pv.Kind() == reflect.Ptr {
            // *T -> T
            pv = pv.Elem()
        }

        if !pv.Type().AssignableTo(reqType) {
            return nil, fmt.Errorf("payload type %v not assignable to %v", pv.Type(), reqType)
        }

        out := v.Call([]reflect.Value{reflect.ValueOf(ctx), pv})

        var res reflect.Value
        var err error

        if withErr {
            if !out[1].IsNil() {
                err = out[1].Interface().(error)
            }
            res = out[0]
        } else {
            res = out[0]
        }

        if err != nil {
            return nil, err
        }
        return res.Interface(), nil
    }, nil
}

3) Transport en gob (pour avoir le type concret dans Payload)

type RPCRequest struct {
ID      string
OpName  string
Payload any // le type concret voyage via gob
}
type RPCResponse struct {
ReplyTo string
Payload any
}

Transport :

func (s *SyncPipe) Send(v any) error { return gob.NewEncoder(s.w).Encode(v) }
func (s *SyncPipe) Recv(v any) error { return gob.NewDecoder(s.r).Decode(v) }

Pense à enregistrer tous les types concrets utilisés :

func init() {
gob.Register(&FirstReq{})
gob.Register(&FirstRes{})
gob.Register(&SecondReq{})
gob.Register(&SecondRes{})
}

Astuce : envoie toujours des pointeurs (&FirstReq{} / &FirstRes{}) pour simplifier.

4) Boucle serveur

func (r *rpcPipe) HandleOnce(ctx context.Context) error {
var req RPCRequest
if err := r.ipc.Recv(&req); err != nil {
return err
}
h := r.handlers[req.OpName]
if h == nil {
return xerr.Op("no handler", nil, xerr.KV{"op": req.OpName})
}
out, err := h(ctx, req.Payload)
if err != nil {
return err
}
return r.ipc.Send(RPCResponse{
ReplyTo: req.ID,
Payload: out,
})
}

5) Côté client

func (r *rpcPipe) Op(ctx context.Context, name string, payload any, dest any) error {
id := fastID() // peu importe
if err := r.ipc.Send(RPCRequest{ID: id, OpName: name, Payload: payload}); err != nil {
return err
}
var resp RPCResponse
if err := r.ipc.Recv(&resp); err != nil {
return err
}
// resp.Payload est déjà du bon type concret (merci gob)
// si dest != nil, on copie dedans pour rester explicites côté appelant
if dest != nil {
rv := reflect.ValueOf(dest)
if rv.Kind() != reflect.Ptr || rv.IsNil() {
return fmt.Errorf("dest must be non-nil pointer")
}
val := reflect.ValueOf(resp.Payload)
// gérer U vs *U
if rv.Elem().Type().AssignableTo(val.Type()) {
rv.Elem().Set(val)
} else if val.Kind() == reflect.Ptr && val.Elem().Type().AssignableTo(rv.Elem().Type()) {
rv.Elem().Set(val.Elem())
} else {
return fmt.Errorf("cannot assign %v to %v", val.Type(), rv.Elem().Type())
}
}
return nil
}


⸻

Exemple d’usage

type FirstReq struct{ A, B int }
type FirstRes struct{ Sum int }

func First(ctx context.Context, req FirstReq) FirstRes {
return FirstRes{Sum: req.A + req.B}
}

type SecondReq struct{ S string }
type SecondRes struct{ Len int }

func Second(ctx context.Context, req *SecondReq) (*SecondRes, error) {
if req == nil { return nil, fmt.Errorf("nil req") }
return &SecondRes{Len: len(req.S)}, nil
}

// init gob + registre
func init() {
gob.Register(&FirstReq{})
gob.Register(&FirstRes{})
gob.Register(&SecondReq{})
gob.Register(&SecondRes{})
}

Enregistrement :

_ = pipe.RegisterFunc("First", First)
_ = pipe.RegisterFunc("Second", Second)

Appel côté client :

var out1 FirstRes
_ = pipe.Op(ctx, "First", &FirstReq{A: 2, B: 3}, &out1)

var out2 *SecondRes
_ = pipe.Op(ctx, "Second", &SecondReq{S: "hello"}, &out2)


⸻

Pourquoi ça marche
•	gob transporte le type concret de Payload (pointeur vers FirstReq, etc.) grâce à gob.Register.
•	L’adaptateur ajuste pointeur/valeur pour matcher la signature réelle du handler.
•	Tu gardes une interface unique côté infra, avec des handlers joliment typés côté métier. No generics 🎉

Si tu veux, je peux te fournir une version compacte « copiable/collable » de makeAdapter + un test unitaire qui couvre les 4 variantes (valeur/pointeur × avec/sans error).