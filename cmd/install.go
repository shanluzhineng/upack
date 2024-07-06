package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/shanluzhineng/upack/pkg"
)

const (
	_defaultOverwrite          = true
	_defaultPrerelease         = false
	_defaultPreserveTimestamps = true
	_defaultCachePackages      = false
)

type Install struct {
	//模块所属组、名称、版本的组合名称, 格式使用: 所属组/名称@版本，版本可为空，如system/quartz@2.2.0,system/quartz@2.*"
	PackageName    string
	SourceFeedName string

	Type PackageType
	//下载的包的元数据
	_metadata        *pkg.UniversalPackageMetadata
	_registry        pkg.Registry
	_packageInfo     *packageInfo
	_targetDirectory string

	//配置信息
	_configuration Configuration
}

func (*Install) Name() string { return "install" }
func (*Install) Description() string {
	return "从模块仓储中下载并安装模块到插件目录."
}

func (i *Install) Help() string  { return pkg.DefaultCommandHelp(i) }
func (i *Install) Usage() string { return pkg.DefaultCommandUsage(i) }

func (*Install) PositionalArguments() []pkg.PositionalArgument {
	return []pkg.PositionalArgument{
		{
			Name:        "package",
			Description: "模块所属组、名称、版本的组合名称, 格式使用: 所属组/名称@版本,版本可为空,如system/quartz@2.2.0,system/quartz@2.*",
			Index:       0,
			TrySetValue: pkg.TrySetStringValue("package", func(cmd pkg.Command) *string {
				return &cmd.(*Install).PackageName
			}),
		},
	}
}

func (*Install) ExtraArguments() []pkg.ExtraArgument {
	return nil
}

// 设置默认属性
func (i *Install) setupDefaultProperties() {
	if len(i.SourceFeedName) > 0 {
		i._configuration = defaultConfigurationWithFeedName(i.SourceFeedName)
	} else {
		i._configuration = *defaultConfiguration()
	}
	if len(i.Type) <= 0 {
		i.Type = PackageType_Plugin
		i._registry = pkg.PlugIns
	}
}

func (i *Install) Run() int {
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

	i._targetDirectory = i.formatTargetPath(i._packageInfo)
	err = pkg.UnpackZip(i._targetDirectory, _defaultOverwrite, zip, _defaultPrerelease)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

func (i *Install) OpenPackage() (io.ReaderAt, int64, func() error, error) {
	var version *pkg.UniversalPackageVersion

	newPackageInfo, err := parsePackageNameWithVersion(i.PackageName)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("无效的模块名:%s", i.PackageName)
	}

	//保存解析的packageInfo
	i._packageInfo = newPackageInfo

	versionString, err := pkg.GetVersion(i._configuration.SourceFeedUrl,
		newPackageInfo.group,
		newPackageInfo.name,
		newPackageInfo.version,
		i._configuration.Authentication,
		_defaultPrerelease)
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

	//version
	newPackageInfo.version = version.String()

	targetDirectory := i.formatTargetPath(newPackageInfo)

	f, done, err := i._registry.GetOrDownload(newPackageInfo.group,
		newPackageInfo.name,
		version,
		i._configuration.SourceFeedUrl,
		i._configuration.Authentication,
		_defaultCachePackages)
	if err != nil {
		return nil, 0, nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		_ = done()
		return nil, 0, nil, err
	}

	zip, err := zip.NewReader(f, fi.Size())
	if err == nil {
		//check upack is app or plugin?
		metadata, err := i.readManifest(zip)
		if metadata != nil && err == nil {
			i._metadata = metadata
			packageType, ok := (*metadata)[_metaPropertyName_Type].(string)
			if ok && len(packageType) > 0 {
				i.Type = PackageType(packageType)
			}
		}
	}

	if i.Type == PackageType_Plugin {
		err = i._registry.RegisterPackage(newPackageInfo.group,
			newPackageInfo.name,
			version,
			targetDirectory,
			i._configuration.SourceFeedUrl,
			i._configuration.Authentication,
			nil,
			nil,
			userName)
		if err != nil {
			return nil, 0, nil, err
		}
	}
	return f, fi.Size(), done, nil
}

func (i *Install) InstalledPath() string {
	return i._targetDirectory
}

func (i *Install) formatTargetPath(info *packageInfo) string {
	if i.Type != PackageType_Plugin {
		//app，install current folder
		return ""
	}
	return filepath.Join(string(i._registry), info.group, info.name, info.version)
}

func (i *Install) readManifest(zip *zip.Reader) (*pkg.UniversalPackageMetadata, error) {
	for _, entry := range zip.File {
		if entry.Name == "upack.json" {
			r, err := entry.Open()
			if err != nil {
				return nil, err
			}
			return pkg.ReadManifest(r)
		}
	}
	return nil, nil
}
