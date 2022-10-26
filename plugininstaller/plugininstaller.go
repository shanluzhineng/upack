package main

import (
	"os"

	"github.com/abmpio/upack/cmd"
)

func main() {
	cmd.DefaultDispatcher.Run(os.Args[1:])
}
