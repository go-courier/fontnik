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
				img := DrawGlyph(g, true)
				SavePNG(fmt.Sprintf("./testdata/%d.png", i), img)
			}
		}
	})

	t.Run("#Glyphs", func(t *testing.T) {
		s := builder.Glyphs(0, 255)
		bytes, err := proto.Marshal(s)
		require.NoError(t, err)
		ioutil.WriteFile("./testdata/0-255.pbf", bytes, os.ModePerm)
	})
}
