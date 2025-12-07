package gif

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/nfnt/resize"
)

// ProgressUpdate represents rendering progress
type ProgressUpdate struct {
	PartialFrame string
	RowsComplete int
	TotalRows    int
}

// renderHalfBlock renders an image using halfblock characters
func (p *Processor) renderHalfBlock(img image.Image, progressChan chan<- ProgressUpdate) string {
	width, height := p.calculateImageSize(img)
	resized := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)
	bounds := resized.Bounds()

	var sb strings.Builder
	totalRows := (bounds.Max.Y - bounds.Min.Y + 1) / 2
	currentRow := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			topColor := resized.At(x, y)

			var bottomColor color.Color
			if y+1 < bounds.Max.Y {
				bottomColor = resized.At(x, y+1)
			} else {
				bottomColor = color.Transparent
			}

			sb.WriteString(renderHalfBlockChar(topColor, bottomColor))
		}
		sb.WriteString("\n")
		currentRow++

		// Send progress updates (throttled to every 2 rows, plus always send last row)
		if progressChan != nil && (currentRow%2 == 0 || y+2 >= bounds.Max.Y) {
			progressChan <- ProgressUpdate{
				PartialFrame: sb.String(),
				RowsComplete: currentRow,
				TotalRows:    totalRows,
			}
		}
	}

	return sb.String()
}

// renderHalfBlockChar converts two vertically stacked pixels into a halfblock character
func renderHalfBlockChar(topColor, bottomColor color.Color) string {
	topR, topG, topB, topA := topColor.RGBA()
	bottomR, bottomG, bottomB, bottomA := bottomColor.RGBA()

	// Both transparent - render nothing
	if topA == 0 && bottomA == 0 {
		return "  "
	}

	// Only bottom pixel visible
	if topA == 0 {
		hex := fmt.Sprintf("#%02x%02x%02x", uint8(bottomR>>8), uint8(bottomG>>8), uint8(bottomB>>8))
		return lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render("▄▄")
	}

	// Only top pixel visible
	if bottomA == 0 {
		hex := fmt.Sprintf("#%02x%02x%02x", uint8(topR>>8), uint8(topG>>8), uint8(topB>>8))
		return lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render("▀▀")
	}

	// Both pixels visible - use foreground and background colors
	topHex := fmt.Sprintf("#%02x%02x%02x", uint8(topR>>8), uint8(topG>>8), uint8(topB>>8))
	bottomHex := fmt.Sprintf("#%02x%02x%02x", uint8(bottomR>>8), uint8(bottomG>>8), uint8(bottomB>>8))
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(topHex)).
		Background(lipgloss.Color(bottomHex)).
		Render("▀▀")
}

// calculateImageSize determines the target size for the image within terminal bounds
func (p *Processor) calculateImageSize(img image.Image) (width, height int) {
	// Each terminal cell is 2 characters wide (we render ▀▀ or ▄▄)
	maxWidth := p.width / 2
	ratio := float64(img.Bounds().Dy()) / float64(img.Bounds().Dx())
	targetHeight := int(float64(maxWidth) * ratio * 2)

	// If height exceeds terminal, scale down
	if targetHeight > p.height*2 {
		targetHeight = p.height * 2
		maxWidth = int(float64(targetHeight) / ratio / 2)
	}

	return maxWidth, targetHeight
}
