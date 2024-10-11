package export

import (
	"context"
	"encoding/csv"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCsvExporter(t *testing.T) {
	assert := assert.New(t)

	exporter, err := NewCsvExporter("abc")
	assert.NoError(err)

	err = exporter.Export(context.Background(), []any{"title1", "title2", "title3"})
	assert.NoError(err)

	err = exporter.Export(context.Background(), []any{"content1", "content2", "content3"})
	assert.NoError(err)

	exporter.Close()

	f, err := os.Open("./abc.csv")
	assert.NoError(err)

	reader := csv.NewReader(f)
	all, err := reader.ReadAll()
	assert.NoError(err)

	assert.Equal([][]string{
		{"title1", "title2", "title3"},
		{"content1", "content2", "content3"},
	}, all)

	os.Remove("./abc.csv")
}
