package util

import "codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/target"

func SplitTargetsBySize(targets []string, size uint32) ([][]string, error) {
	return target.SplitBySize(targets, size)
}

func SplitTargetsN(targets []string, n int) ([][]string, error) {
	return target.SplitN(targets, n)
}
