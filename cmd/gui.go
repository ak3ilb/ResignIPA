package cmd

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/resignipa/pkg/resigner"
)

// Professional compact theme
type compactTheme struct {
	fyne.Theme
}

func (c compactTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0xfa, G: 0xfa, B: 0xfa, A: 0xff} // Light gray background
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x4a, G: 0x90, B: 0xe2, A: 0xff} // Clean blue
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0xc0, G: 0xc0, B: 0xc0, A: 0xff} // Light gray disabled
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0x2c, G: 0x2c, B: 0x2c, A: 0xff} // Dark gray text
	case theme.ColorNameHover:
		return color.NRGBA{R: 0x3a, G: 0x7e, B: 0xd0, A: 0xff} // Slightly darker blue on hover
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // White input background
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff} // Medium gray placeholder
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 0xdd, G: 0xdd, B: 0xdd, A: 0xff} // Light gray border
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 0xe0, G: 0xe0, B: 0xe0, A: 0xff} // Light separator
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0xe8, G: 0xe8, B: 0xe8, A: 0xff} // Subtle gray scrollbar
	case theme.ColorNameMenuBackground:
		return color.NRGBA{R: 0xf5, G: 0xf5, B: 0xf5, A: 0xff} // Light gray for progress log background
	case theme.ColorNameHeaderBackground:
		return color.NRGBA{R: 0xe6, G: 0xf3, B: 0xff, A: 0xff} // Light blue for progress log header
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0xe0, G: 0xe0, B: 0xe0, A: 0xff} // Light gray for dividers
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0xfa, G: 0xfa, B: 0xfa, A: 0xff} // Light gray for error popup background
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (c compactTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (c compactTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (c compactTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 12 // Compact text
	case theme.SizeNameHeadingText:
		return 16 // Compact heading
	case theme.SizeNameSubHeadingText:
		return 14 // Compact subheading
	case theme.SizeNameCaptionText:
		return 10 // Compact caption
	case theme.SizeNameInlineIcon:
		return 14 // Compact icons
	case theme.SizeNamePadding:
		return 8 // Professional padding
	case theme.SizeNameScrollBar:
		return 6 // Subtle scrollbar
	case theme.SizeNameScrollBarSmall:
		return 4 // Very subtle for small scrollbars
	case theme.SizeNameInputBorder:
		return 1 // Thin borders
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// LaunchGUI starts the GUI application
func LaunchGUI() {
	myApp := app.New()
	myApp.Settings().SetTheme(&compactTheme{})

	window := myApp.NewWindow("ResignIPA")
	window.Resize(fyne.NewSize(700, 750))
	window.SetFixedSize(true) // Prevent resizing for consistent layout

	// Compact input fields with uniform sizing
	sourceEntry := widget.NewEntry()
	sourceEntry.SetPlaceHolder("Select IPA or APP file...")
	sourceEntry.Resize(fyne.NewSize(600, 32))

	certEntry := widget.NewEntry()
	certEntry.SetPlaceHolder("Certificate name from Keychain...")
	certEntry.Resize(fyne.NewSize(600, 32))

	entitlementsEntry := widget.NewEntry()
	entitlementsEntry.SetPlaceHolder("Optional: custom entitlements.plist")
	entitlementsEntry.Resize(fyne.NewSize(600, 32))

	provisionEntry := widget.NewEntry()
	provisionEntry.SetPlaceHolder("Optional: provisioning profile")
	provisionEntry.Resize(fyne.NewSize(600, 32))

	bundleEntry := widget.NewEntry()
	bundleEntry.SetPlaceHolder("Optional: new bundle ID")
	bundleEntry.Resize(fyne.NewSize(600, 32))

	// Progress text with compact styling
	progressText := widget.NewRichText()
	progressText.Wrapping = fyne.TextWrapWord

	// Professional progress log with light blue header background
	progressLogLabel := canvas.NewText("Progress Log", color.NRGBA{R: 0x2c, G: 0x2c, B: 0x2c, A: 0xff})
	progressLogLabel.TextSize = 14
	progressLogLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Light blue header container
	progressHeaderContainer := container.NewBorder(
		nil, nil, nil, nil,
		progressLogLabel,
	)

	// Thin line below progress header
	progressHeaderDivider := widget.NewSeparator()

	// Progress text container
	progressContainer := container.NewBorder(
		nil, nil, nil, nil,
		progressText,
	)
	progressContainer.Resize(fyne.NewSize(660, 140))

	// Scroll container with professional styling
	progressScroll := container.NewVScroll(progressContainer)
	progressScroll.SetMinSize(fyne.NewSize(660, 140))

	// Compact file picker buttons with uniform sizing
	sourceBrowse := widget.NewButton("...", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				sourceEntry.SetText(reader.URI().Path())
				reader.Close()
			}
		}, window)
	})
	sourceBrowse.Resize(fyne.NewSize(40, 32))

	entitlementsBrowse := widget.NewButton("...", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				entitlementsEntry.SetText(reader.URI().Path())
				reader.Close()
			}
		}, window)
	})
	entitlementsBrowse.Resize(fyne.NewSize(40, 32))

	provisionBrowse := widget.NewButton("...", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				provisionEntry.SetText(reader.URI().Path())
				reader.Close()
			}
		}, window)
	})
	provisionBrowse.Resize(fyne.NewSize(40, 32))

	// Professional resign button
	var resignBtn *widget.Button
	resignBtn = widget.NewButton("Resign IPA", func() {
		// Enhanced validation
		errors := validateGUIInputs(sourceEntry.Text, certEntry.Text, entitlementsEntry.Text, provisionEntry.Text, bundleEntry.Text)
		if len(errors) > 0 {
			errorMsg := "Please fix the following errors:\n\n" + strings.Join(errors, "\n")
			dialog.ShowError(fmt.Errorf(errorMsg), window)
			return
		}

		// Disable button during operation
		resignBtn.Disable()
		resignBtn.SetText("Processing...")

		// Clear progress and show starting message
		progressText.ParseMarkdown("**Starting resign process...**\n\n")
		progressScroll.ScrollToTop()

		// Run resign in goroutine
		go func() {
			defer func() {
				resignBtn.Enable()
				resignBtn.SetText("Resign IPA")
			}()

			config := resigner.Config{
				SourceIPA:       sourceEntry.Text,
				Certificate:     certEntry.Text,
				Entitlements:    entitlementsEntry.Text,
				MobileProvision: provisionEntry.Text,
				BundleID:        bundleEntry.Text,
			}

			var logMessages []string
			r := resigner.NewResigner(config, func(message string) {
				// Format message with emoji based on content
				formattedMsg := formatProgressMessage(message)
				logMessages = append(logMessages, formattedMsg)

				// Create markdown content
				content := "**Progress Log**\n\n" + strings.Join(logMessages, "\n")
				progressText.ParseMarkdown(content)
				progressScroll.ScrollToBottom()
			})

			err := r.Resign()
			if err != nil {
				errorMsg := fmt.Sprintf("\n\n**Error:** %v\n\n**Troubleshooting:**\n", err)
				if strings.Contains(err.Error(), "certificate") {
					errorMsg += "• Check certificate name matches Keychain exactly\n"
					errorMsg += "• Run: `security find-identity -v -p codesigning`\n"
				}
				if strings.Contains(err.Error(), "provision") {
					errorMsg += "• Verify provisioning profile is valid\n"
					errorMsg += "• Check profile matches certificate\n"
				}
				logMessages = append(logMessages, errorMsg)
				content := "**Progress Log**\n\n" + strings.Join(logMessages, "\n")
				progressText.ParseMarkdown(content)
				dialog.ShowError(err, window)
			} else {
				successMsg := "\n\n**Success!** IPA has been resigned successfully!\n\n**Output:** Check the 'Resigned' folder.\n"
				logMessages = append(logMessages, successMsg)
				content := "**Progress Log**\n\n" + strings.Join(logMessages, "\n")
				progressText.ParseMarkdown(content)
				dialog.ShowInformation("Success", "IPA has been resigned successfully!\n\nCheck the 'Resigned' folder for your new file.", window)
			}
			progressScroll.ScrollToBottom()
		}()
	})
	resignBtn.Resize(fyne.NewSize(140, 32))

	// Professional header with improved typography
	title := canvas.NewText("ResignIPA", color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}) // Bold black text
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText("iOS IPA Resigning Tool", color.NRGBA{R: 0x66, G: 0x66, B: 0x66, A: 0xff}) // Medium gray text
	subtitle.TextSize = 14
	subtitle.TextStyle = fyne.TextStyle{}

	// Header with proper spacing and divider
	headerContent := container.NewVBox(
		container.NewCenter(title),
		container.NewCenter(subtitle),
	)

	// Add thin divider line below header
	headerDivider := widget.NewSeparator()

	header := container.NewVBox(
		headerContent,
		headerDivider,
	)

	// Professional form layout with consistent styling
	requiredLabel := canvas.NewText("Required Fields", color.NRGBA{R: 0x2c, G: 0x2c, B: 0x2c, A: 0xff})
	requiredLabel.TextSize = 14
	requiredLabel.TextStyle = fyne.TextStyle{Bold: true}

	requiredDivider := widget.NewSeparator()

	requiredSection := container.NewVBox(
		requiredLabel,
		requiredDivider,
		container.NewBorder(nil, nil, widget.NewLabel("Source:"), sourceBrowse, sourceEntry),
		container.NewBorder(nil, nil, widget.NewLabel("Certificate:"), nil, certEntry),
	)

	optionalLabel := canvas.NewText("Optional Fields", color.NRGBA{R: 0x2c, G: 0x2c, B: 0x2c, A: 0xff})
	optionalLabel.TextSize = 14
	optionalLabel.TextStyle = fyne.TextStyle{Bold: true}

	optionalDivider := widget.NewSeparator()

	optionalSection := container.NewVBox(
		optionalLabel,
		optionalDivider,
		container.NewBorder(nil, nil, widget.NewLabel("Entitlements:"), entitlementsBrowse, entitlementsEntry),
		container.NewBorder(nil, nil, widget.NewLabel("Provision:"), provisionBrowse, provisionEntry),
		container.NewBorder(nil, nil, widget.NewLabel("Bundle ID:"), nil, bundleEntry),
		// Add spacing after bundle ID field
		container.NewVBox(),
	)

	form := container.NewVBox(
		requiredSection,
		widget.NewSeparator(),
		optionalSection,
	)

	formScroll := container.NewVScroll(form)
	formScroll.SetMinSize(fyne.NewSize(660, 380))

	// Professional compact layout with proper spacing and padding
	mainContent := container.NewVBox(
		header,
		formScroll,
	)

	// Add spacing between form and progress log
	spacingContainer := container.NewVBox()

	bottomContent := container.NewVBox(
		spacingContainer,
		progressHeaderContainer,
		progressHeaderDivider,
		progressScroll,
		container.NewCenter(resignBtn),
	)

	content := container.NewBorder(
		mainContent,
		bottomContent,
		nil,
		nil,
	)

	window.SetContent(content)
	window.ShowAndRun()
}

// validateGUIInputs validates GUI inputs with detailed error messages
func validateGUIInputs(source, cert, entitlements, provision, bundleID string) []string {
	var errors []string

	if source == "" {
		errors = append(errors, "• Source IPA/APP file is required")
	} else {
		if _, err := os.Stat(source); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("• Source file does not exist: %s", source))
		} else {
			// Check extension
			if !strings.HasSuffix(strings.ToLower(source), ".ipa") && !strings.HasSuffix(strings.ToLower(source), ".app") {
				errors = append(errors, "• Source file must be .ipa or .app")
			}
		}
	}

	if cert == "" {
		errors = append(errors, "• Certificate name is required")
	}

	if entitlements != "" {
		if _, err := os.Stat(entitlements); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("• Entitlements file does not exist: %s", entitlements))
		} else if !strings.HasSuffix(strings.ToLower(entitlements), ".plist") {
			errors = append(errors, "• Entitlements file must be .plist")
		}
	}

	if provision != "" {
		if _, err := os.Stat(provision); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("• Mobile provision file does not exist: %s", provision))
		} else if !strings.HasSuffix(strings.ToLower(provision), ".mobileprovision") {
			errors = append(errors, "• Mobile provision file must be .mobileprovision")
		}
	}

	if bundleID != "" {
		if !isValidBundleID(bundleID) {
			errors = append(errors, fmt.Sprintf("• Invalid bundle ID format: %s (expected: com.company.app)", bundleID))
		}
	}

	return errors
}

// formatProgressMessage formats progress messages with appropriate emojis
func formatProgressMessage(message string) string {
	msg := strings.TrimSpace(message)

	// Add emojis based on message content
	switch {
	case strings.Contains(msg, "Start"):
		return fmt.Sprintf("▶ %s", msg)
	case strings.Contains(msg, "Extract"):
		return fmt.Sprintf("📦 %s", msg)
	case strings.Contains(msg, "Sign"):
		return fmt.Sprintf("✍ %s", msg)
	case strings.Contains(msg, "Creating"):
		return fmt.Sprintf("📁 %s", msg)
	case strings.Contains(msg, "Clear"):
		return fmt.Sprintf("🗑 %s", msg)
	case strings.Contains(msg, "FINISHED"):
		return fmt.Sprintf("✓ %s", msg)
	case strings.Contains(msg, "Error"):
		return fmt.Sprintf("✗ %s", msg)
	default:
		return fmt.Sprintf("• %s", msg)
	}
}
