package resigner

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Config holds the configuration for resigning an IPA
type Config struct {
	SourceIPA       string
	Certificate     string
	Entitlements    string
	MobileProvision string
	BundleID        string
}

// ProgressCallback is called during the resign process
type ProgressCallback func(message string)

// Resigner handles the IPA resigning process
type Resigner struct {
	config   Config
	callback ProgressCallback
	tmpDir   string
	appDir   string
}

// NewResigner creates a new Resigner instance
func NewResigner(config Config, callback ProgressCallback) *Resigner {
	return &Resigner{
		config:   config,
		callback: callback,
	}
}

// logProgress sends a progress message
func (r *Resigner) logProgress(message string) {
	if r.callback != nil {
		r.callback(message)
	}
	fmt.Println(message)
}

// Resign performs the resigning operation
func (r *Resigner) Resign() (err error) {
	// Panic recovery
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("panic occurred: %v", rec)
			r.logProgress(fmt.Sprintf("ERROR: %v", err))
		}
		// Cleanup temp directories
		if r.tmpDir != "" {
			os.RemoveAll(r.tmpDir)
		}
	}()

	// Validate inputs
	if err := r.validate(); err != nil {
		return err
	}

	r.logProgress("Start (re)sign the app...")

	// Setup directories
	if err := r.setupDirectories(); err != nil {
		return fmt.Errorf("failed to setup directories: %w", err)
	}

	// Extract or copy the app
	appPath, err := r.extractApp()
	if err != nil {
		return fmt.Errorf("failed to extract app: %w", err)
	}

	// Handle mobile provision
	if err := r.handleMobileProvision(appPath); err != nil {
		return fmt.Errorf("failed to handle mobile provision: %w", err)
	}

	// Extract entitlements
	entitlementsPath, err := r.extractEntitlements(appPath)
	if err != nil {
		return fmt.Errorf("failed to extract entitlements: %w", err)
	}

	// Handle bundle ID
	if err := r.handleBundleID(appPath); err != nil {
		return fmt.Errorf("failed to handle bundle ID: %w", err)
	}

	// Sign components
	if err := r.signComponents(appPath, entitlementsPath); err != nil {
		return fmt.Errorf("failed to sign components: %w", err)
	}

	// Create resigned IPA
	if err := r.createResignedIPA(appPath); err != nil {
		return fmt.Errorf("failed to create resigned IPA: %w", err)
	}

	r.logProgress("XReSign FINISHED")
	return nil
}

// validate checks if all required inputs are valid
func (r *Resigner) validate() error {
	if r.config.SourceIPA == "" {
		return fmt.Errorf("source IPA path is required")
	}
	if r.config.Certificate == "" {
		return fmt.Errorf("certificate is required")
	}
	if _, err := os.Stat(r.config.SourceIPA); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", r.config.SourceIPA)
	}
	if r.config.MobileProvision != "" {
		if _, err := os.Stat(r.config.MobileProvision); os.IsNotExist(err) {
			return fmt.Errorf("mobile provision file does not exist: %s", r.config.MobileProvision)
		}
	}
	if r.config.Entitlements != "" {
		if _, err := os.Stat(r.config.Entitlements); os.IsNotExist(err) {
			return fmt.Errorf("entitlements file does not exist: %s", r.config.Entitlements)
		}
	}
	return nil
}

// setupDirectories creates temporary directories
func (r *Resigner) setupDirectories() error {
	outDir := filepath.Dir(r.config.SourceIPA)
	tmpDir := filepath.Join(outDir, "tmp")
	appDir := filepath.Join(tmpDir, "app")

	if err := os.MkdirAll(appDir, 0755); err != nil {
		return err
	}

	r.tmpDir = tmpDir
	r.appDir = appDir
	return nil
}

// extractApp extracts IPA or copies .app file
func (r *Resigner) extractApp() (string, error) {
	ext := strings.ToLower(filepath.Ext(r.config.SourceIPA))

	if ext == ".ipa" {
		r.logProgress("Extracting IPA file...")
		if err := unzip(r.config.SourceIPA, r.appDir); err != nil {
			return "", err
		}
	} else if ext == ".app" {
		r.logProgress("Copying .app file...")
		payloadDir := filepath.Join(r.appDir, "Payload")
		if err := os.MkdirAll(payloadDir, 0755); err != nil {
			return "", err
		}
		if err := copyDir(r.config.SourceIPA, filepath.Join(payloadDir, filepath.Base(r.config.SourceIPA))); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("unsupported file type: %s (must be .ipa or .app)", ext)
	}

	// Get application path
	payloadDir := filepath.Join(r.appDir, "Payload")
	entries, err := os.ReadDir(payloadDir)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("no app found in Payload directory")
	}

	appPath := filepath.Join(payloadDir, entries[0].Name())
	return appPath, nil
}

// handleMobileProvision copies the mobile provision file
func (r *Resigner) handleMobileProvision(appPath string) error {
	if r.config.MobileProvision == "" {
		r.logProgress("Sign process using existing provisioning profile from payload")
		return nil
	}

	r.logProgress("Copying provisioning profile into application payload")
	dest := filepath.Join(appPath, "embedded.mobileprovision")
	return copyFile(r.config.MobileProvision, dest)
}

// extractEntitlements extracts entitlements from mobile provision
func (r *Resigner) extractEntitlements(appPath string) (string, error) {
	r.logProgress("Extract entitlements from mobileprovision")

	entitlementsPath := filepath.Join(r.tmpDir, "entitlements.plist")

	if r.config.Entitlements != "" {
		if err := copyFile(r.config.Entitlements, entitlementsPath); err != nil {
			return "", err
		}
		r.logProgress(fmt.Sprintf("Using provided entitlements: %s", r.config.Entitlements))
		return entitlementsPath, nil
	}

	// Extract from embedded.mobileprovision
	provisionPath := filepath.Join(appPath, "embedded.mobileprovision")
	provisioningPlist := filepath.Join(r.tmpDir, "provisioning.plist")

	// security cms -D -i embedded.mobileprovision
	cmd := exec.Command("security", "cms", "-D", "-i", provisionPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to decode provisioning profile: %w", err)
	}

	if err := os.WriteFile(provisioningPlist, output, 0644); err != nil {
		return "", err
	}

	// /usr/libexec/PlistBuddy -x -c 'Print:Entitlements' provisioning.plist
	cmd = exec.Command("/usr/libexec/PlistBuddy", "-x", "-c", "Print:Entitlements", provisioningPlist)
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to extract entitlements: %w", err)
	}

	if err := os.WriteFile(entitlementsPath, output, 0644); err != nil {
		return "", err
	}

	return entitlementsPath, nil
}

// handleBundleID changes the bundle identifier if specified
func (r *Resigner) handleBundleID(appPath string) error {
	if r.config.BundleID == "" {
		r.logProgress("Sign using existing bundle identifier from payload")
		return nil
	}

	r.logProgress(fmt.Sprintf("Changing bundle identifier with: %s", r.config.BundleID))
	infoPlist := filepath.Join(appPath, "Info.plist")
	cmd := exec.Command("/usr/libexec/PlistBuddy", "-c", fmt.Sprintf("Set:CFBundleIdentifier %s", r.config.BundleID), infoPlist)
	return cmd.Run()
}

// signComponents signs all app components
func (r *Resigner) signComponents(appPath, entitlementsPath string) error {
	r.logProgress(fmt.Sprintf("Get list of components and sign with certificate: %s", r.config.Certificate))

	// Find all components
	components, err := findComponents(appPath)
	if err != nil {
		return err
	}

	r.logProgress("Sign plugins, frameworks, dylibs")
	extraCounter := 0
	for _, component := range components {
		ext := filepath.Ext(component)
		switch ext {
		case ".appex":
			if r.config.BundleID != "" {
				newBundleID := fmt.Sprintf("%s.extra%d", r.config.BundleID, extraCounter)
				r.logProgress(fmt.Sprintf("Changing .appex bundle identifier with: %s", newBundleID))
				infoPlist := filepath.Join(component, "Info.plist")
				cmd := exec.Command("/usr/libexec/PlistBuddy", "-c", fmt.Sprintf("Set:CFBundleIdentifier %s", newBundleID), infoPlist)
				if err := cmd.Run(); err != nil {
					r.logProgress(fmt.Sprintf("Warning: Failed to change bundle ID for %s: %v", component, err))
				}
				extraCounter++
			}
			if err := r.codesign(component, entitlementsPath); err != nil {
				return fmt.Errorf("failed to sign %s: %w", component, err)
			}
		case ".framework", ".dylib":
			if err := r.codesign(component, entitlementsPath); err != nil {
				return fmt.Errorf("failed to sign %s: %w", component, err)
			}
		}
	}

	r.logProgress("Sign app")
	for _, component := range components {
		if filepath.Ext(component) == ".app" {
			if err := r.codesign(component, entitlementsPath); err != nil {
				return fmt.Errorf("failed to sign %s: %w", component, err)
			}
		}
	}

	return nil
}

// codesign signs a component
func (r *Resigner) codesign(component, entitlementsPath string) error {
	cmd := exec.Command("/usr/bin/codesign",
		"--continue",
		"--generate-entitlement-der",
		"-f",
		"-s", r.config.Certificate,
		"--entitlements", entitlementsPath,
		component)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("codesign failed: %s - %w", string(output), err)
	}
	return nil
}

// createResignedIPA creates the resigned IPA or copies the .app
func (r *Resigner) createResignedIPA(appPath string) error {
	outDir := filepath.Dir(r.config.SourceIPA)
	resignedDir := filepath.Join(outDir, "Resigned")

	// Remove and recreate Resigned directory
	os.RemoveAll(resignedDir)
	if err := os.MkdirAll(resignedDir, 0755); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(r.config.SourceIPA))

	if ext == ".ipa" {
		appName := filepath.Base(appPath)
		filename := strings.TrimSuffix(appName, filepath.Ext(appName)) + ".ipa"
		outputPath := filepath.Join(resignedDir, filename)

		r.logProgress(fmt.Sprintf("Creating the signed ipa: %s", filename))

		// Create zip from Payload directory
		if err := zipDirectory(r.appDir, outputPath); err != nil {
			return err
		}

		r.logProgress(fmt.Sprintf("Resigned IPA saved to: %s", outputPath))
	} else if ext == ".app" {
		appName := filepath.Base(appPath)
		outputPath := filepath.Join(resignedDir, appName)

		r.logProgress("Moving resigned .app file...")
		if err := copyDir(appPath, outputPath); err != nil {
			return err
		}

		r.logProgress(fmt.Sprintf("Resigned .app saved to: %s", outputPath))
	}

	return nil
}

// Helper functions

// unzip extracts a zip file to a destination
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// zipDirectory creates a zip file from a directory
func zipDirectory(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			return err
		}
		return nil
	})

	return err
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath)
	})
}

// findComponents finds all components that need to be signed
func findComponents(appPath string) ([]string, error) {
	var components []string
	var appComponents []string

	err := filepath.Walk(appPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".app" || ext == ".appex" || ext == ".framework" {
				components = append(components, path)
			}
		} else {
			ext := filepath.Ext(path)
			if ext == ".dylib" {
				components = append(components, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort so that .app files are last
	for _, comp := range components {
		if filepath.Ext(comp) == ".app" {
			appComponents = append(appComponents, comp)
		}
	}

	// Remove .app from components and sort them to be signed last
	var nonAppComponents []string
	for _, comp := range components {
		if filepath.Ext(comp) != ".app" {
			nonAppComponents = append(nonAppComponents, comp)
		}
	}

	// Return non-.app components first, then .app components
	return append(nonAppComponents, appComponents...), nil
}
