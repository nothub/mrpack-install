package download

import (
	"crypto"
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strings"
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

// Download fetches every file into baseDir and returns an error naming
// each file that no URL delivered with a matching hash. Every URL gets at
// least one attempt, also with zero Retries.
func (g *Downloader) Download(baseDir string) error {
	semaphore := make(chan empty, g.Threads) // Create a semaphore to limit concurrency
	var wg sync.WaitGroup
	var mu sync.Mutex
	var failed []string

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

				// TODO: check url host based on modrinth rules
				//    cdn.modrinth.com
				//    github.com
				//    raw.githubusercontent.com
				//    gitlab.com

				// retry when download failed
				for attempt := 1; attempt <= max(1, g.Retries); attempt++ {
					// try download
					f, err := web.DefaultClient.DownloadFile(link, filepath.Dir(absPath), filepath.Base(absPath))
					if err != nil {
						log.Printf("Download failed for %s (attempt %v), because: %s\n", dl.Path, attempt, err.Error())
						continue
					}
					// check hashcode
					_, err = chksum.VerifyFile(f, dl.Hashes.Sha512, crypto.SHA512.New(), encoding.Hex)
					if err != nil {
						log.Printf("Hash check failed for %s (attempt %v), because: %s\n", dl.Path, attempt, err.Error())
						continue
					}
					// success yay
					log.Printf("Downloaded: %s\n", f)
					success = true
					break
				}
				if success {
					break
				}
			}
			if !success {
				log.Printf("Download failed: %s\n", dl.Path)
				mu.Lock()
				failed = append(failed, dl.Path)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	if len(failed) > 0 {
		slices.Sort(failed)
		return fmt.Errorf("failed to download: %s", strings.Join(failed, ", "))
	}
	return nil
}
