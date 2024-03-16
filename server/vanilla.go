package server

import (
	"crypto"
	"errors"
	"github.com/nothub/hashutils/chksum"
	"github.com/nothub/hashutils/encoding"
	"github.com/nothub/mrpack-install/mojang"
	"github.com/nothub/mrpack-install/web"
	"log"
)

type VanillaInstaller struct {
	MinecraftVersion string
}

func (inst *VanillaInstaller) Install(serverDir string, serverFile string) error {
	meta, err := mojang.GetMeta(inst.MinecraftVersion)
	if err != nil {
		return err
	}

	file, err := web.DefaultClient.DownloadFile(meta.Downloads.Server.Url, serverDir, serverFile)
	if err != nil {
		log.Fatalln(err)
	}

	ok, err := chksum.VerifyFile(file, meta.Downloads.Server.Sha1, crypto.SHA1.New(), encoding.Hex)
	if err != nil {
		log.Fatalln(err)
	}
	if !ok {
		return errors.New("invalid file hashsum")
	}

	return nil
}
