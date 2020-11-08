// 小助手函数，辅助业务逻辑的函数
package helper

import "regexp"

var (
	regAndroid = regexp.MustCompile("(a|A)ndroid|dr")
	regIOS     = regexp.MustCompile("i(p|P)(hone|ad|od)|(m|M)ac")
)

// 根据ua获取设备名称
func GetDeviceByUa(ua string) string {
	plat := "web"
	if regAndroid.MatchString(ua) {
		plat = "android"
	} else {
		if regIOS.MatchString(ua) {
			plat = "ios"
		}
	}

	return plat
}
