package tpl

import (
	"testing"

	"github.com/projectdiscovery/nuclei/v3/pkg/model"
	"github.com/projectdiscovery/nuclei/v3/pkg/protocols/javascript"
	"github.com/projectdiscovery/nuclei/v3/pkg/protocols/network"
	"github.com/projectdiscovery/nuclei/v3/pkg/templates"
	"github.com/stretchr/testify/assert"
)

func TestGetPorts(t *testing.T) {
	assert := assert.New(t)

	poc := &POC{
		Template: &templates.Template{
			RequestsNetwork: []*network.Request{
				{
					Port: "80,90,100,101",
				},
			},
		},
	}

	ports := poc.GetPorts()
	assert.Equal([]string{"80", "90", "100", "101"}, ports)

	poc = &POC{
		Template: &templates.Template{
			RequestsJavascript: []*javascript.Request{
				{
					Args: map[string]interface{}{
						"Port": 80,
					},
				},
			},
		},
	}

	ports = poc.GetPorts()
	assert.Equal([]string{"80"}, ports)
}

func TestGetSchemes(t *testing.T) {
	assert := assert.New(t)

	poc := &POC{
		Template: &templates.Template{
			Info: model.Info{
				Metadata: map[string]interface{}{
					"schemes": []interface{}{"http", "https"},
				},
			},
		},
	}

	ret := poc.GetSchemes()
	assert.Equal([]string{"http", "https"}, ret)

	poc = &POC{
		Template: &templates.Template{
			Info: model.Info{
				Metadata: map[string]interface{}{
					"schemes": "https",
				},
			},
		},
	}

	ret = poc.GetSchemes()
	assert.Nil(ret)

	poc = &POC{
		Template: &templates.Template{
			Info: model.Info{
				Metadata: map[string]interface{}{},
			},
		},
	}

	ret = poc.GetSchemes()
	assert.Nil(ret)

	poc = &POC{
		Template: &templates.Template{
			Info: model.Info{
				Metadata: map[string]interface{}{
					"schemes": []interface{}{"https"},
				},
			},
		},
	}

	ret = poc.GetSchemes()
	assert.Equal([]string{"https"}, ret)

	poc = &POC{
		Template: &templates.Template{
			Info: model.Info{
				Metadata: map[string]interface{}{
					"data": "https",
				},
			},
		},
	}

	ret = poc.GetSchemes()
	assert.Nil(ret)

	poc = &POC{
		Template: &templates.Template{
			Info: model.Info{
				Metadata: map[string]interface{}{
					"schemes": []interface{}{"tcp"},
				},
			},
		},
	}

	ret = poc.GetSchemes()
	assert.Nil(ret)
	assert.Equal([]string{}, append([]string{}, ret...))
}
