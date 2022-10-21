package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/abmpio/upack/pkg"
)

const (
	_defaultOverwrite          = true
	_defaultPrerelease         = false
	_defaultPreserveTimestamps = true
	_defaultCachePackages      = false
)

type install struct {
	//模块所属组、名称、版本的组合名称, 格式使用: 所属组/名称@版本，版本可为空，如system/quartz@2.2.0,system/quartz@2.*"
	PackageName    string
	SourceURL      string
	SourceFeedName string
	ApiKey         string

	_targetDirectory string
	_sourceFeedUrl   string
	_authentication  *[2]string
}

func (*install) Name() string { return "install" }
func (*install) Description() string {
	return "从模块仓储中下载并安装模块到插件目录."
}

func (i *install) Help() string  { return pkg.DefaultCommandHelp(i) }
func (i *install) Usage() string { return pkg.DefaultCommandUsage(i) }

func (*install) PositionalArguments() []pkg.PositionalArgument {
	return []pkg.PositionalArgument{
		{
			Name:        "package",
			Description: "模块所属组、名称、版本的组合名称, 格式使用: 所属组/名称@版本,版本可为空,如system/quartz@2.2.0,system/quartz@2.*",
			Index:       0,
			TrySetValue: pkg.TrySetStringValue("package", func(cmd pkg.Command) *string {
				return &cmd.(*install).PackageName
			}),
		},
	}
}

func (*install) ExtraArguments() []pkg.ExtraArgument {
	return []pkg.ExtraArgument{
		{
			Name:        "sourceUrl",
			Description: "远程仓库的url.",
			Required:    false,
			TrySetValue: pkg.TrySetStringValue("source", func(cmd pkg.Command) *string {
				return &cmd.(*install).SourceURL
			}),
		},
		{
			Name:        "sourceFeedName",
			Description: "远程仓库所对应的feed名称.",
			Required:    false,
			TrySetValue: pkg.TrySetStringValue("sourceFeedName", func(cmd pkg.Command) *string {
				return &cmd.(*install).SourceFeedName
			}),
		},
		{
			Name:        "apikey",
			Description: "访问远程仓库所需要的apiKey.",
			Required:    false,
			TrySetValue: pkg.TrySetPathValue("apikey", func(cmd pkg.Command) *string {
				return &cmd.(*install).ApiKey
			}),
		},
	}
}

// 设置默认属性
func (i *install) setupDefaultProperties() {
	if len(i.ApiKey) <= 0 && len(_configuration.Authentication) > 1 {
		i.ApiKey = _configuration.Authentication[1]
	}
	if len(i.SourceURL) <= 0 {
		i.SourceURL = _configuration.SourceUrl
	}
	if len(i.SourceFeedName) <= 0 {
		i.SourceFeedName = _configuration.SourceFeedName
	}
	i._sourceFeedUrl = getSourceFeedUrl(i.SourceURL, i.SourceFeedName)

	if len(i.ApiKey) > 0 {
		i._authentication = getAuthentication(i.ApiKey)
	}
}

func (i *install) Run() int {
	i.setupDefaultProperties()

	r, size, done, err := i.OpenPackage()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer done()

	zip, err := zip.NewReader(r, size)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	err = pkg.UnpackZip(i._targetDirectory, _defaultOverwrite, zip, _defaultPrerelease)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

func (i *install) OpenPackage() (io.ReaderAt, int64, func() error, error) {
	var r pkg.Registry
	var version *pkg.UniversalPackageVersion

	newPackageInfo, err := parsePackageNameWithVersion(i.PackageName)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("无效的模块名:%s", i.PackageName)
	}

	versionString, err := pkg.GetVersion(i._sourceFeedUrl, newPackageInfo.group, newPackageInfo.name, newPackageInfo.version, i._authentication, _defaultPrerelease)
	if err != nil {
		return nil, 0, nil, err
	}
	version, err = pkg.ParseUniversalPackageVersion(versionString)
	if err != nil {
		return nil, 0, nil, err
	}

	var userName *string
	u, err := user.Current()
	if err == nil {
		userName = &u.Username
	}

	r = pkg.PlugIns
	//version
	newPackageInfo.version = version.String()

	err = r.RegisterPackage(newPackageInfo.group, newPackageInfo.name, version, i._targetDirectory, i._sourceFeedUrl, i._authentication, nil, nil, userName)
	if err != nil {
		return nil, 0, nil, err
	}

	f, done, err := r.GetOrDownload(newPackageInfo.group, newPackageInfo.name, version, i._sourceFeedUrl, i._authentication, _defaultCachePackages)
	if err != nil {
		return nil, 0, nil, err
	}

	i._targetDirectory = i.formatTargetPath(r, newPackageInfo)
	fi, err := f.Stat()
	if err != nil {
		_ = done()
		return nil, 0, nil, err
	}

	return f, fi.Size(), done, nil
}

func (i *install) formatTargetPath(registry pkg.Registry, info *packageInfo) string {
	return filepath.Join(string(registry), info.group, info.name, info.version)
}
