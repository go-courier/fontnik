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

func builderFor(fontFamily string) *SDFBuilder {
	ttf, err := ioutil.ReadFile("./testdata/" + fontFamily + ".ttf")
	if err != nil {
		panic(err)
	}

	font, err := truetype.Parse(ttf)

	if err != nil {
		panic(err)
	}

	return NewSDFBuilder(font, SDFBuilderOpt{FontSize: 24, Buffer: 3})
}

func TestSDFBuilder_Glyph(t *testing.T) {
	builder := builderFor("NotoSans-Regular")

	for i := 0; i < 255; i++ {
		g := builder.Glyph(rune(i))
		if g != nil {
			fmt.Printf("%s %d\n", string(*g.Id), *g.Top)
			img := DrawGlyph(g, true)
			SavePNG(fmt.Sprintf("./testdata/NotoSans/%d.png", i), img)
		}
	}
}

func TestSDFBuilder(t *testing.T) {
	t.Run("#Glyphs", func(t *testing.T) {
		builder := builderFor("NotoSans-Regular")

		for _, rng := range [][]int{
			{0, 255},
			{20224, 20479},
			{22784, 23039},
		} {
			s := builder.Glyphs(rng[0], rng[1])
			bytes, err := proto.Marshal(s)
			require.NoError(t, err)
			ioutil.WriteFile(fmt.Sprintf("./testdata/NotoSans/%d-%d.pbf", rng[0], rng[1]), bytes, os.ModePerm)
		}
	})
}
