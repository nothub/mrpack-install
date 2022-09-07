package server

type Provider interface {
	Provide(serverDir string, serverFile string) error
}
