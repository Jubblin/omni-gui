package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoggingJSONFormat tests that all log output is valid JSON
func TestLoggingJSONFormat(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer bytes.Buffer
	
	// Set up JSON logger writing to buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{
		Level:     slog.LevelDebug, // Include debug level to test all levels
		AddSource: true,
	}))
	
	// Temporarily replace the default logger
	oldLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldLogger)
	
	// Test various log levels and messages
	slog.Info("Test info message", "key1", "value1", "key2", 42)
	slog.Debug("Test debug message", "debug_key", "debug_value")
	slog.Warn("Test warning message", "warn_key", "warn_value")
	slog.Error("Test error message", "error_key", "error_value", "error", "test error")
	
	// Test startup banner
	logStartupBanner("1.0.0", false, false, 0)
	
	// Get all log output
	output := logBuffer.String()
	
	// Verify the entire output is valid newline-delimited JSON (NDJSON/JSONL)
	// First, verify it's not empty
	require.NotEmpty(t, output, "Log output should not be empty")
	
	// Verify the entire output can be parsed as a sequence of JSON objects
	lines := strings.Split(strings.TrimSpace(output), "\n")
	require.Greater(t, len(lines), 0, "Should have at least one log line")
	
	// Verify each line is valid JSON
	var allJsonObjects []map[string]interface{}
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}
		
		var jsonObj map[string]interface{}
		err := json.Unmarshal([]byte(line), &jsonObj)
		assert.NoError(t, err, "Line %d should be valid JSON: %s", i+1, line)
		
		// Collect all JSON objects for whole-file validation
		allJsonObjects = append(allJsonObjects, jsonObj)
		
		// Verify required JSON fields exist
		assert.Contains(t, jsonObj, "time", "Line %d should have 'time' field", i+1)
		assert.Contains(t, jsonObj, "level", "Line %d should have 'level' field", i+1)
		assert.Contains(t, jsonObj, "msg", "Line %d should have 'msg' field", i+1)
		
		// Verify level is a valid string
		level, ok := jsonObj["level"].(string)
		assert.True(t, ok, "Line %d 'level' should be a string", i+1)
		assert.Contains(t, []string{"INFO", "DEBUG", "WARN", "ERROR"}, level, "Line %d should have valid log level", i+1)
	}
	
	// Verify the entire file as a whole: all lines together form valid NDJSON
	// Re-parse the entire output to ensure it's valid as a whole
	trimmedOutput := strings.TrimSpace(output)
	allLines := strings.Split(trimmedOutput, "\n")
	
	// Verify we can parse all lines as a sequence
	var parsedObjects []map[string]interface{}
	for _, line := range allLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(line), &obj)
		require.NoError(t, err, "Entire file should be valid NDJSON - failed to parse line: %s", line)
		parsedObjects = append(parsedObjects, obj)
	}
	
	// Verify we parsed the expected number of objects
	assert.Equal(t, len(allJsonObjects), len(parsedObjects), "Should parse same number of objects from entire file")
	assert.Greater(t, len(parsedObjects), 0, "Should have at least one valid JSON object in entire file")
}

// TestStartupBannerJSON tests that the startup banner produces valid JSON
func TestStartupBannerJSON(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer bytes.Buffer
	
	// Set up JSON logger writing to buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	
	// Temporarily replace the default logger
	oldLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldLogger)
	
	// Call startup banner
	logStartupBanner("1.2.3", true, false, 2)
	
	// Get log output
	output := logBuffer.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	// Should have at least one log entry
	require.Greater(t, len(lines), 0, "Startup banner should produce at least one log line")
	
	// Verify each line is valid JSON
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}
		
		var jsonObj map[string]interface{}
		err := json.Unmarshal([]byte(line), &jsonObj)
		require.NoError(t, err, "Line %d should be valid JSON: %s", i+1, line)
		
		// Verify required JSON fields exist
		assert.Contains(t, jsonObj, "time", "Line %d should have 'time' field", i+1)
		assert.Contains(t, jsonObj, "level", "Line %d should have 'level' field", i+1)
		assert.Contains(t, jsonObj, "msg", "Line %d should have 'msg' field", i+1)
	}
	
	// Verify the banner log entry contains expected fields (check first non-empty line)
	var bannerJson map[string]interface{}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			err := json.Unmarshal([]byte(line), &bannerJson)
			if err == nil && bannerJson["msg"] == "Application Startup Banner" {
				break
			}
		}
	}
	
	require.NotNil(t, bannerJson, "Should find startup banner log entry")
	assert.Equal(t, "Application Startup Banner", bannerJson["msg"], "Banner should have correct message")
	assert.Equal(t, "1.2.3", bannerJson["version"], "Banner should include version")
	assert.Equal(t, "Omni GUI Application", bannerJson["application"], "Banner should include application name")
	assert.Equal(t, true, bannerJson["ignore_env"], "Banner should include ignore_env")
	assert.Equal(t, false, bannerJson["show_empty_labels"], "Banner should include show_empty_labels")
	assert.Equal(t, float64(2), bannerJson["grpc_debug_level"], "Banner should include grpc_debug_level")
	assert.Contains(t, bannerJson, "started_at", "Banner should include started_at timestamp")
	
	// Verify the entire output as a whole (NDJSON format)
	trimmedOutput := strings.TrimSpace(output)
	allLines := strings.Split(trimmedOutput, "\n")
	var allObjects []map[string]interface{}
	for _, line := range allLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(line), &obj)
		require.NoError(t, err, "Entire output should be valid NDJSON - failed to parse line: %s", line)
		allObjects = append(allObjects, obj)
	}
	assert.Greater(t, len(allObjects), 0, "Entire output should contain at least one valid JSON object")
}

// TestLoggingToFile tests that file logging produces valid JSON
func TestLoggingToFile(t *testing.T) {
	// Create a temporary file for logging
	tmpFile, err := os.CreateTemp("", "omni-api-test-*.log")
	require.NoError(t, err, "Should create temporary log file")
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	
	// Set up JSON logger writing to file
	logger := slog.New(slog.NewJSONHandler(tmpFile, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	
	// Temporarily replace the default logger
	oldLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldLogger)
	
	// Write various log messages
	slog.Info("File log test 1", "test_key", "test_value")
	slog.Warn("File log test 2", "warn_key", "warn_value")
	slog.Error("File log test 3", "error_key", "error_value")
	
	// Close file to ensure all data is written
	tmpFile.Close()
	
	// Read the file back
	file, err := os.Open(tmpFile.Name())
	require.NoError(t, err, "Should open log file for reading")
	defer file.Close()
	
	// Read and validate each line
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		var jsonObj map[string]interface{}
		err := json.Unmarshal([]byte(line), &jsonObj)
		assert.NoError(t, err, "Line %d in log file should be valid JSON: %s", lineNum, line)
		
		// Verify it's a log entry
		assert.Contains(t, jsonObj, "time", "Line %d should have 'time' field", lineNum)
		assert.Contains(t, jsonObj, "level", "Line %d should have 'level' field", lineNum)
		assert.Contains(t, jsonObj, "msg", "Line %d should have 'msg' field", lineNum)
	}
	
	require.NoError(t, scanner.Err(), "Should read log file without errors")
	assert.Greater(t, lineNum, 0, "Should have at least one log entry in file")
	
	// Verify the entire file as a whole: read all content and validate
	file.Seek(0, 0) // Reset to beginning
	fileContent, err := io.ReadAll(file)
	require.NoError(t, err, "Should read entire log file")
	
	// Verify entire file content is valid NDJSON
	fileLines := strings.Split(strings.TrimSpace(string(fileContent)), "\n")
	var fileObjects []map[string]interface{}
	for _, line := range fileLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(line), &obj)
		require.NoError(t, err, "Entire log file should be valid NDJSON - failed to parse line: %s", line)
		fileObjects = append(fileObjects, obj)
	}
	
	assert.Greater(t, len(fileObjects), 0, "Entire log file should contain at least one valid JSON object")
	assert.Equal(t, lineNum, len(fileObjects), "Should parse same number of objects from entire file as line-by-line")
}

// TestLoggingMultiWriter tests that multi-writer (console + file) produces valid JSON
func TestLoggingMultiWriter(t *testing.T) {
	// Create buffers to simulate console and file
	var consoleBuffer bytes.Buffer
	var fileBuffer bytes.Buffer
	
	// Create multi-writer
	multiWriter := io.MultiWriter(&consoleBuffer, &fileBuffer)
	
	// Set up JSON logger writing to multi-writer
	logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	
	// Temporarily replace the default logger
	oldLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldLogger)
	
	// Write log messages
	slog.Info("Multi-writer test", "key", "value")
	slog.Warn("Multi-writer warning", "warn_key", "warn_value")
	
	// Both buffers should have the same content
	consoleOutput := consoleBuffer.String()
	fileOutput := fileBuffer.String()
	
	assert.Equal(t, consoleOutput, fileOutput, "Console and file should have identical output")
	
	// Verify both outputs are valid JSON line by line
	lines := strings.Split(strings.TrimSpace(consoleOutput), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		var jsonObj map[string]interface{}
		err := json.Unmarshal([]byte(line), &jsonObj)
		assert.NoError(t, err, "Line %d should be valid JSON in both console and file", i+1)
	}
	
	// Verify entire console output as a whole (NDJSON format)
	trimmedConsole := strings.TrimSpace(consoleOutput)
	consoleLines := strings.Split(trimmedConsole, "\n")
	var consoleObjects []map[string]interface{}
	for _, line := range consoleLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(line), &obj)
		require.NoError(t, err, "Entire console output should be valid NDJSON - failed to parse line: %s", line)
		consoleObjects = append(consoleObjects, obj)
	}
	assert.Greater(t, len(consoleObjects), 0, "Entire console output should contain at least one valid JSON object")
	
	// Verify entire file output as a whole (NDJSON format)
	trimmedFile := strings.TrimSpace(fileOutput)
	fileLines := strings.Split(trimmedFile, "\n")
	var fileObjects []map[string]interface{}
	for _, line := range fileLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(line), &obj)
		require.NoError(t, err, "Entire file output should be valid NDJSON - failed to parse line: %s", line)
		fileObjects = append(fileObjects, obj)
	}
	assert.Greater(t, len(fileObjects), 0, "Entire file output should contain at least one valid JSON object")
	assert.Equal(t, len(consoleObjects), len(fileObjects), "Console and file should have same number of JSON objects")
}

// TestLoggingSpecialCharacters tests that logs with special characters are valid JSON
func TestLoggingSpecialCharacters(t *testing.T) {
	var logBuffer bytes.Buffer
	
	logger := slog.New(slog.NewJSONHandler(&logBuffer, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	
	oldLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldLogger)
	
	// Test various special characters that should be properly escaped in JSON
	slog.Info("Test message", 
		"quotes", `"quoted string"`,
		"newline", "line1\nline2",
		"tab", "col1\tcol2",
		"backslash", "path\\to\\file",
		"unicode", "测试 🚀 émoji",
		"url", "https://example.com/path?query=value&other=test",
	)
	
	output := logBuffer.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	require.Greater(t, len(lines), 0, "Should have log output")
	
	// Verify each line is valid JSON (special characters should be escaped)
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}
		
		var jsonObj map[string]interface{}
		err := json.Unmarshal([]byte(line), &jsonObj)
		require.NoError(t, err, "Line %d with special characters should be valid JSON: %s", i+1, line)
		
		// Verify required JSON fields exist
		assert.Contains(t, jsonObj, "time", "Line %d should have 'time' field", i+1)
		assert.Contains(t, jsonObj, "level", "Line %d should have 'level' field", i+1)
		assert.Contains(t, jsonObj, "msg", "Line %d should have 'msg' field", i+1)
	}
	
	// Verify special characters are preserved correctly (check first non-empty line)
	var testJson map[string]interface{}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			err := json.Unmarshal([]byte(line), &testJson)
			if err == nil {
				break
			}
		}
	}
	
	require.NotNil(t, testJson, "Should have at least one valid log entry")
	assert.Equal(t, `"quoted string"`, testJson["quotes"], "Quotes should be preserved")
	assert.Contains(t, testJson["newline"].(string), "\n", "Newlines should be preserved")
	assert.Contains(t, testJson["unicode"].(string), "测试", "Unicode should be preserved")
	
	// Verify the entire output as a whole (NDJSON format)
	trimmedOutput := strings.TrimSpace(output)
	allLines := strings.Split(trimmedOutput, "\n")
	var allObjects []map[string]interface{}
	for _, line := range allLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(line), &obj)
		require.NoError(t, err, "Entire output with special characters should be valid NDJSON - failed to parse line: %s", line)
		allObjects = append(allObjects, obj)
	}
	assert.Greater(t, len(allObjects), 0, "Entire output should contain at least one valid JSON object")
}

// TestActualLogFile tests that the actual log file (omni-api.log) contains valid JSON
func TestActualLogFile(t *testing.T) {
	// Get current working directory
	execDir, err := os.Getwd()
	require.NoError(t, err, "Should get current working directory")
	
	logPath := filepath.Join(execDir, "omni-api.json")
	
	// Check if log file exists (it may not exist if application hasn't been run)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Skipf("Log file does not exist at %s - skipping test (run application first to generate log file)", logPath)
		return
	}
	
	// Open and read the log file
	file, err := os.Open(logPath)
	require.NoError(t, err, "Should open actual log file")
	defer file.Close()
	
	// Read entire file content
	fileContent, err := io.ReadAll(file)
	require.NoError(t, err, "Should read entire log file")
	
	// Verify file is not empty
	require.NotEmpty(t, fileContent, "Log file should not be empty")
	
	// Split into lines
	content := string(fileContent)
	lines := strings.Split(strings.TrimSpace(content), "\n")
	
	// Verify we have log entries
	require.Greater(t, len(lines), 0, "Log file should contain at least one log line")
	
	// Validate each line individually
	var lineObjects []map[string]interface{}
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}
		
		var jsonObj map[string]interface{}
		err := json.Unmarshal([]byte(line), &jsonObj)
		assert.NoError(t, err, "Line %d in actual log file should be valid JSON: %s", i+1, line)
		
		// Verify required JSON fields exist
		assert.Contains(t, jsonObj, "time", "Line %d should have 'time' field", i+1)
		assert.Contains(t, jsonObj, "level", "Line %d should have 'level' field", i+1)
		assert.Contains(t, jsonObj, "msg", "Line %d should have 'msg' field", i+1)
		
		// Verify level is a valid string
		level, ok := jsonObj["level"].(string)
		assert.True(t, ok, "Line %d 'level' should be a string", i+1)
		assert.Contains(t, []string{"INFO", "DEBUG", "WARN", "ERROR"}, level, "Line %d should have valid log level", i+1)
		
		lineObjects = append(lineObjects, jsonObj)
	}
	
	// Verify the entire file as a whole (NDJSON format)
	trimmedContent := strings.TrimSpace(content)
	allLines := strings.Split(trimmedContent, "\n")
	var fileObjects []map[string]interface{}
	for _, line := range allLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(line), &obj)
		require.NoError(t, err, "Entire log file should be valid NDJSON - failed to parse line: %s", line)
		fileObjects = append(fileObjects, obj)
	}
	
	// Verify consistency
	assert.Greater(t, len(fileObjects), 0, "Entire log file should contain at least one valid JSON object")
	assert.Equal(t, len(lineObjects), len(fileObjects), "Should parse same number of objects line-by-line and as whole file")
	
	// Log summary
	t.Logf("Validated %d log entries in %s", len(fileObjects), logPath)
}
