package git

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestNewCommander tests git commander creation
func TestNewCommander(t *testing.T) {
	tests := []struct {
		name   string
		dryRun bool
	}{
		{
			name:   "dry run enabled",
			dryRun: true,
		},
		{
			name:   "dry run disabled",
			dryRun: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commander := NewCommander(tt.dryRun)
			if commander == nil {
				t.Error("Expected commander to be created")
				return
			}
			if commander.dryRun != tt.dryRun {
				t.Errorf("Expected dryRun=%t, got=%t", tt.dryRun, commander.dryRun)
			}
		})
	}
}

// TestRunCommand tests command execution
func TestRunCommand(t *testing.T) {
	tests := []struct {
		name        string
		dryRun      bool
		command     string
		description string
		expectError bool
	}{
		{
			name:        "dry run mode - always succeeds",
			dryRun:      true,
			command:     "invalid-command-that-does-not-exist",
			description: "Testing dry run",
			expectError: false,
		},
		{
			name:        "valid command",
			dryRun:      false,
			command:     "echo 'test'",
			description: "Echo test",
			expectError: false,
		},
		{
			name:        "invalid command",
			dryRun:      false,
			command:     "invalid-command-that-does-not-exist",
			description: "Invalid command test",
			expectError: true,
		},
		{
			name:        "empty description",
			dryRun:      false,
			command:     "echo 'test'",
			description: "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commander := NewCommander(tt.dryRun)
			err := commander.RunCommand(tt.command, tt.description)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestRunCommandWithOutput tests command execution with output capture
func TestRunCommandWithOutput(t *testing.T) {
	tests := []struct {
		name           string
		dryRun         bool
		command        string
		description    string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "dry run mode - returns empty",
			dryRun:         true,
			command:        "echo 'test'",
			description:    "Testing dry run",
			expectedOutput: "",
			expectError:    false,
		},
		{
			name:           "echo command",
			dryRun:         false,
			command:        "echo 'hello world'",
			description:    "Echo test",
			expectedOutput: "hello world",
			expectError:    false,
		},
		{
			name:           "invalid command",
			dryRun:         false,
			command:        "invalid-command-that-does-not-exist",
			description:    "Invalid command test",
			expectedOutput: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commander := NewCommander(tt.dryRun)
			output, err := commander.RunCommandWithOutput(tt.command, tt.description)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// For non-dry-run mode, check output content
			if !tt.dryRun && !tt.expectError {
				if !strings.Contains(output, tt.expectedOutput) {
					t.Errorf("Expected output to contain '%s', got '%s'", tt.expectedOutput, output)
				}
			}

			// For dry-run mode, output should always be empty
			if tt.dryRun && output != "" {
				t.Errorf("Expected empty output in dry-run mode, got '%s'", output)
			}
		})
	}
}

// TestNewOperations tests git operations creation
func TestNewOperations(t *testing.T) {
	tests := []struct {
		name   string
		dryRun bool
	}{
		{
			name:   "dry run enabled",
			dryRun: true,
		},
		{
			name:   "dry run disabled",
			dryRun: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := NewOperations(tt.dryRun)
			if ops == nil {
				t.Error("Expected operations to be created")
				return
			}
			if ops.commander.dryRun != tt.dryRun {
				t.Errorf("Expected dryRun=%t, got=%t", tt.dryRun, ops.commander.dryRun)
			}
		})
	}
}

// TestCreateAndSwitchToBranch tests branch creation and switching
func TestCreateAndSwitchToBranch(t *testing.T) {
	// This test requires a git repository to be present
	// We'll test both dry-run and error cases
	tests := []struct {
		name        string
		dryRun      bool
		branchName  string
		expectError bool
	}{
		{
			name:        "dry run mode",
			dryRun:      true,
			branchName:  "test-branch",
			expectError: false,
		},
		{
			name:        "invalid branch name",
			dryRun:      true,
			branchName:  "",
			expectError: false, // Dry run doesn't validate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := NewOperations(tt.dryRun)
			err := ops.CreateAndSwitchToBranch(tt.branchName)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestSwitchToBranch tests branch switching
func TestSwitchToBranch(t *testing.T) {
	tests := []struct {
		name        string
		dryRun      bool
		branchName  string
		expectError bool
	}{
		{
			name:        "dry run mode",
			dryRun:      true,
			branchName:  "main",
			expectError: false,
		},
		{
			name:        "dry run with empty branch",
			dryRun:      true,
			branchName:  "",
			expectError: false, // Dry run doesn't validate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := NewOperations(tt.dryRun)
			err := ops.SwitchToBranch(tt.branchName)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestAddFile tests file addition to git
func TestAddFile(t *testing.T) {
	tests := []struct {
		name        string
		dryRun      bool
		filePath    string
		expectError bool
	}{
		{
			name:        "dry run mode",
			dryRun:      true,
			filePath:    "README.md",
			expectError: false,
		},
		{
			name:        "dry run with nonexistent file",
			dryRun:      true,
			filePath:    "nonexistent.txt",
			expectError: false, // Dry run doesn't validate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := NewOperations(tt.dryRun)
			err := ops.AddFile(tt.filePath)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestCommit tests git commit
func TestCommit(t *testing.T) {
	tests := []struct {
		name        string
		dryRun      bool
		message     string
		expectError bool
	}{
		{
			name:        "dry run mode",
			dryRun:      true,
			message:     "Test commit",
			expectError: false,
		},
		{
			name:        "dry run with empty message",
			dryRun:      true,
			message:     "",
			expectError: false, // Dry run doesn't validate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := NewOperations(tt.dryRun)
			err := ops.Commit(tt.message)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestPushAllBranches tests branch pushing
func TestPushAllBranches(t *testing.T) {
	tests := []struct {
		name        string
		dryRun      bool
		expectError bool
	}{
		{
			name:        "dry run mode",
			dryRun:      true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := NewOperations(tt.dryRun)
			err := ops.PushAllBranches()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestGetLocalBranches tests local branch detection
func TestGetLocalBranches(t *testing.T) {
	tests := []struct {
		name        string
		dryRun      bool
		expectError bool
	}{
		{
			name:        "dry run mode",
			dryRun:      true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := NewOperations(tt.dryRun)
			branches, err := ops.GetLocalBranches()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// In dry-run mode, should return empty map
			if tt.dryRun && len(branches) != 0 {
				t.Errorf("Expected empty branches in dry-run mode, got %d", len(branches))
			}
		})
	}
}

// TestGetRemoteBranches tests remote branch handling
func TestGetRemoteBranches(t *testing.T) {
	tests := []struct {
		name          string
		dryRun        bool
		defaultBranch string
		expectError   bool
	}{
		{
			name:          "dry run mode",
			dryRun:        true,
			defaultBranch: "main",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := NewOperations(tt.dryRun)
			err := ops.GetRemoteBranches(tt.defaultBranch)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// BenchmarkRunCommand benchmarks command execution
func BenchmarkRunCommand(b *testing.B) {
	commander := NewCommander(true) // Use dry-run for consistent timing
	command := "echo 'benchmark test'"
	description := "Benchmark test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = commander.RunCommand(command, description)
	}
}

// BenchmarkRunCommandWithOutput benchmarks command execution with output
func BenchmarkRunCommandWithOutput(b *testing.B) {
	commander := NewCommander(true) // Use dry-run for consistent timing
	command := "echo 'benchmark test'"
	description := "Benchmark test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = commander.RunCommandWithOutput(command, description)
	}
}

// Integration test that requires an actual git repository
func TestGitIntegration(t *testing.T) {
	// Skip this test if we're not in a git repository
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git not available, skipping integration test")
	}

	// Check if we're in a git repository
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		t.Skip("Not in a git repository, skipping integration test")
	}

	// This test only runs with dry-run to avoid affecting the repository
	ops := NewOperations(true)

	t.Run("get local branches in dry-run", func(t *testing.T) {
		branches, err := ops.GetLocalBranches()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		// In dry-run mode, this should return empty map
		if len(branches) != 0 {
			t.Errorf("Expected empty branches in dry-run mode, got %d", len(branches))
		}
	})

	t.Run("fetch all in dry-run", func(t *testing.T) {
		err := ops.FetchAll()
		if err != nil {
			t.Errorf("Unexpected error in dry-run mode: %v", err)
		}
	})

	t.Run("get remote branches in dry-run", func(t *testing.T) {
		err := ops.GetRemoteBranches("main")
		if err != nil {
			t.Errorf("Unexpected error in dry-run mode: %v", err)
		}
	})
}
