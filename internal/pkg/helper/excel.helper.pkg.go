package helper

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/xuri/excelize/v2"
)

// CreateExcelFile generates an Excel file with given headers and data
func CreateExcelFile(filename string, sheetName string, headers []string, data [][]interface{}) (string, error) {
	// Create a new Excel file
	f := excelize.NewFile()

	// Create a rename sheet
	oldSheetName := f.GetSheetName(0) // Get the default sheet name
	f.SetSheetName(oldSheetName, sheetName)
	f.SetActiveSheet(0)

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i) // A1, B1, C1, ...
		f.SetCellValue(sheetName, cell, header)
	}

	// Write data
	for rowIndex, row := range data {
		for colIndex, value := range row {
			cell := fmt.Sprintf("%c%d", 'A'+colIndex, rowIndex+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// Save file
	err := f.SaveAs(filename)
	if err != nil {
		log.Printf("Failed to save file: %v", err)
		return "", err
	}

	return filename, nil
}

func GetData(url string) ([]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	return io.ReadAll(r.Body)
}
