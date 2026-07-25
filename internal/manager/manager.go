package manager

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"

	"grader/internal/models"
	"grader/internal/utils"
)

type DataManager struct {
	Students  []*models.Student
	Questions []string
	MaxGrades []float64
	OutputDir string
	ExcelPath string
	CSVPath   string
	PDFDir    string
	DataDir   string
}

func NewDataManager() (*DataManager, bool) {
	dm := &DataManager{
		OutputDir: "output",
		PDFDir:    "PDFs",
		DataDir:   "Data",
	}
	dm.ExcelPath = filepath.Join(dm.OutputDir, "grades.xlsx")
	dm.CSVPath = filepath.Join(dm.OutputDir, "grades.csv")

	os.MkdirAll(dm.OutputDir, 0755)
	os.MkdirAll(dm.PDFDir, 0755)
	os.MkdirAll(dm.DataDir, 0755)

	if dm.checkAndCreateTemplates() {
		return nil, true
	}

	dm.loadRubric()
	dm.loadStudents()
	dm.loadSavedData()
	dm.ExportFiles()
	return dm, false
}

func (dm *DataManager) checkAndCreateTemplates() bool {
	needsSetup := false
	studentsFile := filepath.Join(dm.DataDir, "students.csv")
	rubricFile := filepath.Join(dm.DataDir, "rubric.csv")

	if _, err := os.Stat(studentsFile); os.IsNotExist(err) {
		f, _ := os.Create(studentsFile)
		f.WriteString("\xef\xbb\xbfنام,نام خانوادگی,نام کاربری\n")
		f.Close()
		needsSetup = true
	}

	if _, err := os.Stat(rubricFile); os.IsNotExist(err) {
		f, _ := os.Create(rubricFile)
		f.WriteString("\xef\xbb\xbfQuestion,Max Grade\nسوال ۱,20\nسوال ۲,30\nسوال ۳,50\n")
		f.Close()
		needsSetup = true
	}

	if !needsSetup {
		sf, err := os.Open(studentsFile)
		if err == nil {
			reader := csv.NewReader(sf)
			records, _ := reader.ReadAll()
			sf.Close()
			if len(records) < 2 {
				needsSetup = true
			}
		}
	}

	return needsSetup
}

func (dm *DataManager) loadRubric() {
	f, err := os.Open(filepath.Join("Data", "rubric.csv"))
	if err != nil {
		return
	}
	defer f.Close()
	reader := csv.NewReader(f)
	records, _ := reader.ReadAll()

	for i, row := range records {
		if i == 0 {
			continue
		}
		if len(row) >= 2 {
			q := utils.RemoveBOM(strings.TrimSpace(row[0]))
			mg, _ := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
			dm.Questions = append(dm.Questions, q)
			dm.MaxGrades = append(dm.MaxGrades, mg)
		}
	}
}

func (dm *DataManager) findPDF(name, surname string) string {
	files, err := os.ReadDir(dm.PDFDir)
	if err != nil {
		return ""
	}
	target := utils.CleanString(name) + utils.CleanString(surname)
	targetRev := utils.CleanString(surname) + utils.CleanString(name)

	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file.Name()), ".pdf") {
			cleanFile := utils.CleanString(file.Name())
			if strings.Contains(cleanFile, target) || strings.Contains(cleanFile, targetRev) {
				return filepath.Join(dm.PDFDir, file.Name())
			}
		}
	}
	return ""
}

func (dm *DataManager) loadStudents() {
	f, err := os.Open(filepath.Join("Data", "students.csv"))
	if err != nil {
		return
	}
	defer f.Close()
	reader := csv.NewReader(f)
	records, _ := reader.ReadAll()
	if len(records) < 2 {
		return
	}

	headers := make(map[string]int)
	for i, h := range records[0] {
		headers[utils.RemoveBOM(h)] = i
	}

	idx := 0
	for i := 1; i < len(records); i++ {
		row := records[i]
		id := utils.CleanID(row[headers["نام کاربری"]])
		if id == "" || id == "nan" {
			continue
		}
		name := strings.TrimSpace(row[headers["نام"]])
		surname := strings.TrimSpace(row[headers["نام خانوادگی"]])
		pdf := dm.findPDF(name, surname)

		s := &models.Student{
			Index:        idx,
			ID:           id,
			Name:         name,
			Surname:      surname,
			HasPDF:       pdf != "",
			PDFPath:      pdf,
			Grades:       make(map[string]interface{}),
			IsSubmitted:  false,
			NotSubmitted: false,
		}
		for _, q := range dm.Questions {
			s.Grades[q] = ""
		}
		dm.Students = append(dm.Students, s)
		idx++
	}

	for i := range dm.Students {
		dm.Students[i].Index = i
	}
}

func (dm *DataManager) loadSavedData() {
	f, err := os.Open(dm.CSVPath)
	if err != nil {
		return
	}
	defer f.Close()
	reader := csv.NewReader(f)
	records, _ := reader.ReadAll()
	if len(records) < 2 {
		return
	}

	headers := make(map[string]int)
	for i, h := range records[0] {
		headers[h] = i
	}

	for i := 1; i < len(records); i++ {
		row := records[i]
		id := row[headers["Student ID"]]
		for _, s := range dm.Students {
			if s.ID == id {
				if val, ok := headers["_Submitted"]; ok && row[val] == "True" {
					s.IsSubmitted = true
				}
				if val, ok := headers["Not Submitted"]; ok && row[val] == "true" {
					s.NotSubmitted = true
				}
				if val, ok := headers["Description"]; ok {
					s.Description = row[val]
				}
				if val, ok := headers["Total Score"]; ok {
					if t, err := strconv.ParseFloat(row[val], 64); err == nil {
						s.TotalScore = t
					}
				}
				for _, q := range dm.Questions {
					if val, ok := headers[q]; ok {
						if row[val] != "" {
							if g, err := strconv.ParseFloat(row[val], 64); err == nil {
								s.Grades[q] = g
							}
						}
					}
				}
				break
			}
		}
	}
}

func (dm *DataManager) SaveGrade(id string, grades map[string]interface{}, total float64, comments string, notSub bool) {
	for _, s := range dm.Students {
		if s.ID == id {
			for k, v := range grades {
				if valStr, ok := v.(string); ok && valStr == "" {
					s.Grades[k] = ""
				} else if valFloat, ok := v.(float64); ok {
					s.Grades[k] = valFloat
				}
			}
			s.TotalScore = total
			s.Description = comments
			s.NotSubmitted = notSub
			s.IsSubmitted = true
			dm.ExportFiles()
			break
		}
	}
}

func (dm *DataManager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	for _, q := range append(dm.Questions, "Total Score") {
		var vals []float64
		for _, s := range dm.Students {
			if s.IsSubmitted && !s.NotSubmitted {
				if q == "Total Score" {
					vals = append(vals, s.TotalScore)
				} else if val, ok := s.Grades[q].(float64); ok {
					vals = append(vals, val)
				}
			}
		}
		if len(vals) > 0 {
			sort.Float64s(vals)
			sum := 0.0
			for _, v := range vals {
				sum += v
			}
			avg := math.Round((sum/float64(len(vals)))*100) / 100
			max := math.Round(vals[len(vals)-1]*100) / 100
			min := math.Round(vals[0]*100) / 100
			var med float64
			mid := len(vals) / 2
			if len(vals)%2 == 0 {
				med = (vals[mid-1] + vals[mid]) / 2
			} else {
				med = vals[mid]
			}
			med = math.Round(med*100) / 100

			stats[q] = map[string]float64{
				"avg": avg,
				"med": med,
				"max": max,
				"min": min,
			}
		} else {
			stats[q] = map[string]float64{
				"avg": 0,
				"med": 0,
				"max": 0,
				"min": 0,
			}
		}
	}
	return map[string]interface{}{
		"total_students": len(dm.Students),
		"data":           stats,
	}
}

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

	xf.SetDefaultFont("Vazirmatn")

	exHeaders := []string{"First Name", "Last Name", "Student ID"}
	exHeaders = append(exHeaders, dm.Questions...)
	exHeaders = append(exHeaders, "Total Score", "Not Submitted", "Description")

	maxC := len(exHeaders)
	maxR := 7 + len(dm.Students) - 1
	if len(dm.Students) == 0 {
		maxR = 7
	}
	descColIdx := maxC
	totalColIdx := 3 + len(dm.Questions) + 1
	nsColIdx := totalColIdx + 1

	xf.SetColWidth(sheet, "A", "C", 18)
	for i := 4; i <= 3+len(dm.Questions); i++ {
		colName, _ := excelize.ColumnNumberToName(i)
		xf.SetColWidth(sheet, colName, colName, 12)
	}
	totColName, _ := excelize.ColumnNumberToName(totalColIdx)
	nsColName, _ := excelize.ColumnNumberToName(nsColIdx)
	descColName, _ := excelize.ColumnNumberToName(descColIdx)

	xf.SetColWidth(sheet, totColName, totColName, 15)
	xf.SetColWidth(sheet, nsColName, nsColName, 15)
	xf.SetColWidth(sheet, descColName, descColName, 50)

	styleCache := make(map[string]int)
	getStyle := func(r, c int) int {
		bold := r <= 6
		isDesc := c == descColIdx

		align := &excelize.Alignment{Horizontal: "center", Vertical: "center"}
		if isDesc && r > 6 {
			align = &excelize.Alignment{Horizontal: "right", Vertical: "top", WrapText: true}
		}

		top, bottom, left, right := 1, 1, 1, 1

		if r == 1 {
			top, bottom = 2, 2
		}
		if r == 2 {
			top = 2
		}
		if r == 6 {
			top, bottom = 2, 2
		}
		if r == 7 {
			top = 2
		}
		if r == maxR {
			bottom = 2
		}

		if c == 1 {
			left = 2
		}
		if c == 3 {
			right = 2
		}
		if c == 4 {
			left = 2
		}
		if c == maxC {
			right = 2
		}

		key := fmt.Sprintf("%v_%v_%d_%d_%d_%d", bold, isDesc, top, bottom, left, right)
		if sID, ok := styleCache[key]; ok {
			return sID
		}

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

		sID, _ := xf.NewStyle(&excelize.Style{
			Font:      fnt,
			Alignment: align,
			Border:    border,
		})
		styleCache[key] = sID
		return sID
	}

	for r := 1; r <= maxR; r++ {
		for c := 1; c <= maxC; c++ {
			colName, _ := excelize.ColumnNumberToName(c)
			cell := fmt.Sprintf("%s%d", colName, r)
			xf.SetCellStyle(sheet, cell, cell, getStyle(r, c))
		}
	}

	for i, h := range exHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 6)
		xf.SetCellValue(sheet, cell, h)
	}

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

		xf.SetCellValue(sheet, fmt.Sprintf("%s%d", nsColName, row), s.NotSubmitted)
		xf.SetCellValue(sheet, fmt.Sprintf("%s%d", descColName, row), s.Description)

		if len(dm.Questions) > 0 {
			firstQ, _ := excelize.ColumnNumberToName(4)
			lastQ, _ := excelize.ColumnNumberToName(3 + len(dm.Questions))
			formula := fmt.Sprintf(`=IF(%s%d=TRUE, "", SUM(%s%d:%s%d))`, nsColName, row, firstQ, row, lastQ, row)
			xf.SetCellFormula(sheet, fmt.Sprintf("%s%d", totColName, row), formula)
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