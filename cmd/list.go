package cmd

import (
	"fmt"
	"os"

	"github.com/abmpio/upack/pkg"
)

type list struct {
	UserRegistry bool
}

func (*list) Name() string        { return "list" }
func (*list) Description() string { return "查看应用安装的所有模块." }

func (l *list) Help() string  { return pkg.DefaultCommandHelp(l) }
func (l *list) Usage() string { return pkg.DefaultCommandUsage(l) }

func (*list) PositionalArguments() []pkg.PositionalArgument {
	return nil
}
func (*list) ExtraArguments() []pkg.ExtraArgument {
	return nil
}

func (l *list) Run() int {
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
