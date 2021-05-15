package wu2quant

import (
	"image"
	"image/color"
	"math/rand"
	"reflect"
	"testing"
)

func genRGBAWithUniqueRGBPerPixel(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	jump := (1 << 24) / (w * h)

	v := 0
	for idx := 0; idx < len(img.Pix); idx += 4 {
		c := v<<8 | 0xff
		img.Pix[idx] = uint8(c >> 24)
		img.Pix[idx+1] = uint8(c >> 16)
		img.Pix[idx+2] = uint8(c >> 8)
		img.Pix[idx+3] = uint8(c)
		v += jump
	}
	return img
}

func genRandomRGBAPalette(rng *rand.Rand, n int) []color.RGBA {
	pal := make([]color.RGBA, n)
	for i := range pal {
		v := rng.Uint64() & (1<<24 - 1)
		c := v<<8 | 0xff
		pal[i] = color.RGBA{uint8(c >> 24),
			uint8(c >> 16),
			uint8(c >> 8),
			uint8(c),
		}
	}
	return pal
}

func genRGBAWithRandomRGBPerPixel(rng *rand.Rand, w, h int) *image.RGBA {
	if rng == nil {
		rng = rand.New(rand.NewSource(0))
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for idx := 0; idx < len(img.Pix); idx += 4 {
		v := rng.Uint64() & (1<<24 - 1)
		c := v<<8 | 0xff
		img.Pix[idx] = uint8(c >> 24)
		img.Pix[idx+1] = uint8(c >> 16)
		img.Pix[idx+2] = uint8(c >> 8)
		img.Pix[idx+3] = uint8(c)
	}
	return img
}

func TestQuantizePixels(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	pal := genRandomRGBAPalette(rand.New(rand.NewSource(0)), 10)

	for i, j := 0, 9; i < j; i, j = i+1, j-1 {
		for p := i; p <= j; p++ {
			img.SetRGBA(i, p, pal[i])
			img.SetRGBA(j, p, pal[i])
			img.SetRGBA(p, i, pal[i])
			img.SetRGBA(p, j, pal[i])
		}
	}

	result, err := New().ToPaletted(8, img, nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := []uint8{
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
		4, 2, 2, 2, 2, 2, 2, 2, 2, 4,
		4, 2, 3, 3, 3, 3, 3, 3, 2, 4,
		4, 2, 3, 1, 1, 1, 1, 3, 2, 4,
		4, 2, 3, 1, 0, 0, 1, 3, 2, 4,
		4, 2, 3, 1, 0, 0, 1, 3, 2, 4,
		4, 2, 3, 1, 1, 1, 1, 3, 2, 4,
		4, 2, 3, 3, 3, 3, 3, 3, 2, 4,
		4, 2, 2, 2, 2, 2, 2, 2, 2, 4,
		4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	}

	if !reflect.DeepEqual(expected, result.Pix) {
		t.Fatal()
	}
}

func TestQuantizeFromSubimage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	pal := genRandomRGBAPalette(rand.New(rand.NewSource(0)), 10)

	for i, j := 0, 9; i < j; i, j = i+1, j-1 {
		for p := i; p <= j; p++ {
			img.SetRGBA(i, p, pal[i])
			img.SetRGBA(j, p, pal[i])
			img.SetRGBA(p, i, pal[i])
			img.SetRGBA(p, j, pal[i])
		}
	}

	sub := img.SubImage(image.Rect(2, 2, 8, 8))
	result, err := New().ToPaletted(8, sub, nil)
	if err != nil {
		t.Fatal(err)
	}

	expected := []uint8{
		2, 2, 2, 2, 2, 2,
		2, 1, 1, 1, 1, 2,
		2, 1, 0, 0, 1, 2,
		2, 1, 0, 0, 1, 2,
		2, 1, 1, 1, 1, 2,
		2, 2, 2, 2, 2, 2,
	}

	if !reflect.DeepEqual(expected, result.Pix) {
		t.Fatal()
	}
}

func TestQuantizeWithRecycledQuantizer(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	img1 := genRGBAWithRandomRGBPerPixel(rng, 512, 256)
	img2 := genRGBAWithRandomRGBPerPixel(rng, 512, 256)
	if reflect.DeepEqual(img1.Pix, img2.Pix) {
		t.Fatal()
	}

	// Get the expected palette for img2:
	expected := New().QuantizeRGBAToPalette(make(color.Palette, 0, 8), img2)

	// Quantize img1 first to fill the structures up with stuff:
	q := New()
	q.QuantizeRGBAToPalette(make(color.Palette, 0, 8), img1)

	// Quantize img2 with the recycled quantizer:
	result := q.QuantizeRGBAToPalette(make(color.Palette, 0, 8), img2)

	if !reflect.DeepEqual(result, expected) {
		t.Fatal()
	}
}

func TestQuantizeTo4(t *testing.T) {
	img := genRGBAWithUniqueRGBPerPixel(512, 256)
	q := New()

	pal := make(color.Palette, 0, 4)
	out := q.Quantize(pal[:0], img)
	exp := color.Palette{
		color.RGBA{63, 63, 64, 255},
		color.RGBA{191, 63, 64, 255},
		color.RGBA{63, 191, 64, 255},
		color.RGBA{191, 191, 64, 255}}

	if !reflect.DeepEqual(out, exp) {
		t.Fatal()
	}
}

func TestQuantizeTo256(t *testing.T) {
	img := genRGBAWithUniqueRGBPerPixel(512, 256)
	q := New()

	pal := make(color.Palette, 0, 256)
	out := q.Quantize(pal[:0], img)
	exp := color.Palette{
		color.RGBA{7, 15, 0, 255}, color.RGBA{135, 15, 0, 255}, color.RGBA{7, 143, 0, 255}, color.RGBA{135, 143, 0, 255}, color.RGBA{7, 15, 128, 255}, color.RGBA{135, 15, 128, 255}, color.RGBA{7, 143, 128, 255}, color.RGBA{135, 143, 128, 255}, color.RGBA{71, 15, 0, 255}, color.RGBA{199, 15, 0, 255}, color.RGBA{71, 143, 0, 255}, color.RGBA{199, 143, 0, 255}, color.RGBA{71, 15, 128, 255}, color.RGBA{199, 15, 128, 255}, color.RGBA{71, 143, 128, 255}, color.RGBA{199, 143, 128, 255}, color.RGBA{7, 79, 0, 255}, color.RGBA{135, 79, 0, 255}, color.RGBA{7, 207, 0, 255}, color.RGBA{135, 207, 0, 255}, color.RGBA{7, 79, 128, 255}, color.RGBA{135, 79, 128, 255}, color.RGBA{7, 207, 128, 255}, color.RGBA{135, 207, 128, 255}, color.RGBA{71, 79, 0, 255}, color.RGBA{199, 79, 0, 255}, color.RGBA{71, 207, 0, 255}, color.RGBA{199, 207, 0, 255}, color.RGBA{71, 79, 128, 255}, color.RGBA{199, 79, 128, 255}, color.RGBA{71, 207, 128, 255}, color.RGBA{199, 207, 128, 255},
		color.RGBA{39, 15, 0, 255}, color.RGBA{167, 15, 0, 255}, color.RGBA{39, 143, 0, 255}, color.RGBA{167, 143, 0, 255}, color.RGBA{39, 15, 128, 255}, color.RGBA{167, 15, 128, 255}, color.RGBA{39, 143, 128, 255}, color.RGBA{167, 143, 128, 255}, color.RGBA{103, 15, 0, 255}, color.RGBA{231, 15, 0, 255}, color.RGBA{103, 143, 0, 255}, color.RGBA{231, 143, 0, 255}, color.RGBA{103, 15, 128, 255}, color.RGBA{231, 15, 128, 255}, color.RGBA{103, 143, 128, 255}, color.RGBA{231, 143, 128, 255}, color.RGBA{39, 79, 0, 255}, color.RGBA{167, 79, 0, 255}, color.RGBA{39, 207, 0, 255}, color.RGBA{167, 207, 0, 255}, color.RGBA{39, 79, 128, 255}, color.RGBA{167, 79, 128, 255}, color.RGBA{39, 207, 128, 255}, color.RGBA{167, 207, 128, 255}, color.RGBA{103, 79, 0, 255}, color.RGBA{231, 79, 0, 255}, color.RGBA{103, 207, 0, 255}, color.RGBA{231, 207, 0, 255}, color.RGBA{103, 79, 128, 255}, color.RGBA{231, 79, 128, 255}, color.RGBA{103, 207, 128, 255}, color.RGBA{231, 207, 128, 255},
		color.RGBA{7, 47, 0, 255}, color.RGBA{135, 47, 0, 255}, color.RGBA{7, 175, 0, 255}, color.RGBA{135, 175, 0, 255}, color.RGBA{7, 47, 128, 255}, color.RGBA{135, 47, 128, 255}, color.RGBA{7, 175, 128, 255}, color.RGBA{135, 175, 128, 255}, color.RGBA{71, 47, 0, 255}, color.RGBA{199, 47, 0, 255}, color.RGBA{71, 175, 0, 255}, color.RGBA{199, 175, 0, 255}, color.RGBA{71, 47, 128, 255}, color.RGBA{199, 47, 128, 255}, color.RGBA{71, 175, 128, 255}, color.RGBA{199, 175, 128, 255}, color.RGBA{7, 111, 0, 255}, color.RGBA{135, 111, 0, 255}, color.RGBA{7, 239, 0, 255}, color.RGBA{135, 239, 0, 255}, color.RGBA{7, 111, 128, 255}, color.RGBA{135, 111, 128, 255}, color.RGBA{7, 239, 128, 255}, color.RGBA{135, 239, 128, 255}, color.RGBA{71, 111, 0, 255}, color.RGBA{199, 111, 0, 255}, color.RGBA{71, 239, 0, 255}, color.RGBA{199, 239, 0, 255}, color.RGBA{71, 111, 128, 255}, color.RGBA{199, 111, 128, 255}, color.RGBA{71, 239, 128, 255}, color.RGBA{199, 239, 128, 255},
		color.RGBA{39, 47, 0, 255}, color.RGBA{167, 47, 0, 255}, color.RGBA{39, 175, 0, 255}, color.RGBA{167, 175, 0, 255}, color.RGBA{39, 47, 128, 255}, color.RGBA{167, 47, 128, 255}, color.RGBA{39, 175, 128, 255}, color.RGBA{167, 175, 128, 255}, color.RGBA{103, 47, 0, 255}, color.RGBA{231, 47, 0, 255}, color.RGBA{103, 175, 0, 255}, color.RGBA{231, 175, 0, 255}, color.RGBA{103, 47, 128, 255}, color.RGBA{231, 47, 128, 255}, color.RGBA{103, 175, 128, 255}, color.RGBA{231, 175, 128, 255}, color.RGBA{39, 111, 0, 255}, color.RGBA{167, 111, 0, 255}, color.RGBA{39, 239, 0, 255}, color.RGBA{167, 239, 0, 255}, color.RGBA{39, 111, 128, 255}, color.RGBA{167, 111, 128, 255}, color.RGBA{39, 239, 128, 255}, color.RGBA{167, 239, 128, 255}, color.RGBA{103, 111, 0, 255}, color.RGBA{231, 111, 0, 255}, color.RGBA{103, 239, 0, 255}, color.RGBA{231, 239, 0, 255}, color.RGBA{103, 111, 128, 255}, color.RGBA{231, 111, 128, 255}, color.RGBA{103, 239, 128, 255}, color.RGBA{231, 239, 128, 255},
		color.RGBA{23, 15, 0, 255}, color.RGBA{151, 15, 0, 255}, color.RGBA{23, 143, 0, 255}, color.RGBA{151, 143, 0, 255}, color.RGBA{23, 15, 128, 255}, color.RGBA{151, 15, 128, 255}, color.RGBA{23, 143, 128, 255}, color.RGBA{151, 143, 128, 255}, color.RGBA{87, 15, 0, 255}, color.RGBA{215, 15, 0, 255}, color.RGBA{87, 143, 0, 255}, color.RGBA{215, 143, 0, 255}, color.RGBA{87, 15, 128, 255}, color.RGBA{215, 15, 128, 255}, color.RGBA{87, 143, 128, 255}, color.RGBA{215, 143, 128, 255}, color.RGBA{23, 79, 0, 255}, color.RGBA{151, 79, 0, 255}, color.RGBA{23, 207, 0, 255}, color.RGBA{151, 207, 0, 255}, color.RGBA{23, 79, 128, 255}, color.RGBA{151, 79, 128, 255}, color.RGBA{23, 207, 128, 255}, color.RGBA{151, 207, 128, 255}, color.RGBA{87, 79, 0, 255}, color.RGBA{215, 79, 0, 255}, color.RGBA{87, 207, 0, 255}, color.RGBA{215, 207, 0, 255}, color.RGBA{87, 79, 128, 255}, color.RGBA{215, 79, 128, 255}, color.RGBA{87, 207, 128, 255}, color.RGBA{215, 207, 128, 255},
		color.RGBA{55, 15, 0, 255}, color.RGBA{183, 15, 0, 255}, color.RGBA{55, 143, 0, 255}, color.RGBA{183, 143, 0, 255}, color.RGBA{55, 15, 128, 255}, color.RGBA{183, 15, 128, 255}, color.RGBA{55, 143, 128, 255}, color.RGBA{183, 143, 128, 255}, color.RGBA{119, 15, 0, 255}, color.RGBA{247, 15, 0, 255}, color.RGBA{119, 143, 0, 255}, color.RGBA{247, 143, 0, 255}, color.RGBA{119, 15, 128, 255}, color.RGBA{247, 15, 128, 255}, color.RGBA{119, 143, 128, 255}, color.RGBA{247, 143, 128, 255}, color.RGBA{55, 79, 0, 255}, color.RGBA{183, 79, 0, 255}, color.RGBA{55, 207, 0, 255}, color.RGBA{183, 207, 0, 255}, color.RGBA{55, 79, 128, 255}, color.RGBA{183, 79, 128, 255}, color.RGBA{55, 207, 128, 255}, color.RGBA{183, 207, 128, 255}, color.RGBA{119, 79, 0, 255}, color.RGBA{247, 79, 0, 255}, color.RGBA{119, 207, 0, 255}, color.RGBA{247, 207, 0, 255}, color.RGBA{119, 79, 128, 255}, color.RGBA{247, 79, 128, 255}, color.RGBA{119, 207, 128, 255}, color.RGBA{247, 207, 128, 255},
		color.RGBA{23, 47, 0, 255}, color.RGBA{151, 47, 0, 255}, color.RGBA{23, 175, 0, 255}, color.RGBA{151, 175, 0, 255}, color.RGBA{23, 47, 128, 255}, color.RGBA{151, 47, 128, 255}, color.RGBA{23, 175, 128, 255}, color.RGBA{151, 175, 128, 255}, color.RGBA{87, 47, 0, 255}, color.RGBA{215, 47, 0, 255}, color.RGBA{87, 175, 0, 255}, color.RGBA{215, 175, 0, 255}, color.RGBA{87, 47, 128, 255}, color.RGBA{215, 47, 128, 255}, color.RGBA{87, 175, 128, 255}, color.RGBA{215, 175, 128, 255}, color.RGBA{23, 111, 0, 255}, color.RGBA{151, 111, 0, 255}, color.RGBA{23, 239, 0, 255}, color.RGBA{151, 239, 0, 255}, color.RGBA{23, 111, 128, 255}, color.RGBA{151, 111, 128, 255}, color.RGBA{23, 239, 128, 255}, color.RGBA{151, 239, 128, 255}, color.RGBA{87, 111, 0, 255}, color.RGBA{215, 111, 0, 255}, color.RGBA{87, 239, 0, 255}, color.RGBA{215, 239, 0, 255}, color.RGBA{87, 111, 128, 255}, color.RGBA{215, 111, 128, 255}, color.RGBA{87, 239, 128, 255}, color.RGBA{215, 239, 128, 255},
		color.RGBA{55, 47, 0, 255}, color.RGBA{183, 47, 0, 255}, color.RGBA{55, 175, 0, 255}, color.RGBA{183, 175, 0, 255}, color.RGBA{55, 47, 128, 255}, color.RGBA{183, 47, 128, 255}, color.RGBA{55, 175, 128, 255}, color.RGBA{183, 175, 128, 255}, color.RGBA{119, 47, 0, 255}, color.RGBA{247, 47, 0, 255}, color.RGBA{119, 175, 0, 255}, color.RGBA{247, 175, 0, 255}, color.RGBA{119, 47, 128, 255}, color.RGBA{247, 47, 128, 255}, color.RGBA{119, 175, 128, 255}, color.RGBA{247, 175, 128, 255}, color.RGBA{55, 111, 0, 255}, color.RGBA{183, 111, 0, 255}, color.RGBA{55, 239, 0, 255}, color.RGBA{183, 239, 0, 255}, color.RGBA{55, 111, 128, 255}, color.RGBA{183, 111, 128, 255}, color.RGBA{55, 239, 128, 255}, color.RGBA{183, 239, 128, 255}, color.RGBA{119, 111, 0, 255}, color.RGBA{247, 111, 0, 255}, color.RGBA{119, 239, 0, 255}, color.RGBA{247, 239, 0, 255}, color.RGBA{119, 111, 128, 255}, color.RGBA{247, 111, 128, 255}, color.RGBA{119, 239, 128, 255}, color.RGBA{247, 239, 128, 255},
	}

	if !reflect.DeepEqual(out, exp) {
		t.Fatal()
	}
}

func TestIntoPaletted(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	img1 := genRGBAWithRandomRGBPerPixel(rng, 512, 256)
	img2 := genRGBAWithRandomRGBPerPixel(rng, 512, 256)
	if reflect.DeepEqual(img1.Pix, img2.Pix) {
		t.Fatal()
	}

	exp1, err := New().ToPaletted(8, img1, nil)
	if err != nil {
		t.Fatal(err)
	}

	exp2, err := New().ToPaletted(8, img2, nil)
	if err != nil {
		t.Fatal(err)
	}

	var q = New()
	var pal = image.NewPaletted(img1.Bounds(), nil)
	if err := q.IntoPaletted(8, img1, pal, nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp1, pal) {
		t.Fatal()
	}

	if err := q.IntoPaletted(8, img2, pal, nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp2, pal) {
		t.Fatal()
	}

	// Size mismach should return error:
	var palsml = image.NewPaletted(image.Rect(0, 0, 1, 1), nil)
	if err := q.IntoPaletted(8, img1, palsml, nil); err == nil {
		t.Fatal()
	}
}

func TestIntoPalettedWithBuffer(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	img1 := genRGBAWithRandomRGBPerPixel(rng, 100, 100)

	// 100x100 image into buffer:
	buf := NewBuffer(0)
	exp1, err := New().ToPaletted(8, img1, buf)
	if err != nil {
		t.Fatal(err)
	}

	// 200x200 image into buffer:
	img2 := genRGBAWithRandomRGBPerPixel(rng, 200, 200)
	exp2, err := New().ToPaletted(8, img2, buf)
	if err != nil {
		t.Fatal(err)
	}

	// First 100x100 image into buffer:
	exp1Check, err := New().ToPaletted(8, img1, buf)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp1.Pix, exp1Check.Pix) {
		t.Fatal()
	}

	// First 200x200 image into buffer should produce same result:
	exp2Check, err := New().ToPaletted(8, img2, buf)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exp2.Pix, exp2Check.Pix) {
		t.Fatal()
	}
}

func BenchmarkToPaletted(b *testing.B) {
	b.ReportAllocs()
	img := genRGBAWithUniqueRGBPerPixel(512, 256)
	q := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.ToPaletted(256, img, nil)
	}
}

func BenchmarkIntoPaletted(b *testing.B) {
	b.ReportAllocs()
	img := genRGBAWithUniqueRGBPerPixel(512, 256)
	q := New()
	dest := image.NewPaletted(img.Rect, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.IntoPaletted(256, img, dest, nil)
	}
}

func BenchmarkIntoPalettedBuffer(b *testing.B) {
	b.ReportAllocs()
	img := genRGBAWithUniqueRGBPerPixel(512, 256)
	buf := NewBuffer(512 * 256)
	q := New()
	dest := image.NewPaletted(img.Rect, make(color.Palette, 256))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.IntoPaletted(len(dest.Palette), img, dest, buf)
	}
}

func BenchmarkRGBAIntoPalettedBuffer(b *testing.B) {
	b.ReportAllocs()
	img := genRGBAWithUniqueRGBPerPixel(512, 256)
	buf := NewBuffer(512 * 256)
	q := New()
	dest := image.NewPaletted(img.Rect, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.RGBAIntoPaletted(256, img, dest, buf)
	}
}

func BenchmarkQuantize512x256(b *testing.B) {
	b.ReportAllocs()
	img := genRGBAWithUniqueRGBPerPixel(512, 256)
	q := New()

	pal := make(color.Palette, 0, 256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Quantize(pal[:0], img)
	}
}

func BenchmarkQuantize2048x2048(b *testing.B) {
	b.ReportAllocs()
	img := genRGBAWithUniqueRGBPerPixel(2048, 2048)
	q := New()

	pal := make(color.Palette, 0, 256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Quantize(pal[:0], img)
	}
}

func BenchmarkQuantizeRGBA512x256(b *testing.B) {
	b.ReportAllocs()
	img := genRGBAWithUniqueRGBPerPixel(512, 256)
	q := New()

	pal := make([]color.RGBA, 0, 256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.QuantizeRGBA(pal[:0], img)
	}
}
