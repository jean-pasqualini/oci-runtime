package domain

type Mount struct {
	Source string
	Target string
	FSType string
	Flags  uintptr
	Data   string
}
