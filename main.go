package main

import (
	"github.com/nothub/mrpack-install/cmd"
	"log"

	_ "github.com/spf13/cobra"
	_ "github.com/spf13/viper"
)

func main() {
	log.SetFlags(0)
	cmd.Execute()
}
