package jif

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/nfnt/resize"
)

// ============================================================================
// Messages
// ============================================================================

type frameMsg int
type processingCompleteMsg struct{}
type progressMsg struct {
	partialFrame string
	rowsComplete int
	totalRows    int
}

// ============================================================================
// Model
// ============================================================================

type model struct {
	// GIF data
	GIF          *gif.GIF
	Frames       []string
	CurrentFrame int

	// Display state
	Width    int
	Height   int
	Paused   bool
	ShowHelp bool
	Ready    bool

	// Progressive loading state
	Loading      bool
	LoadingFrame string
	LoadingRows  int
	TotalRows    int

	// Reference to program for sending messages
	program *tea.Program
}

// ============================================================================
// Rendering
// ============================================================================

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
func (m *model) calculateImageSize(img image.Image) (width, height int) {
	// Each terminal cell is 2 characters wide (we render ▀▀ or ▄▄)
	maxWidth := m.Width / 2
	ratio := float64(img.Bounds().Dy()) / float64(img.Bounds().Dx())
	targetHeight := int(float64(maxWidth) * ratio * 2)

	// If height exceeds terminal, scale down
	if targetHeight > m.Height*2 {
		targetHeight = m.Height * 2
		maxWidth = int(float64(targetHeight) / ratio / 2)
	}

	return maxWidth, targetHeight
}

// renderImageHalfBlock converts an image to halfblock characters with optional progressive updates
func (m *model) renderImageHalfBlock(img image.Image, progressChan chan<- progressMsg) string {
	width, height := m.calculateImageSize(img)
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
			progressChan <- progressMsg{
				partialFrame: sb.String(),
				rowsComplete: currentRow,
				totalRows:    totalRows,
			}
		}
	}

	return sb.String()
}

// ============================================================================
// GIF Processing
// ============================================================================

// getGifDimensions calculates the total canvas size needed for all frames
func getGifDimensions(g *gif.GIF) (width, height int) {
	var lowestX, lowestY, highestX, highestY int

	for _, img := range g.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}
		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}
		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}
		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}

	return highestX - lowestX, highestY - lowestY
}

// processFrame handles GIF disposal methods and composites the current frame
func processFrame(currentImg, previousImg *image.RGBA, srcImg *image.Paletted, disposal byte) {
	switch disposal {
	case gif.DisposalBackground:
		draw.Draw(currentImg, srcImg.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)
	case gif.DisposalPrevious:
		draw.Draw(currentImg, currentImg.Bounds(), previousImg, image.Point{}, draw.Src)
	}
}

// ProcessGIF renders all frames with progressive loading for the first frame
func (m *model) ProcessGIF(p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		imgWidth, imgHeight := getGifDimensions(m.GIF)
		frames := make([]string, len(m.GIF.Image))

		currentImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
		previousImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

		// Set up progressive loading for first frame
		progressChan := make(chan progressMsg, 100)
		go func() {
			for msg := range progressChan {
				p.Send(msg)
			}
		}()

		// Process each frame
		for i, srcImg := range m.GIF.Image {
			// Save previous state if needed
			if i > 0 && m.GIF.Disposal[i-1] == gif.DisposalPrevious {
				draw.Draw(previousImage, previousImage.Bounds(), currentImage, image.Point{}, draw.Src)
			}

			// Apply disposal method from previous frame
			if i > 0 {
				processFrame(currentImage, previousImage, m.GIF.Image[i-1], m.GIF.Disposal[i-1])
			}

			// Composite current frame
			draw.Draw(currentImage, currentImage.Bounds(), srcImg, image.Point{}, draw.Over)

			// Create a copy for rendering
			imgCopy := image.NewRGBA(currentImage.Bounds())
			draw.Draw(imgCopy, imgCopy.Bounds(), currentImage, image.Point{}, draw.Src)

			// Render with progressive updates only for first frame
			if i == 0 {
				frames[i] = m.renderImageHalfBlock(imgCopy, progressChan)
				close(progressChan)
			} else {
				frames[i] = m.renderImageHalfBlock(imgCopy, nil)
			}
		}

		m.Frames = frames
		return processingCompleteMsg{}
	}
}

// ============================================================================
// Bubbletea Implementation
// ============================================================================

func (m *model) Init() tea.Cmd {
	m.Loading = true
	return m.ProcessGIF(m.program)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case frameMsg:
		return m.handleFrameAdvance()

	case progressMsg:
		return m.handleProgress(msg)

	case processingCompleteMsg:
		return m.handleProcessingComplete()

	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)
	}

	return m, nil
}

// ============================================================================
// Message Handlers
// ============================================================================

func (m *model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "space":
		m.Paused = !m.Paused
		if !m.Paused && m.Ready {
			return m, m.nextFrame()
		}

	case "?":
		m.ShowHelp = !m.ShowHelp

	case "n", "right":
		if len(m.Frames) > 0 {
			m.Paused = true
			m.CurrentFrame = (m.CurrentFrame + 1) % len(m.Frames)
		}

	case "p", "left":
		if len(m.Frames) > 0 {
			m.Paused = true
			m.CurrentFrame = (m.CurrentFrame - 1 + len(m.Frames)) % len(m.Frames)
		}

	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

func (m *model) handleFrameAdvance() (tea.Model, tea.Cmd) {
	if !m.Paused && m.Ready && len(m.Frames) > 0 {
		m.CurrentFrame = (m.CurrentFrame + 1) % len(m.Frames)
		return m, m.nextFrame()
	}
	return m, nil
}

func (m *model) handleProgress(msg progressMsg) (tea.Model, tea.Cmd) {
	if m.Loading && !m.Ready {
		m.LoadingFrame = msg.partialFrame
		m.LoadingRows = msg.rowsComplete
		m.TotalRows = msg.totalRows
	}
	return m, nil
}

func (m *model) handleProcessingComplete() (tea.Model, tea.Cmd) {
	m.Ready = true
	m.Loading = false
	m.CurrentFrame = 0
	if !m.Paused {
		return m, m.nextFrame()
	}
	return m, nil
}

func (m *model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	oldWidth, oldHeight := m.Width, m.Height
	m.Width, m.Height = msg.Width, msg.Height

	// Ignore initial size (handled by Init) or if already processing
	if oldWidth == 0 || oldHeight == 0 {
		return m, nil
	}

	// Ignore if size didn't actually change
	if oldWidth == m.Width && oldHeight == m.Height {
		return m, nil
	}

	// Ignore if we're currently loading - just update dimensions
	// The resize will be handled after current processing completes
	if m.Loading {
		return m, nil
	}

	// Process the resize
	m.Ready = false
	m.Loading = true
	m.LoadingFrame = ""
	m.LoadingRows = 0
	m.TotalRows = 0
	m.Frames = []string{}

	return m, m.ProcessGIF(m.program)
}

// nextFrame schedules the next frame based on GIF delay
func (m *model) nextFrame() tea.Cmd {
	if m.CurrentFrame < 0 || m.CurrentFrame >= len(m.GIF.Delay) {
		return nil
	}

	delay := m.GIF.Delay[m.CurrentFrame]
	if delay == 0 {
		delay = 10 // Default to 100ms if no delay specified
	}

	return tea.Tick(time.Duration(delay)*10*time.Millisecond, func(t time.Time) tea.Msg {
		return frameMsg(0)
	})
}

// ============================================================================
// View Rendering
// ============================================================================

func (m model) View() tea.View {
	var v tea.View
	v.AltScreen = true

	// Progressive loading view
	if m.Loading && m.LoadingFrame != "" {
		v.Content = m.renderLoadingView()
		return v
	}

	// Initial loading message
	if !m.Ready || len(m.Frames) == 0 {
		v.Content = m.renderInitialLoading()
		return v
	}

	// Normal playback view
	v.Content = m.renderPlaybackView()
	return v
}

func (m model) renderLoadingView() *lipgloss.Layer {
	frame := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Top).
		Render(m.LoadingFrame)

	status := fmt.Sprintf(" Loading... %d/%d rows ", m.LoadingRows, m.TotalRows)
	statusText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Render(status)

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(frame).Z(0),
		lipgloss.NewLayer(statusText).X(1).Y(0).Z(5),
	}

	return lipgloss.NewLayer(lipgloss.NewCanvas(layers...).Render())
}

func (m model) renderInitialLoading() *lipgloss.Layer {
	content := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Foreground(lipgloss.Color("86")).
		Render("Loading GIF...")

	return lipgloss.NewLayer(content)
}

func (m model) renderPlaybackView() *lipgloss.Layer {
	frame := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(m.Frames[m.CurrentFrame])

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(frame).Z(0),
		lipgloss.NewLayer(m.renderStatus()).X(1).Y(0).Z(5),
	}

	if m.ShowHelp {
		help := m.renderHelp()
		helpWidth := lipgloss.Width(help)
		helpHeight := lipgloss.Height(help)
		helpLayer := lipgloss.NewLayer(help).
			X(max(0, (m.Width-helpWidth)/2)).
			Y(max(0, (m.Height-helpHeight)/2)).
			Z(10)
		layers = append(layers, helpLayer)
	}

	return lipgloss.NewLayer(lipgloss.NewCanvas(layers...).Render())
}

func (m model) renderStatus() string {
	icon := "▶"
	if m.Paused {
		icon = "⏸"
	}

	status := fmt.Sprintf(" %s %d/%d ", icon, m.CurrentFrame+1, len(m.Frames))
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(status)
}

func (m model) renderHelp() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("213")).
		Bold(true).
		Render("  Keybindings")

	helpText := `
  Space      Pause/Resume
  n / →      Next frame
  p / ←      Previous frame
  ?          Toggle help
  q / Ctrl+C Quit
`

	content := lipgloss.JoinVertical(lipgloss.Left, title, helpText)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("213")).
		Padding(1, 2).
		Render(content)
}

// ============================================================================
// Utilities
// ============================================================================

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func loadGIF(source string) (*gif.GIF, error) {
	var reader io.ReadCloser
	var err error

	if isURL(source) {
		fmt.Printf("Downloading GIF from %s...\n", source)
		resp, err := http.Get(source)
		if err != nil {
			return nil, fmt.Errorf("failed to download: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP error: %s", resp.Status)
		}
		reader = resp.Body
	} else {
		file, err := os.Open(source)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		reader = file
	}
	defer reader.Close()

	gifImage, err := gif.DecodeAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode GIF: %w", err)
	}

	return gifImage, nil
}

// ============================================================================
// Main
// ============================================================================

// Run starts the JIF GIF viewer with the given source (file path or URL)
func Run(source string) error {
	gifImage, err := loadGIF(source)
	if err != nil {
		return fmt.Errorf("loading GIF: %w", err)
	}

	m := model{
		GIF:    gifImage,
		Paused: false,
	}

	p := tea.NewProgram(&m)
	m.program = p

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running viewer: %w", err)
	}

	return nil
}
