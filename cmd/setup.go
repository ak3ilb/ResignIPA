package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
	colorPurple = "\033[0;35m"
	colorCyan   = "\033[0;36m"
)

// SetupChecker encapsulates all setup validation and verification logic
type SetupChecker struct {
	hasErrors     bool
	output        []string
	requiredTools map[string]ToolRequirement
	optionalTools map[string]ToolRequirement
	certificates  []Certificate
	systemInfo    SystemInfo
}

// ToolRequirement represents a required or optional system tool
type ToolRequirement struct {
	Name        string
	Command     string
	CheckFunc   func() (bool, string, error)
	InstallHelp string
	Critical    bool
}

// Certificate represents a code signing certificate
type Certificate struct {
	Hash string
	Name string
	Type string
}

// SystemInfo contains system configuration details
type SystemInfo struct {
	OS           string
	Architecture string
	GoVersion    string
	XcodePath    string
	CertCount    int
	WorkingDir   string
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Verify system prerequisites and setup environment",
	Long: `Comprehensive system verification tool that checks:
- Operating system compatibility
- Required development tools (Go, Xcode)
- Code signing tools (codesign, security, PlistBuddy)
- Available signing certificates
- Project dependencies

This command performs a complete environment audit and provides
actionable feedback for any missing components.`,
	Run: func(cmd *cobra.Command, args []string) {
		checker := NewSetupChecker()
		if err := checker.ExecuteFullSetup(); err != nil {
			fmt.Printf("%s✗ Setup failed: %v%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

// NewSetupChecker creates and initializes a new setup checker instance
func NewSetupChecker() *SetupChecker {
	checker := &SetupChecker{
		hasErrors:     false,
		output:        make([]string, 0),
		requiredTools: make(map[string]ToolRequirement),
		optionalTools: make(map[string]ToolRequirement),
		certificates:  make([]Certificate, 0),
	}

	checker.initializeToolRequirements()
	return checker
}

// initializeToolRequirements sets up the tool verification matrix
func (sc *SetupChecker) initializeToolRequirements() {
	sc.requiredTools = map[string]ToolRequirement{
		"go": {
			Name:    "Go",
			Command: "go",
			CheckFunc: func() (bool, string, error) {
				cmd := exec.Command("go", "version")
				output, err := cmd.CombinedOutput()
				if err != nil {
					return false, "", err
				}
				return true, strings.TrimSpace(string(output)), nil
			},
			InstallHelp: "Install from: https://golang.org/dl/ or run: brew install go",
			Critical:    true,
		},
		"xcode-select": {
			Name:    "Xcode Command Line Tools",
			Command: "xcode-select",
			CheckFunc: func() (bool, string, error) {
				cmd := exec.Command("xcode-select", "-p")
				output, err := cmd.CombinedOutput()
				if err != nil {
					return false, "", err
				}
				return true, strings.TrimSpace(string(output)), nil
			},
			InstallHelp: "Run: xcode-select --install",
			Critical:    true,
		},
		"codesign": {
			Name:    "codesign",
			Command: "codesign",
			CheckFunc: func() (bool, string, error) {
				cmd := exec.Command("which", "codesign")
				output, err := cmd.CombinedOutput()
				if err != nil {
					return false, "", err
				}
				return true, strings.TrimSpace(string(output)), nil
			},
			InstallHelp: "Part of Xcode Command Line Tools",
			Critical:    true,
		},
		"security": {
			Name:    "security",
			Command: "security",
			CheckFunc: func() (bool, string, error) {
				cmd := exec.Command("which", "security")
				output, err := cmd.CombinedOutput()
				if err != nil {
					return false, "", err
				}
				return true, strings.TrimSpace(string(output)), nil
			},
			InstallHelp: "Part of macOS system tools",
			Critical:    true,
		},
		"plistbuddy": {
			Name:    "PlistBuddy",
			Command: "/usr/libexec/PlistBuddy",
			CheckFunc: func() (bool, string, error) {
				if _, err := os.Stat("/usr/libexec/PlistBuddy"); os.IsNotExist(err) {
					return false, "", err
				}
				return true, "/usr/libexec/PlistBuddy", nil
			},
			InstallHelp: "Part of Xcode Command Line Tools",
			Critical:    true,
		},
	}
}

// ExecuteFullSetup runs the complete setup verification process
func (sc *SetupChecker) ExecuteFullSetup() error {
	sc.printHeader()

	// Phase 1: System Information
	if err := sc.gatherSystemInfo(); err != nil {
		return fmt.Errorf("failed to gather system info: %w", err)
	}
	sc.displaySystemInfo()

	// Phase 2: Prerequisites Check
	sc.printSection("Checking Prerequisites")
	sc.verifyOperatingSystem()
	sc.verifyRequiredTools()

	if sc.hasErrors {
		sc.printSection("Setup Failed")
		sc.displayErrorSummary()
		return fmt.Errorf("prerequisites check failed")
	}

	// Phase 3: Project Dependencies
	sc.printSection("Managing Project Dependencies")
	if err := sc.downloadDependencies(); err != nil {
		sc.logError("Failed to download dependencies: %v", err)
		return err
	}

	if err := sc.tidyDependencies(); err != nil {
		sc.logWarning("go mod tidy had issues (may be acceptable): %v", err)
	}

	// Phase 4: Build
	sc.printSection("Building Project")
	binaryPath, err := sc.buildProject()
	if err != nil {
		sc.logError("Build failed: %v", err)
		return err
	}
	sc.logSuccess("Build successful: %s", binaryPath)

	// Phase 5: Certificate Discovery
	sc.printSection("Discovering Signing Certificates")
	sc.discoverCertificates()

	// Phase 6: Final Summary
	sc.printFinalSummary(binaryPath)

	return nil
}

// gatherSystemInfo collects system configuration details
func (sc *SetupChecker) gatherSystemInfo() error {
	sc.systemInfo.OS = runtime.GOOS
	sc.systemInfo.Architecture = runtime.GOARCH

	// Get Go version
	cmd := exec.Command("go", "version")
	if output, err := cmd.CombinedOutput(); err == nil {
		sc.systemInfo.GoVersion = strings.TrimSpace(string(output))
	}

	// Get Xcode path
	cmd = exec.Command("xcode-select", "-p")
	if output, err := cmd.CombinedOutput(); err == nil {
		sc.systemInfo.XcodePath = strings.TrimSpace(string(output))
	}

	// Get working directory
	if wd, err := os.Getwd(); err == nil {
		sc.systemInfo.WorkingDir = wd
	}

	return nil
}

// displaySystemInfo prints collected system information
func (sc *SetupChecker) displaySystemInfo() {
	sc.printSection("System Information")
	sc.logInfo("Operating System: %s", sc.systemInfo.OS)
	sc.logInfo("Architecture: %s", sc.systemInfo.Architecture)
	if sc.systemInfo.GoVersion != "" {
		sc.logInfo("Go Version: %s", sc.systemInfo.GoVersion)
	}
	if sc.systemInfo.XcodePath != "" {
		sc.logInfo("Xcode Path: %s", sc.systemInfo.XcodePath)
	}
	sc.logInfo("Working Directory: %s", sc.systemInfo.WorkingDir)
	fmt.Println()
}

// verifyOperatingSystem ensures the system is running macOS
func (sc *SetupChecker) verifyOperatingSystem() {
	if runtime.GOOS != "darwin" {
		sc.logError("Not running on macOS. ResignIPA requires macOS for code signing tools.")
		sc.hasErrors = true
		return
	}
	sc.logSuccess("Running on macOS")
}

// verifyRequiredTools checks for all required system tools
func (sc *SetupChecker) verifyRequiredTools() {
	for _, tool := range sc.requiredTools {
		sc.verifyTool(tool)
	}
}

// verifyTool checks a single tool's availability
func (sc *SetupChecker) verifyTool(tool ToolRequirement) {
	exists, info, err := tool.CheckFunc()

	if !exists || err != nil {
		sc.logError("%s is not installed", tool.Name)
		if tool.InstallHelp != "" {
			sc.logWarning("  Install: %s", tool.InstallHelp)
		}
		if tool.Critical {
			sc.hasErrors = true
		}
		return
	}

	sc.logSuccess("%s is installed", tool.Name)
	if info != "" && len(info) < 100 {
		sc.logInfo("  Location: %s", info)
	}
}

// downloadDependencies runs go mod download
func (sc *SetupChecker) downloadDependencies() error {
	sc.logInfo("Downloading Go dependencies...")

	cmd := exec.Command("go", "mod", "download")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod download failed: %w", err)
	}

	sc.logSuccess("Dependencies downloaded successfully")
	return nil
}

// tidyDependencies runs go mod tidy
func (sc *SetupChecker) tidyDependencies() error {
	sc.logInfo("Tidying dependencies...")

	cmd := exec.Command("go", "mod", "tidy")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("go mod tidy failed: %s", string(output))
	}

	sc.logSuccess("Dependencies tidied")
	return nil
}

// buildProject compiles the project
func (sc *SetupChecker) buildProject() (string, error) {
	sc.logInfo("Compiling ResignIPA...")

	outputPath := "resignipa"
	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", outputPath, "main.go")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build failed: %s", string(output))
	}

	// Get binary info
	if stat, err := os.Stat(outputPath); err == nil {
		sizeMB := float64(stat.Size()) / (1024 * 1024)
		sc.logInfo("  Binary size: %.2f MB", sizeMB)
	}

	// Get absolute path
	absPath, _ := filepath.Abs(outputPath)
	return absPath, nil
}

// discoverCertificates finds available code signing certificates
func (sc *SetupChecker) discoverCertificates() {
	cmd := exec.Command("security", "find-identity", "-v", "-p", "codesigning")
	output, err := cmd.CombinedOutput()

	if err != nil {
		sc.logWarning("Could not query certificates: %v", err)
		return
	}

	lines := strings.Split(string(output), "\n")
	certCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Apple Development") || strings.Contains(line, "Apple Distribution") {
			certCount++
			if certCount <= 5 {
				sc.logSuccess("Found: %s", line)
			}
		}
	}

	if certCount == 0 {
		sc.logWarning("No signing certificates found")
		sc.logInfo("  Get certificates from: https://developer.apple.com")
	} else {
		sc.logSuccess("Found %d signing certificate(s)", certCount)
		if certCount > 5 {
			sc.logInfo("  (showing first 5, %d more available)", certCount-5)
		}
	}

	sc.systemInfo.CertCount = certCount
}

// printFinalSummary displays the completion summary
func (sc *SetupChecker) printFinalSummary(binaryPath string) {
	fmt.Println()
	sc.printSeparator('═')
	sc.logSuccess("Setup Complete!")
	sc.printSeparator('═')
	fmt.Println()

	fmt.Printf("%sNext Steps:%s\n\n", colorCyan, colorReset)

	fmt.Printf("  %s1. Run GUI mode:%s\n", colorBlue, colorReset)
	fmt.Printf("     %s./resignipa%s\n\n", colorPurple, colorReset)

	fmt.Printf("  %s2. Run CLI mode:%s\n", colorBlue, colorReset)
	fmt.Printf("     %s./resignipa -s /path/to/app.ipa -c \"Certificate Name\"%s\n\n", colorPurple, colorReset)

	fmt.Printf("  %s3. View help:%s\n", colorBlue, colorReset)
	fmt.Printf("     %s./resignipa --help%s\n\n", colorPurple, colorReset)

	fmt.Printf("  %s4. Install system-wide (optional):%s\n", colorBlue, colorReset)
	fmt.Printf("     %ssudo make install%s\n\n", colorPurple, colorReset)

	fmt.Printf("%sHappy Resigning! 🎉%s\n", colorGreen, colorReset)
}

// Logging methods with color support

func (sc *SetupChecker) printHeader() {
	fmt.Println()
	sc.printSeparator('═')
	fmt.Printf("%s🚀 ResignIPA Setup Wizard%s\n", colorCyan, colorReset)
	sc.printSeparator('═')
	fmt.Println()
}

func (sc *SetupChecker) printSection(title string) {
	fmt.Println()
	sc.printSeparator('─')
	fmt.Printf("%s%s%s\n", colorCyan, title, colorReset)
	sc.printSeparator('─')
	fmt.Println()
}

func (sc *SetupChecker) printSeparator(char rune) {
	fmt.Println(strings.Repeat(string(char), 70))
}

func (sc *SetupChecker) logSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s✓%s %s\n", colorGreen, colorReset, msg)
}

func (sc *SetupChecker) logError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s✗%s %s\n", colorRed, colorReset, msg)
	sc.output = append(sc.output, fmt.Sprintf("ERROR: %s", msg))
}

func (sc *SetupChecker) logWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s⚠%s %s\n", colorYellow, colorReset, msg)
	sc.output = append(sc.output, fmt.Sprintf("WARNING: %s", msg))
}

func (sc *SetupChecker) logInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s\n", msg)
}

func (sc *SetupChecker) displayErrorSummary() {
	fmt.Println()
	sc.logError("Prerequisites check failed. Please install missing components:")
	fmt.Println()

	for _, tool := range sc.requiredTools {
		exists, _, err := tool.CheckFunc()
		if !exists || err != nil {
			fmt.Printf("  • %s%s%s\n", colorYellow, tool.Name, colorReset)
			if tool.InstallHelp != "" {
				fmt.Printf("    %s\n", tool.InstallHelp)
			}
		}
	}
	fmt.Println()
}
