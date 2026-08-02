package models

type Student struct {
	Index        int
	ID           string
	Name         string
	Surname      string
	HasPDF       bool
	PDFPath      string
	IsSubmitted  bool
	NotSubmitted bool
	FullyGraded  bool
	Flagged      bool
	Description  string
	Grades       map[string]interface{}
	TotalScore   float64
}