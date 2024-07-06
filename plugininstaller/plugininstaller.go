package main

import (
	"os"

	"github.com/shanluzhineng/upack/cmd"
)

func main() {
	cmd.RegistCommand(&cmd.Install{},
		&cmd.InstallApp{},
		&cmd.Pack{},
		&cmd.PackApp{},
		&cmd.Push{},
		&cmd.List{},
	)
	cmd.DefaultDispatcher.Run(os.Args[1:])
}
