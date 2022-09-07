package requester

import (
	"github.com/nothub/mrpack-install/util"
	"log"
	"sync"
)

type Download struct {
	links       []string
	hash        map[string]string
	FileName    string
	downloadDir string
	Success     bool
}

type DownloadPools struct {
	httpClient *HTTPClient
	Downloads  []*Download
	threads    int
	retryTimes int
}

func NewDownloadPools(httpClient *HTTPClient, downloadPool []*Download, downloadThreads int, retryTimes int) *DownloadPools {
	return &DownloadPools{httpClient, downloadPool, downloadThreads, retryTimes}
}

func NewDownloadPool(downloadLink []string, hash map[string]string, fileName string, downloadDir string) *Download {
	return &Download{downloadLink, hash, fileName, downloadDir, false}
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
				for retries := 0; retries < downloadPools.retryTimes; retries++ {

					// download file
					f, err := downloadPools.httpClient.DownloadFile(link, dl.downloadDir, dl.FileName)
					if err != nil {
						log.Println("Download failed for:", dl.FileName, err, "attempt:", retries+1)
						continue
					}

					// check hashcode
					if sha1code, ok := dl.hash["sha1"]; ok {
						_, err = util.CheckFileSha1(sha1code, f)
					}
					if err != nil {
						log.Println("Hash check failed for:", dl.FileName, err, "attempt:", retries+1)
						continue
					}

					log.Println("Downloaded:", f)
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
