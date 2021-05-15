//+build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/shabbyrobe/wu2quant"
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

	fs := flag.NewFlagSet("", 0)
	fs.IntVar(&colors, "colors", 256, "Number of colors")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}
	args := fs.Args()
	if len(args) != 2 {
		return fmt.Errorf("usage: tool.go [flags] <in.png> <out.png>")
	}

	bts, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}

	img, err := png.Decode(bytes.NewReader(bts))
	if err != nil {
		return err
	}

	q := wu2quant.New()
	pal, err := q.ToPaletted(colors, img, nil)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, pal); err != nil {
		return err
	}

	if err := os.WriteFile(args[1], buf.Bytes(), 0600); err != nil {
		return err
	}

	return nil
}
