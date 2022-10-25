package cmd

import (
	"strings"

	"github.com/abmpio/upack/pkg"
)

type installapp struct {
	//应用名称, 格式使用: 所属组/名称@版本，[所属组]与[版本]可为空，如App/helloworld@2.*，如果不包含所属组，如helloworld，则将使用App组
	PackageName string
	ApiKey      string
}

func (*installapp) Name() string { return "installapp" }
func (*installapp) Description() string {
	return "从模块仓储中下载并安装应用到当前目录."
}

func (i *installapp) Help() string  { return pkg.DefaultCommandHelp(i) }
func (i *installapp) Usage() string { return pkg.DefaultCommandUsage(i) }

func (*installapp) PositionalArguments() []pkg.PositionalArgument {
	return []pkg.PositionalArgument{
		{
			Name:        "package",
			Description: "应用名称, 格式使用: 所属组/名称@版本，[所属组]与[版本]可为空,如App/helloworld@2.*,如果不包含所属组,如helloworld,则将使用App组",
			Index:       0,
			TrySetValue: pkg.TrySetStringValue("package", func(cmd pkg.Command) *string {
				return &cmd.(*installapp).PackageName
			}),
		},
	}
}

func (*installapp) ExtraArguments() []pkg.ExtraArgument {
	return []pkg.ExtraArgument{
		{
			Name:        "apikey",
			Description: "访问远程仓库所需要的apiKey.",
			Required:    false,
			TrySetValue: pkg.TrySetPathValue("apikey", func(cmd pkg.Command) *string {
				return &cmd.(*installapp).ApiKey
			}),
		},
	}
}

func (i *installapp) Run() int {

	packageName := i.PackageName
	index := strings.Index(packageName, "/")
	if index == -1 {
		//不包含组名
		packageName = _defaultAppGroupName + "/" + packageName
	}
	installCmd := new(install)
	installCmd.PackageName = packageName
	installCmd.ApiKey = i.ApiKey
	installCmd.SourceFeedName = _defaultAppSourceFeedName
	installCmd.Type = PackageType_App

	return installCmd.Run()
}
