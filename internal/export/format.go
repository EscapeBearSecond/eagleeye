package export

import "path/filepath"

func Format(a string) string {
	var format string

	extension := filepath.Ext(a)
	if extension == ".xlsx" {
		format = "excel"
	} else if extension == ".csv" {
		format = "csv"
	}

	return format
}
