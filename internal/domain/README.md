Super question. Pense “cœur du métier, pur et stable”.
Dans internal/domain, tu mets ce qui décrit ce que fait ton système, pas comment il le fait.

Ce qu’on met (général)
•	Entités & Value Objects : structs avec invariants (ex: User, Money, Status, ID).
•	Interfaces (ports) : contrats côté outside world (ex: Store, Clock, Publisher).
•	Règles métier & invariants : fonctions pures qui valident/changent l’état (sans I/O).
•	Erreurs métier : types d’erreurs sémantiques (ErrAlreadyExists, ErrInvalidState).
•	Événements métier (optionnel) : DomainEvent (ex: ContainerStarted).
•	Policies / Specifications (optionnel) : règles combinables en pur Go.
•	Factories (optionnel) : création d’objets valides (NewContainer(...)).

Ce qu’on ne met PAS
•	Pas d’I/O (fichiers, réseau, DB), pas de syscalls.
•	Pas de logs, pas de context, pas d’HTTP, pas de SQL.
•	Pas de dépendance à un framework/driver. (Dépend juste de la stdlib.)

Mini-exemples (généraux)

Entité + invariants

package domain

import "time"

type Status string
const (
StatusCreated Status = "created"
StatusRunning Status = "running"
StatusStopped Status = "stopped"
)

type Container struct {
ID        string
Status    Status
Pid       int
Bundle    string
CreatedAt time.Time
}

func (c *Container) CanStart() bool  { return c.Status == StatusCreated }
func (c *Container) CanDelete() bool { return c.Status == StatusStopped || c.Status == StatusCreated }

Ports (interfaces)

package domain

type ContainerStore interface {
Save(Container) error
FindByID(id string) (Container, error)
Delete(id string) error
List() ([]Container, error)
}

type BundleLoader interface {
Load(path string) (Bundle, error)
}

type Clock interface { Now() time.Time } // utile pour tests déterministes

Value object + factory + erreurs

package domain

import "errors"

var (
ErrInvalidID     = errors.New("invalid id")
ErrInvalidStatus = errors.New("invalid status transition")
)

type Bundle struct {
Path   string
Rootfs string
}

func NewContainer(id, bundlePath string, now time.Time) (Container, error) {
if id == "" {
return Container{}, ErrInvalidID
}
return Container{
ID:        id,
Status:    StatusCreated,
Bundle:    bundlePath,
CreatedAt: now,
}, nil
}

Domain event (optionnel)

package domain

import "time"

type Event interface {
Name() string
At() time.Time
}

type ContainerStarted struct {
ID string
Ts time.Time
}

func (e ContainerStarted) Name() string { return "container.started" }
func (e ContainerStarted) At() time.Time { return e.Ts }

Règle d’or
•	Stable & testable : si demain tu changes de DB, d’OS, d’API, le domain ne bouge (presque) pas.
•	Pur : idéalement testable sans mocks lourds (juste remplacer les ports par des fakes).

Si tu veux, je te propose un starter pack de fichiers vides pour domain/ (IDs, Status, Container, ports Store/BundleLoader, erreurs).