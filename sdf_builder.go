package fontnik

import (
	"fmt"
	"image"
	"math"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

//go:generate protoc --go_out=. glyph.proto

func NewSDFBuilder(font *truetype.Font, opts ...SDFBuilderOpt) *SDFBuilder {
	sdfBuilder := &SDFBuilder{
		Font: font,
	}

	for _, opt := range opts {
		if opt.FontSize != 0 {
			sdfBuilder.FontSize = opt.FontSize
		}
		if opt.Buffer != 0 {
			sdfBuilder.Buffer = opt.Buffer
		}
	}

	sdfBuilder.Init()

	return sdfBuilder
}

type SDFBuilderOpt struct {
	FontSize float64
	Buffer   float64
}

type SDFBuilder struct {
	Font *truetype.Font
	Face font.Face
	SDFBuilderOpt
	dotStartY int
}

func (b *SDFBuilder) Init() {
	if b.FontSize == 0 {
		b.FontSize = 64
	}
	if b.Buffer == 0 {
		b.Buffer = 3
	}

	b.Face = truetype.NewFace(b.Font, &truetype.Options{
		Size:    b.FontSize,
		Hinting: font.HintingFull,
	})

	metrics := b.Face.Metrics()

	// https://developer.apple.com/library/archive/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyph_metrics_2x.png

	fontDesignedHeight := metrics.Ascent.Floor() + metrics.Descent.Floor()
	fixed := int(math.Round(float64(metrics.Height.Floor()-fontDesignedHeight)/2)) + 1

	b.dotStartY = metrics.Height.Floor() + metrics.Descent.Floor() + fixed
}

func (b *SDFBuilder) Glyphs(min int, max int) *Glyphs {
	rng := fmt.Sprintf("%d-%d", min, max)
	fontFamily := b.Font.Name(truetype.NameIDFontFullName)

	stack := &Fontstack{}
	stack.Range = &rng
	stack.Name = &fontFamily

	for i := min; i < max; i++ {
		g := b.Glyph(rune(i))
		if g != nil {
			stack.Glyphs = append(stack.Glyphs, g)
		}
	}

	return &Glyphs{
		Stacks: []*Fontstack{
			stack,
		},
	}
}

func (b *SDFBuilder) Glyph(x rune) *Glyph {
	if x == 0 {
		return nil
	}

	i := b.Font.Index(x)
	if i == 0 {
		return nil
	}

	bounds, mask, maskp, advance, ok := b.Face.Glyph(fixed.P(0, b.dotStartY), x)
	if !ok {
		return nil
	}

	size := bounds.Size()

	width := uint32(size.X)
	height := uint32(size.Y)

	if width == 0 || height == 0 {
		return nil
	}

	buffer := int(b.Buffer)
	id := uint32(x)

	top := -int32(bounds.Min.Y)
	left := int32(bounds.Min.X)
	a := uint32(advance.Floor())

	g := &Glyph{
		Id:      &id,
		Width:   &width,
		Height:  &height,
		Left:    &left,
		Top:     &top,
		Advance: &a,
	}

	w := int(*g.Width) + buffer*2
	h := int(*g.Height) + buffer*2

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.DrawMask(dst, dst.Bounds(), &image.Uniform{image.Black}, image.ZP, mask, maskp.Sub(image.Pt(buffer, buffer)), draw.Over)

	g.Bitmap = CalcSDF(dst, 8, 0.25)

	return g
}

const INF = 1e20

func CalcSDF(img image.Image, radius float64, cutoff float64) []uint8 {
	size := img.Bounds().Size()
	w, h := size.X, size.Y

	gridOuter := make([]float64, w*h)
	gridInner := make([]float64, w*h)

	f := make([]float64, w*h)
	d := make([]float64, w*h)
	v := make([]float64, w*h)
	z := make([]float64, w*h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := x + y*w
			_, _, _, a := img.At(x, y).RGBA()

			alpha := float64(a) / math.MaxUint16

			outer := float64(0)
			inner := INF

			if alpha != 1 {
				if alpha == 0 {
					outer = INF
					inner = 0
				} else {
					outer = math.Pow(math.Max(0, 0.5-alpha), 2)
					inner = math.Pow(math.Max(0, alpha-0.5), 2)
				}
			}

			gridOuter[i] = outer
			gridInner[i] = inner
		}
	}

	edt(gridOuter, w, h, f, d, v, z)
	edt(gridInner, w, h, f, d, v, z)

	alphas := make([]uint8, w*h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := x + y*w
			d := gridOuter[i] - gridInner[i]

			a := math.Max(0, math.Min(255, math.Round(255-255*(d/radius+cutoff))))

			alphas[i] = uint8(a)
		}
	}

	return alphas
}

// 2D Euclidean distance transform by Felzenszwalb & Huttenlocher https://cs.brown.edu/~pff/papers/dt-final.pdf
func edt(data []float64, width int, height int, f []float64, d []float64, v []float64, z []float64) {
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			f[y] = data[y*width+x]
		}

		edt1d(f, d, v, z, height)

		for y := 0; y < height; y++ {
			data[y*width+x] = d[y]
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			f[x] = data[y*width+x]
		}

		edt1d(f, d, v, z, width)

		for x := 0; x < width; x++ {
			data[y*width+x] = math.Sqrt(d[x])
		}
	}
}

// 1D squared distance transform
func edt1d(f []float64, d []float64, v []float64, z []float64, n int) {
	v[0] = 0
	z[0] = -INF
	z[1] = +INF

	for q, k := 1, 0; q < (n); q++ {
		getS := func() float64 {
			return ((f[q] + float64(q)*float64(q)) - (f[int(v[k])] + v[k]*v[k])) / (2*float64(q) - 2*v[k])
		}

		s := getS()

		for {
			if s <= float64(z[k]) {
				k--
				s = getS()
				continue
			}
			break
		}

		k++

		v[k] = float64(q)
		z[k] = float64(s)
		z[k+1] = +INF
	}

	for q, k := 0, 0; q < n; q++ {
		for {
			if z[k+1] < float64(q) {
				k++
				continue
			}
			break
		}

		d[q] = (float64(q)-v[k])*(float64(q)-v[k]) + f[int(v[k])]
	}
}
