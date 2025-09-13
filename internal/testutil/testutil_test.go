package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTempWorkspace(t *testing.T) {
	ws := NewTempWorkspace(t)

	// Test basic workspace creation
	if ws.RootDir == "" {
		t.Error("Expected RootDir to be set")
	}

	// Test that workspace directory exists
	AssertDirExists(t, ws.RootDir)
}

func TestCreateAssignment(t *testing.T) {
	ws := NewTempWorkspace(t)

	assignmentPath := "test/assignment-example"
	instructions := "# Test Assignment\nThis is a test assignment with instructions."

	ws.CreateAssignment(assignmentPath, instructions)

	// Check assignment directory exists
	fullPath := filepath.Join(ws.RootDir, assignmentPath)
	AssertDirExists(t, fullPath)

	// Check instructions file exists and has correct content
	instructionsFile := filepath.Join(fullPath, "instructions.md")
	AssertFileExists(t, instructionsFile)

	content := ReadFileContent(t, instructionsFile)
	ContainsString(t, content, "# Test Assignment", "instructions content")
	ContainsString(t, content, "This is a test assignment", "instructions content")
}

func TestCreateAssignmentWithImages(t *testing.T) {
	ws := NewTempWorkspace(t)

	assignmentPath := "test/assignment-with-images"
	instructions := "# Assignment with Images\n![Test](static/test.png)"
	images := []string{"test.png", "diagram.jpg"}

	ws.CreateAssignmentWithImages(assignmentPath, instructions, images)

	// Check assignment directory
	fullPath := filepath.Join(ws.RootDir, assignmentPath)
	AssertDirExists(t, fullPath)

	// Check static directory
	staticPath := filepath.Join(fullPath, "static")
	AssertDirExists(t, staticPath)

	// Check images exist
	for _, img := range images {
		imgPath := filepath.Join(staticPath, img)
		AssertFileExists(t, imgPath)
	}

	// Check instructions content
	instructionsFile := filepath.Join(fullPath, "instructions.md")
	content := ReadFileContent(t, instructionsFile)
	ContainsString(t, content, "![Test](static/test.png)", "image reference")
}

func TestCreateStandardStructure(t *testing.T) {
	ws := NewTempWorkspace(t)
	ws.CreateStandardStructure()

	// Test that various assignment types are created
	expectedDirs := []string{
		"test/fixtures/assignments/assignment-1",
		"test/fixtures/assignments/assignment-2",
		"test/fixtures/homework/hw-1",
		"test/fixtures/homework/hw-2",
		"test/fixtures/labs/lab-1",
		"test/fixtures/projects/project-1",
		"test/fixtures/courses/CS101/week-01/assignment-fibonacci",
		"test/fixtures/courses/CS102/week-02/assignment-sorting",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(ws.RootDir, dir)
		AssertDirExists(t, fullPath)

		// Check instructions.md exists
		instructionsFile := filepath.Join(fullPath, "instructions.md")
		AssertFileExists(t, instructionsFile)
	}

	// Test specific content for assignment with images
	project1Path := filepath.Join(ws.RootDir, "test/fixtures/projects/project-1")
	staticPath := filepath.Join(project1Path, "static")
	AssertDirExists(t, staticPath)

	// Check specific images exist
	AssertFileExists(t, filepath.Join(staticPath, "architecture.png"))
	AssertFileExists(t, filepath.Join(staticPath, "flowchart.png"))
}

func TestChangeToWorkspace(t *testing.T) {
	ws := NewTempWorkspace(t)

	// Remember original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get original directory: %v", err)
	}

	// Change to workspace and test
	restore := ws.ChangeToWorkspace()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if currentDir != ws.RootDir {
		t.Errorf("Expected to be in workspace dir %s, but in %s", ws.RootDir, currentDir)
	}

	// Restore and test
	restore()

	restoredDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get restored directory: %v", err)
	}

	if restoredDir != originalDir {
		t.Errorf("Expected to be restored to %s, but in %s", originalDir, restoredDir)
	}
}

func TestEnvSetup(t *testing.T) {
	env := NewEnvSetup()

	// Set a test variable
	testKey := "TEST_VARIABLE_FOR_TESTING"
	testValue := "test-value"
	originalValue := os.Getenv(testKey)

	env.Set(testKey, testValue)

	// Check it was set
	if os.Getenv(testKey) != testValue {
		t.Errorf("Expected %s to be %s, but got %s", testKey, testValue, os.Getenv(testKey))
	}

	// Restore and check
	env.Restore()

	restoredValue := os.Getenv(testKey)
	if restoredValue != originalValue {
		t.Errorf("Expected %s to be restored to %s, but got %s", testKey, originalValue, restoredValue)
	}
}

func TestSetTestDefaults(t *testing.T) {
	env := NewEnvSetup()
	defer env.Restore()

	env.SetTestDefaults()

	expectedVars := map[string]string{
		"GITHUB_TOKEN":      "test-token",
		"GITHUB_REPOSITORY": "test/repo",
		"DEFAULT_BRANCH":    "main",
		"DRY_RUN":           "true",
	}

	for key, expectedValue := range expectedVars {
		actualValue := os.Getenv(key)
		if actualValue != expectedValue {
			t.Errorf("Expected %s to be %s, but got %s", key, expectedValue, actualValue)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		content  string
		substr   string
		expected bool
		name     string
	}{
		{"hello world", "world", true, "basic substring"},
		{"hello world", "hello", true, "substring at start"},
		{"hello world", "hello world", true, "exact match"},
		{"hello world", "xyz", false, "not found"},
		{"", "test", false, "empty content"},
		{"test", "", false, "empty substring"},
		{"", "", false, "both empty"},
		{"TEST", "test", false, "case sensitive"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Contains(test.content, test.substr)
			if result != test.expected {
				t.Errorf("Contains(%q, %q) = %v, expected %v",
					test.content, test.substr, result, test.expected)
			}
		})
	}
}

func TestSampleInstructions(t *testing.T) {
	// Test that sample instructions are available
	if len(SampleInstructions) == 0 {
		t.Error("Expected SampleInstructions to have content")
	}

	// Test specific samples
	expectedSamples := []string{"basic", "with-images", "complex"}
	for _, sample := range expectedSamples {
		if instructions, exists := SampleInstructions[sample]; !exists {
			t.Errorf("Expected sample %s to exist", sample)
		} else if instructions == "" {
			t.Errorf("Expected sample %s to have content", sample)
		}
	}

	// Test that image sample contains image references
	withImages := SampleInstructions["with-images"]
	ContainsString(t, withImages, "![Overview diagram](static/overview.png)", "image reference")
	ContainsString(t, withImages, "![Process workflow](static/workflow.png)", "image reference")
}

func TestAssertHelpers(t *testing.T) {
	ws := NewTempWorkspace(t)

	// Create a test file and directory for validation
	testFile := filepath.Join(ws.RootDir, "test.txt")
	testDir := filepath.Join(ws.RootDir, "testdir")

	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test AssertFileExists - should not fail
	AssertFileExists(t, testFile)

	// Test AssertDirExists - should not fail
	AssertDirExists(t, testDir)

	// Test ReadFileContent
	content := ReadFileContent(t, testFile)
	if content != "test content" {
		t.Errorf("Expected 'test content', got %s", content)
	}

	// Test ContainsString
	ContainsString(t, content, "test", "file content")
}
