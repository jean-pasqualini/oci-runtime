Super question — c’est exactement la règle de l’inversion de dépendances (Clean/Hexagonal).

Règle d’or
•	Les interfaces sont déclarées côté “cœur” (domain/app).
•	L’infrastructure les implémente et est injectée depuis l’extérieur.

Ainsi, le cœur ne dépend pas de l’infra ; c’est l’infra qui dépend du cœur.

Où mettre quoi ?
•	domain/ : entités, value objects, règles métier, éventuellement des ports très métier (ex: ContainerRepository si c’est un concept métier).
•	app/ (application/use cases) : orchestrations de cas d’usage + ports (interfaces) vers l’extérieur : horloge, logger, storage, réseau, IPC/RPC, etc.
•	infra/ : implémentations concrètes de ces ports (files, db, http, ipc, rpc…).

Heuristique simple
•	Concept métier (vocabulaire du domaine) → interface dans domain.
ex. type ContainerRepository interface { Save(Container); FindByID(id) }
•	Dépendance technique / service externe (transport, horloge, UUID, RPC, IPC…) → interface dans app.
ex. type SyncChannel interface { Send(Envelope) error; Recv() (Envelope, error) }

Arborescence suggérée

internal/
domain/
container.go
events.go
// éventuellement : ContainerRepository (si très métier)
app/
usecase/
create_container.go
start_container.go
port/               // <- interfaces déclarées ici
sync_channel.go   // (IPC/RPC abstrait)
clock.go
logger.go
infra/
ipc/
ndjson_codec.go
pipes_sync_channel.go   // implémente app/port.SyncChannel
rpc/
client.go               // idem, impl d’un port
persistence/
container_repo_sql.go   // implémente domain.ContainerRepository (si placé en domain)
cmd/
runc-custom/
init/

Mini exemple en Go

app/port/sync_channel.go

package port

type Envelope struct {
Op      string
Payload []byte
Err     string
}

type SyncChannel interface {
Send(Envelope) error
Recv() (Envelope, error)
}

app/usecase/create_container.go

type CreateContainer struct {
Sync port.SyncChannel
}

func (uc CreateContainer) Execute(spec []byte) error {
if err := uc.Sync.Send(port.Envelope{Op: "spec", Payload: spec}); err != nil {
return err
}
rep, err := uc.Sync.Recv()
if err != nil { return err }
if rep.Err != "" { return errors.New(rep.Err) }
return nil
}

infra/ipc/pipes_sync_channel.go

type PipesSyncChannel struct { enc *json.Encoder; dec *json.Decoder }

func (p *PipesSyncChannel) Send(e port.Envelope) error { return p.enc.Encode(e) }
func (p *PipesSyncChannel) Recv() (port.Envelope, error) {
var e port.Envelope
return e, p.dec.Decode(&e)
}

composition (cmd/runc-custom/main.go) :
•	construire infra/ipc.PipesSyncChannel
•	l’injecter dans app/usecase.CreateContainer

TL;DR
•	Interfaces dans le cœur : domain si vocabulaire métier, sinon app/port.
•	Implémentations dans infra/.
•	Dépendances pointent vers le cœur, jamais l’inverse.
•	Ton internal/ipc et internal/rpc restent en infra ; leurs interfaces vivent dans app/port (ou domain si c’est vraiment métier).