package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/resignipa/pkg/resigner"
	"github.com/spf13/cobra"
)

var (
	sourceIPA       string
	certificate     string
	entitlements    string
	mobileProvision string
	bundleID        string
)

var rootCmd = &cobra.Command{
	Use:   "resignipa",
	Short: "ResignIPA - A tool to resign iOS IPA files",
	Long: `ResignIPA is a tool that allows you to resign iOS IPA files with a new certificate,
provisioning profile, bundle identifier, and entitlements.

If no arguments are provided, the GUI will be launched.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no flags are set, launch GUI
		if sourceIPA == "" && certificate == "" {
			LaunchGUI()
			return
		}

		// Otherwise, run CLI mode
		runCLI()
	},
}

var resignCmd = &cobra.Command{
	Use:   "resign",
	Short: "Resign an IPA file",
	Long: `Resign an IPA file with the specified certificate and options.

Example:
  resignipa resign -s /path/to/app.ipa -c "Apple Development: Name" -p /path/to/provision.mobileprovision -b com.example.app`,
	Run: func(cmd *cobra.Command, args []string) {
		runCLI()
	},
}

func init() {
	// Add flags to both root and resign commands
	for _, cmd := range []*cobra.Command{rootCmd, resignCmd} {
		cmd.Flags().StringVarP(&sourceIPA, "source", "s", "", "Path to IPA file which you want to sign/resign (required)")
		cmd.Flags().StringVarP(&certificate, "certificate", "c", "", "Signing certificate Common Name from Keychain (required)")
		cmd.Flags().StringVarP(&entitlements, "entitlements", "e", "", "New entitlements to change (optional)")
		cmd.Flags().StringVarP(&mobileProvision, "provision", "p", "", "Path to mobile provisioning file (optional)")
		cmd.Flags().StringVarP(&bundleID, "bundle", "b", "", "Bundle identifier (optional)")
	}

	rootCmd.AddCommand(resignCmd)
}

func runCLI() {
	// Validate required flags
	if err := validateCLIArguments(); err != nil {
		fmt.Printf("\n❌ Error: %v\n\n", err)
		printUsageExamples()
		os.Exit(1)
	}

	// Create config
	config := resigner.Config{
		SourceIPA:       sourceIPA,
		Certificate:     certificate,
		Entitlements:    entitlements,
		MobileProvision: mobileProvision,
		BundleID:        bundleID,
	}

	// Create resigner with progress callback
	r := resigner.NewResigner(config, func(message string) {
		fmt.Println(message)
	})

	// Run resign
	if err := r.Resign(); err != nil {
		fmt.Printf("\n❌ Resign failed: %v\n", err)
		printTroubleshootingHelp(err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Successfully resigned IPA!")
}

// validateCLIArguments validates all CLI arguments and checks file existence
func validateCLIArguments() error {
	// Check required flags
	if sourceIPA == "" {
		return fmt.Errorf("source IPA path is required (use -s flag)")
	}

	if certificate == "" {
		return fmt.Errorf("certificate is required (use -c flag)")
	}

	// Check if source file exists
	if _, err := os.Stat(sourceIPA); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", sourceIPA)
	} else if err != nil {
		return fmt.Errorf("cannot access source file %s: %v", sourceIPA, err)
	}

	// Check file extension
	validExtensions := []string{".ipa", ".app"}
	hasValidExt := false
	for _, ext := range validExtensions {
		if len(sourceIPA) >= len(ext) && sourceIPA[len(sourceIPA)-len(ext):] == ext {
			hasValidExt = true
			break
		}
	}
	if !hasValidExt {
		return fmt.Errorf("source file must be .ipa or .app, got: %s", sourceIPA)
	}

	// Check optional files if provided
	if entitlements != "" {
		if _, err := os.Stat(entitlements); os.IsNotExist(err) {
			return fmt.Errorf("entitlements file does not exist: %s", entitlements)
		} else if err != nil {
			return fmt.Errorf("cannot access entitlements file %s: %v", entitlements, err)
		}
		// Check if it's a plist file
		if len(entitlements) < 6 || entitlements[len(entitlements)-6:] != ".plist" {
			return fmt.Errorf("entitlements file must be .plist, got: %s", entitlements)
		}
	}

	if mobileProvision != "" {
		if _, err := os.Stat(mobileProvision); os.IsNotExist(err) {
			return fmt.Errorf("mobile provision file does not exist: %s", mobileProvision)
		} else if err != nil {
			return fmt.Errorf("cannot access mobile provision file %s: %v", mobileProvision, err)
		}
		// Check if it's a mobileprovision file
		if len(mobileProvision) < 17 || mobileProvision[len(mobileProvision)-17:] != ".mobileprovision" {
			return fmt.Errorf("mobile provision file must be .mobileprovision, got: %s", mobileProvision)
		}
	}

	// Validate bundle ID format if provided
	if bundleID != "" {
		if len(bundleID) < 3 || !isValidBundleID(bundleID) {
			return fmt.Errorf("invalid bundle ID format: %s (expected format: com.company.app)", bundleID)
		}
	}

	return nil
}

// isValidBundleID checks if bundle ID has valid format
func isValidBundleID(bundleID string) bool {
	// Basic validation: should contain at least one dot and only valid characters
	if len(bundleID) == 0 {
		return false
	}

	hasDot := false
	for _, ch := range bundleID {
		if ch == '.' {
			hasDot = true
		} else if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return false
		}
	}

	return hasDot
}

// printUsageExamples prints usage examples
func printUsageExamples() {
	fmt.Println("Usage Examples:")
	fmt.Println("───────────────")
	fmt.Println()
	fmt.Println("Basic resign:")
	fmt.Println("  resignipa -s /path/to/app.ipa -c \"Apple Development: Name\"")
	fmt.Println()
	fmt.Println("With provisioning profile:")
	fmt.Println("  resignipa -s app.ipa -c \"Apple Development: Name\" -p profile.mobileprovision")
	fmt.Println()
	fmt.Println("With bundle ID:")
	fmt.Println("  resignipa -s app.ipa -c \"Cert\" -b com.company.newapp")
	fmt.Println()
	fmt.Println("All options:")
	fmt.Println("  resignipa -s app.ipa -c \"Cert\" -p profile.mobileprovision -b com.app.id -e entitlements.plist")
	fmt.Println()
	fmt.Println("Resign .app bundle:")
	fmt.Println("  resignipa -s MyApp.app -c \"Apple Development: Name\"")
	fmt.Println()
	fmt.Println("Required:")
	fmt.Println("  -s, --source       Path to .ipa or .app file")
	fmt.Println("  -c, --certificate  Certificate name from Keychain")
	fmt.Println()
	fmt.Println("Optional:")
	fmt.Println("  -p, --provision    Mobile provisioning file (.mobileprovision)")
	fmt.Println("  -b, --bundle       New bundle identifier")
	fmt.Println("  -e, --entitlements Custom entitlements file (.plist)")
	fmt.Println()
	fmt.Println("Find your certificate:")
	fmt.Println("  security find-identity -v -p codesigning")
	fmt.Println()
}

// printTroubleshootingHelp prints context-specific troubleshooting help
func printTroubleshootingHelp(err error) {
	errStr := err.Error()
	fmt.Println()
	fmt.Println("Troubleshooting:")
	fmt.Println("────────────────")

	if strings.Contains(errStr, "certificate") || strings.Contains(errStr, "codesign") {
		fmt.Println("• Verify certificate exists:")
		fmt.Println("  security find-identity -v -p codesigning")
		fmt.Println("• Certificate name must match exactly (including team ID)")
		fmt.Println("• Check if certificate is expired")
	}

	if strings.Contains(errStr, "provision") {
		fmt.Println("• Check provisioning profile is valid")
		fmt.Println("• Ensure profile matches the certificate")
		fmt.Println("• Profile must not be expired")
	}

	if strings.Contains(errStr, "entitlements") {
		fmt.Println("• Entitlements must match provisioning profile capabilities")
		fmt.Println("• Check entitlements file is valid XML/plist format")
	}

	if strings.Contains(errStr, "bundle") {
		fmt.Println("• Bundle ID must match format: com.company.app")
		fmt.Println("• If using provisioning profile, bundle ID must match")
	}

	fmt.Println()
}

// Execute runs the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
