//usr/bin/env -S go run "$0" "$@" ; exit
//go:build exclude

package main

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"os/exec"
	"text/template"

	"hub.lol/mrpack-install/cmd"
)

//go:embed readme.tmpl
var fs embed.FS

func init() {
	log.SetFlags(0)
}

type CmdEntry struct {
	Name string
	Help string
}

func NewCmdEntry(name string, cmd string) CmdEntry {
	help, err := exec.Command("./out/mrpack-install", cmd, "--help").CombinedOutput()
	if err != nil {
		log.Fatalln(err.Error())
	}
	return CmdEntry{
		Name: name,
		Help: string(help),
	}
}

func main() {

	err := exec.Command("go", "build", "-o", "out/mrpack-install").Run()
	if err != nil {
		log.Fatalln(err.Error())
	}

	var data []CmdEntry
	data = append(data, NewCmdEntry("root", ""))
	for _, sc := range cmd.RootCmd.Commands() {
		data = append(data, NewCmdEntry(sc.Name(), sc.Name()))
	}

	tmpl, err := template.ParseFS(fs, "readme.tmpl")
	if err != nil {
		log.Fatalln(err.Error())
	}

	var buf = bytes.Buffer{}
	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Print(buf.String())

}
