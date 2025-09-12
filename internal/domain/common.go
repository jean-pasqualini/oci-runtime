package domain

type ContainerConfiguration struct {
	Process struct {
		Args []string
	}
	Root struct {
		Path string
	}
}
