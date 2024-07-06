package cmd

import (
	"fmt"
	"os"

	"github.com/shanluzhineng/upack/pkg"
)

type List struct {
	UserRegistry bool
}

func (*List) Name() string        { return "list" }
func (*List) Description() string { return "查看应用安装的所有模块." }

func (l *List) Help() string  { return pkg.DefaultCommandHelp(l) }
func (l *List) Usage() string { return pkg.DefaultCommandUsage(l) }

func (*List) PositionalArguments() []pkg.PositionalArgument {
	return nil
}
func (*List) ExtraArguments() []pkg.ExtraArgument {
	return nil
}

func (l *List) Run() int {
	r := pkg.PlugIns

	packages, err := r.ListInstalledPackages()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	for _, pkg := range packages {
		fmt.Println(pkg.GroupAndName() + " " + pkg.Version.String())
		if pkg.FeedURL != nil && *pkg.FeedURL != "" {
			fmt.Println("From", *pkg.FeedURL)
		}
		if (pkg.Path != nil && *pkg.Path != "") || pkg.InstallationDate != nil {
			path, date := "<unknown path>", "<unknown date>"
			if pkg.Path != nil && *pkg.Path != "" {
				path = *pkg.Path
			}
			if pkg.InstallationDate != nil {
				date = pkg.InstallationDate.Date.String()
			}
			fmt.Println("Installed to", path, "on", date)
		}
		if (pkg.InstalledBy != nil && *pkg.InstalledBy != "") || (pkg.InstalledUsing != nil && *pkg.InstalledUsing != "") {
			user, application := "<unknown user>", "<unknown application>"
			if pkg.InstalledBy != nil && *pkg.InstalledBy != "" {
				user = *pkg.InstalledBy
			}
			if pkg.InstalledUsing != nil && *pkg.InstalledUsing != "" {
				application = *pkg.InstalledUsing
			}
			fmt.Println("Installed by", user, "using", application)
		}
		if pkg.InstallationReason != nil && *pkg.InstallationReason != "" {
			fmt.Println("Comment:", *pkg.InstallationReason)
		}
		fmt.Println()
	}

	fmt.Println(len(packages), "packages")

	return 0
}
