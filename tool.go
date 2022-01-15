//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/shabbyrobe/wu2quant"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"os"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var colors int
	var euclidean bool

	fs := flag.NewFlagSet("", 0)
	fs.IntVar(&colors, "colors", 256, "Number of colors")
	fs.BoolVar(&euclidean, "euclidean", false, "Use wu2 to build palette only; use Euclidean distance to update colours")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}
	args := fs.Args()
	if len(args) != 2 {
		return fmt.Errorf("usage: tool.go [flags] <in> <out.png>")
	}

	bts, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	img, err := decode(bts)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	q := wu2quant.New()

	var out image.Image

	if euclidean {
		rg := convertImageToRGBA(img)
		out = rg
		pal := make(color.Palette, 0, 16)
		pal = q.QuantizeRGBAToPalette(pal, rg)

		for y := 0; y < img.Bounds().Dy(); y++ {
			for x := 0; x < img.Bounds().Dx(); x++ {
				rg.Set(x, y, pal.Convert(rg.At(x, y)))
			}
		}
	} else {
		out, err = q.ToPaletted(colors, img, nil)
		if err != nil {
			return err
		}
	}

	if err := png.Encode(&buf, out); err != nil {
		return err
	}

	if args[1] == "-" {
		os.Stdout.Write(buf.Bytes())
	} else {
		if err := os.WriteFile(args[1], buf.Bytes(), 0600); err != nil {
			return err
		}
	}

	return nil
}

func decode(bts []byte) (image.Image, error) {
	if bytes.HasPrefix(bts, []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}) {
		return png.Decode(bytes.NewReader(bts))
	} else if bytes.HasPrefix(bts, []byte{0xff, 0xd8}) {
		return jpeg.Decode(bytes.NewReader(bts))
	} else {
		return nil, fmt.Errorf("unknown format")
	}
}

func convertImageToRGBA(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	size := bounds.Size()
	pix := make([]uint8, size.X*size.Y*4)

	var idx int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pix[idx] = uint8(r >> 8)
			pix[idx+1] = uint8(g >> 8)
			pix[idx+2] = uint8(b >> 8)
			pix[idx+3] = uint8(a >> 8)
			idx += 4
		}
	}

	return &image.RGBA{Rect: bounds, Stride: size.X * 4, Pix: pix}
}
