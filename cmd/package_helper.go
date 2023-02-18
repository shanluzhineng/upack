package cmd

import (
	"errors"
	"strings"

	"github.com/abmpio/upack/pkg"
)

type PackageType string

func (t PackageType) IsValid() bool {
	return t == PackageType_App || t == PackageType_Plugin || t == PackageType_Tools
}

const (
	_defaultAppGroupName    = "App"
	_defaultPluginGroupName = "plugins"

	_defaultAppSourceFeedName = "app"

	PackageType_Plugin PackageType = "plugin"
	PackageType_App    PackageType = "app"
	PackageType_Tools  PackageType = "tools"

	//元数据属性名称之type
	_metaPropertyName_Type = "_type"
)

type packageInfo struct {
	group   string
	name    string
	version string
}

// / 根据<模块组/模块名@版本号>格式的字符串解析出模块id与版本信息
func parsePackageNameWithVersion(packageName string) (*packageInfo, error) {
	if len(packageName) <= 0 {
		return nil, errors.New("packageName不能为空")
	}
	versionIndex := strings.Index(packageName, "@")
	if versionIndex == -1 {
		return parseGroupAndName(packageName), nil
	}

	newPackage := parseGroupAndName(packageName[:versionIndex])
	newPackage.version = packageName[versionIndex+1:]
	return newPackage, nil
}

func parseGroupAndName(packageName string) *packageInfo {
	packageInfo := &packageInfo{}
	parts := strings.Split(strings.Replace(packageName, ":", "/", -1), "/")
	if len(parts) == 1 {
		packageInfo.name = parts[0]
		return packageInfo
	}
	packageInfo.group = strings.Join(parts[:len(parts)-1], "/")
	packageInfo.name = parts[len(parts)-1]
	return packageInfo
}

// find latest version in source feed
func getLatestVersion(source, group, name, version string, credentials *[2]string, prerelease bool) (latestVersion *pkg.UniversalPackageVersion, err error) {
	versionString, err := pkg.GetVersion(source, group, name, version, credentials, prerelease)
	if err != nil {
		return nil, err
	}
	latestVersion, err = pkg.ParseUniversalPackageVersion(versionString)
	if err != nil {
		return nil, err
	}
	return latestVersion, nil
}
