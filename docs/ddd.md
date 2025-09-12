Super idée. Voilà à quoi ressemblerait ton runtime OCI en DDD “léger” — concret, exploitable, sans sur-ingénierie.

1) Ubiquitous Language (vocabulaire métier)
   •	Container : unité à lancer.
   •	RootFS : système de fichiers racine du container.
   •	Mount : montage (source/target/type/flags).
   •	NamespacePlan : isolements (UTS, PID, NET, etc.).
   •	InitProcess : PID1 du container.
   •	ResourceLimits : CPU/Mem/PIDs/IO.
   •	Lifecycle : states Created → Initialized → Running → Exited.
   •	Events : ContainerInitialized, ProcessStarted, ProcessExited.

2) Bounded Contexts (sous-domaines)
   •	Container Lifecycle (cœur)
   Orchestration de l’état et des transitions.
   •	Filesystem (FS & Mounts)
   Politique de quels montages sont requis selon la config.
   •	Isolation (Namespaces & Capabilities)
   Politique d’isolement, choix UTS/PID/NET, drops de caps.
   •	Process Supervision
   Démarrage, Exec ou supervision PID1, propagation de signaux.
   •	Resource Governance (Cgroups)
   Application des limites, classes de QoS.

Chacun a ses ports (interfaces) vers le kernel/libcontainer (implémentés par tes adapters).

3) Modèle du domaine (exemple)

// domain/container.go
type ContainerID string

type Container struct {
ID        ContainerID
Spec      Spec              // VO : options choisies (hostname, mounts, entrypoint…)
State     State             // Created/Initialized/Running/Exited
}

type State string

type Spec struct {
Hostname      string
RootFS        RootFS
Mounts        []Mount
NS            NamespacePlan
Limits        ResourceLimits
Entrypoint    []string
Env           []string
RootMode      RootMode // Auto|Pivot|Chroot
}

type RootFS struct { Path string; ReadOnly bool }
type Mount struct   { Source, Target, FSType string; Flags uintptr; Data string }
type NamespacePlan  struct { UTS, IPC, PID, Mount, Net, User, Cgroup bool }
type ResourceLimits struct { CPUQuota int; MemoryBytes int64; PIDs int }
type RootMode string

Value Objects utiles
•	ContainerID, Hostname, Path, CapabilitySet, MountFlags (typer pour éviter les erreurs).
•	Règles d’invariants dans les constructeurs (ex: NewContainerID valide le format).

4) Domain Services (règles métier pures, pas de syscalls)

// domain/services/lifecycle.go
type KernelPorts struct {
Mount   MountPort        // ports (interfaces) vers le monde
NS      NamespacePort
Root    RootPort
Proc    ProcessPort
Cgroup  CgroupPort
}

// Orchestration métier (pas de x/sys/unix ici)
func Initialize(ctx context.Context, c *Container, k KernelPorts) ([]DomainEvent, error) {
// Exemples de règles "métier"
if c.State != "Created" { return nil, ErrInvalidState }

    // Politique d’isolement
    if err := k.NS.Configure(ctx, c.Spec.NS, c.Spec.Hostname); err != nil { return nil, err }

    // Politique de filesystem
    if err := k.Mount.MakePrivate(ctx, "/"); err != nil { return nil, err }
    if err := k.Mount.ApplySet(ctx, c.Spec.Mounts); err != nil { return nil, err }

    // Politique de root switch
    if c.Spec.RootMode == "Pivot" || c.Spec.RootMode == "Auto" {
        if err := k.Root.TryPivotOrChroot(ctx, c.Spec.RootFS.Path, c.Spec.RootMode == "Auto"); err != nil { return nil, err }
    } else {
        if err := k.Root.Chroot(ctx, c.Spec.RootFS.Path); err != nil { return nil, err }
    }

    // Politique de ressources
    if err := k.Cgroup.Apply(ctx, string(c.ID), c.Spec.Limits); err != nil { return nil, err }

    c.State = "Initialized"
    return []DomainEvent{ContainerInitialized{ID: c.ID}}, nil
}

Ici, MountPort, NamespacePort, etc. sont tes ports (interfaces). Les implémentations concrètes sont dans adapters/ (syscalls, libcontainer).

Ports (exemples)

// domain/ports.go
type MountPort interface {
MakePrivate(ctx context.Context, path string) error
ApplySet(ctx context.Context, mounts []Mount) error
}
type NamespacePort interface {
Configure(ctx context.Context, plan NamespacePlan, hostname string) error
}
type RootPort interface {
PivotRoot(ctx context.Context, path string) error
Chroot(ctx context.Context, path string) error
TryPivotOrChroot(ctx context.Context, path string, fallback bool) error // peut rester côté app si tu préfères
}
type ProcessPort interface {
SetComm(ctx context.Context, name string) error
Exec(ctx context.Context, argv, env []string) error
}
type CgroupPort interface {
Apply(ctx context.Context, cid string, limits ResourceLimits) error
}

5) Domain Events (pour tracer/observer sans coupler)

type DomainEvent interface{ Name() string }

type ContainerInitialized struct{ ID ContainerID }
func (ContainerInitialized) Name() string { return "container.initialized" }

type ProcessStarted struct{ ID ContainerID; Pid int }
type ProcessExited struct{ ID ContainerID; ExitCode int }

	•	Émis par les services du domaine.
	•	Observés par une couche “application” pour logger, métriques, audit, etc.

6) Context Map (anti-corruption avec OCI)
   •	Bounded Context “OCI Spec” : un mapper convertit oci.Spec → ton domain.Spec.
   •	Anti-Corruption Layer : protège ton langage métier de la forme externe de la spec.

[ OCI Spec ] --(mapper)--> [ Domain Spec ]

7) Organisation des packages (exemple)

internal/
domain/
container.go          // entité + VO
ports.go              // interfaces
events.go             // domain events
services/
lifecycle.go        // Initialize / Start / Stop
policy_fs.go        // politique de FS/mounts (pure)
policy_isolation.go // politique namespaces/caps (pure)
app/
init/
usecase.go          // charge Spec, appelle services, publie événements
adapters/
linux/
mount/   (unix.Mount, mkdir, errno…)
ns/      (unshare, sethostname…)
root/    (pivot_root, chroot…)
proc/    (prctl, exec…)
cgroups/ (v2)
oci/
mapper/  (oci.Spec → domain.Spec)
platform/
logging, xerr, metrics, config
cmd/oci-runtime/
main.go (composition root : wiring adapters + app)

8) À quoi ça te sert (avantages DDD ici)
   •	Langage commun clair (plus de “où va ce code ?”)
   •	Règles exprimées en code pur (faciles à tester, sans syscalls)
   •	Adaptateurs échangeables (syscalls ↔︎ libcontainer)
   •	Événements pour instrumenter sans couplage (observabilité)

9) “DDD-ifier” ton code en 5 étapes rapides
    1.	Figer le langage : écris un mini glossaire (les 8–10 termes au §1).
    2.	Créer domain.Container + Spec (VO) et State.
    3.	Déplacer la politique dans domain/services (pure logique, pas d’unix).
    4.	Garder les syscalls en adapters/* via les ports.
    5.	Mapper OCI → Domain dans un petit adapter oci/mapper.

⸻

Est-ce “trop” DDD ?

Pas besoin d’agrégats/repositories lourds : ton “dépôt”, c’est le kernel. Fais un DDD pragmatique :
•	Domain services pures ✔️
•	Ports/adapters ✔️
•	Glossaire + events ✔️
•	Mapper OCI ✔️

Si tu veux, je te fournis un exemple complet domain/services/lifecycle.go (initialization + start) avec ses tests unitaires (mocks de ports) pour que tu voies concrètement à quoi ça ressemble.