package server

type DownloadSupplier interface {
	GetUrl() (string, error)
}
