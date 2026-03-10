package report

import (
	"encoding/json"
	"io"

	"github.com/runkids/mdproof/internal/core"
)

// WriteJSONReport writes the report as indented JSON.
func WriteJSONReport(w io.Writer, report core.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// WriteJSONReports writes multiple reports as a JSON array.
func WriteJSONReports(w io.Writer, reports []core.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(reports)
}
