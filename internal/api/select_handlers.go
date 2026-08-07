package api

import (
	"net/http"
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