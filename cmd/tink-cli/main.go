package main

import (
	"fmt"
	"os"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd"
)

// version is set at build time
var version = "devel"

func main() {
	conn, err := client.GetConnection()
	if err != nil {
		panic(err)
	}
	if err := cmd.Execute(version, client.NewFullClient(conn)); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
