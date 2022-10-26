package cmd

import (
	"archive/zip"
	"fmt"
	"net/http"
	"os"

	"github.com/abmpio/upack/pkg"
)

type Push struct {
	Package string

	SourceFeedName string
	Type           PackageType

	_configuration Configuration
}

func (*Push) Name() string { return "push" }
func (*Push) Description() string {
	return "发布一个模块包.upack文件到模块仓储中."
}

func (p *Push) Help() string  { return pkg.DefaultCommandHelp(p) }
func (p *Push) Usage() string { return pkg.DefaultCommandUsage(p) }

func (*Push) PositionalArguments() []pkg.PositionalArgument {
	return []pkg.PositionalArgument{
		{
			Name:        "package",
			Description: ".upack文件路径.",
			Index:       0,
			TrySetValue: pkg.TrySetStringValue("package", func(cmd pkg.Command) *string {
				return &cmd.(*Push).Package
			}),
		},
	}
}

func (*Push) ExtraArguments() []pkg.ExtraArgument {
	return nil
}

// 设置默认属性
func (i *Push) setupDefaultProperties() {
	if len(i.SourceFeedName) > 0 {
		i._configuration = defaultConfigurationWithFeedName(i.SourceFeedName)
	} else {
		i._configuration = *defaultConfiguration()
	}
	i.Type = PackageType_Plugin
}

func (p *Push) Run() int {
	p.setupDefaultProperties()

	packageStream, err := os.Open(p.Package)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer packageStream.Close()

	var info *pkg.UniversalPackageMetadata

	fi, err := packageStream.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	zipFile, err := zip.NewReader(packageStream, fi.Size())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	for _, entry := range zipFile.File {
		if entry.Name == "upack.json" {
			r, err := entry.Open()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}

			info, err = pkg.ReadManifest(r)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
			break
		}
	}

	if info == nil {
		fmt.Fprintln(os.Stderr, "upack.json missing from upack file!")
		return 1
	}

	err = pkg.ValidateManifest(info)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid upack.json:", err)
		return 2
	}

	pkg.PrintManifest(info)

	req, err := http.NewRequest("PUT", p._configuration.SourceFeedUrl, packageStream)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	req.Header.Set("Content-Length", fmt.Sprintf("%d", fi.Size()))
	req.Header.Set("Content-Type", "application/octet-stream")

	if p._configuration.Authentication != nil {
		req.SetBasicAuth(p._configuration.Authentication[0], p._configuration.Authentication[1])
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		fmt.Fprintln(os.Stderr, resp.Status)
		return 1
	}

	fmt.Println(info.GroupAndName(), info.Version(), "published!")

	return 0
}
