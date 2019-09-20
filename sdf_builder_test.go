package fontnik

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/golang/freetype/truetype"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
)

func TestSDFBuilder(t *testing.T) {
	ttf, err := ioutil.ReadFile("./testdata/Lato-Regular.ttf")
	require.NoError(t, err)

	font, err := truetype.Parse(ttf)
	require.NoError(t, err)

	builder := NewSDFBuilder(font, SDFBuilderOpt{FontSize: 24, Buffer: 3})

	t.Run("#Glyph", func(t *testing.T) {
		for i := 0; i < 255; i++ {
			g := builder.Glyph(rune(i))
			if g != nil {
				fmt.Printf("%s %d\n", string(*g.Id), *g.Top)
				img := DrawGlyph(g, true)
				SavePNG(fmt.Sprintf("./testdata/Lato/%d.png", i), img)
			}
		}
	})

	t.Run("#Glyphs", func(t *testing.T) {
		for _, rng := range [][]int{
			{0, 255},
			//{20224, 20479},
			//{22784, 23039},
		} {
			s := builder.Glyphs(rng[0], rng[1])
			bytes, err := proto.Marshal(s)
			require.NoError(t, err)
			ioutil.WriteFile(fmt.Sprintf("./testdata/Lato/%d-%d.pbf", rng[0], rng[1]), bytes, os.ModePerm)
		}
	})
}
