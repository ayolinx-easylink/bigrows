package main

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestListCSVFilesEmptyFolder(t *testing.T) {
	dir := t.TempDir()

	_, err := listCSVFiles(dir)
	if err == nil {
		t.Fatal("expected error for empty folder")
	}
	if !strings.Contains(err.Error(), "no CSV files found") {
		t.Fatalf("expected no CSV files error, got %q", err)
	}
}

func TestCountRowsSmallCSVWithHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "customers.csv")
	writeFile(t, path, "id,name\n1,Ana\n2,Budi\n")

	totalRows, header, err := countRows(path)
	if err != nil {
		t.Fatalf("countRows returned error: %v", err)
	}

	if totalRows != 2 {
		t.Fatalf("expected 2 data rows, got %d", totalRows)
	}
	if !reflect.DeepEqual(header, []string{"id", "name"}) {
		t.Fatalf("unexpected header: %#v", header)
	}
}

func TestSplitCSVIntoTwoParts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "transactions.csv")
	writeFile(t, path, "id,amount\n1,100\n2,200\n3,300\n4,400\n5,500\n")

	outputDir, createdParts, err := splitCSV(path, 2)
	if err != nil {
		t.Fatalf("splitCSV returned error: %v", err)
	}

	if createdParts != 2 {
		t.Fatalf("expected 2 created parts, got %d", createdParts)
	}

	assertCSVRecords(t, filepath.Join(outputDir, "transactions_part_1.csv"), [][]string{
		{"id", "amount"},
		{"1", "100"},
		{"2", "200"},
		{"3", "300"},
	})
	assertCSVRecords(t, filepath.Join(outputDir, "transactions_part_2.csv"), [][]string{
		{"id", "amount"},
		{"4", "400"},
		{"5", "500"},
	})
}

func TestSplitCSVPartsGreaterThanRows(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "small.csv")
	writeFile(t, path, "id,name\n1,Ana\n2,Budi\n")

	outputDir, createdParts, err := splitCSV(path, 5)
	if err != nil {
		t.Fatalf("splitCSV returned error: %v", err)
	}

	if createdParts != 2 {
		t.Fatalf("expected 2 created parts, got %d", createdParts)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "small_part_3.csv")); !os.IsNotExist(err) {
		t.Fatalf("expected part 3 to not exist, stat error: %v", err)
	}
}

func TestSplitCSVInconsistentColumnCount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mixed.csv")
	writeFile(t, path, "id,name,city\n1,Ana\n2,Budi,Jakarta,Extra\n")

	outputDir, createdParts, err := splitCSV(path, 2)
	if err != nil {
		t.Fatalf("splitCSV returned error: %v", err)
	}

	if createdParts != 2 {
		t.Fatalf("expected 2 created parts, got %d", createdParts)
	}

	assertCSVRecords(t, filepath.Join(outputDir, "mixed_part_1.csv"), [][]string{
		{"id", "name", "city"},
		{"1", "Ana"},
	})
	assertCSVRecords(t, filepath.Join(outputDir, "mixed_part_2.csv"), [][]string{
		{"id", "name", "city"},
		{"2", "Budi", "Jakarta", "Extra"},
	})
}

func TestSplitCSVOuterQuotedRowsWithBOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "export.csv")
	writeFile(t, path, "\uFEFF\"id,\"\"merchant name\"\",amount\"\r\n\"1,\"\"PT Langit\"\",10000\"\r\n\"2,\"\"PT Bumi\"\",20000\"\r\n")

	outputDir, createdParts, err := splitCSV(path, 2)
	if err != nil {
		t.Fatalf("splitCSV returned error: %v", err)
	}

	if createdParts != 2 {
		t.Fatalf("expected 2 created parts, got %d", createdParts)
	}

	assertCSVRecords(t, filepath.Join(outputDir, "export_part_1.csv"), [][]string{
		{"id", "merchant name", "amount"},
		{"1", "PT Langit", "10000"},
	})
	assertCSVRecords(t, filepath.Join(outputDir, "export_part_2.csv"), [][]string{
		{"id", "merchant name", "amount"},
		{"2", "PT Bumi", "20000"},
	})
}

func TestSplitCSVStandardQuotedRowsWithBOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "standard-export.csv")
	writeFile(t, path, "\uFEFF\"id\",\"merchant name\",amount\n\"1\",\"PT Langit\",10000\n\"2\",\"PT Bumi\",20000\n")

	outputDir, createdParts, err := splitCSV(path, 2)
	if err != nil {
		t.Fatalf("splitCSV returned error: %v", err)
	}

	if createdParts != 2 {
		t.Fatalf("expected 2 created parts, got %d", createdParts)
	}

	assertCSVRecords(t, filepath.Join(outputDir, "standard-export_part_1.csv"), [][]string{
		{"id", "merchant name", "amount"},
		{"1", "PT Langit", "10000"},
	})
	assertCSVRecords(t, filepath.Join(outputDir, "standard-export_part_2.csv"), [][]string{
		{"id", "merchant name", "amount"},
		{"2", "PT Bumi", "20000"},
	})
}

func TestSplitCSVOuterQuotedRowsWithoutHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "headerless.csv")
	writeFile(t, path, "\"07-06-2026 23:59:59,627,\"\"PT Langit\"\"\"\r\n\"08-06-2026 00:00:00,628,\"\"PT Bumi\"\"\"\r\n")

	outputDir, createdParts, err := splitCSV(path, 2)
	if err != nil {
		t.Fatalf("splitCSV returned error: %v", err)
	}
	if createdParts != 2 {
		t.Fatalf("expected 2 created parts, got %d", createdParts)
	}

	assertCSVRecords(t, filepath.Join(outputDir, "headerless_part_1.csv"), [][]string{
		{"column_1", "column_2", "column_3"},
		{"07-06-2026 23:59:59", "627", "PT Langit"},
	})
	assertCSVRecords(t, filepath.Join(outputDir, "headerless_part_2.csv"), [][]string{
		{"column_1", "column_2", "column_3"},
		{"08-06-2026 00:00:00", "628", "PT Bumi"},
	})
}

func TestRunFileModeWithParts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "direct.csv")
	writeFile(t, path, "id,name\n1,Ana\n2,Budi\n3,Citra\n")

	err := run(cliOptions{
		file:  path,
		parts: 2,
	})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	assertCSVRecords(t, filepath.Join(dir, "direct_split", "direct_part_1.csv"), [][]string{
		{"id", "name"},
		{"1", "Ana"},
		{"2", "Budi"},
	})
	assertCSVRecords(t, filepath.Join(dir, "direct_split", "direct_part_2.csv"), [][]string{
		{"id", "name"},
		{"3", "Citra"},
	})
}

func TestSplitCSVOverwritesExistingOutputFolder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "repeat.csv")
	writeFile(t, path, "id,name\n1,Ana\n2,Budi\n")

	outputDir, _, err := splitCSV(path, 2)
	if err != nil {
		t.Fatalf("first splitCSV returned error: %v", err)
	}

	staleFile := filepath.Join(outputDir, "stale.txt")
	writeFile(t, staleFile, "old output")

	outputDir, createdParts, err := splitCSV(path, 1)
	if err != nil {
		t.Fatalf("second splitCSV returned error: %v", err)
	}

	if createdParts != 1 {
		t.Fatalf("expected 1 created part, got %d", createdParts)
	}
	if _, err := os.Stat(staleFile); !os.IsNotExist(err) {
		t.Fatalf("expected stale file to be removed, stat error: %v", err)
	}
	assertCSVRecords(t, filepath.Join(outputDir, "repeat_part_1.csv"), [][]string{
		{"id", "name"},
		{"1", "Ana"},
		{"2", "Budi"},
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
}

func assertCSVRecords(t *testing.T, path string, expected [][]string) {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open %s: %v", path, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	var records [][]string
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}
		records = append(records, record)
	}

	if !reflect.DeepEqual(records, expected) {
		t.Fatalf("unexpected records in %s:\nexpected: %#v\nactual:   %#v", path, expected, records)
	}
}
