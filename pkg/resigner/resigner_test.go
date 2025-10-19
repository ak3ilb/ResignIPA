package resigner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewResigner(t *testing.T) {
	config := Config{
		SourceIPA:   "test.ipa",
		Certificate: "Test Certificate",
	}

	var messages []string
	callback := func(msg string) {
		messages = append(messages, msg)
	}

	r := NewResigner(config, callback)

	if r == nil {
		t.Fatal("NewResigner returned nil")
	}

	if r.config.SourceIPA != config.SourceIPA {
		t.Errorf("Expected SourceIPA %s, got %s", config.SourceIPA, r.config.SourceIPA)
	}

	if r.config.Certificate != config.Certificate {
		t.Errorf("Expected Certificate %s, got %s", config.Certificate, r.config.Certificate)
	}

	if r.callback == nil {
		t.Error("Callback should not be nil")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				SourceIPA:   "testdata/test.ipa",
				Certificate: "Apple Development",
			},
			wantErr: true, // File doesn't exist, so should error
		},
		{
			name: "missing source",
			config: Config{
				Certificate: "Apple Development",
			},
			wantErr: true,
		},
		{
			name: "missing certificate",
			config: Config{
				SourceIPA: "test.ipa",
			},
			wantErr: true,
		},
		{
			name: "nonexistent file",
			config: Config{
				SourceIPA:   "/nonexistent/file.ipa",
				Certificate: "Apple Development",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResigner(tt.config, nil)
			err := r.validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	content := []byte("test content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Copy file
	dstPath := filepath.Join(tmpDir, "dest.txt")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile() failed: %v", err)
	}

	// Verify destination file exists and has same content
	gotContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(gotContent) != string(content) {
		t.Errorf("Content mismatch: got %s, want %s", gotContent, content)
	}
}

func TestCopyDir(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "source")
	os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)

	// Create test files
	testFile1 := filepath.Join(srcDir, "file1.txt")
	testFile2 := filepath.Join(srcDir, "subdir", "file2.txt")
	os.WriteFile(testFile1, []byte("content1"), 0644)
	os.WriteFile(testFile2, []byte("content2"), 0644)

	// Copy directory
	dstDir := filepath.Join(tmpDir, "dest")
	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copyDir() failed: %v", err)
	}

	// Verify files exist
	if _, err := os.Stat(filepath.Join(dstDir, "file1.txt")); os.IsNotExist(err) {
		t.Error("file1.txt not copied")
	}

	if _, err := os.Stat(filepath.Join(dstDir, "subdir", "file2.txt")); os.IsNotExist(err) {
		t.Error("subdir/file2.txt not copied")
	}
}

func TestFindComponents(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "Test.app")
	os.MkdirAll(appDir, 0755)

	// Create test components
	frameworkDir := filepath.Join(appDir, "Frameworks", "Test.framework")
	dylibPath := filepath.Join(appDir, "test.dylib")
	appexDir := filepath.Join(appDir, "PlugIns", "Widget.appex")

	os.MkdirAll(frameworkDir, 0755)
	os.WriteFile(dylibPath, []byte(""), 0644)
	os.MkdirAll(appexDir, 0755)

	// Find components
	components, err := findComponents(appDir)
	if err != nil {
		t.Fatalf("findComponents() failed: %v", err)
	}

	// Check if components were found
	foundFramework := false
	foundDylib := false
	foundAppex := false
	foundApp := false

	for _, comp := range components {
		switch filepath.Ext(comp) {
		case ".framework":
			foundFramework = true
		case ".dylib":
			foundDylib = true
		case ".appex":
			foundAppex = true
		case ".app":
			foundApp = true
		}
	}

	if !foundFramework {
		t.Error("Framework not found")
	}
	if !foundDylib {
		t.Error("Dylib not found")
	}
	if !foundAppex {
		t.Error("Appex not found")
	}
	if !foundApp {
		t.Error("App not found")
	}
}

func TestLogProgress(t *testing.T) {
	var messages []string
	callback := func(msg string) {
		messages = append(messages, msg)
	}

	config := Config{
		SourceIPA:   "test.ipa",
		Certificate: "Test",
	}

	r := NewResigner(config, callback)

	testMsg := "Test progress message"
	r.logProgress(testMsg)

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0] != testMsg {
		t.Errorf("Expected message %s, got %s", testMsg, messages[0])
	}
}

func TestPanicRecovery(t *testing.T) {
	config := Config{
		SourceIPA:   "nonexistent.ipa",
		Certificate: "Test",
	}

	r := NewResigner(config, nil)

	// This should not panic, even though the file doesn't exist
	err := r.Resign()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// Benchmark tests

func BenchmarkCopyFile(b *testing.B) {
	tmpDir := b.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	content := make([]byte, 1024*1024) // 1MB
	os.WriteFile(srcPath, content, 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dstPath := filepath.Join(tmpDir, "dest", "file.txt")
		copyFile(srcPath, dstPath)
	}
}

func BenchmarkFindComponents(b *testing.B) {
	tmpDir := b.TempDir()
	appDir := filepath.Join(tmpDir, "Test.app")
	os.MkdirAll(appDir, 0755)

	// Create some test components
	for i := 0; i < 10; i++ {
		frameworkDir := filepath.Join(appDir, "Frameworks", "Test.framework")
		os.MkdirAll(frameworkDir, 0755)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findComponents(appDir)
	}
}
