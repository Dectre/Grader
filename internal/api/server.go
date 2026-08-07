package api

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"grader/internal/manager"
	"grader/internal/models"

	"github.com/gin-gonic/gin"
)

const (
	defaultMoodleID   = "0000000"
	pdfFileNameFormat = "%s %s_%s_assignsubmission_file_%s.pdf"
)

type Server struct {
	router   *gin.Engine
	dm       *manager.DataManager
	staticFS embed.FS
}

func NewServer(staticFS embed.FS) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	s := &Server{router: r, staticFS: staticFS}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.GET("/", s.handleHome)

	subFS, err := fs.Sub(s.staticFS, "fonts")
	if err == nil {
		s.router.StaticFS("/fonts", http.FS(subFS))
	}

	s.router.GET("/api/assignments", s.handleListAssignments)
	s.router.POST("/api/assignments", s.handleCreateAssignment)
	s.router.POST("/api/select", s.handleSelectAssignment)
	s.router.POST("/api/switch", s.handleSwitchAssignment)

	grading := s.router.Group("/", s.requireDM())
	grading.GET("/api/init", s.handleInit)
	grading.GET("/api/student/:id", s.handleGetStudent)
	grading.POST("/api/student/:id/submit", s.handleSubmit)
	grading.POST("/api/student/:id/flag", s.handleFlag)
	grading.POST("/api/student/:id/upload", s.handleUpload)
	grading.GET("/api/pdf/:id", s.handleGetPDF)
	grading.GET("/api/download/grades", s.handleDownloadGrades)
	grading.GET("/api/download/pdf_view", s.handlePDFView)
}

func (s *Server) requireDM() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.dm == nil {
			c.JSON(http.StatusPreconditionRequired, gin.H{"error": "No assignment selected"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) findStudent(id string) *models.Student {
	for _, st := range s.dm.Students {
		if st.ID == id {
			return st
		}
	}
	return nil
}

func (s *Server) handleInit(c *gin.Context) {
	var students []map[string]interface{}
	for _, st := range s.dm.Students {
		students = append(students, map[string]interface{}{
			"index":         st.Index,
			"id":            st.ID,
			"name":          st.Name,
			"surname":       st.Surname,
			"has_pdf":       st.HasPDF,
			"is_submitted":  st.IsSubmitted,
			"not_submitted": st.NotSubmitted,
			"fully_graded":  st.FullyGraded,
			"flagged":       st.Flagged,
		})
	}
	c.JSON(200, gin.H{
		"questions":  s.dm.Questions,
		"max_grades": s.dm.MaxGrades,
		"students":   students,
		"stats":      s.dm.GetStats(),
	})
}

func (s *Server) handleGetStudent(c *gin.Context) {
	id := c.Param("id")
	st := s.findStudent(id)
	if st == nil {
		c.JSON(404, gin.H{"error": "Not Found"})
		return
	}
	c.JSON(200, gin.H{
		"grades":        st.Grades,
		"description":   st.Description,
		"not_submitted": st.NotSubmitted,
		"is_submitted":  st.IsSubmitted,
		"fully_graded":  st.FullyGraded,
		"flagged":       st.Flagged,
	})
}

func (s *Server) handleSubmit(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Grades       map[string]interface{} `json:"grades"`
		Comments     string                 `json:"comments"`
		NotSubmitted bool                   `json:"not_submitted"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}
	total := s.calculateTotal(body.Grades)
	s.dm.SaveGrade(id, body.Grades, total, body.Comments, body.NotSubmitted)
	c.JSON(200, gin.H{"status": "success", "stats": s.dm.GetStats()})
}

func (s *Server) calculateTotal(grades map[string]interface{}) float64 {
	total := 0.0
	for _, v := range grades {
		switch val := v.(type) {
		case float64:
			total += val
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				total += f
			}
		}
	}
	return total
}

func (s *Server) handleFlag(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Flagged bool `json:"flagged"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}
	s.dm.SetFlag(id, body.Flagged)
	c.JSON(200, gin.H{"status": "success"})
}

func (s *Server) handleUpload(c *gin.Context) {
	id := c.Param("id")
	student := s.findStudent(id)
	if student == nil {
		c.JSON(404, gin.H{"error": "Not Found"})
		return
	}
	file, _ := c.FormFile("file")
	if file == nil {
		c.JSON(400, gin.H{"error": "No file"})
		return
	}
	moodleID := s.extractMoodleID()
	newName := s.buildPDFFileName(student, moodleID)
	path := filepath.Join(s.dm.PDFDir, newName)
	c.SaveUploadedFile(file, path)
	student.PDFPath = path
	student.HasPDF = true
	c.JSON(200, gin.H{"status": "success"})
}

func (s *Server) extractMoodleID() string {
	files, _ := os.ReadDir(s.dm.PDFDir)
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".pdf") {
			continue
		}
		parts := strings.Split(f.Name(), "_")
		if len(parts) > 2 {
			if _, err := strconv.Atoi(parts[1]); err == nil {
				return parts[1]
			}
		}
	}
	return defaultMoodleID
}

func (s *Server) buildPDFFileName(student *models.Student, moodleID string) string {
	return fmt.Sprintf(pdfFileNameFormat, student.Name, student.Surname, moodleID, student.ID)
}

func (s *Server) handleGetPDF(c *gin.Context) {
	id := c.Param("id")
	st := s.findStudent(id)
	if st == nil || !st.HasPDF {
		c.JSON(404, gin.H{"error": "Not Found"})
		return
	}
	c.Header("Content-Disposition", "inline")
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.File(st.PDFPath)
}

func (s *Server) handleDownloadGrades(c *gin.Context) {
	c.Header("Content-Disposition", "attachment; filename=grades.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.File(s.dm.ExcelPath)
}

func (s *Server) handlePDFView(c *gin.Context) {
	html := s.buildHTMLHead()
	html += s.renderStatRows()
	html += s.renderStudentRows()
	c.Data(200, "text/html; charset=utf-8", []byte(html))
}

func (s *Server) buildHTMLHead() string {
	return `<!DOCTYPE html><html lang="fa" dir="rtl"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>Grades Report</title><style>@font-face { font-family: 'Vazirmatn'; src: url('/fonts/Vazirmatn.ttf'); }body { font-family: 'Vazirmatn', Tahoma, Arial, sans-serif; padding: 20px; background: #fff; color: #000; direction: rtl; }h2 { text-align: center; color: #333; margin-bottom: 20px; }.table { width: 100%; border-collapse: collapse; font-size: 13px; text-align: center; border: 2px solid #000; }.table th, .table td { border: 1px solid #000; padding: 6px; vertical-align: middle; }.table th { background-color: #f4f4f4; color: #000; font-weight: bold; }.stat-row th { text-align: left; padding-left: 15px; border-bottom: 1px solid #000; }.header-row th { border-top: 2px solid #000; border-bottom: 2px solid #000; }.desc-col { text-align: right; max-width: 400px; line-height: 1.6; }.flagged-cell { background-color: #FFFF00; }.table tbody tr:last-child td { border-bottom: 2px solid #000; }@media print {body { padding: 0; }.no-print { display: none; }}</style></head><body><div class="no-print" style="text-align: center; margin-bottom: 20px;"><button onclick="window.print()" style="padding: 10px 20px; font-size: 16px; cursor: pointer; background: #1f6feb; color: white; border: none; border-radius: 5px; font-family: 'Vazirmatn';">Save as PDF / Print</button></div><h2>گزارش نمرات دانشجویان</h2><table class="table"><thead>`
}

func (s *Server) renderStatRows() string {
	statsData := s.dm.GetStats()["data"].(map[string]interface{})
	html := s.renderGradedCountRow(statsData)
	html += s.renderStatLabelRows(statsData)
	html += "<tr class='header-row'><th>First Name</th><th>Last Name</th><th>Student ID</th>"
	for _, q := range s.dm.Questions {
		html += fmt.Sprintf("<th>%s</th>", q)
	}
	html += "<th>Total Score</th><th>Description</th><th>Not Submitted</th><th>Fully Graded</th><th>Flagged</th></tr></thead><tbody>"
	return html
}

func (s *Server) renderGradedCountRow(statsData map[string]interface{}) string {
	html := "<tr class='stat-row'><th colspan='3'>Graded Count</th>"
	for _, q := range s.dm.Questions {
		val := statsData[q].(map[string]float64)["count"]
		html += fmt.Sprintf("<td>%.0f</td>", val)
	}
	val := statsData["Total Score"].(map[string]float64)["count"]
	html += fmt.Sprintf("<td>%.0f</td><td></td><td></td><td></td><td></td></tr>", val)
	return html
}

func (s *Server) renderStatLabelRows(statsData map[string]interface{}) string {
	html := ""
	labels := []string{"Mean", "Median", "Max", "Min"}
	keys := []string{"avg", "med", "max", "min"}
	for i, label := range labels {
		html += fmt.Sprintf("<tr class='stat-row'><th colspan='3'>%s</th>", label)
		for _, q := range s.dm.Questions {
			val := statsData[q].(map[string]float64)[keys[i]]
			html += fmt.Sprintf("<td>%.2f</td>", val)
		}
		val := statsData["Total Score"].(map[string]float64)[keys[i]]
		html += fmt.Sprintf("<td>%.2f</td><td></td><td></td><td></td><td></td></tr>", val)
	}
	return html
}

func (s *Server) renderStudentRows() string {
	html := ""
	for _, st := range s.dm.Students {
		html += s.renderStudentRow(st)
	}
	html += "</tbody></table></body></html>"
	return html
}

func (s *Server) renderStudentRow(st *models.Student) string {
	html := "<tr>"
	flagClass := ""
	if st.Flagged {
		flagClass = " class='flagged-cell'"
	}
	html += fmt.Sprintf("<td%s>%s</td><td%s>%s</td><td%s>%s</td>", flagClass, st.Name, flagClass, st.Surname, flagClass, st.ID)
	for _, q := range s.dm.Questions {
		if st.NotSubmitted {
			html += "<td></td>"
		} else if val, ok := st.Grades[q].(float64); ok {
			html += fmt.Sprintf("<td>%.2f</td>", val)
		} else {
			html += "<td></td>"
		}
	}
	if st.NotSubmitted {
		html += "<td>0.00</td>"
	} else if st.FullyGraded {
		html += fmt.Sprintf("<td>%.2f</td>", st.TotalScore)
	} else {
		html += "<td></td>"
	}
	desc := strings.ReplaceAll(st.Description, "\n", "<br>")
	desc = strings.ReplaceAll(desc, "\\n", "<br>")
	html += fmt.Sprintf("<td class='desc-col'>%s</td>", desc)
	if st.NotSubmitted {
		html += "<td>☑</td>"
	} else {
		html += "<td>☐</td>"
	}
	if st.FullyGraded {
		html += "<td>☑</td>"
	} else {
		html += "<td>☐</td>"
	}
	if st.Flagged {
		html += "<td>☑</td>"
	} else {
		html += "<td>☐</td>"
	}
	html += "</tr>"
	return html
}

