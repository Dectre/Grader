package manager

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

func (dm *DataManager) ExportFiles() {
	f, _ := os.Create(dm.CSVPath)
	defer f.Close()
	writer := csv.NewWriter(f)
	f.WriteString("\xef\xbb\xbf")

	header := []string{"First Name", "Last Name", "Student ID"}
	header = append(header, dm.Questions...)
	header = append(header, "Total Score", "Not Submitted", "Description", "_Submitted")
	writer.Write(header)

	for _, s := range dm.Students {
		row := []string{s.Name, s.Surname, s.ID}
		for _, q := range dm.Questions {
			if s.NotSubmitted {
				row = append(row, "")
			} else {
				if val, ok := s.Grades[q].(float64); ok {
					row = append(row, fmt.Sprintf("%.2f", val))
				} else {
					row = append(row, "")
				}
			}
		}
		if s.NotSubmitted {
			row = append(row, "")
		} else {
			row = append(row, fmt.Sprintf("%.2f", s.TotalScore))
		}
		row = append(row, fmt.Sprintf("%t", s.NotSubmitted))
		row = append(row, s.Description)
		row = append(row, fmt.Sprintf("%t", s.IsSubmitted))
		writer.Write(row)
	}
	writer.Flush()

	xf := excelize.NewFile()
	sheet := "Sheet1"

	vazirFont := &excelize.Font{Family: "Vazirmatn", Size: 11}
	vazirBold := &excelize.Font{Family: "Vazirmatn", Size: 11, Bold: true}
	centerAlign := &excelize.Alignment{Horizontal: "center", Vertical: "center"}
	descAlign := &excelize.Alignment{Horizontal: "right", Vertical: "top", WrapText: true}
	borderThin := []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
	}
	borderThickTop := []excelize.Border{
		{Type: "top", Color: "000000", Style: 2},
	}

	styleHeader, _ := xf.NewStyle(&excelize.Style{Font: vazirBold, Alignment: centerAlign, Border: borderThin})
	styleData, _ := xf.NewStyle(&excelize.Style{Font: vazirFont, Alignment: centerAlign, Border: borderThin})
	styleDesc, _ := xf.NewStyle(&excelize.Style{Font: vazirFont, Alignment: descAlign, Border: borderThin})
	styleThickTop, _ := xf.NewStyle(&excelize.Style{Border: borderThickTop})

	xf.SetColWidth(sheet, "A", "C", 18)

	colIdx := 4
	for range dm.Questions {
		colName, _ := excelize.ColumnNumberToName(colIdx)
		xf.SetColWidth(sheet, colName, colName, 12)
		colIdx++
	}
	totColName, _ := excelize.ColumnNumberToName(colIdx)
	xf.SetColWidth(sheet, totColName, totColName, 15)
	colIdx++
	nsColName, _ := excelize.ColumnNumberToName(colIdx)
	xf.SetColWidth(sheet, nsColName, nsColName, 15)
	colIdx++
	descColName, _ := excelize.ColumnNumberToName(colIdx)
	xf.SetColWidth(sheet, descColName, descColName, 50)

	exHeaders := []string{"First Name", "Last Name", "Student ID"}
	exHeaders = append(exHeaders, dm.Questions...)
	exHeaders = append(exHeaders, "Total Score", "Not Submitted", "Description")

	for i, h := range exHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 6)
		xf.SetCellValue(sheet, cell, h)
		xf.SetCellStyle(sheet, cell, cell, styleHeader)
	}

	stats := []string{"Mean", "Median", "Max", "Min"}
	for i, s := range stats {
		row := i + 2
		xf.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row))
		xf.SetCellValue(sheet, fmt.Sprintf("A%d", row), s)
		xf.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), styleHeader)
	}

	dataStartRow := 7
	dataEndRow := dataStartRow + len(dm.Students) - 1
	if len(dm.Students) == 0 {
		dataEndRow = dataStartRow
	}

	for i := range dm.Questions {
		c := i + 4
		colName, _ := excelize.ColumnNumberToName(c)
		dr := fmt.Sprintf("%s%d:%s%d", colName, dataStartRow, colName, dataEndRow)

		xf.SetCellFormula(sheet, fmt.Sprintf("%s2", colName), fmt.Sprintf("=IFERROR(AVERAGE(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s3", colName), fmt.Sprintf("=IFERROR(MEDIAN(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s4", colName), fmt.Sprintf("=IFERROR(MAX(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s5", colName), fmt.Sprintf("=IFERROR(MIN(%s), 0)", dr))

		xf.SetCellStyle(sheet, fmt.Sprintf("%s2", colName), fmt.Sprintf("%s5", colName), styleData)
	}

	if len(dm.Questions) > 0 {
		totColIdx := len(dm.Questions) + 4
		totCol, _ := excelize.ColumnNumberToName(totColIdx)
		dr := fmt.Sprintf("%s%d:%s%d", totCol, dataStartRow, totCol, dataEndRow)
		xf.SetCellFormula(sheet, fmt.Sprintf("%s2", totCol), fmt.Sprintf("=IFERROR(AVERAGE(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s3", totCol), fmt.Sprintf("=IFERROR(MEDIAN(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s4", totCol), fmt.Sprintf("=IFERROR(MAX(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s5", totCol), fmt.Sprintf("=IFERROR(MIN(%s), 0)", dr))
		xf.SetCellStyle(sheet, fmt.Sprintf("%s2", totCol), fmt.Sprintf("%s5", totCol), styleData)
	}

	for rIdx, s := range dm.Students {
		row := rIdx + dataStartRow
		xf.SetCellValue(sheet, fmt.Sprintf("A%d", row), s.Name)
		xf.SetCellValue(sheet, fmt.Sprintf("B%d", row), s.Surname)
		xf.SetCellValue(sheet, fmt.Sprintf("C%d", row), s.ID)

		for i, q := range dm.Questions {
			c := i + 4
			colName, _ := excelize.ColumnNumberToName(c)
			cell := fmt.Sprintf("%s%d", colName, row)
			if s.NotSubmitted {
				xf.SetCellValue(sheet, cell, "")
			} else {
				if val, ok := s.Grades[q].(float64); ok {
					xf.SetCellValue(sheet, cell, val)
				} else {
					xf.SetCellValue(sheet, cell, "")
				}
			}
		}

		totColIdx := len(dm.Questions) + 4
		totColName, _ := excelize.ColumnNumberToName(totColIdx)
		nsColName, _ := excelize.ColumnNumberToName(totColIdx + 1)
		descColName, _ := excelize.ColumnNumberToName(totColIdx + 2)

		xf.SetCellValue(sheet, fmt.Sprintf("%s%d", nsColName, row), s.NotSubmitted)
		xf.SetCellValue(sheet, fmt.Sprintf("%s%d", descColName, row), s.Description)

		if len(dm.Questions) > 0 {
			firstQ, _ := excelize.ColumnNumberToName(4)
			lastQ, _ := excelize.ColumnNumberToName(3 + len(dm.Questions))
			formula := fmt.Sprintf(`=IF(%s%d=TRUE, "", SUM(%s%d:%s%d))`, nsColName, row, firstQ, row, lastQ, row)
			xf.SetCellFormula(sheet, fmt.Sprintf("%s%d", totColName, row), formula)
		}

		for c := 1; c <= len(exHeaders); c++ {
			colName, _ := excelize.ColumnNumberToName(c)
			cell := fmt.Sprintf("%s%d", colName, row)
			if c == len(exHeaders) {
				xf.SetCellStyle(sheet, cell, cell, styleDesc)
			} else {
				xf.SetCellStyle(sheet, cell, cell, styleData)
			}
		}
	}

	xf.SetCellStyle(sheet, "A7", fmt.Sprintf("%s7", descColName), styleThickTop)

	xf.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      3,
		YSplit:      6,
		TopLeftCell: "D7",
		ActivePane:  "bottomRight",
	})
	xf.SetSheetView(sheet, 0, &excelize.ViewOptions{RightToLeft: func(b bool) *bool { return &b }(false)})

	xf.SaveAs(dm.ExcelPath)
}