package common

// Report is a structured output produced by a Module.
// The console runner will render these into colored tables and blocks.

type Report struct {
	Title    string    // Human-friendly title
	Summary  string    // One-paragraph summary
	Sections []Section // Detailed sections with tables
	Warnings []string
	Errors   []string
	Meta     map[string]any // Optional metadata for cross-module use
}

type Section struct {
	Header string
	Table  *Table   // Optional table data for compact presentation
	Lines  []string // Free-form lines if no table
}

type Table struct {
	Headers []string
	Rows    [][]string
}
