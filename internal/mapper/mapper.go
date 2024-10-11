package mapper

import "codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/pkg/types"

type Src struct {
	TemplateID string `yaml:"template_id"`
}

type Dest struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Severity    string `yaml:"severity"`
	Description string `yaml:"description"`
	Remediation string `yaml:"remediation"`
}

func (dest Dest) Assign(result *types.JobResultItem) *types.JobResultItem {
	result.TemplateID = dest.ID
	result.TemplateName = dest.Name
	result.Severity = dest.Severity
	result.Description = dest.Description
	result.Remediation = dest.Remediation
	return result
}
