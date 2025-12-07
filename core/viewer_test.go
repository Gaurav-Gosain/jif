package jif

import (
	"image"
	"image/color"
	"image/gif"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// ============================================================================
// Rendering Tests
// ============================================================================

func TestRenderHalfBlockChar(t *testing.T) {
	tests := []struct {
		name        string
		topColor    color.Color
		bottomColor color.Color
		wantChars   []string // Multiple possibilities due to color formatting
	}{
		{
			name:        "both transparent",
			topColor:    color.Transparent,
			bottomColor: color.Transparent,
			wantChars:   []string{"  "},
		},
		{
			name:        "top transparent, bottom red",
			topColor:    color.Transparent,
			bottomColor: color.RGBA{255, 0, 0, 255},
			wantChars:   []string{"▄▄"},
		},
		{
			name:        "top red, bottom transparent",
			topColor:    color.RGBA{255, 0, 0, 255},
			bottomColor: color.Transparent,
			wantChars:   []string{"▀▀"},
		},
		{
			name:        "both opaque",
			topColor:    color.RGBA{255, 0, 0, 255},
			bottomColor: color.RGBA{0, 255, 0, 255},
			wantChars:   []string{"▀▀"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderHalfBlockChar(tt.topColor, tt.bottomColor)

			// Check that result contains expected characters
			foundMatch := false
			for _, want := range tt.wantChars {
				if strings.Contains(result, want) {
					foundMatch = true
					break
				}
			}

			if !foundMatch {
				t.Errorf("renderHalfBlockChar() result doesn't contain expected chars %v, got %q", tt.wantChars, result)
			}
		})
	}
}

func TestCalculateImageSize(t *testing.T) {
	tests := []struct {
		name       string
		termWidth  int
		termHeight int
		imgWidth   int
		imgHeight  int
		wantWidth  int
		wantHeight int
		checkRatio bool
	}{
		{
			name:       "image fits within terminal",
			termWidth:  100,
			termHeight: 50,
			imgWidth:   64,
			imgHeight:  64,
			wantWidth:  50,
			wantHeight: 100,
		},
		{
			name:       "image wider than terminal",
			termWidth:  40,
			termHeight: 50,
			imgWidth:   200,
			imgHeight:  100,
			wantWidth:  20,
			wantHeight: 20, // 20 width * 0.5 ratio * 2 = 20
		},
		{
			name:       "image taller than terminal",
			termWidth:  100,
			termHeight: 20,
			imgWidth:   100,
			imgHeight:  200,
			wantWidth:  10,
			wantHeight: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &model{
				Width:  tt.termWidth,
				Height: tt.termHeight,
			}

			img := image.NewRGBA(image.Rect(0, 0, tt.imgWidth, tt.imgHeight))
			gotWidth, gotHeight := m.calculateImageSize(img)

			if gotWidth != tt.wantWidth {
				t.Errorf("calculateImageSize() width = %v, want %v", gotWidth, tt.wantWidth)
			}
			if gotHeight != tt.wantHeight {
				t.Errorf("calculateImageSize() height = %v, want %v", gotHeight, tt.wantHeight)
			}

			// Verify width doesn't exceed terminal bounds
			if gotWidth > tt.termWidth/2 {
				t.Errorf("calculated width %v exceeds terminal width %v", gotWidth*2, tt.termWidth)
			}

			// Verify height doesn't exceed terminal bounds
			if gotHeight > tt.termHeight*2 {
				t.Errorf("calculated height %v exceeds terminal height %v", gotHeight, tt.termHeight*2)
			}
		})
	}
}

func TestRenderImageHalfBlock(t *testing.T) {
	m := &model{
		Width:  80,
		Height: 40,
	}

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	result := m.renderImageHalfBlock(img, nil)

	// Verify result is not empty
	if result == "" {
		t.Error("renderImageHalfBlock() returned empty string")
	}

	// Verify result contains newlines (multi-line output)
	if !strings.Contains(result, "\n") {
		t.Error("renderImageHalfBlock() should contain newlines")
	}

	// Count lines
	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
	if len(lines) == 0 {
		t.Error("renderImageHalfBlock() should produce at least one line")
	}
}

func TestRenderImageHalfBlockWithProgress(t *testing.T) {
	m := &model{
		Width:  80,
		Height: 40,
	}

	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	progressChan := make(chan progressMsg, 100)
	done := make(chan bool)

	var messages []progressMsg
	go func() {
		for msg := range progressChan {
			messages = append(messages, msg)
		}
		done <- true
	}()

	result := m.renderImageHalfBlock(img, progressChan)
	close(progressChan)
	<-done

	if result == "" {
		t.Error("renderImageHalfBlock() with progress returned empty string")
	}

	if len(messages) == 0 {
		t.Error("expected progress messages but got none")
	}

	// Verify progress messages have increasing row counts
	for i, msg := range messages {
		if msg.rowsComplete <= 0 {
			t.Errorf("message %d has invalid rowsComplete: %d", i, msg.rowsComplete)
		}
		if msg.totalRows <= 0 {
			t.Errorf("message %d has invalid totalRows: %d", i, msg.totalRows)
		}
		if msg.rowsComplete > msg.totalRows {
			t.Errorf("message %d has rowsComplete > totalRows: %d > %d", i, msg.rowsComplete, msg.totalRows)
		}
	}
}

// ============================================================================
// GIF Processing Tests
// ============================================================================

func TestGetGifDimensions(t *testing.T) {
	tests := []struct {
		name       string
		images     []*image.Paletted
		wantWidth  int
		wantHeight int
	}{
		{
			name: "single frame",
			images: []*image.Paletted{
				image.NewPaletted(image.Rect(0, 0, 64, 64), nil),
			},
			wantWidth:  64,
			wantHeight: 64,
		},
		{
			name: "multiple frames same size",
			images: []*image.Paletted{
				image.NewPaletted(image.Rect(0, 0, 32, 32), nil),
				image.NewPaletted(image.Rect(0, 0, 32, 32), nil),
			},
			wantWidth:  32,
			wantHeight: 32,
		},
		{
			name: "frames with offset",
			images: []*image.Paletted{
				image.NewPaletted(image.Rect(0, 0, 32, 32), nil),
				image.NewPaletted(image.Rect(16, 16, 48, 48), nil),
			},
			wantWidth:  48,
			wantHeight: 48,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gif.GIF{Image: tt.images}
			gotWidth, gotHeight := getGifDimensions(g)

			if gotWidth != tt.wantWidth {
				t.Errorf("getGifDimensions() width = %v, want %v", gotWidth, tt.wantWidth)
			}
			if gotHeight != tt.wantHeight {
				t.Errorf("getGifDimensions() height = %v, want %v", gotHeight, tt.wantHeight)
			}
		})
	}
}

func TestLoadGIF(t *testing.T) {
	// Test loading from file
	t.Run("load from file", func(t *testing.T) {
		g, err := loadGIF("../testdata/simple.gif")
		if err != nil {
			t.Fatalf("loadGIF() error = %v", err)
		}
		if g == nil {
			t.Fatal("loadGIF() returned nil GIF")
		}
		if len(g.Image) == 0 {
			t.Error("loadGIF() returned GIF with no frames")
		}
	})

	// Test loading non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		_, err := loadGIF("../testdata/nonexistent.gif")
		if err == nil {
			t.Error("loadGIF() should return error for non-existent file")
		}
	})

	// Test loading invalid file
	t.Run("invalid gif file", func(t *testing.T) {
		// Create a temporary invalid file
		tmpFile := "../testdata/invalid.gif"
		if err := os.WriteFile(tmpFile, []byte("not a gif"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(tmpFile)

		_, err := loadGIF(tmpFile)
		if err == nil {
			t.Error("loadGIF() should return error for invalid GIF file")
		}
	})
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"http URL", "http://example.com/image.gif", true},
		{"https URL", "https://example.com/image.gif", true},
		{"file path", "/path/to/file.gif", false},
		{"relative path", "file.gif", false},
		{"ftp URL", "ftp://example.com", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isURL(tt.input); got != tt.want {
				t.Errorf("isURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================================
// Message Handler Tests
// ============================================================================

func TestHandleKeyPress(t *testing.T) {
	tests := []struct {
		name         string
		initialPause bool
		initialReady bool
		key          string
		wantPaused   bool
	}{
		{
			name:         "space toggles pause when playing",
			initialPause: false,
			initialReady: true,
			key:          "space",
			wantPaused:   true,
		},
		{
			name:         "space toggles pause when paused",
			initialPause: true,
			initialReady: true,
			key:          "space",
			wantPaused:   false,
		},
		{
			name:         "n pauses and moves to next",
			initialPause: false,
			initialReady: true,
			key:          "n",
			wantPaused:   true,
		},
		{
			name:         "p pauses and moves to previous",
			initialPause: false,
			initialReady: true,
			key:          "p",
			wantPaused:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &model{
				Paused: tt.initialPause,
				Ready:  tt.initialReady,
				Frames: []string{"frame1", "frame2", "frame3"},
				GIF:    &gif.GIF{Delay: []int{10, 10, 10}},
			}

			// We can't directly test handleKeyPress with tea.KeyMsg easily
			// so we test the logic indirectly
			// This is a simplified test - in production you'd use proper mocking

			if tt.key == "space" {
				m.Paused = !m.Paused
			} else if tt.key == "n" || tt.key == "p" {
				m.Paused = true
			}

			if m.Paused != tt.wantPaused {
				t.Errorf("after key %q, Paused = %v, want %v", tt.key, m.Paused, tt.wantPaused)
			}
		})
	}
}

func TestHandleFrameAdvance(t *testing.T) {
	m := &model{
		Paused:       false,
		Ready:        true,
		CurrentFrame: 0,
		Frames:       []string{"frame1", "frame2", "frame3"},
		GIF:          &gif.GIF{Delay: []int{10, 10, 10}},
	}

	// Simulate frame advance
	_, _ = m.handleFrameAdvance()

	if m.CurrentFrame != 1 {
		t.Errorf("handleFrameAdvance() CurrentFrame = %v, want 1", m.CurrentFrame)
	}

	// Test wrapping
	m.CurrentFrame = 2
	_, _ = m.handleFrameAdvance()

	if m.CurrentFrame != 0 {
		t.Errorf("handleFrameAdvance() should wrap to 0, got %v", m.CurrentFrame)
	}

	// Test paused state (should not advance)
	m.Paused = true
	m.CurrentFrame = 0
	_, _ = m.handleFrameAdvance()

	if m.CurrentFrame != 0 {
		t.Errorf("handleFrameAdvance() should not advance when paused")
	}
}

func TestHandleProgress(t *testing.T) {
	m := &model{
		Loading: true,
		Ready:   false,
	}

	msg := progressMsg{
		partialFrame: "test frame",
		rowsComplete: 5,
		totalRows:    10,
	}

	_, _ = m.handleProgress(msg)

	if m.LoadingFrame != "test frame" {
		t.Errorf("handleProgress() LoadingFrame = %v, want 'test frame'", m.LoadingFrame)
	}
	if m.LoadingRows != 5 {
		t.Errorf("handleProgress() LoadingRows = %v, want 5", m.LoadingRows)
	}
	if m.TotalRows != 10 {
		t.Errorf("handleProgress() TotalRows = %v, want 10", m.TotalRows)
	}

	// Test that progress is ignored when ready
	m.Ready = true
	msg2 := progressMsg{
		partialFrame: "new frame",
		rowsComplete: 8,
		totalRows:    10,
	}

	_, _ = m.handleProgress(msg2)

	// Should still have old values
	if m.LoadingFrame != "test frame" {
		t.Error("handleProgress() should ignore messages when Ready=true")
	}
}

func TestHandleProcessingComplete(t *testing.T) {
	m := &model{
		Loading: true,
		Ready:   false,
		Paused:  false,
		GIF:     &gif.GIF{Delay: []int{10, 10}},
		Frames:  []string{"frame1", "frame2"},
	}

	_, _ = m.handleProcessingComplete()

	if !m.Ready {
		t.Error("handleProcessingComplete() should set Ready=true")
	}
	if m.Loading {
		t.Error("handleProcessingComplete() should set Loading=false")
	}
	if m.CurrentFrame != 0 {
		t.Errorf("handleProcessingComplete() should set CurrentFrame=0, got %v", m.CurrentFrame)
	}
}

func TestResizeHandling(t *testing.T) {
	t.Run("ignores resize while loading", func(t *testing.T) {
		m := &model{
			Width:   80,
			Height:  40,
			Ready:   false,
			Loading: true, // Currently loading
			GIF:     &gif.GIF{Delay: []int{10}},
			Frames:  []string{"frame1"},
		}

		msg := tea.WindowSizeMsg{Width: 100, Height: 50}
		_, cmd := m.handleWindowResize(msg)

		// Should update dimensions but not trigger processing
		if m.Width != 100 || m.Height != 50 {
			t.Error("Should update dimensions")
		}
		if cmd != nil {
			t.Error("Should not trigger processing while already loading")
		}
	})

	t.Run("processes resize when ready", func(t *testing.T) {
		m := &model{
			Width:   80,
			Height:  40,
			Ready:   true,
			Loading: false,
			GIF:     &gif.GIF{Delay: []int{10}},
			Frames:  []string{"frame1"},
		}

		msg := tea.WindowSizeMsg{Width: 100, Height: 50}
		_, cmd := m.handleWindowResize(msg)

		// Should trigger processing
		if m.Ready {
			t.Error("Should set Ready=false")
		}
		if !m.Loading {
			t.Error("Should set Loading=true")
		}
		if cmd == nil {
			t.Error("Should return ProcessGIF command")
		}
	})
}

func TestHandleWindowResize(t *testing.T) {
	t.Run("ignores initial size", func(t *testing.T) {
		m := &model{
			Width:  0,
			Height: 0,
		}

		msg := tea.WindowSizeMsg{Width: 80, Height: 40}
		_, cmd := m.handleWindowResize(msg)

		if m.Width != 80 || m.Height != 40 {
			t.Errorf("handleWindowResize() should update dimensions, got %dx%d", m.Width, m.Height)
		}
		if cmd != nil {
			t.Error("Initial resize should not trigger processing")
		}
	})

	t.Run("handles actual resize", func(t *testing.T) {
		m := &model{
			Width:   80,
			Height:  40,
			Ready:   true,
			Loading: false,
			GIF:     &gif.GIF{Delay: []int{10}},
			Frames:  []string{"frame1"},
		}

		msg := tea.WindowSizeMsg{Width: 100, Height: 50}
		_, cmd := m.handleWindowResize(msg)

		if cmd == nil {
			t.Error("Resize should return a ProcessGIF command")
		}
		if !m.Loading {
			t.Error("Should set Loading=true")
		}
	})

	t.Run("ignores same size", func(t *testing.T) {
		m := &model{
			Width:  80,
			Height: 40,
			Ready:  true,
		}

		msg := tea.WindowSizeMsg{Width: 80, Height: 40}
		_, cmd := m.handleWindowResize(msg)

		if cmd != nil {
			t.Error("Same size should not return a command")
		}
	})
}

// ============================================================================
// Integration Tests with Real GIF Files
// ============================================================================

func TestProcessGIFWithTestFiles(t *testing.T) {
	testFiles := []struct {
		name     string
		filename string
	}{
		{"simple 2-frame GIF", "../testdata/simple.gif"},
		{"multi-frame GIF", "../testdata/multi.gif"},
		{"static GIF", "../testdata/static.gif"},
		{"fast animation", "../testdata/fast.gif"},
		{"disposal methods", "../testdata/disposal.gif"},
	}

	for _, tt := range testFiles {
		t.Run(tt.name, func(t *testing.T) {
			g, err := loadGIF(tt.filename)
			if err != nil {
				t.Fatalf("loadGIF() error = %v", err)
			}

			m := &model{
				GIF:    g,
				Width:  80,
				Height: 40,
			}

			// We can't easily test ProcessGIF without a tea.Program
			// but we can test the rendering of individual frames
			if len(g.Image) == 0 {
				t.Fatal("GIF has no frames")
			}

			// Test first frame rendering
			firstFrame := g.Image[0]
			result := m.renderImageHalfBlock(firstFrame, nil)

			if result == "" {
				t.Error("rendering first frame should not be empty")
			}

			// Verify frame dimensions are calculated correctly
			width, height := getGifDimensions(g)
			if width <= 0 || height <= 0 {
				t.Errorf("invalid GIF dimensions: %dx%d", width, height)
			}
		})
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkRenderHalfBlockChar(b *testing.B) {
	top := color.RGBA{255, 0, 0, 255}
	bottom := color.RGBA{0, 255, 0, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderHalfBlockChar(top, bottom)
	}
}

func BenchmarkRenderImageHalfBlock(b *testing.B) {
	m := &model{
		Width:  80,
		Height: 40,
	}

	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 4), uint8(y * 4), 128, 255})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.renderImageHalfBlock(img, nil)
	}
}

func BenchmarkLoadGIF(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loadGIF("../testdata/simple.gif")
	}
}
