package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"grader/internal/manager"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleHome(c *gin.Context) {
	name := "index.html"
	if s.dm == nil {
		name = "select.html"
	}
	file, err := s.staticFS.ReadFile(name)
	if err != nil {
		c.String(500, "Error loading "+name)
		return
	}
	c.Data(200, "text/html; charset=utf-8", file)
}

func (s *Server) handleListAssignments(c *gin.Context) {
	c.JSON(200, gin.H{"assignments": manager.GetAssignments()})
}

func (s *Server) handleCreateAssignment(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}
	if manager.AssignmentExists(name) {
		c.JSON(http.StatusConflict, gin.H{"error": "duplicate"})
		return
	}
	manager.CreateAssignment(name)
	c.JSON(200, gin.H{"status": "success"})
}


func (s *Server) handleSelectAssignment(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	if err := c.BindJSON(&body); err != nil || body.Name == "" {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}
	dm, needsSetup := manager.NewDataManager(manager.AssignmentPath(body.Name))
	if needsSetup {
		c.JSON(200, gin.H{
			"status":   "needs_setup",
			"data_dir": filepath.Join(manager.AssignmentPath(body.Name), "Data"),
		})
		return
	}
	s.dm = dm
	dm.Touch()
	c.JSON(200, gin.H{"status": "ok"})
}

func (s *Server) handleSwitchAssignment(c *gin.Context) {
	s.dm = nil
	c.JSON(200, gin.H{"status": "ok"})
}

func dataFileName(kind string) (string, bool) {
	switch kind {
	case "rubric":
		return "rubric.csv", true
	case "students":
		return "students.csv", true
	}
	return "", false
}

func (s *Server) handleGetDataFile(c *gin.Context) {
	name := filepath.Base(c.Param("name"))
	fname, ok := dataFileName(c.Param("kind"))
	if !ok {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}
	data, err := os.ReadFile(filepath.Join(manager.AssignmentPath(name), "Data", fname))
	if err != nil {
		c.JSON(200, gin.H{"content": ""})
		return
	}
	c.JSON(200, gin.H{"content": string(data)})
}

func (s *Server) handleUploadDataFile(c *gin.Context) {
	name := filepath.Base(c.Param("name"))
	fname, ok := dataFileName(c.Param("kind"))
	if !ok {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil || file == nil {
		c.JSON(400, gin.H{"error": "No file"})
		return
	}
	dataDir := filepath.Join(manager.AssignmentPath(name), "Data")
	os.MkdirAll(dataDir, 0755)
	if err := c.SaveUploadedFile(file, filepath.Join(dataDir, fname)); err != nil {
		c.JSON(500, gin.H{"error": "Save failed"})
		return
	}
	c.JSON(200, gin.H{"status": "success"})
}

func (s *Server) handleSetDataText(c *gin.Context) {
	name := filepath.Base(c.Param("name"))
	fname, ok := dataFileName(c.Param("kind"))
	if !ok {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}
	dataDir := filepath.Join(manager.AssignmentPath(name), "Data")
	os.MkdirAll(dataDir, 0755)
	if err := os.WriteFile(filepath.Join(dataDir, fname), []byte(body.Content), 0644); err != nil {
		c.JSON(500, gin.H{"error": "Write failed"})
		return
	}
	c.JSON(200, gin.H{"status": "success"})
}

func (s *Server) handleUploadPDFs(c *gin.Context) {
	name := filepath.Base(c.Param("name"))
	pdfDir := filepath.Join(manager.AssignmentPath(name), "PDFs")
	form, _ := c.MultipartForm()
	if form == nil {
		c.JSON(400, gin.H{"error": "No files"})
		return
	}
	added := 0
	for _, fh := range form.File["files"] {
		lower := strings.ToLower(fh.Filename)
		if strings.HasSuffix(lower, ".zip") {
			tmp := filepath.Join(os.TempDir(), fh.Filename)
			if err := c.SaveUploadedFile(fh, tmp); err != nil {
				continue
			}
			n, _ := manager.ImportZipToPDFs(tmp, pdfDir)
			os.Remove(tmp)
			added += n
		} else if strings.HasSuffix(lower, ".pdf") {
			src, err := fh.Open()
			if err != nil {
				continue
			}
			manager.SavePDFToDir(src, pdfDir, fh.Filename)
			src.Close()
			added++
		}
	}
	c.JSON(200, gin.H{"status": "success", "added": added})
}