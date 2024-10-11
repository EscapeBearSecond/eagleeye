package vuln

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVulnMapper(t *testing.T) {
	assert := assert.New(t)

	f, err := os.CreateTemp("", "vm.demo.yaml")
	assert.NoError(err)
	_, err = f.WriteString(`- template_id: SampleTemplateID
  vulnerabilities:
    - id: SampleActualID
      name: SampleName
      severity: critical
      description: SampleDescription
      remediation: SampleRemediation
      expressions:
        - ">7.5, <9.6"

    - id: SampleActualID2
      name: SampleName2
      severity: critical
      description: SampleDescription2
      remediation: SampleRemediation2
      expressions:
        - ">5.0"
        - "<3.2"`)
	assert.NoError(err)

	filename := f.Name()
	defer os.Remove(filename)
	f.Close()

	vm, err := New(filename)
	assert.Nil(err)

	r, err := vm.Get("SampleTemplateID").By("8.0")
	assert.Nil(err)
	assert.Len(r, 2)
	assert.Equal(r[0].ID, "SampleActualID")
	assert.Equal(r[1].ID, "SampleActualID2")

	r, err = vm.Get("SampleTemplateID").By("10.0")
	assert.Nil(err)
	assert.Len(r, 1)
	assert.Equal(r[0].ID, "SampleActualID2")

	r, err = vm.Get("SampleTemplateID").By("1.0")
	assert.Nil(err)
	assert.Len(r, 1)
	assert.Equal(r[0].ID, "SampleActualID2")

	_, err = vm.Get("SampleTemplateID").By("")
	assert.Error(err)
	assert.ErrorContains(err, "[] invalid versions format")
}
