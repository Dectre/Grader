package api

import (
	"path/filepath"

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
	c.JSON(200, gin.H{"assignments": manager.ListAssignments()})
}

func (s *Server) handleCreateAssignment(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	if err := c.BindJSON(&body); err != nil || body.Name == "" {
		c.JSON(400, gin.H{"error": "Bad Request"})
		return
	}
	manager.CreateAssignment(body.Name)
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
	c.JSON(200, gin.H{"status": "ok"})
}