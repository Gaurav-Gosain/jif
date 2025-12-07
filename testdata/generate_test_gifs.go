package main

import (
	"image"
	"image/color"
	"image/gif"
	"os"
)

// Generate simple test GIFs for testing purposes
func main() {
	// Create testdata directory if it doesn't exist
	os.MkdirAll("testdata", 0755)

	// Test GIF 1: Simple 2-frame animation with different colors
	generateSimpleGIF("testdata/simple.gif", 2)

	// Test GIF 2: Multi-frame animation (10 frames)
	generateSimpleGIF("testdata/multi.gif", 10)

	// Test GIF 3: Single frame (static image)
	generateSimpleGIF("testdata/static.gif", 1)

	// Test GIF 4: Fast animation (short delays)
	generateFastGIF("testdata/fast.gif")

	// Test GIF 5: Different disposal methods
	generateDisposalTestGIF("testdata/disposal.gif")

	println("Test GIFs generated successfully!")
}

func generateSimpleGIF(filename string, frames int) {
	const size = 64
	palette := color.Palette{
		color.RGBA{0x00, 0x00, 0x00, 0xff}, // black
		color.RGBA{0xff, 0x00, 0x00, 0xff}, // red
		color.RGBA{0x00, 0xff, 0x00, 0xff}, // green
		color.RGBA{0x00, 0x00, 0xff, 0xff}, // blue
		color.RGBA{0xff, 0xff, 0x00, 0xff}, // yellow
		color.RGBA{0xff, 0x00, 0xff, 0xff}, // magenta
		color.RGBA{0x00, 0xff, 0xff, 0xff}, // cyan
		color.RGBA{0xff, 0xff, 0xff, 0xff}, // white
	}

	var images []*image.Paletted
	var delays []int

	for i := 0; i < frames; i++ {
		img := image.NewPaletted(image.Rect(0, 0, size, size), palette)

		// Fill with different colors per frame
		colorIndex := uint8((i % 7) + 1)
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				img.SetColorIndex(x, y, colorIndex)
			}
		}

		// Draw a pattern to make it interesting
		for x := 0; x < size; x++ {
			img.SetColorIndex(x, x, 0) // diagonal line
			if x < size-1 {
				img.SetColorIndex(x, size-1-x, 7) // other diagonal
			}
		}

		images = append(images, img)
		delays = append(delays, 20) // 200ms delay
	}

	f, _ := os.Create(filename)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})
}

func generateFastGIF(filename string) {
	const size = 32
	palette := color.Palette{
		color.RGBA{0x00, 0x00, 0x00, 0xff},
		color.RGBA{0xff, 0x00, 0x00, 0xff},
		color.RGBA{0x00, 0xff, 0x00, 0xff},
		color.RGBA{0x00, 0x00, 0xff, 0xff},
	}

	var images []*image.Paletted
	var delays []int

	for i := 0; i < 4; i++ {
		img := image.NewPaletted(image.Rect(0, 0, size, size), palette)

		colorIndex := uint8(i % 4)
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				img.SetColorIndex(x, y, colorIndex)
			}
		}

		images = append(images, img)
		delays = append(delays, 5) // 50ms - fast!
	}

	f, _ := os.Create(filename)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})
}

func generateDisposalTestGIF(filename string) {
	const size = 48
	palette := color.Palette{
		color.RGBA{0x00, 0x00, 0x00, 0x00}, // transparent
		color.RGBA{0xff, 0x00, 0x00, 0xff}, // red
		color.RGBA{0x00, 0xff, 0x00, 0xff}, // green
		color.RGBA{0x00, 0x00, 0xff, 0xff}, // blue
	}

	var images []*image.Paletted
	var delays []int
	var disposals []byte

	// Frame 1: Red square
	img1 := image.NewPaletted(image.Rect(0, 0, size, size), palette)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img1.SetColorIndex(x, y, 1)
		}
	}

	// Frame 2: Green circle (smaller)
	img2 := image.NewPaletted(image.Rect(0, 0, size, size), palette)
	center := size / 2
	radius := size / 4
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - center
			dy := y - center
			if dx*dx+dy*dy < radius*radius {
				img2.SetColorIndex(x, y, 2)
			}
		}
	}

	// Frame 3: Blue square (smaller)
	img3 := image.NewPaletted(image.Rect(0, 0, size, size), palette)
	for y := size / 4; y < 3*size/4; y++ {
		for x := size / 4; x < 3*size/4; x++ {
			img3.SetColorIndex(x, y, 3)
		}
	}

	images = append(images, img1, img2, img3)
	delays = append(delays, 30, 30, 30)
	disposals = append(disposals, gif.DisposalNone, gif.DisposalBackground, gif.DisposalPrevious)

	f, _ := os.Create(filename)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image:    images,
		Delay:    delays,
		Disposal: disposals,
	})
}
