Je vois le problème : avec type Rpc[Req, Res] interface { Op(name string, payload Req) Res }, dès que tu changes de couple (Req, Res), tu dois fabriquer une nouvelle instance/implémentation. En Go, on ne peut pas mettre des paramètres de type sur une méthode d’interface, donc pas de Op[Req,Res] dans une interface. Mais il existe plusieurs patterns pour s’en sortir, chacun avec ses compromis.

4 façons de faire (avec exemples)

1) Faire de Rpc un type concret avec une méthode générique

Tu perds l’interface, mais tu gagnes la flexibilité au call-site.

type Rpc struct {
// deps, transport, codec...
}

func (r *Rpc) Op[Req any, Res any](name string, payload Req) (Res, error) {
var zero Res
// ... encode(payload), send(name), decode into Res ...
return zero, nil
}

// usage
var r Rpc
res, err := r.Op[MyReq, MyRes]("getUser", MyReq{ID: "42"})

+ Appels ultra flexibles, un seul client.
  − Pas d’abstraction par interface (sauf à ruser, voir #2).

⸻

2) Interface non générique + fonction générique d’adaptation

On garde une interface simple et on fournit une fonction générique qui fait le cast.
On perd un peu de sécurité à l’intérieur, mais pas au call-site.

type RawRPC interface {
Op(name string, payload any) (any, error)
}

func Call[Req any, Res any](r RawRPC, name string, payload Req) (Res, error) {
v, err := r.Op(name, payload)
if err != nil {
var zero Res
return zero, err
}
return v.(Res), nil // cast runtime (panique si mismatch)
}

// usage
res, err := Call[MyReq, MyRes](raw, "getUser", MyReq{ID: "42"})

+ Une seule implémentation, call-site générique.
  − Cast runtime (pas 100% type-safe à l’exécution).

Astuce : sécurise en stockant aussi le type attendu et vérifie via reflect.TypeOf ou en transportant un codec typé.

⸻

3) Router avec enregistrement de handlers génériques (typage fort à l’enregistrement et à l’appel)

Très pratique si tu as un bus interne.

type Handler[Req any, Res any] func(ctx context.Context, req Req) (Res, error)

type Router struct {
m map[string]any
}

func NewRouter() *Router { return &Router{m: make(map[string]any)} }

func (rt *Router) Register[Req any, Res any](name string, h Handler[Req, Res]) {
rt.m[name] = h
}

func (rt *Router) Call[Req any, Res any](ctx context.Context, name string, req Req) (Res, error) {
var zero Res
h, ok := rt.m[name].(Handler[Req, Res])
if !ok {
return zero, fmt.Errorf("handler type mismatch for %s", name)
}
return h(ctx, req)
}

// usage
rt := NewRouter()
rt.Register[MyReq, MyRes]("getUser", func(ctx context.Context, r MyReq) (MyRes, error) {
return MyRes{ID: r.ID, Name: "Alice"}, nil
})
res, err := rt.Call[MyReq, MyRes](ctx, "getUser", MyReq{ID: "42"})

+ API ergonomique, compile-time sur l’enregistrement, erreurs claires à l’appel.
  − Internes stockés en any, donc cast au centre (mais contrôlé).

⸻

4) Codegen (si tu veux 100% type-safe partout)

Tu génères RpcUser, RpcOrder, etc., via go:generate (ou genny, ou un petit outil maison).
+ Zéro cast, interfaces propres par domaine.
  − Boilerplate généré et chaîne de build un peu plus lourde.

⸻

Quel choisir ?
•	Tu peux vivre sans interface ? → #1 (client concret avec méthode générique) : simple et efficace.
•	Tu tiens à une interface unique ? → #2 (adapter générique) ou #3 (router + handlers).
•	Projet long terme / gros domaine stable ? → #4 (codegen) pour une robustesse maximale.

Si tu me dis ton contexte (client réseau ? codec ? erreurs ? besoin d’async ?), je te propose une version prête à coller adaptée à ton cas.