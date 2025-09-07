package docsParfait — voici un exemple hors contexte (pas OCI) qui montre comment les couches se parlent : un petit cas “envoyer un email de bienvenue”.

Arborescence

example/
├─ cmd/example/main.go          # wiring + point d’entrée
├─ internal/
│  ├─ app/notify.go             # use case (orchestration)
│  ├─ domain/
│  │  ├─ user.go                # entités/erreurs
│  │  └─ ports.go               # interfaces (ports)
│  ├─ adapters/
│  │  └─ mail/smtp_mailer.go    # impl concrète d’un port (Mailer)
│  └─ platform/
│     ├─ config/config.go       # config générique
│     └─ logging/logger.go      # logger générique
└─ go.mod


⸻

domain/ports.go

package domain

// Port (contrat) : l'app sait "j'ai besoin d'envoyer un mail", pas comment.
type Mailer interface {
	Send(to, subject, body string) error
}

domain/user.go

package domain

import "errors"

type User struct {
	ID    string
	Email string
	Name  string
}

var ErrInvalidEmail = errors.New("invalid email")

app/notify.go (use case)

package app

import (
"fmt"
"example/internal/domain"
)

type Notifier struct {
	mailer domain.Mailer // dépend d’une interface du domain
}

func NewNotifier(m domain.Mailer) *Notifier { return &Notifier{mailer: m} }

func (n *Notifier) Welcome(u domain.User) error {
	if u.Email == "" { return domain.ErrInvalidEmail }
	subject := "Welcome!"
	body := fmt.Sprintf("Hi %s, glad you're here.", u.Name)
	return n.mailer.Send(u.Email, subject, body)
}

adapters/mail/smtp_mailer.go (impl du port)

package mail

import "fmt"

// Impl concrète du port domain.Mailer (via SMTP, API provider, etc.)
type SMTPMailer struct {
	Server string
}

func NewSMTPMailer(server string) *SMTPMailer { return &SMTPMailer{Server: server} }

func (m *SMTPMailer) Send(to, subject, body string) error {
	// ici tu mettrais le vrai SMTP; on simule
	fmt.Printf("[SMTP %s] to=%s subj=%q body=%q\n", m.Server, to, subject, body)
	return nil
}

platform/config/config.go (générique)

package config

import "os"

type Config struct {
	SMTPServer string
}

func Load() Config {
	return Config{SMTPServer: getenv("SMTP_SERVER", "smtp.example.local:25")}
}

func getenv(k, d string) string { if v := os.Getenv(k); v != "" { return v }; return d }

platform/logging/logger.go (générique)

package logging

import (
"log/slog"
"os"
)

func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
}

cmd/example/main.go (wiring)

package main

import (
"example/internal/app"
"example/internal/adapters/mail"
"example/internal/domain"
"example/internal/platform/config"
"example/internal/platform/logging"
)

func main() {
	cfg := config.Load()
	log := logging.New()

	// Adapters (impl des ports)
	mailer := mail.NewSMTPMailer(cfg.SMTPServer)

	// App (use case) -> dépend des ports du domain
	notifier := app.NewNotifier(mailer)

	// Domain (données métier)
	u := domain.User{ID: "u1", Email: "dev@acme.test", Name: "Dev"}

	// Exécuter le use case
	if err := notifier.Welcome(u); err != nil {
		log.Error("welcome failed", "err", err)
	}
}


⸻

Ce que ça illustre
•	domain: définit le port Mailer + types métier (User) — pas d’I/O.
•	app: use case Welcome orchestre et ne connaît que l’interface.
•	adapters: implémentent le port (SMTPMailer).
•	platform: générique (config/log) — réutilisable ailleurs.
•	cmd: wire les dépendances et lance le use case.

Tu peux transposer ce pattern à n’importe quel sujet (paiement, PDF, queue, etc.).

----

Ce que je t’ai montré correspond à une famille d’architectures qu’on appelle selon les contextes :

⸻

📌 Les noms les plus courants
•	Hexagonal Architecture (Ports & Adapters)
Inventée par Alistair Cockburn.
•	Domain = cœur de l’hexagone.
•	Ports = interfaces (ex: Mailer).
•	Adapters = implémentations concrètes (ex: SMTPMailer).
👉 D’où tes dossiers domain/ + adapters/.
•	Clean Architecture (Uncle Bob)
Variante qui met l’accent sur les cercles concentriques :
•	Entities (domain) au centre.
•	Use cases (app) autour.
•	Interface adapters (adapters).
•	Frameworks & drivers (platform/cmd).
•	Onion Architecture (Jeffrey Palermo)
Très proche, insiste sur la dépendance orientée vers le centre (domain).

⸻

🔎 Dans ton exemple Go
•	domain/ → Entities & Ports (cœur métier).
•	app/ → Use Cases (orchestration métier).
•	adapters/ → Adapters (implémentations concrètes des ports).
•	platform/ → Cross-cutting concerns (infra générique : logging, config, primitives Linux).
•	cmd/ → Entrypoint / composition root (où tout est branché).

⸻

👉 Donc tu es en plein dans un mélange Clean Architecture / Hexagonal (Ports & Adapters) adapté au style Go.
Certains appellent ça simplement “layered + hexagonal in Go” ou “modular clean architecture”.

⸻

Veux-tu que je te fasse un schéma visuel simple (cercles ou hexagone) qui montre comment cmd ↔ app ↔ domain ↔ adapters ↔ platform s’articulent ?

----


Parfait — voici deux schémas simples.

Schéma “Hexagonal (Ports & Adapters)”

            +-----------------------+
            |      Adapters         |  (impl concrètes des ports)
            |  - http / cli / smtp |
            |  - repo: fs/sql/etcd |
            +-----------+-----------+
                        ^
                        |  dépend via interfaces (ports)
                        |
+-----------+           |            +-----------+
|  cmd/     | --------- | ---------- | platform/ |  (cross-cutting générique)
| (entry)   |     appelle use cases  | logging   |  logging, config, linux/*
| wiring DI |                        | config    |
+-----------+                        +-----------+

                 +---------------------------+
                 |           app/           |  (use cases / orchestration)
                 |  Create, Start, Run, …   |
                 +-------------+-------------+
                               ^
                               | dépend que de
                               | interfaces
                     +---------+---------+
                     |       domain/     |  (cœur métier)
                     |  Entities & Ports |
                     |  (types+interfaces) 
                     +-------------------+

Schéma “Clean (cercles concentriques)”

+--------------------------------------------------------------+
|                      Frameworks / Drivers                    |  <- cmd/, platform/
|         (entrypoint, logging, config, syscalls, etc.)        |
|   dépend de… (vers l'intérieur)                              |
|   v                                                          |
|   +----------------------------------------------+           |
|   |            Interface Adapters                |           |  <- adapters/
|   | (impl des Ports: SMTP, repo FS/SQL, etc.)    |           |
|   |   dépend de…                                 |           |
|   |   v                                          |           |
|   |   +-------------------------------+          |           |
|   |   |           Use Cases           |          |           |  <- app/
|   |   |     (orchestration métier)    |          |           |
|   |   |     dépend des Ports          |          |           |
|   |   +---------------v---------------+          |           |
|   |                   |                          |           |
|   |   +---------------+---------------+          |           |
|   |   |               Domain           |          |           |  <- domain/
|   |   |  (Entities, Value Obj, Ports) |          |           |
|   |   +-------------------------------+          |           |
|   +----------------------------------------------+           |
+--------------------------------------------------------------+

Légende rapide
•	domain/ : cœur stable (types, règles, interfaces/ports).
•	app/ : use cases (orchestration) — dépend uniquement des ports du domain.
•	adapters/ : implémentations concrètes des ports (FS, SQL, HTTP, SMTP…).
•	platform/ : générique et transversal (logger, config, primitives Linux).
•	cmd/ : point d’entrée / composition root (wiring des dépendances).

Si tu veux, je peux te générer la version mermaid de ces schémas pour l’intégrer dans ton README.

----



Bonne question 👍, parce que c’est justement le cœur du pattern (hexagonal / clean) :
👉 les adapters et platform ne sont jamais appelés directement par le domaine → ce sont les use cases (dans app/) qui les utilisent.

⸻

🔹 Qui appelle qui ?
•	cmd/ : c’est le point d’entrée.
•	Il instancie les dépendances (adapters, platform) et les passe à app/.
•	⚠️ cmd ne fait pas de logique métier.
•	app/ (use cases) :
•	Orchestration → c’est lui qui appelle les interfaces définies dans domain/.
•	Concrètement : app reçoit une implémentation d’adapter injectée par cmd.
•	adapters/ :
•	Implémentent les ports du domain (ex: ContainerStore, Mailer).
•	Ce sont donc les app use cases qui appellent les adapters via leurs interfaces.
•	platform/ :
•	Contient des helpers génériques (logger, config, primitives Linux).
•	Ceux-ci sont appelés soit par app soit par adapters, selon le besoin.
•	Exemple : app utilise platform/config pour lire l’ENV, un adapter peut utiliser platform/linux/mount pour faire un pivot_root.

⸻

🔎 Exemple concret

Disons qu’on a un use case : créer un utilisateur et envoyer un email.
1.	cmd/main.go

mailer := adapters.NewSMTPMailer(cfg.SMTPServer)
notifier := app.NewNotifier(mailer)
notifier.Welcome(user)

	2.	app/notifier.go

func (n *Notifier) Welcome(u domain.User) error {
return n.mailer.Send(u.Email, "Welcome!", "Hello "+u.Name)
}

	3.	domain/ports.go

type Mailer interface {
Send(to, subject, body string) error
}

	4.	adapters/smtp_mailer.go

func (m *SMTPMailer) Send(to, subject, body string) error {
return smtp.SendMail(m.server, nil, "noreply@example.com", []string{to}, []byte(body))
}

➡️ Ici :
•	app appelle le port Mailer (défini dans domain).
•	cmd a injecté une implémentation SMTPMailer (adapter).
•	platform/logging pourrait être utilisé par app et adapters pour loguer.

⸻

📌 Résumé
•	cmd appelle app.
•	app appelle les interfaces du domain.
•	Les adapters implémentent ces interfaces, et sont donc utilisés par app.
•	platform est transversal : app et adapters l’appellent quand ils ont besoin d’un outil générique.

⸻

👉 Tu veux que je fasse un diagramme fléché simple (ex: cmd → app → domain → adapters, et app/adapters → platform) pour que tu voies clairement le sens des appels ?