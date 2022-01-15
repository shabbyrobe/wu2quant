Go implementation of Wu's Color Quantizer
=========================================

Reduces the number of colours in an *image.RGBA to fit in the desired palette
size. Quality is decent for the speed; it's much faster than NeuQuant and I
can't really tell the difference in the output (YMMV).

![](https://github.com/shabbyrobe/wu2quant/blob/main/demo/kodim23.apng?raw=true)

Originally invented by [Xiaolin Wu](https://www.ece.mcmaster.ca/~xwu/)

Supports [image/draw.Quantizer](https://golang.org/pkg/image/draw/#Quantizer):

```go
jpg, err := jpeg.Decode(buf)
wu2 := wu2quant.New()

// Quantize the colors in jpg to a 256 color palette:
palette := wu2.Quantize(make(color.Palette, 0, 256), jpg)
```

Convert an existing image into a quantized, paletted version in a single call::

```go
jpg, err := jpeg.Decode(buf)
wu2 := wu2quant.New()

// Quantize the colors in jpg to a 256 color palette:
paletted := wu2.ToPaletted(256, jpg, nil)
```

Reduce allocs by recycling the quantizer::

```go
wu2 := wu2quant.New()

jpg1, err := jpeg.Decode(buf1)
paletted, err := wu2.ToPaletted(256, jpg1, nil)

jpg2, err := jpeg.Decode(buf2)
paletted2, err := wu2.ToPaletted(256, jpg2, nil)
```

If the paletted image doesn't need to be retained, you can also reduce allocs
by recycling the image.Paletted (if the output is the same size as the input):

```go
wu2 := wu2quant.New()

jpg1, err := jpeg.Decode(buf1)
jpg2, err := jpeg.Decode(buf2)

frame := image.NewPaletted(jpg1.Bounds(), nil)
err := wu2.IntoPaletted(256, jpg1, frame, nil)
err := wu2.IntoPaletted(256, jpg2, frame, nil)
```

If you expect to quantize many images in series, you can reduce allocs substantially
by reusing the temporary buffer (the buffer does not depend on size, it will grow
to the largest `x*y*sizeof(uint16)` of the images you quantize with it):

```go
// Can also pre-size with wu2quant.NewBuffer(sz)
var buf wu2quant.Buffer

wu2 := wu2quant.New()

jpg1, err := jpeg.Decode(buf1)
frame := image.NewPaletted(jpg1.Bounds(), nil)
err := wu2.IntoPaletted(256, jpg1, frame, &buf)
```


## Possible future stuff

We can quantise YCbCr directly without needing an RGBA conversion. We may
be able to do the same with CIELAB colours, if they're represented as 8-bit
ints rather than float64s.

The 'moment' type tolerates being converted to float64 without adversely
affecting performance and without altering the result at all, which _may_
mean we can tolerate float64 CIELAB, but this can wait until Go has generics.


## Expectation Management

The only API in this package that can be considered stable is the one implementing
[image/draw.Quantizer](https://golang.org/pkg/image/draw/#Quantizer). All other APIs
may change at any time without warning. This library is small enough to be
vendored, and unlikely to change in any serious fashion, so if stability is a
concern, this is what you should do.

Issues may be responded to whenever I happen to get around to them, but PRs are
unlikely to be accepted without discussion prior to starting.


## License

This is provided under an MIT license.

