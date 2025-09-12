Tu n’as pas besoin de génériques ici. Le vrai problème est Recv(v *any): le décodeur JSON a besoin d’un pointeur vers un type concret, pas d’un *any. Tu as deux options simples selon ton besoin.

⸻

1) Tu connais le type attendu au point d’appel

Change juste la signature et utilise un pointeur vers la valeur concrète :

package app
type Ipc interface {
Send(data any) error
Recv(v any) error // v DOIT être un pointeur vers le type concret
}

package ipc

import (
"encoding/json"
"io"
)

type SyncPipe struct {
r io.Reader
w io.Writer
}

func (s *SyncPipe) Send(data any) error {
return json.NewEncoder(s.w).Encode(data)
}

func (s *SyncPipe) Recv(v any) error { // ex: &MyStruct{}
return json.NewDecoder(s.r).Decode(v)
}

Utilisation :

var out MyStruct
if err := pipe.Recv(&out); err != nil { /* handle */ }

if err := pipe.Send(in); err != nil { /* handle */ }

Note: j’ai utilisé io.Reader/io.Writer plutôt que *os.File pour ne pas te lier à un type précis.

⸻

2) Tu ne connais PAS le type à l’avance (dispatch dynamique)

Dans ce cas, utilise un enveloppe avec un champ type et un payload brut, puis un petit registre pour instancier le bon type :

type Envelope struct {
Type    string          `json:"type"`
Payload json.RawMessage `json:"payload"`
}

Côté envoi :

func (s *SyncPipe) SendTyped(t string, payload any) error {
b, err := json.Marshal(payload)
if err != nil { return err }
env := Envelope{Type: t, Payload: b}
return json.NewEncoder(s.w).Encode(&env)
}

Côté réception + dispatch :

var registry = map[string]func() any{
"ping": func() any { return &Ping{} },
"pong": func() any { return &Pong{} },
// ...
}

func (s *SyncPipe) RecvEnvelope() (string, any, error) {
var env Envelope
if err := json.NewDecoder(s.r).Decode(&env); err != nil {
return "", nil, err
}
newFn, ok := registry[env.Type]
if !ok {
return env.Type, nil, fmt.Errorf("type inconnu: %s", env.Type)
}
v := newFn()
if err := json.Unmarshal(env.Payload, v); err != nil {
return env.Type, nil, err
}
return env.Type, v, nil
}

Utilisation :

typ, msg, err := pipe.RecvEnvelope()
if err != nil { /* handle */ }

switch v := msg.(type) {
case *Ping:
// ...
case *Pong:
// ...
}


⸻

Alternatives rapides
•	Tu peux restreindre Send/Recv aux types qui implémentent json.Marshaler / json.Unmarshaler, sans génériques, mais ça reste plus verbeux et peu utile si tu utilises déjà encoding/json standard.
•	Si tu veux la sérialisation de type incluse « automatiquement », encoding/gob le fait, mais ce n’est plus du JSON.

⸻

TL;DR
•	Remplace Recv(v *any) par Recv(v any) et passe &MonType{} à l’appel (option 1), ou
•	Ajoute une enveloppe {type, payload} + registre (option 2) si le type est déterminé dynamiquement à l’exécution.