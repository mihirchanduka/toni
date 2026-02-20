package ui

import (
	"image"
	"os"
	"strings"

	"github.com/qeesung/image2ascii/convert"
)

// TerminalCapabilities represents which graphics protocols the terminal supports.
type TerminalCapabilities struct {
	SupportsKitty  bool
	SupportsSixel  bool
	SupportsITerm2 bool
}

// DetectTerminalCapabilities detects which graphics protocols the terminal supports.
func DetectTerminalCapabilities() TerminalCapabilities {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")

	return TerminalCapabilities{
		SupportsKitty:  strings.Contains(term, "kitty") || os.Getenv("KITTY_WINDOW_ID") != "",
		SupportsSixel:  detectSixelSupport(),
		SupportsITerm2: termProgram == "iTerm.app",
	}
}

// detectSixelSupport checks if the terminal supports Sixel graphics.
// This is a simplified check - we look for common Sixel-capable terminals.
func detectSixelSupport() bool {
	term := os.Getenv("TERM")

	// Xterm with Sixel support
	if strings.Contains(term, "xterm") && os.Getenv("XTERM_VERSION") != "" {
		return true
	}

	// MLTerm
	if strings.Contains(term, "mlterm") {
		return true
	}

	// WezTerm
	if os.Getenv("WEZTERM_EXECUTABLE") != "" {
		return true
	}

	// Foot terminal
	if strings.Contains(term, "foot") {
		return true
	}

	return false
}

// RenderMapImage renders a map image using the best available terminal graphics protocol.
// Falls back to ASCII art if no graphics protocols are supported.
func RenderMapImage(img image.Image, caps TerminalCapabilities, targetWidth, targetHeight int) string {
	// TODO: Implement Kitty, Sixel, and iTerm2 protocols when rasterm library is available
	// For now, we'll use ASCII art as the universal fallback

	// Convert to ASCII art
	return convertToASCII(img, targetWidth, targetHeight)
}

// convertToASCII converts an image to colored ASCII art.
func convertToASCII(img image.Image, targetWidth, targetHeight int) string {
	// Create converter with options
	converter := convert.NewImageConverter()

	// Convert options
	opts := convert.DefaultOptions
	opts.FixedWidth = targetWidth
	opts.FixedHeight = targetHeight
	opts.Colored = true // Use ANSI colors
	opts.Ratio = 0.5    // Adjust for terminal character aspect ratio

	// Convert image to ASCII
	ascii := converter.Image2ASCIIString(img, &opts)

	return ascii
}
