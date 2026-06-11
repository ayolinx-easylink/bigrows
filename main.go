package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

var paymentRecordHeader = []string{
	"Date and Time",
	"Merchant ID",
	"Merchant Name",
	"Transaction Type",
	"Payment Channel",
	"Ayolinx Transaction ID",
	"Partner Refference Number",
	"Transaction Amount",
	"Admin Fee",
	"VAT Type",
	"VAT(11%)",
	"Transaction Settled",
	"Transaction Status",
	"Settlement Status",
	"Merchant Transaction ID",
	"Merchant Refference Number",
	"Shop ID",
	"Settlement Type",
	"Create Time",
}

func main() {
	dir := flag.String("dir", ".", "folder containing CSV files")
	file := flag.String("file", "", "CSV file path to split directly")
	parts := flag.Int("parts", 0, "number of output parts; if omitted, the CLI asks interactively")
	flag.Parse()

	options := cliOptions{
		dir:   *dir,
		file:  *file,
		parts: *parts,
	}

	if err := run(options); err != nil {
		log.Fatal(err)
	}
}

type cliOptions struct {
	dir   string
	file  string
	parts int
}

type csvRecordReader interface {
	Read() ([]string, error)
}

type standardCSVReader struct {
	reader      *csv.Reader
	firstRecord bool
}

type outerQuotedCSVReader struct {
	reader      *bufio.Reader
	firstRecord bool
}

func run(options cliOptions) error {
	if options.parts < 0 {
		return fmt.Errorf("split count must be greater than 0")
	}

	if options.file != "" {
		parts, err := getParts(options.parts)
		if err != nil {
			return err
		}
		return splitAndPrintResult(options.file, parts)
	}

	files, err := listCSVFiles(options.dir)
	if err != nil {
		return err
	}

	fmt.Println("CSV files found:")
	for i, file := range files {
		fmt.Printf("%d. %s\n", i+1, file)
	}
	fmt.Println()

	chosenFile, err := askFileChoice(files)
	if err != nil {
		return err
	}

	parts, err := getParts(options.parts)
	if err != nil {
		return err
	}

	inputPath := filepath.Join(options.dir, chosenFile)
	return splitAndPrintResult(inputPath, parts)
}

func getParts(parts int) (int, error) {
	if parts == 0 {
		return askParts()
	}
	if parts < 1 {
		return 0, fmt.Errorf("split count must be greater than 0")
	}
	return parts, nil
}

func splitAndPrintResult(inputPath string, parts int) error {
	outputDir, createdParts, err := splitCSV(inputPath, parts)
	if err != nil {
		return err
	}

	fmt.Printf("Split complete: %d file(s) created in %s\n", createdParts, outputDir)
	return nil
}

func listCSVFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("folder not found: %s", dir)
		}
		return nil, fmt.Errorf("failed to read folder %s: %w", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".csv") {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no CSV files found in folder: %s", dir)
	}

	return files, nil
}

func askFileChoice(files []string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Choose file number: ")

	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read file choice: %w", err)
	}

	choice, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return "", fmt.Errorf("file choice must be a number")
	}
	if choice < 1 || choice > len(files) {
		return "", fmt.Errorf("file choice must be between 1 and %d", len(files))
	}

	return files[choice-1], nil
}

func askParts() (int, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Split into how many parts: ")

	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, fmt.Errorf("failed to read split count: %w", err)
	}

	parts, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return 0, fmt.Errorf("split count must be a number")
	}
	if parts < 1 {
		return 0, fmt.Errorf("split count must be greater than 0")
	}

	return parts, nil
}

func countRows(filePath string) (totalRows int, header []string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to open CSV file %s: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close CSV file %s: %w", filePath, closeErr)
		}
	}()

	reader, err := newCSVRecordReader(file)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to inspect CSV format in %s: %w", filePath, err)
	}

	firstRecord, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return 0, nil, fmt.Errorf("CSV file is empty: %s", filePath)
	}
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read CSV header from %s: %w", filePath, err)
	}
	if len(firstRecord) == 0 {
		return 0, nil, fmt.Errorf("CSV file has no header: %s", filePath)
	}

	header = firstRecord
	if firstFieldLooksLikeDateTime(firstRecord[0]) {
		header = generatedHeader(len(firstRecord))
		totalRows = 1
	}

	for {
		_, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, nil, fmt.Errorf("failed to read CSV row from %s: %w", filePath, err)
		}
		totalRows++
	}

	return totalRows, header, nil
}

func generatedHeader(columnCount int) []string {
	if columnCount == len(paymentRecordHeader) {
		return slices.Clone(paymentRecordHeader)
	}

	header := make([]string, columnCount)
	for i := range header {
		header[i] = fmt.Sprintf("column_%d", i+1)
	}
	return header
}

func firstFieldLooksLikeDateTime(value string) bool {
	layouts := []string{
		"02-01-2006 15:04:05",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if _, err := time.Parse(layout, value); err == nil {
			return true
		}
	}
	return false
}

func splitCSV(filePath string, parts int) (outputDir string, createdParts int, err error) {
	if parts < 1 {
		return "", 0, fmt.Errorf("split count must be greater than 0")
	}
	if !strings.EqualFold(filepath.Ext(filePath), ".csv") {
		return "", 0, fmt.Errorf("input file must have .csv extension: %s", filePath)
	}

	totalRows, header, err := countRows(filePath)
	if err != nil {
		return "", 0, err
	}
	if totalRows == 0 {
		return "", 0, fmt.Errorf("CSV file has no data rows: %s", filePath)
	}

	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	outputDir = filepath.Join(filepath.Dir(filePath), baseName+"_split")
	if err := os.Mkdir(outputDir, 0o755); err != nil {
		if errors.Is(err, os.ErrExist) {
			if err := os.RemoveAll(outputDir); err != nil {
				return "", 0, fmt.Errorf("failed to remove existing output folder %s: %w", outputDir, err)
			}
			if err := os.Mkdir(outputDir, 0o755); err != nil {
				return "", 0, fmt.Errorf("failed to recreate output folder %s: %w", outputDir, err)
			}
		} else {
			return "", 0, fmt.Errorf("failed to create output folder %s: %w", outputDir, err)
		}
	}

	rowsPerPart := int(math.Ceil(float64(totalRows) / float64(parts)))

	input, err := os.Open(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to reopen CSV file %s: %w", filePath, err)
	}
	defer func() {
		if closeErr := input.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close CSV file %s: %w", filePath, closeErr)
		}
	}()

	reader, err := newCSVRecordReader(input)
	if err != nil {
		return "", 0, fmt.Errorf("failed to inspect CSV format in %s: %w", filePath, err)
	}

	firstRecord, err := reader.Read()
	if err != nil {
		return "", 0, fmt.Errorf("failed to read CSV header from %s: %w", filePath, err)
	}
	var pendingRecord []string
	if !slices.Equal(firstRecord, header) {
		pendingRecord = firstRecord
	}

	var currentFile *os.File
	var currentWriter *csv.Writer
	rowsInCurrentPart := 0

	defer func() {
		if currentFile == nil {
			return
		}
		if closeErr := closeCSVOutput(currentFile, currentWriter); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	for {
		record := pendingRecord
		pendingRecord = nil
		var readErr error
		if record == nil {
			record, readErr = reader.Read()
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return "", 0, fmt.Errorf("failed to read CSV row from %s: %w", filePath, readErr)
		}

		if currentWriter == nil || rowsInCurrentPart >= rowsPerPart {
			if currentFile != nil {
				if err := closeCSVOutput(currentFile, currentWriter); err != nil {
					return "", 0, err
				}
			}

			createdParts++
			currentFile, currentWriter, err = createOutputFile(outputDir, baseName, createdParts, header)
			if err != nil {
				return "", 0, err
			}
			rowsInCurrentPart = 0
		}

		if err := currentWriter.Write(record); err != nil {
			return "", 0, fmt.Errorf("failed to write CSV row to part %d: %w", createdParts, err)
		}
		rowsInCurrentPart++
	}

	return outputDir, createdParts, nil
}

func newCSVRecordReader(file *os.File) (csvRecordReader, error) {
	outerQuoted, err := isOuterQuotedCSV(file)
	if err != nil {
		return nil, err
	}

	if outerQuoted {
		return &outerQuotedCSVReader{
			reader:      bufio.NewReader(file),
			firstRecord: true,
		}, nil
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	return &standardCSVReader{
		reader:      reader,
		firstRecord: true,
	}, nil
}

func isOuterQuotedCSV(file *os.File) (bool, error) {
	reader := bufio.NewReader(file)
	line, readErr := reader.ReadString('\n')
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return false, readErr
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return false, err
	}

	line = strings.TrimPrefix(line, "\uFEFF")
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	if len(line) < 2 || line[0] != '"' || line[len(line)-1] != '"' {
		return false, nil
	}

	normalized := strings.ReplaceAll(line[1:len(line)-1], `""`, `"`)
	probe := csv.NewReader(strings.NewReader(normalized))
	probe.FieldsPerRecord = -1
	record, err := probe.Read()
	return err == nil && len(record) > 1, nil
}

func (r *standardCSVReader) Read() ([]string, error) {
	record, err := r.reader.Read()
	if err != nil {
		return nil, err
	}
	if r.firstRecord && len(record) > 0 {
		record[0] = strings.TrimPrefix(record[0], "\uFEFF")
		r.firstRecord = false
	}
	return record, nil
}

func (r *outerQuotedCSVReader) Read() ([]string, error) {
	line, readErr := r.reader.ReadString('\n')
	if errors.Is(readErr, io.EOF) && len(line) == 0 {
		return nil, io.EOF
	}
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return nil, readErr
	}

	if r.firstRecord {
		line = strings.TrimPrefix(line, "\uFEFF")
		r.firstRecord = false
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	if len(line) < 2 || line[0] != '"' || line[len(line)-1] != '"' {
		return nil, fmt.Errorf("expected an outer-quoted CSV row")
	}

	line = strings.ReplaceAll(line[1:len(line)-1], `""`, `"`)
	reader := csv.NewReader(strings.NewReader(line))
	reader.FieldsPerRecord = -1
	return reader.Read()
}

func createOutputFile(outputDir, baseName string, partIndex int, header []string) (*os.File, *csv.Writer, error) {
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_part_%d.csv", baseName, partIndex))
	file, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, nil, fmt.Errorf("output file already exists, refusing to overwrite: %s", outputPath)
		}
		return nil, nil, fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}

	writer := csv.NewWriter(file)
	if err := writer.Write(header); err != nil {
		_ = file.Close()
		return nil, nil, fmt.Errorf("failed to write header to %s: %w", outputPath, err)
	}

	return file, writer, nil
}

func closeCSVOutput(file *os.File, writer *csv.Writer) error {
	writer.Flush()
	if err := writer.Error(); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to flush output file %s: %w", file.Name(), err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close output file %s: %w", file.Name(), err)
	}
	return nil
}
