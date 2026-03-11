package report

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/runkids/mdproof/internal/core"
)

type junitTestsuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Errors     int              `xml:"errors,attr"`
	Time       string           `xml:"time,attr"`
	Testsuites []junitTestsuite `xml:"testsuite"`
}

type junitTestsuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Skipped   int             `xml:"skipped,attr"`
	Time      string          `xml:"time,attr"`
	Testcases []junitTestcase `xml:"testcase"`
}

type junitTestcase struct {
	XMLName   xml.Name      `xml:"testcase"`
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
	Skipped   *junitSkipped `xml:"skipped,omitempty"`
	SystemOut string        `xml:"system-out,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Body    string `xml:",chardata"`
}

type junitSkipped struct {
	Message string `xml:"message,attr,omitempty"`
}

// WriteJUnitReport writes the reports as JUnit XML.
func WriteJUnitReport(w io.Writer, reports []core.Report) error {
	var totalTests, totalFailures int
	var totalMs int64

	suites := make([]junitTestsuite, 0, len(reports))
	for _, r := range reports {
		suite := junitTestsuite{
			Name:     r.Runbook,
			Tests:    r.Summary.Total,
			Failures: r.Summary.Failed,
			Skipped:  r.Summary.Skipped,
			Time:     msToSeconds(r.DurationMs),
		}

		for _, sr := range r.Steps {
			tc := junitTestcase{
				Name:      sr.Step.Title,
				Classname: r.Runbook,
				Time:      msToSeconds(sr.DurationMs),
			}

			switch sr.Status {
			case core.StatusFailed:
				tc.Failure = buildFailure(sr)
			case core.StatusSkipped:
				tc.Skipped = &junitSkipped{}
			}

			if sr.Stdout != "" {
				tc.SystemOut = sr.Stdout
			}

			suite.Testcases = append(suite.Testcases, tc)
		}

		totalTests += r.Summary.Total
		totalFailures += r.Summary.Failed
		totalMs += r.DurationMs
		suites = append(suites, suite)
	}

	root := junitTestsuites{
		Tests:      totalTests,
		Failures:   totalFailures,
		Time:       msToSeconds(totalMs),
		Testsuites: suites,
	}

	io.WriteString(w, xml.Header)
	out, err := xml.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, "\n")
	return err
}

func buildFailure(sr core.StepResult) *junitFailure {
	var msg, typ string
	var details []string

	for _, a := range sr.Assertions {
		if !a.Matched {
			if msg == "" {
				msg = a.Pattern
				typ = a.Type
				if typ == "" {
					typ = "AssertionError"
				}
			}
			line := a.Pattern
			if a.Detail != "" {
				line += " (" + a.Detail + ")"
			}
			details = append(details, line)
		}
	}

	if msg == "" {
		msg = core.StepFailReason(sr)
		typ = "AssertionError"
	}

	return &junitFailure{
		Message: msg,
		Type:    typ,
		Body:    strings.Join(details, "\n"),
	}
}

func msToSeconds(ms int64) string {
	return fmt.Sprintf("%.3f", float64(ms)/1000)
}
