package export

import (
	"github.com/samber/lo"
)

const (
	Positive = true
	Negative = false
	Header   = "header"
)

// styles 为excel中可能存在的pos和neg标记添加背景色
type styles map[any]int

// 格式key:value对
type styleItem struct {
	kind  any
	value int
}

func newStyleItem(kind any, value int) *styleItem {
	return &styleItem{
		kind:  kind,
		value: value,
	}
}

var (
	// 肯定状态
	positives = []string{"是", "T", "True", "t", "true", "TRUE"}
	// 否定状态
	negatives = []string{"否", "F", "False", "f", "false", "FALSE"}
)

// newStyles 实例化格式对象
func newStyles(items ...*styleItem) styles {
	m := map[any]int{}
	for _, item := range items {
		m[item.kind] = item.value
	}
	return m
}

// style 获取格式编号
func (ss styles) style(v any) int {
	switch t := v.(type) {
	case string:
		if lo.Contains(positives, t) {
			return ss[Positive]
		} else if lo.Contains(negatives, t) {
			return ss[Negative]
		} else {
			return ss[t]
		}
	default:
		return ss[t]
	}
}
