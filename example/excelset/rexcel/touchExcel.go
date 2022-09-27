package rexcel

import "github.com/xuri/excelize/v2"

// NewFile 创建文件
func NewFile() *excelize.File {
	return excelize.NewFile()
}

// NewSheet 创建工作表
func NewSheet(f *excelize.File, sheetName string) int {
	return f.NewSheet(sheetName)
}
