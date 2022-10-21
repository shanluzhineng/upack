package cmd

import (
	"os"

	"github.com/abmpio/upack/pkg"
)

var (
	// 应用标题
	AppTitle string = "plugininstaller"

	// 应用版本
	AppVersion string = pkg.Version

	// 应用描述
	AppDescription string
)

func Main() {
	DefaultDispatcher.Main(os.Args[1:])
}
