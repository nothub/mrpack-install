package server

import "net/url"

type DownloadSupplier interface {
	GetUrl() (*url.URL, error)
}
