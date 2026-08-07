package manager

import (
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

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
	dm.recalcFullyGraded()
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

func (dm *DataManager) findPDF(name, surname, id string) string {
	files, err := os.ReadDir(dm.PDFDir)
	if err != nil {
		return ""
	}
	firstName := utils.CleanString(name)
	lastName := utils.CleanString(surname)
	cleanID := utils.CleanString(id)
	var bestPath string
	var bestTime time.Time
	var idPath string
	for _, file := range files {
		if !strings.HasSuffix(strings.ToLower(file.Name()), ".pdf") {
			continue
		}
		cleanFile := utils.CleanString(file.Name())
		if match := dm.matchPDFByName(cleanFile, firstName, lastName); match {
			info, err := file.Info()
			if err == nil && (bestPath == "" || info.ModTime().After(bestTime)) {
				bestPath = filepath.Join(dm.PDFDir, file.Name())
				bestTime = info.ModTime()
			}
		}
		if idPath == "" && cleanID != "" && strings.Contains(cleanFile, cleanID) {
			idPath = filepath.Join(dm.PDFDir, file.Name())
		}
	}
	if bestPath != "" {
		return bestPath
	}
	return idPath
}

func (dm *DataManager) matchPDFByName(cleanFile, firstName, lastName string) bool {
	return firstName != "" && lastName != "" && strings.Contains(cleanFile, firstName) && strings.Contains(cleanFile, lastName)
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
	headers := dm.parseHeaders(records[0])
	idx := 0
	for i := 1; i < len(records); i++ {
		s := dm.parseStudentRow(records[i], headers, idx)
		if s == nil {
			continue
		}
		dm.Students = append(dm.Students, s)
		idx++
	}
	for i := range dm.Students {
		dm.Students[i].Index = i
	}
}

func (dm *DataManager) parseHeaders(row []string) map[string]int {
	headers := make(map[string]int)
	for i, h := range row {
		headers[utils.RemoveBOM(strings.TrimSpace(h))] = i
	}
	return headers
}

func (dm *DataManager) parseStudentRow(row []string, headers map[string]int, idx int) *models.Student {
	id := utils.CleanID(row[headers["نام کاربری"]])
	if id == "" || id == "nan" {
		return nil
	}
	name := strings.TrimSpace(row[headers["نام"]])
	surname := strings.TrimSpace(row[headers["نام خانوادگی"]])
	pdf := dm.findPDF(name, surname, id)
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
	return s
}

func (dm *DataManager) loadSavedData() {
	records := dm.loadFromCSV()
	if records == nil {
		records = dm.loadFromExcel()
	}
	if records == nil || len(records) < 2 {
		return
	}
	headers := dm.parseHeaders(records[0])
	if len(records[0]) >= 3 {
		headers["First Name"] = 0
		headers["Last Name"] = 1
		headers["Student ID"] = 2
	}
	for i := 1; i < len(records); i++ {
		dm.applySavedRow(records[i], headers)
	}
}

func (dm *DataManager) loadFromCSV() [][]string {
	if _, err := os.Stat(dm.CSVPath); err != nil {
		return nil
	}
	f, err := os.Open(dm.CSVPath)
	if err != nil {
		return nil
	}
	defer f.Close()
	reader := csv.NewReader(f)
	records, _ := reader.ReadAll()
	return records
}

func (dm *DataManager) loadFromExcel() [][]string {
	if _, err := os.Stat(dm.ExcelPath); err != nil {
		return nil
	}
	xf, err := excelize.OpenFile(dm.ExcelPath)
	if err != nil {
		return nil
	}
	defer xf.Close()
	sheet := xf.GetSheetName(0)
	records, _ := xf.GetRows(sheet)
	var filtered [][]string
	for _, row := range records {
		if len(row) == 0 {
			continue
		}
		firstCol := strings.TrimSpace(row[0])
		if dm.isStatRow(firstCol) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func (dm *DataManager) isStatRow(firstCol string) bool {
	stats := []string{"Mean", "Median", "Max", "Min", "First Name", "Graded Count"}
	for _, s := range stats {
		if firstCol == s {
			return true
		}
	}
	return false
}

func (dm *DataManager) applySavedRow(row []string, headers map[string]int) {
	idIdx, hasID := headers["Student ID"]
	if !hasID {
		return
	}
	id := utils.CleanID(row[idIdx])
	if id == "" || id == "nan" {
		return
	}
	s := dm.findStudent(id)
	if s == nil {
		return
	}
	if val, ok := headers["_Submitted"]; ok {
		s.IsSubmitted = strings.EqualFold(strings.TrimSpace(row[val]), "true")
	}
	if val, ok := headers["Not Submitted"]; ok {
		v := strings.TrimSpace(row[val])
		s.NotSubmitted = strings.EqualFold(v, "true") || v == "☑"
	}
	if val, ok := headers["Flagged"]; ok {
		v := strings.TrimSpace(row[val])
		s.Flagged = strings.EqualFold(v, "true") || v == "☑"
	}
	if val, ok := headers["Description"]; ok {
		s.Description = strings.TrimSpace(row[val])
	}
	if val, ok := headers["Total Score"]; ok {
		if t, err := strconv.ParseFloat(strings.TrimSpace(row[val]), 64); err == nil {
			s.TotalScore = t
		}
	}
	for _, q := range dm.Questions {
		if val, ok := headers[q]; ok {
			cleanVal := strings.TrimSpace(row[val])
			if cleanVal != "" {
				if g, err := strconv.ParseFloat(cleanVal, 64); err == nil {
					s.Grades[q] = g
				}
			}
		}
	}
}

func (dm *DataManager) findStudent(id string) *models.Student {
	for _, s := range dm.Students {
		if s.ID == id {
			return s
		}
	}
	return nil
}

func (dm *DataManager) recalcFullyGraded() {
	for _, s := range dm.Students {
		if s.NotSubmitted || len(dm.Questions) == 0 {
			s.FullyGraded = false
			continue
		}
		full := true
		for _, q := range dm.Questions {
			if _, ok := s.Grades[q].(float64); !ok {
				full = false
				break
			}
		}
		s.FullyGraded = full
	}
}

func (dm *DataManager) SaveGrade(id string, grades map[string]interface{}, total float64, comments string, notSub bool) {
	s := dm.findStudent(id)
	if s == nil {
		return
	}
	for _, q := range dm.Questions {
		s.Grades[q] = dm.parseGradeValue(grades[q])
	}
	s.TotalScore = total
	s.Description = comments
	s.NotSubmitted = notSub
	s.IsSubmitted = true
	dm.recalcFullyGraded()
	dm.ExportFiles()
}

func (dm *DataManager) parseGradeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		if val == "" {
			return ""
		}
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
		return ""
	default:
		return ""
	}
}

func (dm *DataManager) SetFlag(id string, flagged bool) {
	s := dm.findStudent(id)
	if s == nil {
		return
	}
	s.Flagged = flagged
	dm.ExportFiles()
}

func (dm *DataManager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	for _, q := range append(dm.Questions, "Total Score") {
		stats[q] = dm.calculateQuestionStats(q)
	}
	return map[string]interface{}{
		"total_students": len(dm.Students),
		"data":           stats,
	}
}

func (dm *DataManager) calculateQuestionStats(question string) map[string]float64 {
	vals, count := dm.collectQuestionValues(question)
	if len(vals) == 0 {
		return map[string]float64{
			"avg":   0,
			"med":   0,
			"max":   0,
			"min":   0,
			"count": count,
		}
	}
	sort.Float64s(vals)
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	avg := math.Round((sum/float64(len(vals)))*100) / 100
	maxVal := math.Round(vals[len(vals)-1]*100) / 100
	minVal := math.Round(vals[0]*100) / 100
	med := dm.calculateMedian(vals)
	return map[string]float64{
		"avg":   avg,
		"med":   med,
		"max":   maxVal,
		"min":   minVal,
		"count": count,
	}
}

func (dm *DataManager) collectQuestionValues(question string) ([]float64, float64) {
	var vals []float64
	count := 0.0
	for _, s := range dm.Students {
		if s.IsSubmitted && !s.NotSubmitted {
			if question == "Total Score" {
				if s.FullyGraded {
					vals = append(vals, s.TotalScore)
				}
			} else if val, ok := s.Grades[question].(float64); ok {
				vals = append(vals, val)
			}
		}
		if !s.NotSubmitted {
			if question == "Total Score" {
				if s.FullyGraded {
					count++
				}
			} else if _, ok := s.Grades[question].(float64); ok {
				count++
			}
		}
	}
	return vals, count
}

func (dm *DataManager) calculateMedian(vals []float64) float64 {
	mid := len(vals) / 2
	var med float64
	if len(vals)%2 == 0 {
		med = (vals[mid-1] + vals[mid]) / 2
	} else {
		med = vals[mid]
	}
	return math.Round(med*100) / 100
}