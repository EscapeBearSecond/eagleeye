package tvalidator

import (
	"io"
	"strings"
)

type memoryCatalog struct {
	templateContents map[string]string
}

func newCatalog(templateContents map[string]string) *memoryCatalog {
	return &memoryCatalog{templateContents: templateContents}
}

func (m *memoryCatalog) OpenFile(templateID string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(m.templateContents[templateID])), nil
}

func (m *memoryCatalog) GetTemplatePath(target string) ([]string, error) { return nil, nil }
func (m *memoryCatalog) GetTemplatesPath(definitions []string) ([]string, map[string]error) {
	return nil, nil
}
func (m *memoryCatalog) ResolvePath(templateName, second string) (string, error) { return "", nil }
