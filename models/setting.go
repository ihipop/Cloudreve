package model

import (
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"net/url"
	"strconv"

	"github.com/cloudreve/Cloudreve/v3/pkg/cache"
	"github.com/jinzhu/gorm"
)

// Setting 系统设置模型
type Setting struct {
	gorm.Model
	Type  string `gorm:"not null"`
	Name  string `gorm:"unique;not null;index:setting_key"`
	Value string `gorm:"size:‎65535"`
}

// IsTrueVal 返回设置的值是否为真
func IsTrueVal(val string) bool {
	return val == "1" || val == "true"
}

// GetSettingByName 用 Name 获取设置值
func GetSettingByName(name string) string {
	var setting Setting

	// 优先从缓存中查找
	cacheKey := "setting_" + name
	if optionValue, ok := cache.Get(cacheKey); ok {
		return optionValue.(string)
	}
	// 尝试数据库中查找
	result := DB.Where("name = ?", name).First(&setting)
	if result.Error == nil {
		_ = cache.Set(cacheKey, setting.Value, -1)
		return setting.Value
	}
	return ""
}

// GetSettingByNames 用多个 Name 获取设置值
func GetSettingByNames(names ...string) map[string]string {
	var queryRes []Setting
	res, miss := cache.GetSettings(names, "setting_")

	if len(miss) > 0 {
		DB.Where("name IN (?)", miss).Find(&queryRes)
		for _, setting := range queryRes {
			res[setting.Name] = setting.Value
		}
	}

	_ = cache.SetSettings(res, "setting_")
	return res
}

// GetSettingByType 获取一个或多个分组的所有设置值
func GetSettingByType(types []string) map[string]string {
	var queryRes []Setting
	res := make(map[string]string)

	DB.Where("type IN (?)", types).Find(&queryRes)
	for _, setting := range queryRes {
		res[setting.Name] = setting.Value
	}

	return res
}

// GetSiteURL 获取站点地址
func GetSiteURL() *url.URL {
	base, err := url.Parse(GetSettingByName("siteURL"))
	if err != nil {
		base, _ = url.Parse("https://cloudreve.org")
	}
	return base
}

// GetSiteURL 获取请求内的站点地址  并以设置内的站点地址为默认配置
func GetContextURL(c *gin.Context) *url.URL {
	base := GetSiteURL();
	//根据访问的上下文获取完整路径而不是设置里面的站点路径 这样就能支持多域名
	//defaultConfig := location.DefaultConfig()
	location.New(location.Config{
		Host:   base.Host,
		Scheme: base.Scheme,
		Headers: location.Headers {
			Scheme: "X-Forwarded-Proto",
			Host:   "X-Forwarded-Host",//It's an old bug https://github.com/gin-contrib/location/pull/13/files
		},
	})(c)
	return location.Get(c)
}

// GetIntSetting 获取整形设置值，如果转换失败则返回默认值defaultVal
func GetIntSetting(key string, defaultVal int) int {
	res, err := strconv.Atoi(GetSettingByName(key))
	if err != nil {
		return defaultVal
	}
	return res
}
