package domain

import "time"

const (
	StatusCreated = "created"
	StatusRunning = "running"
	StatusStopped = "stopped"
)

type ContainerState struct {
	Name      string
	Status    string `json:"-"`
	Pid       int
	Bundle    string
	CreatedAt time.Time
}
