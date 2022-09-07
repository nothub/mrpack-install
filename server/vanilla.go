package server

import (
	"errors"
	"github.com/nothub/mrpack-install/mojang"
	"github.com/nothub/mrpack-install/requester"
	"github.com/nothub/mrpack-install/util"
	"log"
)

type Vanilla struct {
	MinecraftVersion string
}

func (provider *Vanilla) Provide(serverDir string, serverFile string) error {
	meta, err := mojang.GetMeta(provider.MinecraftVersion)
	if err != nil {
		return err
	}

	file, err := requester.DefaultHttpClient.DownloadFile(meta.Downloads.Server.Url, serverDir, serverFile)
	if err != nil {
		log.Fatalln(err)
	}

	ok, err := util.CheckFileSha1(meta.Downloads.Server.Sha1, file)
	if err != nil {
		log.Fatalln(err)
	}
	if !ok {
		return errors.New("invalid file hashsum")
	}

	return nil
}
