package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/abmpio/upack/pkg"
)

type pack struct {
	SourceDirectory string
	//是否自动push
	AutoPush bool
	ApiKey   string

	Metadata        pkg.UniversalPackageMetadata
	TargetDirectory string

	_configuration Configuration
}

func (*pack) Name() string { return "pack" }
func (*pack) Description() string {
	return "根据元数据生成一个plugin包."
}

func (p *pack) Help() string  { return pkg.DefaultCommandHelp(p) }
func (p *pack) Usage() string { return pkg.DefaultCommandUsage(p) }

func (*pack) PositionalArguments() []pkg.PositionalArgument {
	return nil
}

func (*pack) ExtraArguments() []pkg.ExtraArgument {
	return []pkg.ExtraArgument{
		{
			Name:        "source",
			Description: "包含了插件所有文件的目录.",
			TrySetValue: pkg.TrySetPathValue("source", func(cmd pkg.Command) *string {
				return &cmd.(*pack).SourceDirectory
			}),
		},
		{
			Name:        "name",
			Description: "plugin包名,如果不指定将使用文件夹的名称",
			TrySetValue: pkg.TrySetStringFnValue("name", func(cmd pkg.Command) func(string) {
				return (&cmd.(*pack).Metadata).SetName
			}),
		},
		{
			Name:        "push",
			Description: "是否自动push到仓库中",
			TrySetValue: pkg.TrySetBoolValue("push", func(cmd pkg.Command) *bool {
				return &cmd.(*pack).AutoPush
			}),
		},
		{
			Name:        "ver",
			Description: "应用版本号,如果未指定将自动决定版本号.仓库中不存在此包则版本号为1.0.0,如果已经有,则将版本号的minor部分加1,如已经存在了1.1.0的包,则新的包为1.2.0",
			TrySetValue: pkg.TrySetStringFnValue("version", func(cmd pkg.Command) func(string) {
				return (&cmd.(*pack).Metadata).SetVersion
			}),
		},
	}
}

func (p *pack) setupDefaultProperties() {
	p._configuration = defaultConfigurationWithFeedName(_defaultAppSourceFeedName)
	if p.TargetDirectory == "" {
		p.TargetDirectory, _ = os.Getwd()
	}
	if len(p.SourceDirectory) <= 0 {
		p.SourceDirectory, _ = os.Getwd()
	}
	p.Metadata.SetGroup(_defaultAppGroupName)
	if len(p.Metadata.Name()) <= 0 {
		currentPathName, err := os.Getwd()
		if err == nil {
			p.Metadata.SetName(filepath.Base(currentPathName))
		}
	}
	if len(p.Metadata.Version()) <= 0 {
		latestVersion, err := getLatestVersion(p._configuration.SourceFeedUrl,
			p.Metadata.Group(),
			p.Metadata.Name(),
			"",
			p._configuration.Authentication, false)
		if err == nil {
			latestVersion.Minor.SetInt64(latestVersion.Minor.Int64() + 1)
			p.Metadata.SetVersion(latestVersion.String())
		}
		if latestVersion == nil {
			p.Metadata.SetVersion("1.0.0")
		}
	}
}

func (p *pack) Run() int {
	p.setupDefaultProperties()
	info := &p.Metadata

	err := pkg.ValidateManifest(info)
	if err != nil {
		thing := "parameters:"
		fmt.Fprintln(os.Stderr, "Invalid", thing, err)
		return 2
	}

	pkg.PrintManifest(info)

	(*info)["createdDate"] = time.Now().UTC().Format(time.RFC3339)
	(*info)["createdUsing"] = "plugininstaller/" + pkg.Version
	currentUser, err := user.Current()
	if err == nil {
		(*info)["createdBy"] = currentUser.Name
	}

	fi, err := os.Stat(p.SourceDirectory)
	if os.IsNotExist(err) || (err == nil && !fi.IsDir()) {
		fmt.Fprintf(os.Stderr, "The source directory '%s' does not exist.\n", p.SourceDirectory)
		return 2
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	_, err = os.Stat(filepath.Join(p.SourceDirectory, info.Name()+"-"+info.BareVersion()+".upack"))
	if err == nil {
		fmt.Fprintln(os.Stderr, "Warning: output file already exists in source directory and may be included inadvertently in the package contents.")
	} else if !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	targetFileName := filepath.Join(p.TargetDirectory, info.Name()+"-"+info.BareVersion()+".upack")
	tmpFile, err := os.CreateTemp("", "upack")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if tmpFile != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	zipFile := zip.NewWriter(tmpFile)

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(&p.Metadata)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	err = pkg.CreateEntryFromStream(zipFile, &buf, "upack.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	err = pkg.AddDirectory(zipFile, p.SourceDirectory, "package/")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	err = zipFile.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	err = os.MkdirAll(filepath.Dir(targetFileName), 0755)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	// err = os.Remove(targetFileName)
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// 	return 1
	// }
	err = tmpFile.Close()
	tmpFile = nil
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if isWindows() {
		//windows,cp first,then remove
		_, err = copyFile(tmpPath, targetFileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		err = os.Remove(tmpPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	} else {
		err = os.Rename(tmpPath, targetFileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

	// fileName := pathfile. targetFileName
	if p.AutoPush {
		pushCmd := new(push)
		pushCmd.Package = filepath.Base(targetFileName)
		pushCmd.ApiKey = p.ApiKey
		pushCmd.SourceFeedName = _defaultAppSourceFeedName
		pushCmd.Type = PackageType_App
		return pushCmd.Run()
	}
	return 0
}
