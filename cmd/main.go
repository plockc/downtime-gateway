package main

import (
	"fmt"
	"os"

	"github.com/plockc/gateway/exec"
	"github.com/plockc/gateway/handle"
)

func main() {
	if _, out, err := exec.ExecLine("id -u"); err != nil {
		fmt.Println("Could not determine user id: " + err.Error())
		os.Exit(1)
	} else if out != "0" {
		fmt.Println("must be run as root")
		os.Exit(1)
	}
	handle.Serve()
}
