package domain

type ContainerMountConfiguration struct {
	Destination string
	Type        string
	Source      string
	Options     []string
}

type ContainerConfiguration struct {
	Process struct {
		Args []string
	}
	Root struct {
		Path string
	}
	Mounts []ContainerMountConfiguration
}
