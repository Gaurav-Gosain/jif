package gif

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io"
	"net/http"
	"os"
	"strings"
)

// Processor handles GIF loading and processing
type Processor struct {
	gif    *gif.GIF
	width  int
	height int
}

// NewProcessor creates a new GIF processor
func NewProcessor(g *gif.GIF, width, height int) *Processor {
	return &Processor{
		gif:    g,
		width:  width,
		height: height,
	}
}

// LoadFromSource loads a GIF from either a file path or URL
func LoadFromSource(source string) (*gif.GIF, error) {
	var reader io.ReadCloser
	var err error

	if isURL(source) {
		resp, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, err
		}
		reader = resp.Body
	} else {
		file, err := os.Open(source)
		if err != nil {
			return nil, err
		}
		reader = file
	}
	defer reader.Close()

	gifImage, err := gif.DecodeAll(reader)
	if err != nil {
		return nil, err
	}

	return gifImage, nil
}

// ProcessAllFrames processes all frames with disposal methods
func (p *Processor) ProcessAllFrames(progressChan chan<- ProgressUpdate) []string {
	imgWidth, imgHeight := GetGIFDimensions(p.gif)
	frames := make([]string, len(p.gif.Image))

	currentImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	previousImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	for i, srcImg := range p.gif.Image {
		// Handle disposal from previous frame
		if i > 0 && p.gif.Disposal[i-1] == gif.DisposalPrevious {
			draw.Draw(previousImage, previousImage.Bounds(), currentImage, image.Point{}, draw.Src)
		}

		if i > 0 {
			switch p.gif.Disposal[i-1] {
			case gif.DisposalBackground:
				draw.Draw(currentImage, p.gif.Image[i-1].Bounds(),
					&image.Uniform{color.Transparent}, image.Point{}, draw.Src)
			case gif.DisposalPrevious:
				draw.Draw(currentImage, currentImage.Bounds(),
					previousImage, image.Point{}, draw.Src)
			}
		}

		draw.Draw(currentImage, currentImage.Bounds(), srcImg, image.Point{}, draw.Over)

		imgCopy := image.NewRGBA(currentImage.Bounds())
		draw.Draw(imgCopy, imgCopy.Bounds(), currentImage, image.Point{}, draw.Src)

		// Render with progress updates only for first frame
		if i == 0 && progressChan != nil {
			frames[i] = p.renderHalfBlock(imgCopy, progressChan)
			close(progressChan)
		} else {
			frames[i] = p.renderHalfBlock(imgCopy, nil)
		}
	}

	return frames
}

// GetGIFDimensions calculates the total canvas size needed for all frames
func GetGIFDimensions(g *gif.GIF) (width, height int) {
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

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
