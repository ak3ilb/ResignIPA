package cmd

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/resignipa/pkg/resigner"
)

// Custom dark theme inspired by Docker
type dockerTheme struct {
	fyne.Theme
}

func (d dockerTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x0d, G: 0x1b, B: 0x2a, A: 0xff} // Dark blue background
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x0d, G: 0x6e, B: 0xfd, A: 0xff} // Docker blue
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x4a, G: 0x5c, B: 0x6a, A: 0xff}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0x1e, G: 0x7f, B: 0xff, A: 0xff}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x1a, G: 0x2b, B: 0x3c, A: 0xff}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x8b, G: 0x94, B: 0x9e, A: 0xff}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (d dockerTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (d dockerTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (d dockerTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// LaunchGUI starts the GUI application
func LaunchGUI() {
	myApp := app.New()
	myApp.Settings().SetTheme(&dockerTheme{})

	window := myApp.NewWindow("ResignIPA - Docker Edition")
	window.Resize(fyne.NewSize(800, 600))

	// Input fields
	sourceEntry := widget.NewEntry()
	sourceEntry.SetPlaceHolder("Path to IPA or APP file")

	certEntry := widget.NewEntry()
	certEntry.SetPlaceHolder("Certificate Common Name (e.g., Apple Development: Name)")

	entitlementsEntry := widget.NewEntry()
	entitlementsEntry.SetPlaceHolder("Path to entitlements file (optional)")

	provisionEntry := widget.NewEntry()
	provisionEntry.SetPlaceHolder("Path to mobileprovision file (optional)")

	bundleEntry := widget.NewEntry()
	bundleEntry.SetPlaceHolder("Bundle identifier (optional)")

	// Progress text
	progressText := widget.NewLabel("")
	progressText.Wrapping = fyne.TextWrapWord

	// Scroll container for progress
	progressScroll := container.NewVScroll(progressText)
	progressScroll.SetMinSize(fyne.NewSize(780, 200))

	// File picker buttons
	sourceBrowse := widget.NewButton("Browse", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				sourceEntry.SetText(reader.URI().Path())
				reader.Close()
			}
		}, window)
	})

	entitlementsBrowse := widget.NewButton("Browse", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				entitlementsEntry.SetText(reader.URI().Path())
				reader.Close()
			}
		}, window)
	})

	provisionBrowse := widget.NewButton("Browse", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				provisionEntry.SetText(reader.URI().Path())
				reader.Close()
			}
		}, window)
	})

	// Resign button
	var resignBtn *widget.Button
	resignBtn = widget.NewButton("Resign IPA", func() {
		// Validate inputs
		if sourceEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("source IPA path is required"), window)
			return
		}
		if certEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("certificate is required"), window)
			return
		}

		// Disable button during operation
		resignBtn.Disable()
		progressText.SetText("Starting resign process...\n")

		// Run resign in goroutine
		go func() {
			defer resignBtn.Enable()

			config := resigner.Config{
				SourceIPA:       sourceEntry.Text,
				Certificate:     certEntry.Text,
				Entitlements:    entitlementsEntry.Text,
				MobileProvision: provisionEntry.Text,
				BundleID:        bundleEntry.Text,
			}

			var logMessages []string
			r := resigner.NewResigner(config, func(message string) {
				logMessages = append(logMessages, message)
				progressText.SetText(strings.Join(logMessages, "\n") + "\n")
				progressScroll.ScrollToBottom()
			})

			err := r.Resign()
			if err != nil {
				logMessages = append(logMessages, fmt.Sprintf("\n❌ Error: %v", err))
				progressText.SetText(strings.Join(logMessages, "\n") + "\n")
				dialog.ShowError(err, window)
			} else {
				logMessages = append(logMessages, "\n✓ Successfully resigned IPA!")
				progressText.SetText(strings.Join(logMessages, "\n") + "\n")
				dialog.ShowInformation("Success", "Successfully resigned IPA!", window)
			}
			progressScroll.ScrollToBottom()
		}()
	})

	// Create header with logo/title
	title := canvas.NewText("ResignIPA", color.NRGBA{R: 0x0d, G: 0x6e, B: 0xfd, A: 0xff})
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Resign iOS IPA files with ease")
	subtitle.Alignment = fyne.TextAlignCenter

	header := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(title),
		container.NewCenter(subtitle),
		layout.NewSpacer(),
		widget.NewSeparator(),
	)

	// Form layout
	form := container.NewVBox(
		widget.NewLabel("Required Fields:"),
		widget.NewSeparator(),
		widget.NewLabel("Source IPA/APP:"),
		container.NewBorder(nil, nil, nil, sourceBrowse, sourceEntry),
		widget.NewLabel("Certificate:"),
		certEntry,
		widget.NewSeparator(),
		widget.NewLabel("Optional Fields:"),
		widget.NewSeparator(),
		widget.NewLabel("Entitlements:"),
		container.NewBorder(nil, nil, nil, entitlementsBrowse, entitlementsEntry),
		widget.NewLabel("Mobile Provision:"),
		container.NewBorder(nil, nil, nil, provisionBrowse, provisionEntry),
		widget.NewLabel("Bundle ID:"),
		bundleEntry,
	)

	formScroll := container.NewVScroll(form)
	formScroll.SetMinSize(fyne.NewSize(780, 300))

	// Main layout
	content := container.NewBorder(
		container.NewVBox(header, formScroll),
		container.NewVBox(
			widget.NewSeparator(),
			widget.NewLabel("Progress:"),
			progressScroll,
			container.NewCenter(resignBtn),
		),
		nil,
		nil,
	)

	window.SetContent(content)
	window.ShowAndRun()
}

