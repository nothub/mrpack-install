package requester

import (
	"github.com/nothub/mrpack-install/util"
	"log"
	"sync"
)

type DownloadPool struct {
	downloadLink []string
	hash         map[string]string
	fileName     string
	downloadDir  string
}

type DownloadPools struct {
	httpClient      *HTTPClient
	downloadPool    []*DownloadPool
	downloadThreads int
	retryTimes      int
}

func NewDownloadPools(httpClient *HTTPClient, downloadPool []*DownloadPool, downloadThreads int, retryTimes int) *DownloadPools {
	return &DownloadPools{httpClient, downloadPool, downloadThreads, retryTimes}
}

func NewDownloadPool(downloadLink []string, hash map[string]string, fileName string, downloadDir string) *DownloadPool {
	return &DownloadPool{downloadLink, hash, fileName, downloadDir}
}

func (downloadPools *DownloadPools) Do() map[string]string {
	var wg sync.WaitGroup
	ch := make(chan struct{}, downloadPools.downloadThreads)
	downloadFailFiles := make(map[string]string)
	for i := range downloadPools.downloadPool {
		file := downloadPools.downloadPool[i]

		//goroutine
		ch <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			success := false
			for _, downloadLink := range file.downloadLink {
				// when download failed retry
				for retryTime := 0; retryTime < downloadPools.retryTimes; retryTime++ {

					//download file
					f, err := downloadPools.httpClient.DownloadFile(downloadLink, file.downloadDir, file.fileName)
					if err != nil {
						log.Println("Downloaded:", file.fileName, err, "retry times:", retryTime)
						downloadFailFiles[file.fileName] = "Download Failed"
						continue
					}

					//check hashcode
					if sha1code, ok := file.hash["sha1"]; ok {
						_, err = util.CheckFileSha1(sha1code, f)
					}
					if err != nil {
						log.Println("Downloaded:", file.fileName, err, "retry times:", retryTime)
						downloadFailFiles[file.fileName] = "Hash Check Failed"
						continue
					}

					log.Println("Downloaded:", f)
					success = true
					break
				}
				if success {
					delete(downloadFailFiles, file.fileName)
					break
				}
			}
			<-ch
		}()
	}
	wg.Wait()
	return downloadFailFiles
}
