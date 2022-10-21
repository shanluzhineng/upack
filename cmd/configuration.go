package cmd

import (
	"encoding/json"
	"os"
	"path"
	"strings"

	"github.com/abmpio/upack/pkg"
)

const (
	ConfigurationKey = "plugininstaller"

	_envKeySourceUrl string = ConfigurationKey + ".sourceUrl"
	_envKeyFeedName  string = ConfigurationKey + ".feedName"
	_envKeyApiKey    string = ConfigurationKey + ".apiKey"
)

var (
	_configuration *Configuration = defaultConfiguration()
)

func getConfigKey(key string) string {
	return strings.ToLower(key)
}

type Configuration struct {
	// 模块仓储的授权信息
	Authentication *[2]string
	// 模块仓储api url
	SourceFeedUrl string
	// 应用模块目录，绝对路径
	AppPackageRegistryPath string
	// 获取应用包注册对象
	AppPackageRegistry pkg.Registry
	/// 获取不带feed名称的endpoint url
	SourceUrl string
	// 获取feed名称
	SourceFeedName string
}

func defaultConfiguration() *Configuration {
	config := &Configuration{}
	//先读取环境变量
	config.SetSourceFeedUrl(getEnvKey(_envKeySourceUrl), getEnvKey(_envKeyFeedName))
	config.SetAppPackageRegistryPath("plugins")
	config.Authentication = getAuthentication(getEnvKey(_envKeyApiKey))

	m := make(map[string]interface{})
	data, err := readJsonFile(getCurrentDirectory() + "/plugininstaller.json")
	if err != nil {
		return config
	}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return config
	}
	insensitiviseMap(m)
	config.readFromConfig(m)

	return config
}

// 从key/value配置中读取配置信息
func (c *Configuration) readFromConfig(properties map[string]interface{}) {
	apiKey, ok := properties[getConfigKey("apiKey")].(string)
	if ok && len(apiKey) > 0 {
		if c.Authentication == nil {
			c.Authentication = &[2]string{}
		}
		c.Authentication[0] = "api"
		c.Authentication[1] = apiKey
	}

	//url
	sourceUrl, _ := properties[getConfigKey("sourceUrl")].(string)
	sourceFeedName, _ := properties[getConfigKey("sourceFeedName")].(string)
	if len(sourceUrl) > 0 || len(sourceFeedName) > 0 {
		c.SetSourceFeedUrl(sourceUrl, sourceFeedName)
	}
}

func (c *Configuration) SetAppPackageRegistryPath(relativePath string) {
	if len(relativePath) <= 0 {
		return
	}
	c.AppPackageRegistryPath = path.Join(getCurrentDirectory(), "plugins")
	c.AppPackageRegistry = pkg.Registry(c.AppPackageRegistryPath)
}

// 设置sourceUrl与feedName
func (c *Configuration) SetSourceFeedUrl(sourceUrl string, sourceFeedName string) {
	c.SourceFeedUrl = sourceUrl
	c.SourceFeedName = sourceFeedName
	c.SourceFeedUrl = getSourceFeedUrl(sourceUrl, sourceFeedName)
}

func getSourceFeedUrl(sourceUrl string, sourceFeedName string) string {
	if len(sourceUrl) <= 0 || len(sourceFeedName) <= 0 {
		return ""
	}
	if !strings.HasSuffix(sourceUrl, "/") {
		sourceUrl = sourceUrl + "/upack/"
	}
	return strings.Join([]string{sourceUrl, sourceFeedName, "/"}, "")
}

func getEnvKey(key string) string {
	return os.Getenv(key)
}

func getAuthentication(value string) *[2]string {
	if len(value) <= 0 {
		return nil
	}
	return &[2]string{"api", value}
}
