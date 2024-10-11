package tpl

import (
	"strings"

	"github.com/projectdiscovery/nuclei/v3/pkg/templates"
	"github.com/projectdiscovery/nuclei/v3/pkg/types"
)

// POC nuclei的template
type POC struct {
	*templates.Template
}

type Result struct {
	Pocs               []*POC
	SkipHeadlessSize   int
	SkipHeadlessReason string
}

func (poc *POC) GetPorts() []string {
	var ports []string

	if len(poc.RequestsNetwork) > 0 {
		for _, request := range poc.RequestsNetwork {
			ports = append(ports, strings.Split(request.Port, ",")...)
		}
	} else if len(poc.RequestsJavascript) > 0 {
		for _, request := range poc.RequestsJavascript {
			for k, v := range request.Args {
				if strings.EqualFold(k, "Port") {
					ports = append(ports, types.ToString(v))
					break
				}
			}
		}
	}
	return ports
}

func (poc *POC) GetSchemes() []string {
	var schemes []string

	ss, ok := poc.Info.Metadata["schemes"]
	if !ok {
		return nil
	}

	switch v := ss.(type) {
	case []interface{}:
		for _, vv := range v {
			switch vv {
			case "http":
				schemes = append(schemes, "http")
			case "https":
				schemes = append(schemes, "https")
			}
		}
	default:
		return nil
	}

	return schemes
}
