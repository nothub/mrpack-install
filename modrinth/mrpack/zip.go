package mrpack

import (
	"archive/zip"
	"fmt"
)

func IterZip(zipPath string, callback func(file *zip.File) error) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}

	defer func(r *zip.ReadCloser) {
		err := r.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(r)

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		err := callback(f)
		if err != nil {
			return err
		}
	}

	return nil
}
