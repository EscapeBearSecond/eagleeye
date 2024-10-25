package util

import "github.com/EscapeBearSecond/falcon/internal/target"

func SplitTargetsBySize(targets []string, size uint32) ([][]string, error) {
	return target.SplitBySize(targets, size)
}

func SplitTargetsN(targets []string, n int) ([][]string, error) {
	return target.SplitN(targets, n)
}
