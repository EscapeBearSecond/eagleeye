package util

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func ReadXlsxAll(f *excelize.File) ([][]string, error) {
	var contents [][]string

	for _, sheetName := range f.GetSheetList() {
		sheetContents, err := f.GetRows(sheetName)
		if err != nil {
			return nil, fmt.Errorf("get [%s] rows failed: %w", sheetName, err)
		}

		contents = append(contents, sheetContents...)
	}

	return contents, nil
}

func XlsxRowsSize(f *excelize.File) (uint32, error) {
	var size uint32
	for _, sheetName := range f.GetSheetList() {
		rows, err := f.Rows(sheetName)
		if err != nil {
			return 0, fmt.Errorf("get [%s] rows failed: %w", sheetName, err)
		}

		for rows.Next() {
			size++
		}

		if err := rows.Close(); err != nil {
			return 0, fmt.Errorf("close [%s] rows failed: %w", sheetName, err)
		}
	}

	return size, nil
}
