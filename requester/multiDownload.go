package requester

import (
	"github.com/nothub/mrpack-install/util"
	"log"
	"sync"
)

type DownloadPool struct {
	downloadLink []string
	hash         map[string]string
	FileName     string
	downloadDir  string
	Success      bool
}

type DownloadPools struct {
	httpClient      *HTTPClient
	DownloadPool    []*DownloadPool
	downloadThreads int
	retryTimes      int
}

func NewDownloadPools(httpClient *HTTPClient, downloadPool []*DownloadPool, downloadThreads int, retryTimes int) *DownloadPools {
	return &DownloadPools{httpClient, downloadPool, downloadThreads, retryTimes}
}

func NewDownloadPool(downloadLink []string, hash map[string]string, fileName string, downloadDir string) *DownloadPool {
	return &DownloadPool{downloadLink, hash, fileName, downloadDir, false}
}

func (downloadPools *DownloadPools) Do() {
	var wg sync.WaitGroup
	ch := make(chan struct{}, downloadPools.downloadThreads)
	for i := range downloadPools.DownloadPool {
		file := downloadPools.DownloadPool[i]

		//goroutine
		ch <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, downloadLink := range file.downloadLink {
				// when download failed retry
				for retryTime := 0; retryTime < downloadPools.retryTimes; retryTime++ {

					//download file
					f, err := downloadPools.httpClient.DownloadFile(downloadLink, file.downloadDir, file.FileName)
					if err != nil {
						log.Println("Download failed for:", file.FileName, err, "retry times:", retryTime)
						continue
					}

					//check hashcode
					if sha1code, ok := file.hash["sha1"]; ok {
						_, err = util.CheckFileSha1(sha1code, f)
					}
					if err != nil {
						log.Println("Hash check failed for:", file.FileName, err, "retry times:", retryTime)
						continue
					}

					log.Println("Downloaded:", f)
					file.Success = true
					break
				}
				if file.Success {
					break
				}
			}
			<-ch
		}()
	}
	wg.Wait()
}
