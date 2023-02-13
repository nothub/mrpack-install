package download

import (
	"crypto"
	"fmt"
	"github.com/nothub/hashutils/chksum"
	"github.com/nothub/hashutils/encoding"
	"github.com/nothub/mrpack-install/http"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"log"
	"path"
	"path/filepath"
	"sync"
)

type Download struct {
	Path   string
	Urls   []string
	Hashes modrinth.Hashes
}

type Downloader struct {
	Downloads []*Download
	Threads   int // TODO
	Retries   int
}

func (g *Downloader) Download(baseDir string) {
	var wg sync.WaitGroup
	for i := range g.Downloads {
		wg.Add(1)
		dl := g.Downloads[i]
		go func() {
			defer wg.Done()
			absPath, _ := filepath.Abs(path.Join(baseDir, dl.Path))
			success := false
			for _, link := range dl.Urls {
				// retry when download failed
				for retries := 0; retries < g.Retries; retries++ {
					// try download
					f, err := http.DefaultClient.DownloadFile(link, path.Dir(absPath), path.Base(absPath))
					if err != nil {
						log.Printf("Download failed for %s (attempt %v), because: %s\n", dl.Path, retries+1, err.Error())
						continue
					}
					// check hashcode
					_, err = chksum.VerifyFile(f, dl.Hashes.Sha512, crypto.SHA512.New(), encoding.Hex)
					if err != nil {
						log.Printf("Hash check failed for %s (attempt %v), because: %s\n", dl.Path, retries+1, err.Error())
						continue
					}
					// success yay
					fmt.Printf("Download: %s\n", f)
					success = true
					break
				}
				if success {
					break
				}
			}
			if !success {
				log.Printf("Downloaded failed: %s\n", dl.Path)
			}
		}()
	}
	wg.Wait()
}
