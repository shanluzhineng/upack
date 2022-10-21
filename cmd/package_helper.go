package cmd

import (
	"errors"
	"strings"
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
	newPackage.version = packageName[versionIndex:]
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
