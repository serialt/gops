package biz

import "testing"

func TestSuffixedFilename(t *testing.T) {
	tests := map[string]string{
		"report.txt":     "report_1.txt",
		"archive.tar.gz": "archive.tar_1.gz",
		"README":         "README_1",
		".env":           ".env_1",
	}
	for filename, expected := range tests {
		if actual := suffixedFilename(filename, 1); actual != expected {
			t.Fatalf("suffixedFilename(%q, 1) = %q, want %q", filename, actual, expected)
		}
	}
}
