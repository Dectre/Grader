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
		if strings.HasSuffix(strings.ToLower(file.Name()), ".pdf") {
			cleanFile := utils.CleanString(file.Name())
			if firstName != "" && lastName != "" && strings.Contains(cleanFile, firstName) && strings.Contains(cleanFile, lastName) {
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
	}
	if bestPath != "" {
		return bestPath
	}
	return idPath
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
		headers[utils.RemoveBOM(strings.TrimSpace(h))] = i
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
		dm.Students = append(dm.Students, s)
		idx++
	}
	for i := range dm.Students {
		dm.Students[i].Index = i
	}
}

func (dm *DataManager) loadSavedData() {
	var records [][]string
	if _, err := os.Stat(dm.CSVPath); err == nil {
		f, err := os.Open(dm.CSVPath)
		if err != nil {
			return
		}
		reader := csv.NewReader(f)
		records, _ = reader.ReadAll()
		f.Close()
	} else if _, err := os.Stat(dm.ExcelPath); err == nil {
		xf, err := excelize.OpenFile(dm.ExcelPath)
		if err != nil {
			return
		}
		sheet := xf.GetSheetName(0)
		records, _ = xf.GetRows(sheet)
		xf.Close()
		var filtered [][]string
		for _, row := range records {
			if len(row) > 0 {
				firstCol := strings.TrimSpace(row[0])
				if firstCol != "Mean" && firstCol != "Median" && firstCol != "Max" && firstCol != "Min" && firstCol != "First Name" {
					filtered = append(filtered, row)
				}
			}
		}
		records = filtered
	} else {
		return
	}
	if len(records) < 2 {
		return
	}
	headers := make(map[string]int)
	for i, h := range records[0] {
		headers[utils.RemoveBOM(strings.TrimSpace(h))] = i
	}
	if len(records[0]) >= 3 {
		headers["First Name"] = 0
		headers["Last Name"] = 1
		headers["Student ID"] = 2
	}
	for i := 1; i < len(records); i++ {
		row := records[i]
		idIdx, hasID := headers["Student ID"]
		if !hasID {
			continue
		}
		id := utils.CleanID(row[idIdx])
		if id == "" || id == "nan" {
			continue
		}
		for _, s := range dm.Students {
			if s.ID == id {
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
				break
			}
		}
	}
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
	for _, s := range dm.Students {
		if s.ID == id {
			for _, q := range dm.Questions {
				v, exists := grades[q]
				if !exists {
					s.Grades[q] = ""
					continue
				}
				switch val := v.(type) {
				case float64:
					s.Grades[q] = val
				case string:
					if val == "" {
						s.Grades[q] = ""
					} else if f, err := strconv.ParseFloat(val, 64); err == nil {
						s.Grades[q] = f
					} else {
						s.Grades[q] = ""
					}
				default:
					s.Grades[q] = ""
				}
			}
			s.TotalScore = total
			s.Description = comments
			s.NotSubmitted = notSub
			s.IsSubmitted = true
			dm.recalcFullyGraded()
			dm.ExportFiles()
			break
		}
	}
}

func (dm *DataManager) SetFlag(id string, flagged bool) {
	for _, s := range dm.Students {
		if s.ID == id {
			s.Flagged = flagged
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
			maxVal := math.Round(vals[len(vals)-1]*100) / 100
			minVal := math.Round(vals[0]*100) / 100
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
				"max": maxVal,
				"min": minVal,
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