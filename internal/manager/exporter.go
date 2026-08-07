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
	header = append(header, "Total Score", "Description", "Not Submitted", "Fully Graded", "Flagged", "_Submitted")
	writer.Write(header)
	for _, s := range dm.Students {
		row := []string{s.Name, s.Surname, s.ID}
		for _, q := range dm.Questions {
			if val, ok := s.Grades[q].(float64); ok {
				row = append(row, fmt.Sprintf("%.2f", val))
			} else {
				row = append(row, "")
			}
		}
		if s.NotSubmitted {
			row = append(row, "")
		} else {
			row = append(row, fmt.Sprintf("%.2f", s.TotalScore))
		}
		row = append(row, s.Description)
		row = append(row, fmt.Sprintf("%t", s.NotSubmitted))
		row = append(row, fmt.Sprintf("%t", s.FullyGraded))
		row = append(row, fmt.Sprintf("%t", s.Flagged))
		row = append(row, fmt.Sprintf("%t", s.IsSubmitted))
		writer.Write(row)
	}
	writer.Flush()

	xf := excelize.NewFile()
	sheet := "Sheet1"
	xf.SetDefaultFont("Vazirmatn")

	exHeaders := []string{"First Name", "Last Name", "Student ID"}
	exHeaders = append(exHeaders, dm.Questions...)
	exHeaders = append(exHeaders, "Total Score", "Description", "Not Submitted", "Fully Graded", "Flagged")
	maxC := len(exHeaders)
	maxR := 7 + len(dm.Students) - 1
	if len(dm.Students) == 0 {
		maxR = 7
	}

	flagColIdx := maxC
	fgColIdx := maxC - 1
	nsColIdx := maxC - 2
	descColIdx := maxC - 3
	totColIdx := maxC - 4

	for i, q := range dm.Questions {
		c := i + 4
		colName, _ := excelize.ColumnNumberToName(c)
		xf.SetCellValue(sheet, fmt.Sprintf("%s1", colName), q)
	}
	totColName, _ := excelize.ColumnNumberToName(totColIdx)
	xf.SetCellValue(sheet, fmt.Sprintf("%s1", totColName), "Total Score")
	nsColName, _ := excelize.ColumnNumberToName(nsColIdx)
	xf.SetCellValue(sheet, fmt.Sprintf("%s1", nsColName), "Not Submitted")
	descColName, _ := excelize.ColumnNumberToName(descColIdx)
	xf.SetCellValue(sheet, fmt.Sprintf("%s1", descColName), "Description")
	fgColName, _ := excelize.ColumnNumberToName(fgColIdx)
	xf.SetCellValue(sheet, fmt.Sprintf("%s1", fgColName), "Fully Graded")
	flagColName, _ := excelize.ColumnNumberToName(flagColIdx)
	xf.SetCellValue(sheet, fmt.Sprintf("%s1", flagColName), "Flagged")

	xf.SetCellValue(sheet, "A6", "First Name")
	xf.SetCellValue(sheet, "B6", "Last Name")
	xf.SetCellValue(sheet, "C6", "Student ID")

	stats := []string{"Mean", "Median", "Max", "Min"}
	for i, s := range stats {
		row := i + 2
		xf.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row))
		xf.SetCellValue(sheet, fmt.Sprintf("A%d", row), s)
	}

	dataStartRow := 7
	dataEndRow := maxR

	for i := range dm.Questions {
		c := i + 4
		colName, _ := excelize.ColumnNumberToName(c)
		dr := fmt.Sprintf("%s%d:%s%d", colName, dataStartRow, colName, dataEndRow)
		xf.SetCellFormula(sheet, fmt.Sprintf("%s2", colName), fmt.Sprintf("=IFERROR(AVERAGE(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s3", colName), fmt.Sprintf("=IFERROR(MEDIAN(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s4", colName), fmt.Sprintf("=IFERROR(MAX(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s5", colName), fmt.Sprintf("=IFERROR(MIN(%s), 0)", dr))
	}
	if len(dm.Questions) > 0 {
		dr := fmt.Sprintf("%s%d:%s%d", totColName, dataStartRow, totColName, dataEndRow)
		xf.SetCellFormula(sheet, fmt.Sprintf("%s2", totColName), fmt.Sprintf("=IFERROR(AVERAGE(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s3", totColName), fmt.Sprintf("=IFERROR(MEDIAN(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s4", totColName), fmt.Sprintf("=IFERROR(MAX(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s5", totColName), fmt.Sprintf("=IFERROR(MIN(%s), 0)", dr))
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
		if s.NotSubmitted {
			xf.SetCellValue(sheet, fmt.Sprintf("%s%d", nsColName, row), "☐")
		} else {
			xf.SetCellValue(sheet, fmt.Sprintf("%s%d", nsColName, row), "☐")
		}
		xf.SetCellValue(sheet, fmt.Sprintf("%s%d", descColName, row), s.Description)
		if s.FullyGraded {
			xf.SetCellValue(sheet, fmt.Sprintf("%s%d", fgColName, row), "☑")
		} else {
			xf.SetCellValue(sheet, fmt.Sprintf("%s%d", fgColName, row), "☐")
		}
		if s.Flagged {
			xf.SetCellValue(sheet, fmt.Sprintf("%s%d", flagColName, row), "☑")
		} else {
			xf.SetCellValue(sheet, fmt.Sprintf("%s%d", flagColName, row), "☐")
		}
		if len(dm.Questions) > 0 {
			firstQ, _ := excelize.ColumnNumberToName(4)
			lastQ, _ := excelize.ColumnNumberToName(3 + len(dm.Questions))
			formula := fmt.Sprintf(`=IF(%s%d="☑", "", SUM(%s%d:%s%d))`, nsColName, row, firstQ, row, lastQ, row)
			xf.SetCellFormula(sheet, fmt.Sprintf("%s%d", totColName, row), formula)
		}
	}

	xf.SetColWidth(sheet, "A", "C", 18)
	for c := 4; c < descColIdx; c++ {
		colName, _ := excelize.ColumnNumberToName(c)
		xf.SetColWidth(sheet, colName, colName, 12)
	}
	xf.SetColWidth(sheet, descColName, descColName, 50)
	xf.SetColWidth(sheet, nsColName, nsColName, 14)
	xf.SetColWidth(sheet, fgColName, fgColName, 14)
	xf.SetColWidth(sheet, flagColName, flagColName, 12)

	if len(dm.Students) > 0 {
		dv := excelize.NewDataValidation(true)
		dv.SetSqref(fmt.Sprintf("%s%d:%s%d", nsColName, dataStartRow, flagColName, maxR))
		if err := dv.SetDropList([]string{"☐", "☑"}); err == nil {
			xf.AddDataValidation(sheet, dv)
		}
	}

	styleCache := make(map[string]int)
	for r := 1; r <= maxR; r++ {
		for c := 1; c <= maxC; c++ {
			bold := r <= 6
			isDesc := (c == descColIdx)
			isYellow := false
			if r >= dataStartRow && r-dataStartRow < len(dm.Students) && c <= 3 {
				isYellow = dm.Students[r-dataStartRow].Flagged
			}
			align := &excelize.Alignment{Horizontal: "center", Vertical: "center"}
			if isDesc && r > 6 {
				align = &excelize.Alignment{Horizontal: "right", Vertical: "top", WrapText: true}
			}
			top, bottom, left, right := 1, 1, 1, 1
			if r == 1 {
				top, bottom = 5, 5
			}
			if r == 2 {
				top = 5
			}
			if r == 6 {
				top, bottom = 5, 5
			}
			if r == 7 {
				top = 5
			}
			if r == maxR {
				bottom = 5
			}
			if c == 1 {
				left = 5
			}
			if c == 3 {
				right = 5
			}
			if c == 4 {
				left = 5
			}
			if c == maxC {
				right = 5
			}
			key := fmt.Sprintf("%v_%v_%v_%d_%d_%d_%d", bold, isDesc, isYellow, top, bottom, left, right)
			sID, exists := styleCache[key]
			if !exists {
				border := []excelize.Border{
					{Type: "top", Color: "000000", Style: top},
					{Type: "bottom", Color: "000000", Style: bottom},
					{Type: "left", Color: "000000", Style: left},
					{Type: "right", Color: "000000", Style: right},
				}
				fnt := &excelize.Font{Family: "Vazirmatn", Size: 11}
				if bold {
					fnt.Bold = true
				}
				cellStyle := &excelize.Style{
					Font:      fnt,
					Alignment: align,
					Border:    border,
				}
				if isYellow {
					cellStyle.Fill = excelize.Fill{Type: "pattern", Color: []string{"FFFF00"}, Pattern: 1}
				}
				sID, _ = xf.NewStyle(cellStyle)
				styleCache[key] = sID
			}
			colName, _ := excelize.ColumnNumberToName(c)
			cell := fmt.Sprintf("%s%d", colName, r)
			xf.SetCellStyle(sheet, cell, cell, sID)
		}
	}

	xf.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      3,
		YSplit:      1,
		TopLeftCell: "D2",
		ActivePane:  "bottomRight",
	})
	xf.SetSheetView(sheet, 0, &excelize.ViewOptions{RightToLeft: func(b bool) *bool { return &b }(false)})
	xf.SaveAs(dm.ExcelPath)
}