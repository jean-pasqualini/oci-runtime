Tu as raison : en Go on ne peut pas mettre de méthodes génériques sur une interface, donc ton SyncPipe ne peut pas avoir Recv[T any]() si tu veux qu’il satisfasse une interface. La solution classique pour garder une API « type-safe » sans générique côté interface, c’est d’utiliser une enveloppe avec un champ de type + json.RawMessage, et un petit registry qui sait instancier le bon payload à la réception.

Voici un pattern simple et robuste avec ton pipe JSON (newline-delimited, compatible avec json.Encoder/Decoder) :

package ipc

import (
"encoding/json"
"errors"
"os"
)

// Enveloppe de message
type Envelope struct {
Kind    string          `json:"kind"`
Payload json.RawMessage `json:"payload"`
}

// Optionnel: interface pour déduire le "Kind" automatiquement
type Kinded interface{ Kind() string }

// Registry pour mapper Kind -> constructeur
type Registry struct {
m map[string]func() any
}

func NewRegistry() *Registry { return &Registry{m: make(map[string]func() any)} }

// On peut utiliser des génériques ici (hors interface) pour l’ergonomie de l’inscription.
func (r *Registry) Register[T any](kind string) {
r.m[kind] = func() any { var zero T; return newT(&zero) }
}

func newT[T any](ptr *T) any { return new(T) }

type SyncPipe struct {
rPipe *os.File
wPipe *os.File
}

func NewSyncPipe(r, w *os.File) *SyncPipe {
return &SyncPipe{rPipe: r, wPipe: w}
}

// Envoi: on enveloppe avec le "kind"
func (s *SyncPipe) Send(kind string, data any) error {
var payload []byte
var err error

	// Si l'appelant te passe déjà un RawMessage, on évite le remarshal
	if rm, ok := data.(json.RawMessage); ok {
		payload = rm
	} else {
		payload, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	enc := json.NewEncoder(s.wPipe)
	return enc.Encode(Envelope{Kind: kind, Payload: payload})
}

// Variante ergonomique si tes structs implémentent Kinded
func (s *SyncPipe) SendAuto(data Kinded) error {
b, err := json.Marshal(data)
if err != nil { return err }
enc := json.NewEncoder(s.wPipe)
return enc.Encode(Envelope{Kind: data.Kind(), Payload: b})
}

// Réception: on lit l’enveloppe puis on instancie le bon type via le registry.
func (s *SyncPipe) Recv(reg *Registry) (kind string, msg any, err error) {
var env Envelope
dec := json.NewDecoder(s.rPipe)
if err = dec.Decode(&env); err != nil {
return "", nil, err
}
ctor := reg.m[env.Kind]
if ctor == nil {
// Si type inconnu: on renvoie le RawMessage brut
return env.Kind, env.Payload, nil
}
dst := ctor() // pointeur vers le type enregistré
if err = json.Unmarshal(env.Payload, dst); err != nil {
return "", nil, err
}
return env.Kind, dst, nil
}

Exemple d’usage

type Ping struct{ Msg string }
func (Ping) Kind() string { return "ping" }

type Pong struct{ OK bool }
func (Pong) Kind() string { return "pong" }

reg := ipc.NewRegistry()
reg.Register[Ping]("ping")
reg.Register[Pong]("pong")

// Envoi
_ = pipe.SendAuto(Ping{Msg: "hello"}) // ou: pipe.Send("ping", Ping{...})

// Réception
kind, v, err := pipe.Recv(reg)
if err != nil { /* handle */ }

switch msg := v.(type) {
case *Ping:
// ...
case *Pong:
// ...
case json.RawMessage:
// Type inconnu : à router autrement ou loguer
default:
// Sécurité
}

Pourquoi ça marche bien ici
•	Pas de génériques dans l’interface (Send/Recv restent non génériques).
•	Dispatch par type sans réflexion lourde : Kind + Registry.
•	Évolutif : tu ajoutes un type → Register[T]("kind"); pas besoin de toucher au pipe.
•	Tolérant : si le kind n’est pas connu, tu récupères un json.RawMessage que tu peux router/forwarder/loguer.

Variantes possibles
•	Contrat d’interface de (dé)serialization : si tu veux complètement décorréler JSON, tu peux définir un contrat MarshalBinary/UnmarshalBinary (ou json.Marshaler/Unmarshaler) sur tes payloads et garder la même enveloppe.
•	Gob / MsgPack : si c’est interne et tu veux éviter les tags JSON, gob ou msgpack marchent très bien avec le même schéma enveloppe + registry.
•	Framing : json.Encoder.Encode ajoute un newline, donc ton pipe est déjà correctement délimité message-par-message.

Si tu me dis les types de messages que tu veux transporter, je te pose le squelette Register[...] et les handlers tout prêts.