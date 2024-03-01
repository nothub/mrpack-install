//usr/bin/env -S go run "$0" "$@" ; exit
////go:build exclude

package main

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"os/exec"
	"text/template"
)

//go:embed readme.tmpl
var fs embed.FS

func init() {
	log.SetFlags(0)
}

func main() {

	err := exec.Command("go", "build", "-o", "out/mrpack-install").Run()
	if err != nil {
		log.Fatalln(err.Error())
	}

	root, err := exec.Command("./out/mrpack-install", "--help").CombinedOutput()
	if err != nil {
		log.Fatalln(err.Error())
	}

	server, err := exec.Command("./out/mrpack-install", "server", "--help").CombinedOutput()
	if err != nil {
		log.Fatalln(err.Error())
	}

	update, err := exec.Command("./out/mrpack-install", "update", "--help").CombinedOutput()
	if err != nil {
		log.Fatalln(err.Error())
	}

	tmpl, err := template.ParseFS(fs, "readme.tmpl")
	if err != nil {
		log.Fatalln(err.Error())
	}

	var buf = bytes.Buffer{}
	err = tmpl.Execute(&buf, map[string]string{
		"root":   string(root),
		"server": string(server),
		"update": string(update),
	})
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Print(buf.String())

}
