package excelset

import (
	"excelset/rexcel"
	"testing"
)

func TestName(t *testing.T) {
	f := rexcel.NewFile()
	index := f.NewSheet("Sheet2")
	f.SetCellValue("Sheet2", "A2", "Hello world.")
	f.SetCellValue("Sheet1", "B2", 100)
	f.SetActiveSheet(index)
	if err := f.SaveAs("Book1.xlsx"); err != nil {
		t.Fatal(err)
	}
}
