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
	"grader/internal/models"
)

type ExcelBuilder struct {
	xf             *excelize.File
	sheet          string
	dataStartRow   int
	maxR           int
	maxC           int
	nsColIdx       int
	nsColName      string
	descColIdx     int
	descColName    string
	fgColIdx       int
	fgColName      string
	flagColIdx     int
	flagColName    string
	totColIdx      int
	totColName     string
	questions      []string
	stylesCache    map[string]int
	defaultStyleID int
}

func (dm *DataManager) ExportFiles() {
	dm.writeCSVReport()
	dm.writeExcelReport()
}

func (dm *DataManager) writeCSVReport() {
	f, _ := os.Create(dm.CSVPath)
	defer f.Close()
	f.WriteString("\xef\xbb\xbf")
	writer := csv.NewWriter(f)
	dm.writeCSVHeader(writer)
	for _, s := range dm.Students {
		dm.writeCSVRow(writer, s)
	}
	writer.Flush()
}

func (dm *DataManager) writeCSVHeader(w *csv.Writer) {
	header := []string{"First Name", "Last Name", "Student ID"}
	header = append(header, dm.Questions...)
	header = append(header, "Total Score", "Description", "Not Submitted", "Fully Graded", "Flagged", "_Submitted")
	w.Write(header)
}

func (dm *DataManager) writeCSVRow(w *csv.Writer, s *models.Student) {
	row := []string{s.Name, s.Surname, s.ID}
	for _, q := range dm.Questions {
		if val, ok := s.Grades[q].(float64); ok {
			row = append(row, fmt.Sprintf("%.2f", val))
		} else {
			row = append(row, "")
		}
	}
	row = append(row, dm.studentTotalCSV(s))
	row = append(row, s.Description)
	row = append(row, fmt.Sprintf("%t", s.NotSubmitted))
	row = append(row, fmt.Sprintf("%t", s.FullyGraded))
	row = append(row, fmt.Sprintf("%t", s.Flagged))
	row = append(row, fmt.Sprintf("%t", s.IsSubmitted))
	w.Write(row)
}

func (dm *DataManager) studentTotalCSV(s *models.Student) string {
	if s.NotSubmitted {
		return "0.00"
	}
	if s.FullyGraded {
		return fmt.Sprintf("%.2f", s.TotalScore)
	}
	return ""
}

func (dm *DataManager) writeExcelReport() {
	b := newExcelBuilder(dm.Questions, len(dm.Students))
	b.initFile()
	b.writeQuestionHeaders()
	b.writeMetaHeaders()
	b.writeStatLabels()
	b.writeDataTableHeaders()
	b.writeQuestionStatFormulas()
	b.writeTotalStatFormulas()
	b.writeStudentRows(dm.Students)
	b.setColumnWidths()
	b.addCheckboxValidation()
	b.applyCellStyles()
	b.configurePanes()
	b.save(dm.ExcelPath)
	if len(dm.Students) > 0 {
		injectFlagHighlight(dm.ExcelPath, b.flagSqref(), b.flagFormula())
	}
}

func newExcelBuilder(questions []string, studentCount int) *ExcelBuilder {
	dataStartRow := 8
	maxR := dataStartRow + studentCount - 1
	if studentCount == 0 {
		maxR = dataStartRow
	}
	exHeaders := append(append([]string{"First Name", "Last Name", "Student ID"}, questions...),
		"Total Score", "Description", "Not Submitted", "Fully Graded", "Flagged")
	maxC := len(exHeaders)
	return &ExcelBuilder{
		sheet:        "Sheet1",
		dataStartRow: dataStartRow,
		maxR:         maxR,
		maxC:         maxC,
		nsColIdx:     maxC - 2,
		descColIdx:   maxC - 3,
		fgColIdx:     maxC - 1,
		flagColIdx:   maxC,
		totColIdx:    maxC - 4,
		questions:    questions,
		stylesCache:  make(map[string]int),
	}
}

func (b *ExcelBuilder) initFile() {
	b.xf = excelize.NewFile()
	b.xf.SetDefaultFont("Vazirmatn")
	b.nsColName, _ = excelize.ColumnNumberToName(b.nsColIdx)
	b.descColName, _ = excelize.ColumnNumberToName(b.descColIdx)
	b.fgColName, _ = excelize.ColumnNumberToName(b.fgColIdx)
	b.flagColName, _ = excelize.ColumnNumberToName(b.flagColIdx)
	b.totColName, _ = excelize.ColumnNumberToName(b.totColIdx)
}

func (b *ExcelBuilder) writeQuestionHeaders() {
	for i, q := range b.questions {
		c := i + 4
		colName, _ := excelize.ColumnNumberToName(c)
		b.xf.SetCellValue(b.sheet, fmt.Sprintf("%s1", colName), q)
	}
	b.xf.SetCellValue(b.sheet, b.totColName+"1", "Total Score")
	b.xf.SetCellValue(b.sheet, b.nsColName+"1", "Not Submitted")
	b.xf.SetCellValue(b.sheet, b.descColName+"1", "Description")
	b.xf.SetCellValue(b.sheet, b.fgColName+"1", "Fully Graded")
	b.xf.SetCellValue(b.sheet, b.flagColName+"1", "Flagged")
}

func (b *ExcelBuilder) writeMetaHeaders() {
	b.xf.MergeCell(b.sheet, "A2", "C2")
	b.xf.SetCellValue(b.sheet, "A2", "Graded Count")
	for i, label := range []string{"Mean", "Median", "Max", "Min"} {
		row := i + 3
		b.xf.MergeCell(b.sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row))
		b.xf.SetCellValue(b.sheet, fmt.Sprintf("A%d", row), label)
	}
}

func (b *ExcelBuilder) writeStatLabels() {}

func (b *ExcelBuilder) writeDataTableHeaders() {
	b.xf.SetCellValue(b.sheet, "A7", "First Name")
	b.xf.SetCellValue(b.sheet, "B7", "Last Name")
	b.xf.SetCellValue(b.sheet, "C7", "Student ID")
}

func (b *ExcelBuilder) writeQuestionStatFormulas() {
	for i := range b.questions {
		c := i + 4
		colName, _ := excelize.ColumnNumberToName(c)
		dr := b.dataRange(colName)
		b.xf.SetCellFormula(b.sheet, colName+"2", fmt.Sprintf("=COUNT(%s)", dr))
		b.xf.SetCellFormula(b.sheet, colName+"3", fmt.Sprintf("=IFERROR(AVERAGE(%s), 0)", dr))
		b.xf.SetCellFormula(b.sheet, colName+"4", fmt.Sprintf("=IFERROR(MEDIAN(%s), 0)", dr))
		b.xf.SetCellFormula(b.sheet, colName+"5", fmt.Sprintf("=IFERROR(MAX(%s), 0)", dr))
		b.xf.SetCellFormula(b.sheet, colName+"6", fmt.Sprintf("=IFERROR(MIN(%s), 0)", dr))
	}
}

func (b *ExcelBuilder) writeTotalStatFormulas() {
	if len(b.questions) == 0 {
		return
	}
	dr := b.dataRange(b.totColName)
	fgr := b.dataRange(b.fgColName)
	n := fmt.Sprintf(`SUMPRODUCT((%s="☑")*1)`, fgr)
	b.xf.SetCellFormula(b.sheet, b.totColName+"2", fmt.Sprintf(`=COUNTIF(%s, "☑")`, fgr))
	b.xf.SetCellFormula(b.sheet, b.totColName+"3", fmt.Sprintf(`=IFERROR(AVERAGEIF(%s, "☑", %s), 0)`, fgr, dr))
	b.xf.SetCellFormula(b.sheet, b.totColName+"4", fmt.Sprintf(
		`=IFERROR(IF(%[1]s=0, 0, (SUMPRODUCT(SMALL(IF(%[2]s="☑", %[3]s, 10^9), INT((%[1]s+1)/2))) + SUMPRODUCT(SMALL(IF(%[2]s="☑", %[3]s, 10^9), INT((%[1]s+2)/2)))) / 2), 0)`,
		n, fgr, dr))
	b.xf.SetCellFormula(b.sheet, b.totColName+"5", fmt.Sprintf(`=IFERROR(SUMPRODUCT(MAX((%s="☑")*%s)), 0)`, fgr, dr))
	b.xf.SetCellFormula(b.sheet, b.totColName+"6", fmt.Sprintf(`=IFERROR(IF(%s=0, 0, SUMPRODUCT(MIN(IF(%s="☑", %s, 10^9)))), 0)`, n, fgr, dr))
}

func (b *ExcelBuilder) dataRange(colName string) string {
	return fmt.Sprintf("%s%d:%s%d", colName, b.dataStartRow, colName, b.maxR)
}

func (b *ExcelBuilder) writeStudentRows(students []*models.Student) {
	for rIdx, s := range students {
		b.writeStudentRow(rIdx+b.dataStartRow, s)
	}
}

func (b *ExcelBuilder) writeStudentRow(row int, s *models.Student) {
	b.xf.SetCellValue(b.sheet, fmt.Sprintf("A%d", row), s.Name)
	b.xf.SetCellValue(b.sheet, fmt.Sprintf("B%d", row), s.Surname)
	b.xf.SetCellValue(b.sheet, fmt.Sprintf("C%d", row), s.ID)
	for i, q := range b.questions {
		c := i + 4
		colName, _ := excelize.ColumnNumberToName(c)
		cell := fmt.Sprintf("%s%d", colName, row)
		if s.NotSubmitted {
			b.xf.SetCellValue(b.sheet, cell, "")
		} else if val, ok := s.Grades[q].(float64); ok {
			b.xf.SetCellValue(b.sheet, cell, val)
		} else {
			b.xf.SetCellValue(b.sheet, cell, "")
		}
	}
	b.writeStudentFlags(row, s)
	b.writeStudentTotalFormula(row)
}

func (b *ExcelBuilder) writeStudentFlags(row int, s *models.Student) {
	b.xf.SetCellValue(b.sheet, fmt.Sprintf("%s%d", b.nsColName, row), checkSymbol(s.NotSubmitted))
	b.xf.SetCellValue(b.sheet, fmt.Sprintf("%s%d", b.descColName, row), s.Description)
	b.xf.SetCellValue(b.sheet, fmt.Sprintf("%s%d", b.fgColName, row), checkSymbol(s.FullyGraded))
	b.xf.SetCellValue(b.sheet, fmt.Sprintf("%s%d", b.flagColName, row), checkSymbol(s.Flagged))
}

func checkSymbol(checked bool) string {
	if checked {
		return "☑"
	}
	return "☐"
}

func (b *ExcelBuilder) writeStudentTotalFormula(row int) {
	if len(b.questions) == 0 {
		return
	}
	firstQ, _ := excelize.ColumnNumberToName(4)
	lastQ, _ := excelize.ColumnNumberToName(3 + len(b.questions))
	formula := fmt.Sprintf(`=IF(%s%d="☑", 0, IF(%s%d="☑", SUM(%s%d:%s%d), ""))`,
		b.nsColName, row, b.fgColName, row, firstQ, row, lastQ, row)
	b.xf.SetCellFormula(b.sheet, fmt.Sprintf("%s%d", b.totColName, row), formula)
}

func (b *ExcelBuilder) setColumnWidths() {
	b.xf.SetColWidth(b.sheet, "A", "C", 18)
	for c := 4; c < b.descColIdx; c++ {
		colName, _ := excelize.ColumnNumberToName(c)
		b.xf.SetColWidth(b.sheet, colName, colName, 12)
	}
	b.xf.SetColWidth(b.sheet, b.descColName, b.descColName, 50)
	b.xf.SetColWidth(b.sheet, b.nsColName, b.nsColName, 14)
	b.xf.SetColWidth(b.sheet, b.fgColName, b.fgColName, 14)
	b.xf.SetColWidth(b.sheet, b.flagColName, b.flagColName, 12)
}

func (b *ExcelBuilder) addCheckboxValidation() {
	dv := excelize.NewDataValidation(true)
	dv.SetSqref(fmt.Sprintf("%s%d:%s%d", b.nsColName, b.dataStartRow, b.flagColName, b.maxR))
	if err := dv.SetDropList([]string{"☐", "☑"}); err == nil {
		b.xf.AddDataValidation(b.sheet, dv)
	}
}

func (b *ExcelBuilder) applyCellStyles() {
	for r := 1; r <= b.maxR; r++ {
		for c := 1; c <= b.maxC; c++ {
			styleID := b.computeCellStyle(r, c)
			colName, _ := excelize.ColumnNumberToName(c)
			b.xf.SetCellStyle(b.sheet, fmt.Sprintf("%s%d", colName, r), fmt.Sprintf("%s%d", colName, r), styleID)
		}
	}
}

func (b *ExcelBuilder) computeCellStyle(row, col int) int {
	bold := row <= 7
	isDesc := col == b.descColIdx
	align := b.cellAlignment(row, isDesc)
	top, bottom, left, right := b.cellBorders(row, col)
	key := fmt.Sprintf("%v_%v_%d_%d_%d_%d", bold, isDesc, top, bottom, left, right)
	if sID, exists := b.stylesCache[key]; exists {
		return sID
	}
	sID := b.createStyle(bold, align, top, bottom, left, right)
	b.stylesCache[key] = sID
	return sID
}

func (b *ExcelBuilder) cellAlignment(row int, isDesc bool) *excelize.Alignment {
	if isDesc && row >= b.dataStartRow {
		return &excelize.Alignment{Horizontal: "right", Vertical: "top", WrapText: true}
	}
	return &excelize.Alignment{Horizontal: "center", Vertical: "center"}
}

func (b *ExcelBuilder) cellBorders(row, col int) (top, bottom, left, right int) {
	top, bottom, left, right = 1, 1, 1, 1
	switch row {
	case 1:
		top, bottom = 5, 5
	case 2:
		top = 5
	case 7:
		top, bottom = 5, 5
	case b.dataStartRow:
		top = 5
	case b.maxR:
		bottom = 5
	}
	switch col {
	case 1:
		left = 5
	case 3:
		right = 5
	case 4:
		left = 5
	case b.maxC:
		right = 5
	}
	return
}

func (b *ExcelBuilder) createStyle(bold bool, align *excelize.Alignment, top, bottom, left, right int) int {
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
	sID, _ := b.xf.NewStyle(&excelize.Style{
		Font:      fnt,
		Alignment: align,
		Border:    border,
	})
	return sID
}

func (b *ExcelBuilder) configurePanes() {
	b.xf.SetPanes(b.sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      3,
		YSplit:      1,
		TopLeftCell: "D2",
		ActivePane:  "bottomRight",
	})
	b.xf.SetSheetView(b.sheet, 0, &excelize.ViewOptions{RightToLeft: func(b bool) *bool { return &b }(false)})
}

func (b *ExcelBuilder) save(path string) {
	b.xf.SaveAs(path)
}

func (b *ExcelBuilder) flagSqref() string {
	return fmt.Sprintf("A%d:C%d", b.dataStartRow, b.maxR)
}

func (b *ExcelBuilder) flagFormula() string {
	return fmt.Sprintf(`$%s%d="☑"`, b.flagColName, b.dataStartRow)
}

func injectFlagHighlight(xlsxPath, sqref, formula string) error {
	raw, err := os.ReadFile(xlsxPath)
	if err != nil {
		return err
	}
	entries, order, err := unzipBytes(raw)
	if err != nil {
		return err
	}
	dxfID := injectDxfStyle(entries)
	injectConditionalRule(entries, sqref, formula, dxfID)
	return repackZip(entries, order, xlsxPath)
}

func unzipBytes(raw []byte) (map[string][]byte, []string, error) {
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, nil, err
	}
	entries := make(map[string][]byte)
	var order []string
	for _, zf := range zr.File {
		rc, err := zf.Open()
		if err != nil {
			return nil, nil, err
		}
		b, _ := io.ReadAll(rc)
		rc.Close()
		entries[zf.Name] = b
		order = append(order, zf.Name)
	}
	return entries, order, nil
}

func injectDxfStyle(entries map[string][]byte) int {
	dxf := `<dxf><fill><patternFill patternType="solid"><fgColor rgb="FFFFFF00"/><bgColor rgb="FFFFFF00"/></patternFill></fill></dxf>`
	styles := string(entries["xl/styles.xml"])
	if idx := strings.Index(styles, "<dxfs"); idx != -1 {
		rest := styles[idx:]
		end := strings.Index(rest, ">")
		tag := rest[:end+1]
		count := 0
		if ci := strings.Index(tag, `count="`); ci != -1 {
			fmt.Sscanf(tag[ci+7:], "%d", &count)
		}
		newTag := strings.Replace(tag, fmt.Sprintf(`count="%d"`, count), fmt.Sprintf(`count="%d"`, count+1), 1)
		if strings.HasSuffix(tag, "/>") {
			styles = strings.Replace(styles, tag, strings.TrimSuffix(newTag, "/>")+dxf+`</dxfs>`, 1)
		} else {
			styles = strings.Replace(styles, tag, newTag, 1)
			styles = strings.Replace(styles, `</dxfs>`, dxf+`</dxfs>`, 1)
		}
		entries["xl/styles.xml"] = []byte(styles)
		return count
	}
	styles = strings.Replace(styles, `</cellStyles>`, `</cellStyles><dxfs count="1">`+dxf+`</dxfs>`, 1)
	entries["xl/styles.xml"] = []byte(styles)
	return 0
}

func injectConditionalRule(entries map[string][]byte, sqref, formula string, dxfID int) {
	escFormula := strings.ReplaceAll(formula, `"`, "&quot;")
	cf := fmt.Sprintf(`<conditionalFormatting sqref="%s"><cfRule type="expression" priority="1" dxfId="%d"><formula>%s</formula></cfRule></conditionalFormatting>`,
		sqref, dxfID, escFormula)
	sheet := string(entries["xl/worksheets/sheet1.xml"])
	if idx := strings.Index(sheet, "<dataValidations"); idx != -1 {
		sheet = sheet[:idx] + cf + sheet[idx:]
	} else {
		sheet = strings.Replace(sheet, `</worksheet>`, cf+`</worksheet>`, 1)
	}
	entries["xl/worksheets/sheet1.xml"] = []byte(sheet)
}

func repackZip(entries map[string][]byte, order []string, destPath string) error {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, name := range order {
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
	return os.WriteFile(destPath, buf.Bytes(), 0644)
}