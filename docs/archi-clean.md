https://romanglushach.medium.com/understanding-hexagonal-clean-onion-and-traditional-layered-architectures-a-deep-dive-c0f93b8a1b96
https://lixtec.fr/hexagonal-vs-onion-vs-clean/
https://medium.com/@dorinbaba/n-tier-vs-hexagonal-vs-onion-vs-clean-architecture-in-very-simple-terms-68f66c4dba22
https://programmingpulse.vercel.app/blog/hexagonal-vs-clean-vs-onion-architectures
Très bonne question 👌 Ces trois termes sont souvent confondus, car ils partagent beaucoup de principes.
Je vais te faire un petit comparatif clair :

⸻

1. Clean Architecture (Uncle Bob)
   •	Idée principale : tout tourne autour des use cases (application) et du domain (métier).
   •	Couche centrale = règles métier (domain).
   •	Couche suivante = application (use cases, orchestrations).
   •	Autour = interfaces (ports) définies par l’intérieur.
   •	Tout autour = implémentations concrètes (adapters : DB, OS, réseau…).
   •	Règle d’or : dépendances pointent vers le centre (jamais l’inverse).
   •	Diagramme typique avec les cercles concentriques (entités → use cases → interface adapters → frameworks).

👉 C’est une philosophie assez complète (et un peu dogmatique) : séparation stricte, testabilité maximale.

⸻

2. Architecture en oignon (Onion Architecture, Jeffrey Palermo)
   •	Proche de Clean Architecture mais plus simple et plus “couches concentriques” :
   •	Domain (au centre).
   •	Application Services (logique applicative).
   •	Infrastructure (persistance, IO, impl. externes).
   •	UI (le plus à l’extérieur).
   •	La métaphore de l’oignon : tu traverses les couches de l’extérieur vers le centre.
   •	Même règle : dépendances vers l’intérieur seulement.

👉 L’Onion est plus orientée domain-driven design (DDD) : le domain au centre, entouré par des services, puis l’infrastructure.

⸻

3. Ports & Adapters (a.k.a Hexagonal Architecture, Alistair Cockburn)
   •	Métaphore : l’application est comme une “boîte hexagonale” avec :
   •	Ports = interfaces offertes par l’application (ce dont elle a besoin ou ce qu’elle expose).
   •	Adapters = implémentations pour brancher la boîte au monde extérieur (DB adapter, HTTP adapter, CLI adapter, OS adapter…).
   •	Mets l’accent sur la testabilité : on peut brancher des adapters factices (in-memory, mocks).
   •	Pas forcément en couches concentriques : c’est plus une métaphore de connecteurs.

👉 C’est souvent vu comme la version “pragmatique” de la Clean Architecture. Beaucoup d’équipes disent “ports & adapters” et organisent simplement en application + domain + adapters.

⸻

🔍 Comparaison

Aspect	Clean Architecture	Onion Architecture	Ports & Adapters (Hexagonal)
Métaphore	Cercles concentriques	Oignon (couches)	Hexagone (ports/adapters)
Centre	Domain (Entities)	Domain Model	Application / Domain
Couches	Domain → Use Cases → Interfaces → Infra	Domain → Services → Infra → UI	Pas de couches strictes, juste ports (interfaces) et adapters
Règle des dépendances	Vers l’intérieur	Vers l’intérieur	Vers l’application
Mise en avant	Séparation nette + règles	Domain-driven design	Isolation par interfaces
Pragmatisme	Très dogmatique	Plus simple (DDD-friendly)	Très pragmatique


⸻

🚀 Concrètement (dans ton projet runtime/container)
•	Si tu dis Clean Archi → tu auras :
•	domain/ (ContainerConfiguration, Messages…)
•	application/ (RunHandler orchestrant les ports)
•	platform/ (impls Linux : spawner, namespaces, io…)
•	Si tu dis Onion → c’est quasi pareil, mais tu insistes sur le domain au centre et tu ranges l’infra dehors.
•	Si tu dis Ports & Adapters → tu simplifies le vocabulaire :
•	Ports = interfaces dans application/ports
•	Adapters = impls concrètes dans platform/ ou infra/

👉 En pratique : Clean Archi = Onion + Ports&Adapters (c’est un peu une “fusion + formalisée”).

⸻

Veux-tu que je te fasse un schéma ASCII pour visualiser les trois architectures avec ton cas (OCI runtime) ?


------


Top — voici trois bootstraps minimalistes (en Go) pour le même use case (lancer un container init), montrant surtout où vivent les interfaces, où vivent les implémentations, et où on “wire” le tout.

⸻

Clean Architecture

Idée : domain (métier) au centre, application (use cases) dépend de domain, les ports sont définis côté application, les adapters (infra) sont dehors.

internal/
domain/
container.go          // Entités/DTO
application/
ports/
process.go          // Ports (interfaces)
cfgloader.go
usecase/
run_handler.go      // Orchestration (use case)
platform/               // Adapters concrets (OS/FS)
linux/
process_spawner.go
fs/
cfg_loader_fs.go
cmd/oci-runtime/
main.go                 // Composition root (wiring)

Ports (application/ports/process.go)

package ports

type Process interface {
PID() int
Wait() error
}

type ProcessSpawner interface {
StartInit(ctx context.Context, bundlePath string) (Process, error)
}

Use case (application/usecase/run_handler.go)

type RunHandler struct {
Spawner   ports.ProcessSpawner
CfgLoader ports.CfgLoader
}

func (h RunHandler) Handle(ctx context.Context, bundle string) error {
cfg, err := h.CfgLoader.Load(filepath.Join(bundle, "config.json"))
if err != nil { return err }
_ = cfg // appliquer la logique (hooks, validation…)
proc, err := h.Spawner.StartInit(ctx, bundle)
if err != nil { return err }
return proc.Wait()
}

Adapter (platform/linux/process_spawner.go)

type LinuxSpawner struct{}
func (LinuxSpawner) StartInit(ctx context.Context, bundle string) (ports.Process, error) {
cmd := exec.CommandContext(ctx, "/proc/self/exe", "init")
// SysProcAttr, ExtraFiles, env...
if err := cmd.Start(); err != nil { return nil, err }
return execProcess{cmd}, nil
}

Wiring (cmd/oci-runtime/main.go)

func main() {
spawner := platformlinux.LinuxSpawner{}
loader  := platformfs.NewCfgLoader()
handler := usecase.RunHandler{Spawner: spawner, CfgLoader: loader}
if err := handler.Handle(context.Background(), os.Args[1]); err != nil { log.Fatal(err) }
}


⸻

Onion Architecture

Idée : semblable à Clean, mais le vocabulaire insiste sur Domain au centre puis Services d’application, puis Infrastructure.

internal/
domain/
container.go          // modèle + invariants métier
app/
services/
run_service.go      // service applicatif (use case)
ports/
process.go          // interfaces requises par le service
infrastructure/
linux/
process_spawner.go
fs/
cfg_loader_fs.go
cmd/oci-runtime/
main.go

Service (app/services/run_service.go)

type RunService struct {
Spawner appports.ProcessSpawner
Loader  appports.ConfigLoader
}

func (s RunService) Run(ctx context.Context, bundle string) error {
cfg, err := s.Loader.Load(filepath.Join(bundle, "config.json"))
if err != nil { return err }
// règles métier Domain.* si besoin
proc, err := s.Spawner.StartInit(ctx, bundle)
if err != nil { return err }
return proc.Wait()
}

Wiring (cmd/oci-runtime/main.go)

spawner := infra.LinuxSpawner{}
loader  := infra.NewFSLoader()
svc     := appservices.RunService{Spawner: spawner, Loader: loader}
if err := svc.Run(ctx, bundle); err != nil { log.Fatal(err) }

Différence clé vs Clean : noms des couches (Domain → App Services → Infra). Le code reste très proche.

⸻

Ports & Adapters (Hexagonal)

Idée : l’“hexagone” expose/consomme des ports ; le monde externe se branche via des adapters. Moins de couches nommées, plus de clarté autour des ports.

internal/
hex/
port/
process.go          // ports côté hexagone
cfgloader.go
app/
run.go              // application core (use case)
adapter/
os/                   // adapters “drivers”
linux_spawner.go
fs/
loader.go
cmd/oci-runtime/
main.go

Ports (hex/port/process.go)

package port
type Spawner interface { StartInit(ctx context.Context, bundle string) (Process, error) }
type Process interface   { PID() int; Wait() error }
type ConfigLoader interface { Load(path string) (domain.ContainerConfiguration, error) }

Core (hex/app/run.go)

type RunApp struct {
Spawner port.Spawner
Loader  port.ConfigLoader
}
func (a RunApp) Run(ctx context.Context, bundle string) error {
_, err := a.Loader.Load(filepath.Join(bundle, "config.json"))
if err != nil { return err }
p, err := a.Spawner.StartInit(ctx, bundle)
if err != nil { return err }
return p.Wait()
}

Wiring (cmd/oci-runtime/main.go)

spawner := adapteros.NewLinuxSpawner()
loader  := adapterfs.NewLoader()
app     := hexapp.RunApp{Spawner: spawner, Loader: loader}
if err := app.Run(ctx, bundle); err != nil { log.Fatal(err) }


⸻

Ce qui change vraiment entre les trois
•	Vocabulaire & dossiers :
•	Clean : domain / application / interface adapters / frameworks & drivers.
•	Onion : domain / application services / infrastructure / ui.
•	Hexagonal : core (app+domain) / ports / adapters.
•	Règle des dépendances : identique — vers l’intérieur (core/use cases/domain).
•	Pragmatisme : Hexagonal est le plus “léger” en conventions ; Clean est le plus prescriptif.

⸻

Conseil pratique

Choisis surtout une arbo lisible par ton équipe.
Si tu fais beaucoup d’OS-level (namespaces, FDs), garde un répertoire platform/linux (ou adapter/os) pour clarifier que ce sont des adapters — et définis toujours les ports côté application/core.

Si tu veux, je te génère un squelette de repo (fichiers + go.mod) prêt à copier-coller pour l’un des trois styles.

------

Voici trois schémas ASCII, côte à côte dans l’esprit, pour comparer Clean Architecture, Ports & Adapters (Hexagonal) et Onion Architecture.
Chaque schéma indique les couches/éléments, la direction des dépendances (→ vers le centre) et où placer Ports/Adapters.

⸻

Clean Architecture

          +-----------------------------+
          |  Frameworks & Drivers       |  (DB, HTTP, CLI, OS, UI)
          +--------------+--------------+
                         |
                         v
          +--------------+--------------+
          | Interface Adapters          |  (controllers, presenters, gateways)
          +--------------+--------------+
                         |
                         v
          +--------------+--------------+
          | Application (Use Cases)     |  (orchestration, règles d'app)
          +--------------+--------------+
                         |
                         v
          +--------------+--------------+
          | Domain (Entities)           |  (règles métier pures)
          +-----------------------------+

Dépendances : Frameworks → Adapters → Application → Domain
Ports : définis côté Application (use cases)
Adapters : dans Interface Adapters / Frameworks & Drivers


⸻

Ports & Adapters (Hexagonal)

          [ Adapter UI ]     [ Adapter CLI ]     [ Adapter HTTP ]
                 \                 |                   /
                  \                |                  /
                   \               |                 /
                   +---------------+----------------+
                   |            PORTS               |  (interfaces du core)
                   +---------------+----------------+
                                   |
                                   v
                           +-------+-------+
                           |   CORE        |  (Domain + Application)
                           | (use cases)   |
                           +---------------+
                                   ^
                   +---------------+----------------+
                   |            PORTS               |  (interfaces du core)
                   +---------------+----------------+
                   /               |                 \
                  /                |                  \
         [ Adapter DB ]     [ Adapter FS ]      [ Adapter OS ]

Dépendances : Tous les adapters dépendent des Ports; le Core ne dépend d’aucun adapter.
Ports : dans le Core (exposés/consommés par l’app)
Adapters : implémentations concrètes branchées autour


⸻

Onion Architecture

+---------------------------------------------------------+
|                  UI / Infrastructure                    |  (Frameworks, DB, HTTP, OS)
+------------------------------+--------------------------+
|
v
+------------------------------+--------------------------+
|            Application Services                         |  (orchestration, cas d’usage)
+------------------------------+--------------------------+
|
v
+------------------------------+--------------------------+
|                 Domain Model                            |  (entités, invariants)
+---------------------------------------------------------+

Dépendances : Infra/UI → Application Services → Domain
Ports : souvent définis par Application Services (interfaces requises)
Adapters : dans Infrastructure/UI, implémentent ces interfaces


⸻

Repères rapides
•	Centre :
•	Clean : Domain
•	Hexagonal : Core (Domain + Use cases)
•	Onion : Domain
•	Où vivent les Ports ?
•	Clean : côté Application (Use Cases)
•	Hexagonal : dans le Core (interfaces du hexagone)
•	Onion : souvent dans Application Services
•	Où vivent les Adapters ?
•	Clean : Interface Adapters / Frameworks & Drivers
•	Hexagonal : autour du Core (DB/HTTP/UI/OS adapters)
•	Onion : Infrastructure/UI
•	Règle d’or commune : les dépendances pointent vers l’intérieur (vers le domaine/les use cases), jamais l’inverse.

⸻

Mini mapping concret (ex. runtime OCI)
•	Ports (interfaces) : ProcessSpawner, ControlChannel, ConfigLoader
•	Adapters : impl Linux (namespaces, exec, FDs), FS loader, HTTP/CLI
•	Placement selon l’archi :
•	Clean : Ports dans Application, Adapters dans Interface Adapters/Frameworks (ex. platform/linux)
•	Hexagonal : Ports dans le Core, Adapters dans adapter/* (OS, FS, HTTP)
•	Onion : Ports dans Application Services, Adapters dans Infrastructure

Si tu veux, je peux te générer un squelette de dossiers/fichiers prêt à coller pour l’un des trois modèles.
----


Voici un résumé clair (et actionnable) de l’article que tu as partagé, avec les différences qui comptent en pratique et comment les appliquer en Go.

# Ce que les trois architectures partagent

* **Règle de dépendances** : le **domaine** (règles métier) ne dépend **jamais** de l’infrastructure (DB, HTTP, queues, UI). Les dépendances pointent **vers l’intérieur**. ([Programming Pulse][1])
* **Séparation des préoccupations**, **testabilité**, **modularité** : on isole la logique métier, on « branche » les détails techniques via des interfaces. ([Programming Pulse][1])

# Les trois “saveurs”, en une image mentale

## Hexagonal (Ports & Adapters)

* **Idée-clé** : le cœur expose des **Ports** (interfaces). Les **Adapters** implémentent ces ports pour parler au monde (DB, REST, CLI, MQ).
* Utile quand tu as **plusieurs interfaces** (REST + gRPC + CLI) ou **plusieurs backends** (Postgres vs Dynamo, Stripe vs Adyen). ([Programming Pulse][1])

## Clean Architecture (Uncle Bob)

* **Couches concentriques nommées** : Entities → Use Cases → Interface Adapters → Frameworks & Drivers.
* Très didactique pour **gros domaines métiers** car elle distingue clairement **règles d’entreprise** (Entities) et **règles d’application** (Use Cases). ([Programming Pulse][1])

## Onion

* Similaire au clean/hexagonal, vocabulaire un peu différent : **Domain** au centre, puis **Application Services**, puis **Infrastructure/Presentation** autour.
* Bonne variante quand tu veux une **stratification simple** sans insister sur les “ports/adapters” comme concept premier. ([Programming Pulse][1])

# Quand choisir quoi (règle pratique)

* **Beaucoup d’intégrations externes / plusieurs modes d’I/O** → **Hexagonal** (ports/adapters très explicites). ([Programming Pulse][1])
* **Domaine complexe** (beaucoup de règles, invariants, orchestrations) → **Clean Architecture** (Use Cases au centre du jeu). ([Programming Pulse][1])
* **Projet de complexité moyenne** où tu veux surtout **isoler le domaine** proprement → **Onion** (strates simples). ([Programming Pulse][1])

# Pièges fréquents (peu importe la saveur)

* **Sur-abstraction** : multiplier les interfaces “au cas où”. Crée un port **uniquement** s’il stabilise une frontière.
* **DTO/mapper mania** : n’en fais pas une usine à gaz ; aligne les modèles quand c’est raisonnable.
* **Domaine anémique** : mets la **logique** près des **entités/valeurs**, pas dans 1000 services utilitaires.
* **Fausse indépendance** : si ton domaine “connait” Postgres via des types SQL, tu as déjà fui la règle.

# Squelette minimal en Go (style hexagonal/clean, concret)

```go
// Domaine (entités + règles) — ne dépend de rien
package domain

type OrderID string
type Money int64

type Order struct {
    ID     OrderID
    Amount Money
}

func (o *Order) Pay(amount Money) error { /* invariants & règles */ return nil }

// Port (côté domaine) — abstrait
type OrderRepo interface {
    ByID(id OrderID) (*Order, error)
    Save(o *Order) error
}

// Cas d’usage (application) — orchestre le domaine via des ports
package app

import "myapp/domain"

type PayOrder struct{ Repo domain.OrderRepo }

func (uc PayOrder) Exec(id domain.OrderID, amt domain.Money) error {
    o, err := uc.Repo.ByID(id)
    if err != nil { return err }
    if err := o.Pay(amt); err != nil { return err }
    return uc.Repo.Save(o)
}
```

```go
// Adapter infra (impl du Port) — dépend vers l’intérieur
package postgres

type PgOrderRepo struct { /* db handle */ }

func (r PgOrderRepo) ByID(id domain.OrderID) (*domain.Order, error) { /* SQL */ }
func (r PgOrderRepo) Save(o *domain.Order) error { /* SQL */ }
```

```go
// Adapter delivery (HTTP) — convertit HTTP ↔ use case
func PayOrderHandler(w http.ResponseWriter, r *http.Request) {
    id := domain.OrderID(mux.Vars(r)["id"])
    amt := parseMoney(r)
    err := payOrderUC.Exec(id, amt)
    writeHTTP(err, w)
}
```

* **Dépendances** : HTTP adapter → app.UseCase → domain (+ ports) ← infra adapter
* Tu peux remplacer Postgres par Dynamo (nouvel adapter) ou exposer gRPC en plus (nouvel adapter) **sans toucher au domaine**.

# TL;DR

Les trois architectures visent la **même cible** (domaine indépendant, tests faciles, modularité). La différence est surtout **terminologique et pédagogique** : **Hexagonal** met l’accent sur **Ports/Adapters**, **Clean** sur des **couches nommées** et la **dépendance vers l’intérieur**, **Onion** sur une **stratification simple** autour du **domaine**. Choisis la saveur qui **rendra la frontière la plus claire** pour ton équipe et ton contexte. ([Programming Pulse][1])

Si tu veux, je te propose un **arborescence Go** (modules/packages) prête à copier-coller pour ton prochain repo.

[1]: https://programmingpulse.vercel.app/blog/hexagonal-vs-clean-vs-onion-architectures "Hexagonal vs Clean vs Onion Architectures: Choosing the Right Architecture | Programming Pulse"
