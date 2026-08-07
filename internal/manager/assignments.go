package manager

import (
	"os"
	"path/filepath"
)

const (
	assignmentsDir = "Assignments"
	assignmentPerm = 0755
)

func ListAssignments() []string {
	os.MkdirAll(assignmentsDir, assignmentPerm)
	entries, _ := os.ReadDir(assignmentsDir)
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

func AssignmentPath(name string) string {
	return filepath.Join(assignmentsDir, name)
}

func CreateAssignment(name string) string {
	dir := AssignmentPath(name)
	os.MkdirAll(filepath.Join(dir, "Data"), assignmentPerm)
	os.MkdirAll(filepath.Join(dir, "PDFs"), assignmentPerm)
	os.MkdirAll(filepath.Join(dir, "output"), assignmentPerm)
	return dir
}