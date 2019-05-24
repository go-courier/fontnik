package fontnik

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

func DrawGlyph(glyph *Glyph, smoothstep bool) image.Image {
	width := int(*glyph.Width + 6)
	height := int(*glyph.Height + 6)

	img := image.NewRGBA(image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{width, height}})

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := x + y*width
			a := glyph.Bitmap[i]

			if smoothstep {
				img.Set(x, y, color.RGBA{0, 0, 0,
					alpha(136, 168, float64(a)),
				})
			} else {
				img.Set(x, y, color.RGBA{0, 0, 0,
					uint8(a),
				})
			}

		}
	}

	return img
}

func alpha(e0 float64, e1 float64, x float64) uint8 {
	a := math.Max(math.Min((x-e1)/(e1-e0), 1), 0)
	return uint8((a * a * (3 - 2*a)) * float64(x))
}

func SavePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	err = png.Encode(f, img)
	if err != nil {
		log.Fatal(err)
	}
}
