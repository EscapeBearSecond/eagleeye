package vuln

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/EscapeBearSecond/falcon/internal/mapper"
	"github.com/hashicorp/go-version"
	"gopkg.in/yaml.v3"
)

type Mapper struct {
	core *sync.Map
}

type Vulnerabilities []Vulnerability

type Mapping struct {
	mapper.Src      `yaml:",inline"`
	Vulnerabilities Vulnerabilities `yaml:"vulnerabilities"`
}

type Vulnerability struct {
	mapper.Dest `yaml:",inline"`
	Expressions []string `yaml:"expressions"`
}

func New(filename string) (*Mapper, error) {
	vm := &Mapper{}
	if filename != "" {
		fBytes, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("read version vulns file failed: %w", err)
		}

		var mappings []Mapping
		err = yaml.Unmarshal(fBytes, &mappings)
		if err != nil {
			return nil, fmt.Errorf("unmarshal version vulns file failed: %w", err)
		}

		syncMap := &sync.Map{}
		for i := range mappings {
			syncMap.Store(mappings[i].TemplateID, mappings[i].Vulnerabilities)
		}

		vm.core = syncMap
	}
	return vm, nil
}

func (vm *Mapper) Get(templateID string) Vulnerabilities {
	if vm.core == nil {
		return nil
	}

	v, ok := vm.core.Load(templateID)
	if !ok {
		return nil
	}

	return v.(Vulnerabilities)
}

func (vulns Vulnerabilities) By(vers ...string) ([]mapper.Dest, error) {
	if len(vulns) == 0 {
		// 此处不需要返回错误
		// vulns为nil的情况只有两种
		// 原因1：如果是vm的core为nil，表示映射文件不存在，自然不用映射
		// 原因2：如果是未加载出对应template id的漏洞，表示对应id的模板没有映射漏洞，也不用映射
		return nil, nil
	}

	if len(vers) == 0 {
		return nil, errors.New("input versions are not specified")
	}

	var (
		ver *version.Version
		err error
	)
	for _, v := range vers {
		ver, err = version.NewVersion(v)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("%v invalid versions format: %w", vers, err)
	}

	var results []mapper.Dest

	// 遍历当前模板对应的所有漏洞
	for i := range vulns {
		// 遍历漏洞版本判断表达式，判断对应版本是否满足
		// 如果满足则添加到结果中，并跳出循环（只要有一个表达式满足即可）
		for _, expression := range vulns[i].Expressions {
			constraints, err := version.NewConstraint(expression)
			if err != nil {
				return nil, fmt.Errorf("invalid constraint %s format: %w", expression, err)
			}

			if constraints.Check(ver) {
				results = append(results, vulns[i].Dest)
				break
			}
		}
	}
	return results, nil
}
