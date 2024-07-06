package cmd

import (
	"strings"

	"github.com/shanluzhineng/upack/pkg"
)

type InstallApp struct {
	//应用名称, 格式使用: 所属组/名称@版本，[所属组]与[版本]可为空，如App/helloworld@2.*，如果不包含所属组，如helloworld，则将使用App组
	PackageName string
}

func (*InstallApp) Name() string { return "installapp" }
func (*InstallApp) Description() string {
	return "从模块仓储中下载并安装应用到当前目录."
}

func (i *InstallApp) Help() string  { return pkg.DefaultCommandHelp(i) }
func (i *InstallApp) Usage() string { return pkg.DefaultCommandUsage(i) }

func (*InstallApp) PositionalArguments() []pkg.PositionalArgument {
	return []pkg.PositionalArgument{
		{
			Name:        "package",
			Description: "应用名称, 格式使用: 所属组/名称@版本，[所属组]与[版本]可为空,如App/helloworld@2.*,如果不包含所属组,如helloworld,则将使用App组",
			Index:       0,
			TrySetValue: pkg.TrySetStringValue("package", func(cmd pkg.Command) *string {
				return &cmd.(*InstallApp).PackageName
			}),
		},
	}
}

func (*InstallApp) ExtraArguments() []pkg.ExtraArgument {
	return nil
}

func (i *InstallApp) Run() int {

	packageName := i.PackageName
	index := strings.Index(packageName, "/")
	if index == -1 {
		//不包含组名
		packageName = _defaultAppGroupName + "/" + packageName
	}
	installCmd := new(Install)
	installCmd.PackageName = packageName
	installCmd.SourceFeedName = _defaultAppSourceFeedName
	installCmd.Type = PackageType_App

	return installCmd.Run()
}
