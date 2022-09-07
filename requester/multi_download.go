package requester

import (
	"fmt"
	"github.com/nothub/mrpack-install/util"
	"sync"
)

type Download struct {
	links       []string
	hashes      map[string]string
	FileName    string
	downloadDir string
	Success     bool
}

type DownloadPools struct {
	httpClient *HTTPClient
	Downloads  []*Download
	threads    int
	maxRetries int
}

func NewDownloadPools(httpClient *HTTPClient, downloads []*Download, threads int, maxRetries int) *DownloadPools {
	return &DownloadPools{httpClient, downloads, threads, maxRetries}
}

func NewDownload(links []string, hashes map[string]string, fileName string, downloadDir string) *Download {
	return &Download{links, hashes, fileName, downloadDir, false}
}

func (downloadPools *DownloadPools) Do() {
	var wg sync.WaitGroup
	ch := make(chan struct{}, downloadPools.threads)
	for i := range downloadPools.Downloads {
		dl := downloadPools.Downloads[i]

		//goroutine
		ch <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, link := range dl.links {
				// retry when download failed
				for retries := 0; retries < downloadPools.maxRetries; retries++ {

					// download file
					f, err := downloadPools.httpClient.DownloadFile(link, dl.downloadDir, dl.FileName)
					if err != nil {
						fmt.Println("Download failed for:", dl.FileName, err, "attempt:", retries+1)
						continue
					}

					// check hashcode
					if sha1code, ok := dl.hashes["sha1"]; ok {
						_, err = util.CheckFileSha1(sha1code, f)
					}
					if err != nil {
						fmt.Println("Hash check failed for:", dl.FileName, err, "attempt:", retries+1)
						continue
					}

					fmt.Println("Downloaded:", f)
					dl.Success = true
					break
				}
				if dl.Success {
					break
				}
			}
			<-ch
		}()
	}
	wg.Wait()
}
