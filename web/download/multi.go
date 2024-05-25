package download

import (
	"crypto"
	"log"
	"path/filepath"
	"sync"

	"github.com/nothub/hashutils/chksum"
	"github.com/nothub/hashutils/encoding"
	modrinth "github.com/nothub/mrpack-install/modrinth/api"
	"github.com/nothub/mrpack-install/web"
)

type Download struct {
	Path   string
	Urls   []string
	Hashes modrinth.Hashes
}

type Downloader struct {
	Downloads []*Download
	Threads   int // Maximum number of concurrent downloads
	Retries   int
}

type empty struct{}

func (g *Downloader) Download(baseDir string) {
	semaphore := make(chan empty, g.Threads) // Create a semaphore to limit concurrency
	var wg sync.WaitGroup

	for i := range g.Downloads {
		wg.Add(1)
		dl := g.Downloads[i]

		semaphore <- empty{} // Acquire a slot in the semaphore
		go func() {
			defer func() {
				<-semaphore // Release the slot in the semaphore
				wg.Done()
			}()

			absPath, _ := filepath.Abs(filepath.Join(baseDir, dl.Path))
			success := false
			for _, link := range dl.Urls {
				// retry when download failed
				for retries := 0; retries < g.Retries; retries++ {
					// try download
					f, err := web.DefaultClient.DownloadFile(link, filepath.Dir(absPath), filepath.Base(absPath))
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
					log.Printf("Download: %s\n", f)
					success = true
					break
				}
				if success {
					break
				}
			}
			if !success {
				log.Printf("Downloaded failed: %s\n", dl.Path)
				// TODO: tell user what manual actions are required?
			}
		}()
	}

	wg.Wait()
}
