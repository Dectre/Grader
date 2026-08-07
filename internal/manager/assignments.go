package manager

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	assignmentsDir = "Assignments"
	assignmentPerm = 0755
	registryFile   = "assignments.json"
	defaultType    = "pdf"
)

type Assignment struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	CreatedAt  string `json:"created_at"`
	ModifiedAt string `json:"modified_at"`
}

type registry struct {
	Assignments []Assignment `json:"assignments"`
}

func registryPath() string {
	return filepath.Join(assignmentsDir, registryFile)
}

func AssignmentPath(name string) string {
	return filepath.Join(assignmentsDir, name)
}

func ListAssignments() []string {
	var names []string
	for _, a := range loadRegistry().Assignments {
		names = append(names, a.Name)
	}
	return names
}

func CreateAssignment(name string) string {
	dir := AssignmentPath(name)
	os.MkdirAll(filepath.Join(dir, "Data"), assignmentPerm)
	os.MkdirAll(filepath.Join(dir, "PDFs"), assignmentPerm)
	os.MkdirAll(filepath.Join(dir, "output"), assignmentPerm)
	registerAssignment(name)
	return dir
}

func registerAssignment(name string) {
	r := loadRegistry()
	for _, a := range r.Assignments {
		if a.Name == name {
			return
		}
	}
	now := time.Now().Format(time.RFC3339)
	r.Assignments = append(r.Assignments, Assignment{
		Name:       name,
		Type:       defaultType,
		CreatedAt:  now,
		ModifiedAt: now,
	})
	saveRegistry(r)
}

func loadRegistry() registry {
	data, err := os.ReadFile(registryPath())
	if err != nil {
		return migrateRegistry()
	}
	var r registry
	if err := json.Unmarshal(data, &r); err != nil {
		return migrateRegistry()
	}
	return r
}

func saveRegistry(r registry) {
	os.MkdirAll(assignmentsDir, assignmentPerm)
	data, _ := json.MarshalIndent(r, "", "  ")
	os.WriteFile(registryPath(), data, 0644)
}

func migrateRegistry() registry {
	var r registry
	entries, _ := os.ReadDir(assignmentsDir)
	for _, e := range entries {
		if e.IsDir() && hasDataDir(e.Name()) {
			now := time.Now().Format(time.RFC3339)
			r.Assignments = append(r.Assignments, Assignment{
				Name:       e.Name(),
				Type:       defaultType,
				CreatedAt:  now,
				ModifiedAt: now,
			})
		}
	}
	saveRegistry(r)
	return r
}

func hasDataDir(name string) bool {
	_, err := os.Stat(filepath.Join(assignmentsDir, name, "Data"))
	return err == nil
}

func GetAssignments() []Assignment {
	return loadRegistry().Assignments
}

func AssignmentExists(name string) bool {
	for _, a := range loadRegistry().Assignments {
		if a.Name == name {
			return true
		}
	}
	return false
}

func TouchAssignment(name string) {
	r := loadRegistry()
	for i := range r.Assignments {
		if r.Assignments[i].Name == name {
			r.Assignments[i].ModifiedAt = time.Now().Format(time.RFC3339)
			break
		}
	}
	saveRegistry(r)
}

func (dm *DataManager) Touch() {
	TouchAssignment(filepath.Base(dm.AssignmentDir))
}
