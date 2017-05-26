package main

import (
	"fmt"
	"os"

	"github.com/off-sync/platform-proxy/cmd/certgen/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
