package server

import "net/url"

type DownloadSupplier interface {
	get() (*url.URL, error)
}
