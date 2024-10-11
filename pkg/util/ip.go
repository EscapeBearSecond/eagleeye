package util

import "codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/util"

func IsIP(str string) bool {
	return util.IsIP(str)
}

func IsIPv4(str string) bool {
	return util.IsIPv4(str)
}

func IsPort(str string) bool {
	return util.IsPort(str)
}

func IsHostPort(str string) bool {
	return util.IsHostPort(str)
}

func IsCIDR(str string) bool {
	return util.IsCIDR(str)
}

func IsIPRange(input string) bool {
	return util.IsIPRange(input)
}
