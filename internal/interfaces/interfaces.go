package interfaces

type Bot interface {
	Start() error
	Stop() error
}
