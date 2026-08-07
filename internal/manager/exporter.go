package manager

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

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
			row = append(row, "0.00")
		} else if s.FullyGraded {
			row = append(row, fmt.Sprintf("%.2f", s.TotalScore))
		} else {
			row = append(row, "")
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
	dataStartRow := 8
	maxR := dataStartRow + len(dm.Students) - 1
	if len(dm.Students) == 0 {
		maxR = dataStartRow
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

	xf.MergeCell(sheet, "A2", "C2")
	xf.SetCellValue(sheet, "A2", "Graded Count")
	stats := []string{"Mean", "Median", "Max", "Min"}
	for i, s := range stats {
		row := i + 3
		xf.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row))
		xf.SetCellValue(sheet, fmt.Sprintf("A%d", row), s)
	}

	xf.SetCellValue(sheet, "A7", "First Name")
	xf.SetCellValue(sheet, "B7", "Last Name")
	xf.SetCellValue(sheet, "C7", "Student ID")

	for i := range dm.Questions {
		c := i + 4
		colName, _ := excelize.ColumnNumberToName(c)
		dr := fmt.Sprintf("%s%d:%s%d", colName, dataStartRow, colName, maxR)
		xf.SetCellFormula(sheet, fmt.Sprintf("%s2", colName), fmt.Sprintf("=COUNT(%s)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s3", colName), fmt.Sprintf("=IFERROR(AVERAGE(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s4", colName), fmt.Sprintf("=IFERROR(MEDIAN(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s5", colName), fmt.Sprintf("=IFERROR(MAX(%s), 0)", dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s6", colName), fmt.Sprintf("=IFERROR(MIN(%s), 0)", dr))
	}
	if len(dm.Questions) > 0 {
		dr := fmt.Sprintf("%s%d:%s%d", totColName, dataStartRow, totColName, maxR)
		fgr := fmt.Sprintf("%s%d:%s%d", fgColName, dataStartRow, fgColName, maxR)
		n := fmt.Sprintf("SUMPRODUCT((%s=\"☑\")*1)", fgr)
		xf.SetCellFormula(sheet, fmt.Sprintf("%s2", totColName), fmt.Sprintf(`=COUNTIF(%s, "☑")`, fgr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s3", totColName), fmt.Sprintf(`=IFERROR(AVERAGEIF(%s, "☑", %s), 0)`, fgr, dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s4", totColName), fmt.Sprintf(`=IFERROR(IF(%[1]s=0, 0, (SUMPRODUCT(SMALL(IF(%[2]s="☑", %[3]s, 10^9), INT((%[1]s+1)/2))) + SUMPRODUCT(SMALL(IF(%[2]s="☑", %[3]s, 10^9), INT((%[1]s+2)/2)))) / 2), 0)`, n, fgr, dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s5", totColName), fmt.Sprintf(`=IFERROR(SUMPRODUCT(MAX((%s="☑")*%s)), 0)`, fgr, dr))
		xf.SetCellFormula(sheet, fmt.Sprintf("%s6", totColName), fmt.Sprintf(`=IFERROR(IF(%s=0, 0, SUMPRODUCT(MIN(IF(%s="☑", %s, 10^9)))), 0)`, n, fgr, dr))
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
			xf.SetCellValue(sheet, fmt.Sprintf("%s%d", nsColName, row), "☑")
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
			formula := fmt.Sprintf(`=IF(%s%d="☑", 0, IF(%s%d="☑", SUM(%s%d:%s%d), ""))`, nsColName, row, fgColName, row, firstQ, row, lastQ, row)
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
			bold := r <= 7
			isDesc := (c == descColIdx)
			align := &excelize.Alignment{Horizontal: "center", Vertical: "center"}
			if isDesc && r >= dataStartRow {
				align = &excelize.Alignment{Horizontal: "right", Vertical: "top", WrapText: true}
			}
			top, bottom, left, right := 1, 1, 1, 1
			if r == 1 {
				top, bottom = 5, 5
			}
			if r == 2 {
				top = 5
			}
			if r == 7 {
				top, bottom = 5, 5
			}
			if r == dataStartRow {
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
			key := fmt.Sprintf("%v_%v_%d_%d_%d_%d", bold, isDesc, top, bottom, left, right)
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

	if len(dm.Students) > 0 {
		sqref := fmt.Sprintf("A%d:C%d", dataStartRow, maxR)
		formula := fmt.Sprintf(`$%s%d="☑"`, flagColName, dataStartRow)
		injectFlagHighlight(dm.ExcelPath, sqref, formula)
	}
}

func injectFlagHighlight(xlsxPath, sqref, formula string) error {
	raw, err := os.ReadFile(xlsxPath)
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return err
	}
	entries := make(map[string][]byte)
	var names []string
	for _, zf := range zr.File {
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		b, _ := io.ReadAll(rc)
		rc.Close()
		entries[zf.Name] = b
		names = append(names, zf.Name)
	}

	dxf := `<dxf><fill><patternFill patternType="solid"><fgColor rgb="FFFFFF00"/><bgColor rgb="FFFFFF00"/></patternFill></fill></dxf>`
	dxfID := 0
	styles := string(entries["xl/styles.xml"])
	if idx := strings.Index(styles, "<dxfs"); idx != -1 {
		rest := styles[idx:]
		end := strings.Index(rest, ">")
		tag := rest[:end+1]
		count := 0
		if ci := strings.Index(tag, `count="`); ci != -1 {
			fmt.Sscanf(tag[ci+7:], "%d", &count)
		}
		dxfID = count
		newTag := strings.Replace(tag, fmt.Sprintf(`count="%d"`, count), fmt.Sprintf(`count="%d"`, count+1), 1)
		selfClosing := strings.HasSuffix(tag, "/>")
		if selfClosing {
			newTag = strings.TrimSuffix(newTag, "/>") + dxf + `</dxfs>`
			styles = strings.Replace(styles, tag, newTag, 1)
		} else {
			styles = strings.Replace(styles, tag, newTag, 1)
			styles = strings.Replace(styles, `</dxfs>`, dxf+`</dxfs>`, 1)
		}
	} else {
		styles = strings.Replace(styles, `<cellStyles`, `<dxfs count="1">`+dxf+`</dxfs><cellStyles`, 1)
	}
	entries["xl/styles.xml"] = []byte(styles)

	escFormula := strings.ReplaceAll(formula, `"`, "&quot;")
	cf := fmt.Sprintf(`<conditionalFormatting sqref="%s"><cfRule type="expression" priority="1" dxfId="%d"><formula>%s</formula></cfRule></conditionalFormatting>`, sqref, dxfID, escFormula)
	sheet := string(entries["xl/worksheets/sheet1.xml"])
	if idx := strings.Index(sheet, "<dataValidations"); idx != -1 {
		sheet = sheet[:idx] + cf + sheet[idx:]
	} else {
		sheet = strings.Replace(sheet, `</worksheet>`, cf+`</worksheet>`, 1)
	}
	entries["xl/worksheets/sheet1.xml"] = []byte(sheet)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, name := range names {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		if _, err := w.Write(entries[name]); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return os.WriteFile(xlsxPath, buf.Bytes(), 0644)
}