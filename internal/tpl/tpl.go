package tpl

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/EscapeBearSecond/eagleeye/pkg/types"
	"github.com/projectdiscovery/nuclei/v3/pkg/protocols"
	"github.com/projectdiscovery/nuclei/v3/pkg/templates"
)

// LoadWithLoader 通过遍历template文件加载template
func LoadWithFileWalk(template string, eOptions *protocols.ExecutorOptions, opts ...Option) (*Result, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	result := &Result{}
	err := filepath.WalkDir(template, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk dir failed: %w", err)
		}

		if d.IsDir() {
			return nil
		}

		template, err := templates.Parse(path, nil, *eOptions)
		if err != nil {
			return fmt.Errorf("template [%s] parse failed: %w", path, err)
		}

		// 如果浏览器为nil或者不是headless模式，则不加载headless模版
		if shouldSkipBrowser(template, eOptions, &o) {
			result.SkipHeadlessSize++
			return nil
		}
		// if shouldSkipInteractsh(template, eOptions, &o) {
		// 	return nil
		// }

		result.Pocs = append(result.Pocs, &POC{template})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("create engine failed: %w", err)
	}

	if result.SkipHeadlessSize > 0 {
		if eOptions.Browser == nil {
			result.SkipHeadlessReason = "browser not found"
		} else {
			result.SkipHeadlessReason = "headless disabled"
		}
	}

	return result, nil
}

// LoadWithFunc 通过函数加载template
func LoadWithFunc(fn types.GetTemplates, eOptions *protocols.ExecutorOptions, opts ...Option) (*Result, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	result := &Result{}
	for _, rt := range fn() {
		template, err := templates.ParseTemplateFromReader(strings.NewReader(rt.Original), nil, *eOptions)
		if err != nil {
			return nil, fmt.Errorf("template [%s] parse failed: %w", rt.ID, err)
		}

		if shouldSkipBrowser(template, eOptions, &o) {
			result.SkipHeadlessSize++
			continue
		}
		// if shouldSkipInteractsh(template, eOptions, &o) {
		// 	continue
		// }

		result.Pocs = append(result.Pocs, &POC{template})
	}

	if result.SkipHeadlessSize > 0 {
		if eOptions.Browser == nil {
			result.SkipHeadlessReason = "browser not found"
		} else {
			result.SkipHeadlessReason = "headless disabled"
		}
	}

	return result, nil
}

func shouldSkipBrowser(template *templates.Template, eOptions *protocols.ExecutorOptions, o *options) bool {
	return (eOptions.Browser == nil || !o.headless) && len(template.RequestsHeadless) > 0
}

// func shouldSkipInteractsh(template *templates.Template, _ *protocols.ExecutorOptions, _ *options) bool {
// 	var skip bool
// 	template.Variables.ForEach(func(key string, data interface{}) {
// 		if strings.Contains(ntypes.ToString(data), "interactsh-url") {
// 			skip = true
// 			return
// 		}
// 	})
// 	return skip
// }
