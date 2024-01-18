package test

import (
	"jit/internal"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInitialBranchSetupWithValidRepoRoot(t *testing.T) {
	// Create a temporary directory to simulate the repository root.
	tempDir, tempDirErr := os.MkdirTemp("", "repo")
	if tempDirErr != nil {
		t.Fatalf("Failed to create temporary directory: %v", tempDirErr)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir) // Clean up after the test.

	// Create the 'branches' directory with proper directory permissions.
	if mkDirErr := os.Mkdir(filepath.Join(tempDir, "branches"), 0755); mkDirErr != nil {
		t.Fatalf("Failed to create branches directory: %v", mkDirErr)
	}

	// Set up the initial branch and check for errors.
	_, err := internal.SetUpInitialBranch(tempDir, "main")
	if err != nil {
		t.Fatalf("SetUpInitialBranch failed: %v", err)
	}

	// Check if the 'main' branch file was created.
	if _, infoErr := os.Stat(filepath.Join(tempDir, "branches", "main")); infoErr != nil {
		if os.IsNotExist(infoErr) {
			t.Errorf("Expected 'main' branch file to exist, but it does not.")
		} else {
			t.Fatalf("Error checking 'main' branch file: %v", infoErr)
		}
	}

	// Read and check the content of the 'head' file.
	content, readErr := os.ReadFile(filepath.Join(tempDir, "head"))
	if readErr != nil {
		t.Fatalf("Failed to read head file: %v", readErr)
	}
	expectedPath := filepath.Join(tempDir, "branches", "main")
	if string(content) != expectedPath {
		t.Errorf("Expected head content to be '%s', got '%s'", expectedPath, string(content))
	}
}

func TestWriteToConfigFileWithValidRepoRoot(t *testing.T) {
	// Create a temporary directory to simulate the repository root.
	tempDir, tempDirErr := os.MkdirTemp("", "repo")
	if tempDirErr != nil {
		t.Fatalf("Failed to create temporary directory: %v", tempDirErr)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir) // Clean up after the test.

	config := map[string]string{
		"TEMPLATE": "/usr/template",
		"BRANCH":   "main",
	}

	_, err := internal.WriteToConfigFile(config, tempDir)
	if err != nil {
		t.Fatalf("WriteToConfigFile failed: %v", err)
	}

	content, readErr := os.ReadFile(filepath.Join(tempDir, "config"))
	if readErr != nil {
		t.Fatalf("Failed to read config file: %v", readErr)
	}

	// Checking for the exact key-value pair format.
	expectedContent := []string{"TEMPLATE=/usr/template", "BRANCH=main"}
	for _, expected := range expectedContent {
		if !strings.Contains(string(content), expected) {
			t.Errorf("Expected config to contain '%s', but it was not found", expected)
		}
	}
}

func TestCreateJitDirWithSeparateBareDir(t *testing.T) {
	// Create a temporary directory to simulate the repository root.
	tempDir, tempDirErr := os.MkdirTemp("", "repo")
	if tempDirErr != nil {
		t.Fatalf("Failed to create temporary directory: %v", tempDirErr)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir) // Clean up after the test.

	tests := []struct {
		wkDir  string
		sepDir bool
		bare   bool
		desc   string
	}{
		{tempDir, true, true, "SeparateAndBare"},
		{tempDir, true, false, "SeparateAndNotBare"},
		{tempDir, false, true, "NotSeparateAndBare"},
		{tempDir, false, false, "NotSeparateAndNotBare"},
	}

	dirContents := []string{"main", "head", "stage", "config", "logs", "info", "branches", "snapshots", "objects"}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, err := internal.CreateJitDir(test.wkDir, test.sepDir, test.bare, 0755)
			if err != nil {
				t.Fatalf("CreateJitDir failed: %v", err)
			}

			expectedDir := tempDir
			if !test.bare && !test.sepDir {
				expectedDir = filepath.Join(tempDir, ".jit")
			}

			for _, content := range dirContents {
				if _, err := os.Stat(filepath.Join(expectedDir, content)); err != nil {
					t.Fatalf("%s does not exist in %s directory", content, expectedDir)
				}
			}
		})
	}
}

func TestValidateDirPath(t *testing.T) {
	tempDir, tempDirErr := os.MkdirTemp("", "test")
	if tempDirErr != nil {
		t.Fatalf("Failed to create temporary directory: %v", tempDirErr)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir)

	tests := []struct {
		dir   string
		valid bool
		desc  string
	}{
		{tempDir, true, "validTemporalDir"},
		{".", true, "validCurrentDir"},
		{"..", true, "validParentDir"},
		{"/path/to/nonexistent/dir/", false, "nonExistentDir"}, // More obviously invalid path
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			err := internal.ValidateDirPath(test.dir)
			if (err == nil) != test.valid {
				t.Errorf("TestValidateDirPath %s: expected %v, got %v", test.desc, test.valid, err != nil)
			}
		})
	}
}

func TestCheckWritePermission(t *testing.T) {
	// Test with a writable directory
	tempDir, err := os.MkdirTemp("", "testDir")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir) // Cleanup for writable directory

	err = internal.CheckWritePermission(tempDir)
	if err != nil {
		t.Errorf("Expected writable directory to pass, got error: %v", err)
	}

	// Test with a non-existent directory
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	err = internal.CheckWritePermission(nonExistentDir)
	if err == nil {
		t.Errorf("Expected non-existent directory to fail, but got no error")
	}

	// Test with a non-writable directory (Unix-like systems only)
	if runtime.GOOS != "windows" {
		nonWritableDir := filepath.Join(tempDir, "nonWritable")
		if mkDirErr := os.Mkdir(nonWritableDir, 0555); mkDirErr != nil { // Read and execute permissions only
			t.Fatalf("Failed to create non-writable directory: %v", mkDirErr)
		}
		defer func(name string, mode os.FileMode) {
			_ = os.Chmod(name, mode)
		}(nonWritableDir, 0755) // Restore permissions for cleanup
		defer func(path string) {
			_ = os.RemoveAll(path)
		}(nonWritableDir) // Cleanup for non-writable directory

		err = internal.CheckWritePermission(nonWritableDir)
		if err == nil {
			t.Errorf("Expected non-writable directory to fail, but got no error")
		}
	}
}

func TestGetJitRootDir(t *testing.T) {
	// Set up a valid temporary directory
	tempDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir) // Cleanup for temporary directory

	// Define the test table
	tests := []struct {
		name        string
		dirPath     string
		expectError bool
		expectedDir string
	}{
		{
			name:        "Valid Directory",
			dirPath:     tempDir,
			expectError: false,
			expectedDir: tempDir,
		},
		{
			name:        "Non-existent Directory",
			dirPath:     filepath.Join(tempDir, "nonexistent"),
			expectError: true,
		},
		{
			name:        "Empty Directory Path",
			dirPath:     "",
			expectError: false,
			expectedDir: func() string { dir, _ := os.Getwd(); return dir }(), // Get current working directory
		},

		{
			name:        "Dot Directory Path",
			dirPath:     ".",
			expectError: false,
			expectedDir: ".", // Get current working directory
		},
	}

	// Run the test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rootDir, err := internal.GetJitRootDir(tc.dirPath)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got %v", err)
				} else if rootDir != tc.expectedDir {
					t.Errorf("Expected rootDir to be %v, got %v", tc.expectedDir, rootDir)
				}
			}
		})
	}
}

func TestConstructFinalJitDir(t *testing.T) {
	// Set up a working directory for the test
	wkDir, err := os.MkdirTemp("", "testWkDir")
	if err != nil {
		t.Fatalf("Failed to create working directory: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(wkDir) // Cleanup for working directory

	// Define the test table
	tests := []struct {
		name     string
		wkDir    string
		sepDir   string
		bare     bool
		expected string
	}{
		{
			name:     "Separate Directory",
			wkDir:    wkDir,
			sepDir:   "/path/to/sep/dir",
			bare:     false,
			expected: "/path/to/sep/dir",
		},
		{
			name:     "Bare Repository",
			wkDir:    wkDir,
			sepDir:   "",
			bare:     true,
			expected: wkDir,
		},
		{
			name:     "Non-Bare Repository",
			wkDir:    wkDir,
			sepDir:   "",
			bare:     false,
			expected: filepath.Join(wkDir, ".jit"),
		},
	}

	// Run the test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			finalDir := internal.ConstructFinalJitDir(tc.wkDir, tc.sepDir, tc.bare)
			if finalDir != tc.expected {
				t.Errorf("Test %s: expected %v, got %v", tc.name, tc.expected, finalDir)
			}
		})
	}
}

func TestInitializeJitRepository(t *testing.T) {
	wkDir, err := os.MkdirTemp("", "testWkDir")
	if err != nil {
		t.Fatalf("Failed to create working directory: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(wkDir) // Cleanup for working directory

	sepDir, err := os.MkdirTemp("", "sepDir")
	if err != nil {
		t.Fatalf("Failed to create working directory: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(sepDir) // Cleanup for working directory
	// Define test cases
	tests := []struct {
		name    string
		options map[string]any
		dir     string
		wantErr bool
	}{
		{
			name: "Standard Repository",
			options: map[string]any{
				"quiet":            false,
				"bare":             false,
				"separate-jit-dir": "",
				"initial-branch":   "main",
				"perm":             "0755",
			},
			dir:     "",
			wantErr: false,
		},
		{
			name: "Separate Directory Repository",
			options: map[string]any{
				"separate-jit-dir": sepDir,
				"initial-branch":   "develop",
				"perm":             "0755",
			},
			dir:     "",
			wantErr: false,
		},
		{
			name: "Quiet Mode Repository",
			options: map[string]any{
				"quiet":          true,
				"initial-branch": "feature-branch",
				"perm":           "0755",
			},
			dir:     "",
			wantErr: false,
		},
		{
			name: "Invalid Permissions",
			options: map[string]any{
				"perm": "invalid",
			},
			dir:     "",
			wantErr: true, // Expecting an error due to invalid permissions format
		},
		{
			name: "Non-existent Separate Directory",
			options: map[string]any{
				"separate-jit-dir": "/non/existent/path",
				"perm":             "0755",
			},
			dir:     "",
			wantErr: true, // Expecting an error due to non-existent separate directory
		},
		{
			name: "Repository With Template",
			options: map[string]any{
				"template":       wkDir,
				"initial-branch": "template-branch",
				"perm":           "0755",
			},
			dir:     "",
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "testrepo")
			if err != nil {
				t.Fatalf("Failed to create temporary directory: %v", err)
			}
			defer func(path string) {
				_ = os.RemoveAll(path)
			}(tempDir)

			tc.dir = tempDir // Use the temporary directory as the working directory

			ok, err := internal.InitializeJitRepository(tc.options, tc.dir)
			if (err != nil) != tc.wantErr {
				t.Errorf("Test %s: InitializeJitRepository() error = %v, wantErr %v", tc.name, err, tc.wantErr)
				return
			}
			if !ok && !tc.wantErr {
				t.Errorf("Test %s: InitializeJitRepository() was not successful", tc.name)
			}

			// Add additional checks here to verify the creation of files and directories
		})
	}
}
