package domain

const (
	StatusRunning = "running"
)

type ContainerState struct {
	Name   string
	Status string
}
