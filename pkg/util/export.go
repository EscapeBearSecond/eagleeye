package util

import "github.com/EscapeBearSecond/falcon/internal/export"

func ExportFormat(filename string) string {
	return export.Format(filename)
}
