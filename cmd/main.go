package main

import (
	"fmt"
	"os"

	"github.com/plockc/gateway"
	"github.com/plockc/gateway/handle"
)

func main() {
	if _, out, err := gateway.ExecLine("id -u"); err != nil {
		fmt.Println("Could not determine user id: " + err.Error())
		os.Exit(1)
	} else if out != "0" {
		fmt.Println("must be run as root")
		os.Exit(1)
	}
	handle.Serve()
}
