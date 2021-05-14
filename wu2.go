package wu2quant

import (
	"fmt"
	"image"
	"image/color"
)

type tags [33 * 33 * 33]paletteIndex

type Quantizer struct {
	hist  histogram3D
	tag   tags
	dirty bool
}

func New() *Quantizer {
	return &Quantizer{}
}

// Quantizes the color palette of an image.Image and returns the
// palette as a color.Palette.
//
// It appends up to cap(p) - len(p) colors to p and returns the
// updated palette suitable for converting m to a paletted image.
//
// Quantize satisfies the image/draw.Quantizer interface.
//
// If m is not an *image.RGBA, it will be converted to one before quantization.
// Depending on the image type, this may trigger very slow code paths.
func (q *Quantizer) Quantize(p color.Palette, m image.Image) color.Palette {
	rgbImg := convertToRGBA(m)
	return q.QuantizeRGBAToPalette(p, rgbImg)
}

// QuantizeRGBA quantizes the color palette of an *image.RGBA image and
// returns the palette of a slice of color.RGBA types.
//
// It appends up to cap(p) - len(p) colors to p and returns the
// updated palette suitable for converting m to a paletted image.
func (q *Quantizer) QuantizeRGBA(p []color.RGBA, m *image.RGBA) []color.RGBA {
	var cols quantizedColors

	if err := q.quantize(&cols, m, cap(p)-len(p), nil); err != nil {
		panic(err)
	}

	idx := paletteIndex(len(p))
	p = append(p, make([]color.RGBA, cols.paletteSize)...)
	end := idx + cols.paletteSize
	for i := idx; i < end; i++ {
		p[i] = color.RGBA{R: cols.rLut[i], G: cols.gLut[i], B: cols.bLut[i], A: 0xff}
	}

	return p
}

// QuantizeRGBAToPalette quantizes the color palette of an *image.RGBA image
// and returns the palette as a color.Palette, which can be assigned directly
// into an image.Paletted. No implicit image type conversion is applied.
//
// It appends up to cap(p) - len(p) colors to p and returns the updated palette suitable
// for converting m to a paletted image.
func (q *Quantizer) QuantizeRGBAToPalette(p color.Palette, m *image.RGBA) color.Palette {
	var cols quantizedColors

	if err := q.quantize(&cols, m, cap(p)-len(p), nil); err != nil {
		panic(err)
	}

	idx := paletteIndex(len(p))
	p = append(p, make(color.Palette, cols.paletteSize)...)
	end := idx + cols.paletteSize
	for i := idx; i < end; i++ {
		p[i] = color.RGBA{R: cols.rLut[i], G: cols.gLut[i], B: cols.bLut[i], A: 0xff}
	}

	return p
}

// ToPaletted accepts input image m and returns a paletted version of the image reduced
// to paletteColors.
//
// If m is not an *image.RGBA, it will be converted to one before quantization.
// Depending on the image type, this may trigger very slow code paths.
func (q *Quantizer) ToPaletted(paletteColors int, m image.Image) (*image.Paletted, error) {
	rgbImg := convertToRGBA(m)
	return q.RGBAToPaletted(paletteColors, rgbImg)
}

func (q *Quantizer) RGBAToPaletted(paletteColors int, m *image.RGBA) (*image.Paletted, error) {
	var (
		cols   quantizedColors
		size   = m.Bounds().Size()
		pixels = size.X * size.Y
	)

	// Qadd contains the quantized image (array of table addresses)
	var qadd = make([]paletteIndex, pixels)
	if err := q.quantize(&cols, m, paletteColors, qadd); err != nil {
		return nil, err
	}

	var palette = make(color.Palette, cols.paletteSize)
	for i := paletteIndex(0); i < cols.paletteSize; i++ {
		palette[i] = color.RGBA{R: cols.rLut[i], G: cols.gLut[i], B: cols.bLut[i], A: 0xff}
	}

	var out = image.NewPaletted(image.Rect(0, 0, size.X, size.Y), palette)
	var qaddIdx int
	var vlen = len(m.Pix) / 4

	for i := 0; i < vlen; i++ {
		out.Pix[i] = uint8(q.tag[qadd[qaddIdx]])
		qaddIdx++
	}

	return out, nil
}

// IntoPaletted places a color-quantized copy of m into output image o. If m.Bounds() !=
// o.Bounds(), an error is returned.
//
// If m is not an *image.RGBA, it will be converted to one before quantization.
// Depending on the image type, this may trigger very slow code paths.
func (q *Quantizer) IntoPaletted(paletteColors int, m image.Image, o *image.Paletted) error {
	rgbImg := convertToRGBA(m)
	return q.RGBAIntoPaletted(paletteColors, rgbImg, o)
}

func (q *Quantizer) RGBAIntoPaletted(paletteColors int, m *image.RGBA, o *image.Paletted) error {
	var (
		cols   quantizedColors
		bounds = m.Bounds()
		size   = bounds.Size()
		pixels = size.X * size.Y
	)

	if bounds != o.Bounds() {
		return fmt.Errorf("wu2quant: input image m bounds %v did not match output image bounds %v", bounds, o.Bounds())
	}

	// Qadd contains the quantized image (array of table addresses)
	// FIXME: use a shared buffer in q, resize as needed:
	var qadd = make([]paletteIndex, pixels)
	if err := q.quantize(&cols, m, paletteColors, qadd); err != nil {
		return err
	}

	if cap(o.Palette) < int(cols.paletteSize) {
		o.Palette = make(color.Palette, cols.paletteSize)
	} else {
		o.Palette = o.Palette[:cols.paletteSize]
	}
	for i := paletteIndex(0); i < cols.paletteSize; i++ {
		o.Palette[i] = color.RGBA{R: cols.rLut[i], G: cols.gLut[i], B: cols.bLut[i], A: 0xff}
	}

	var qaddIdx int
	var vlen = len(m.Pix) / 4

	for i := 0; i < vlen; i++ {
		o.Pix[i] = uint8(q.tag[qadd[qaddIdx]])
		qaddIdx++
	}

	return nil
}

func (q *Quantizer) reset() {
	if q.dirty {
		q.hist = histogram3D{}
		for idx := range q.tag {
			q.tag[idx] = 0
		}
	}
	q.dirty = true
}

func (q *Quantizer) quantize(into *quantizedColors, img *image.RGBA, paletteColors int, qadd []paletteIndex) error {
	if paletteColors <= 0 || paletteColors > int(maxColors) {
		return fmt.Errorf("palette size must be 0 < sz < %d; found %d", maxColors, paletteColors)
	}

	q.reset()

	var (
		paletteSize = paletteIndex(paletteColors)
		cube        [maxColors]box
		next        paletteIndex
		vv          [maxColors]float32
		temp        float32
	)

	q.hist.build(img, qadd)
	q.hist.calculateMoments()

	cube[0].rmax, cube[0].gmax, cube[0].bmax = 32, 32, 32

	for i := paletteIndex(1); i < paletteSize; i++ {
		if cut(&cube[next], &cube[i], &q.hist.mr, &q.hist.mg, &q.hist.mb, &q.hist.wt) {
			// volume test ensures we won't try to cut one-cell box
			if cube[next].vol > 1 {
				vv[next] = q.hist.weightedVariance(&cube[next])
			} else {
				vv[next] = 0
			}
			if cube[i].vol > 1 {
				vv[i] = q.hist.weightedVariance(&cube[i])
			} else {
				vv[i] = 0
			}

		} else {
			vv[next] = 0.0 // don't try to split this box again
			i--            // didn't create box i
		}

		next = 0
		temp = vv[0]
		for k := paletteIndex(1); k <= i; k++ {
			if vv[k] > temp {
				temp = vv[k]
				next = k
			}
		}
		if temp <= 0.0 {
			paletteSize = i + 1
			break
		}
	}

	for k := paletteIndex(0); k < paletteSize; k++ {
		mark(&cube[k], k, &q.tag)

		weight := vol(&cube[k], &q.hist.wt)
		if weight != 0 {
			into.rLut[k] = uint8(vol(&cube[k], &q.hist.mr) / weight)
			into.gLut[k] = uint8(vol(&cube[k], &q.hist.mg) / weight)
			into.bLut[k] = uint8(vol(&cube[k], &q.hist.mb) / weight)
		} else {
			// fprintf(stderr, "bogus box %d\n", k)
			into.rLut[k], into.gLut[k], into.bLut[k] = 0, 0, 0
		}
	}

	into.paletteSize = paletteSize

	return nil
}

type box struct {
	rmin, rmax int // (rmin, rmax]
	gmin, gmax int // (gmin, gmax]
	bmin, bmax int // (bmin, bmax]
	vol        int
}

type paletteIndex uint16

const maxColors paletteIndex = 256

var squares [256]int64

func init() {
	for i := int64(0); i < 256; i++ {
		squares[i] = i * i
	}
}

type histogram3D struct {
	mr, mg, mb moment
	m2         momentFloat
	wt         moment
}

// build 3-D color histogram of counts, r/g/b, c^2
//
// At conclusion of the histogram step, we can interpret
//   wt[r][g][b] = sum over voxel of P(c)
//   mr[r][g][b] = sum over voxel of r*P(c)  ,  similarly for mg, mb
//   m2[r][g][b] = sum over voxel of c^2*P(c)
// Actually each of these should be divided by 'size' to give the usual
// interpretation of P() as ranging from 0 to 1, but we needn't do that here.
//
func (hist *histogram3D) build(img *image.RGBA, qadd []paletteIndex) {
	const trunc = 3 // shift each color channel right

	qidx := 0
	plen := len(img.Pix)
	pix := img.Pix
	for idx := 0; idx < plen; idx += 4 {
		var (
			r8, g8, b8    = int64(pix[idx]), int64(pix[idx+1]), int64(pix[idx+2])
			inr, ing, inb = (r8 >> trunc) + 1, (g8 >> trunc) + 1, (b8 >> trunc) + 1
			ind           = (inr << 10) + (inr << 6) + inr + (ing << 5) + ing + inb
		)

		if qadd != nil {
			qadd[qidx] = paletteIndex(ind)
			qidx++
		}

		hist.wt[inr][ing][inb]++
		hist.mr[inr][ing][inb] += r8
		hist.mg[inr][ing][inb] += g8
		hist.mb[inr][ing][inb] += b8
		hist.m2[inr][ing][inb] += (float32)(squares[r8] + squares[g8] + squares[b8])
	}
}

// Convert histogram into moments so that we can rapidly calculate
// the sums of the above quantities over any desired box.
//
func (hist *histogram3D) calculateMoments() {
	for r := 1; r <= 32; r++ {
		var (
			area, rArea, gArea, bArea [33]int64
			area2                     [33]float32
		)

		for g := 1; g <= 32; g++ {
			var (
				line, rLine, gLine, bLine int64
				line2                     float32
			)

			for b := 1; b <= 32; b++ {
				line += hist.wt[r][g][b]
				rLine += hist.mr[r][g][b]
				gLine += hist.mg[r][g][b]
				bLine += hist.mb[r][g][b]
				line2 += hist.m2[r][g][b]

				area[b] += line
				rArea[b] += rLine
				gArea[b] += gLine
				bArea[b] += bLine
				area2[b] += line2

				rx := r - 1
				hist.wt[r][g][b] = hist.wt[rx][g][b] + area[b]
				hist.mr[r][g][b] = hist.mr[rx][g][b] + rArea[b]
				hist.mg[r][g][b] = hist.mg[rx][g][b] + gArea[b]
				hist.mb[r][g][b] = hist.mb[rx][g][b] + bArea[b]
				hist.m2[r][g][b] = hist.m2[rx][g][b] + area2[b]
			}
		}
	}
}

// Compute the weighted variance of a box
// NB: as with the raw statistics, this is really the variance * size
func (hist *histogram3D) weightedVariance(cube *box) float32 {
	dr := float32(vol(cube, &hist.mr))
	dg := float32(vol(cube, &hist.mg))
	db := float32(vol(cube, &hist.mb))
	xx := float32(0 +
		hist.m2[cube.rmax][cube.gmax][cube.bmax] -
		hist.m2[cube.rmax][cube.gmax][cube.bmin] -
		hist.m2[cube.rmax][cube.gmin][cube.bmax] +
		hist.m2[cube.rmax][cube.gmin][cube.bmin] -
		hist.m2[cube.rmin][cube.gmax][cube.bmax] +
		hist.m2[cube.rmin][cube.gmax][cube.bmin] +
		hist.m2[cube.rmin][cube.gmin][cube.bmax] -
		hist.m2[cube.rmin][cube.gmin][cube.bmin])

	return xx - (((dr * dr) + (dg * dg) + (db * db)) / float32(vol(cube, &hist.wt)))
}

const (
	momentSize = 33

	dirR momentDir = 2
	dirG momentDir = 1
	dirB momentDir = 0
)

type (
	momentDir   int
	moment      [momentSize][momentSize][momentSize]int64
	momentFloat [momentSize][momentSize][momentSize]float32
)

// Compute sum over a box of any given statistic
func vol(cube *box, m *moment) int64 {
	return (0 +
		m[cube.rmax][cube.gmax][cube.bmax] -
		m[cube.rmax][cube.gmax][cube.bmin] -
		m[cube.rmax][cube.gmin][cube.bmax] +
		m[cube.rmax][cube.gmin][cube.bmin] -
		m[cube.rmin][cube.gmax][cube.bmax] +
		m[cube.rmin][cube.gmax][cube.bmin] +
		m[cube.rmin][cube.gmin][cube.bmax] -
		m[cube.rmin][cube.gmin][cube.bmin])
}

// Bottom and Top allow a slightly more efficient calculation of Vol() for a proposed
// subbox of a given box.  The sum of Top() and Bottom() is the Vol() of a subbox split in
// the given direction and with the specified new upper bound.
//
// Compute part of Vol(cube, mmt) that doesn't depend on r1, g1, or b1
// (depending on dir)
func bottom(cube *box, dir momentDir, m *moment) int64 {
	switch dir {
	case dirR:
		return (0 -
			m[cube.rmin][cube.gmax][cube.bmax] +
			m[cube.rmin][cube.gmax][cube.bmin] +
			m[cube.rmin][cube.gmin][cube.bmax] -
			m[cube.rmin][cube.gmin][cube.bmin])

	case dirG:
		return (0 -
			m[cube.rmax][cube.gmin][cube.bmax] +
			m[cube.rmax][cube.gmin][cube.bmin] +
			m[cube.rmin][cube.gmin][cube.bmax] -
			m[cube.rmin][cube.gmin][cube.bmin])

	case dirB:
		return (0 -
			m[cube.rmax][cube.gmax][cube.bmin] +
			m[cube.rmax][cube.gmin][cube.bmin] +
			m[cube.rmin][cube.gmax][cube.bmin] -
			m[cube.rmin][cube.gmin][cube.bmin])

	default:
		panic("unkown dir")
	}
}

func top(cube *box, dir momentDir, pos int, m *moment) int64 {
	// Compute remainder of Vol(cube, mmt), substituting pos for
	// r1, g1, or b1 (depending on dir)
	switch dir {
	case dirR:
		return (0 +
			m[pos][cube.gmax][cube.bmax] -
			m[pos][cube.gmax][cube.bmin] -
			m[pos][cube.gmin][cube.bmax] +
			m[pos][cube.gmin][cube.bmin])

	case dirG:
		return (0 +
			m[cube.rmax][pos][cube.bmax] -
			m[cube.rmax][pos][cube.bmin] -
			m[cube.rmin][pos][cube.bmax] +
			m[cube.rmin][pos][cube.bmin])

	case dirB:
		return (0 +
			m[cube.rmax][cube.gmax][pos] -
			m[cube.rmax][cube.gmin][pos] -
			m[cube.rmin][cube.gmax][pos] +
			m[cube.rmin][cube.gmin][pos])

	default:
		panic("unkown dir")
	}
}

// We want to minimize the sum of the variances of two subboxes.
// The sum(c^2) terms can be ignored since their sum over both subboxes
// is the same (the sum for the whole box) no matter where we split.
// The remaining terms have a minus sign in the variance formula,
// so we drop the minus sign and MAXIMIZE the sum of the two terms.
func maximize(
	cube *box, dir momentDir, first, last int,
	rWhole, gWhole, bWhole, wWhole int64,
	mr, mg, mb, wt *moment,
) (max float32, cut int) {

	var (
		rBase = bottom(cube, dir, mr)
		gBase = bottom(cube, dir, mg)
		bBase = bottom(cube, dir, mb)
		wBase = bottom(cube, dir, wt)

		temp float32
	)

	cut = -1

	for i := first; i < last; i++ {
		rHalf := int64(rBase + top(cube, dir, i, mr))
		gHalf := int64(gBase + top(cube, dir, i, mg))
		bHalf := int64(bBase + top(cube, dir, i, mb))
		wHalf := int64(wBase + top(cube, dir, i, wt))

		// now half_x is sum over lower half of box, if split at i
		if wHalf == 0 { // subbox could be empty of pixels!
			continue // never split into an empty box
		} else {
			temp = (0 +
				(float32(rHalf) * float32(rHalf)) +
				(float32(gHalf) * float32(gHalf)) +
				(float32(bHalf) * float32(bHalf))) / float32(wHalf)
		}

		rHalf = rWhole - rHalf
		gHalf = gWhole - gHalf
		bHalf = bWhole - bHalf
		wHalf = wWhole - wHalf
		if wHalf == 0 { // subbox could be empty of pixels!
			continue // never split into an empty box
		} else {
			temp += (0 +
				(float32(rHalf) * float32(rHalf)) +
				(float32(gHalf) * float32(gHalf)) +
				(float32(bHalf) * float32(bHalf))) / float32(wHalf)
		}

		if temp > max {
			max = temp
			cut = i
		}
	}

	return max, cut
}

func cut(
	set1, set2 *box,
	mr, mg, mb, wt *moment,
) bool {

	rWhole := int64(vol(set1, mr))
	gWhole := int64(vol(set1, mg))
	bWhole := int64(vol(set1, mb))
	wWhole := int64(vol(set1, wt))

	maxr, cutr := maximize(set1, dirR, set1.rmin+1, set1.rmax, rWhole, gWhole, bWhole, wWhole, mr, mg, mb, wt)
	maxg, cutg := maximize(set1, dirG, set1.gmin+1, set1.gmax, rWhole, gWhole, bWhole, wWhole, mr, mg, mb, wt)
	maxb, cutb := maximize(set1, dirB, set1.bmin+1, set1.bmax, rWhole, gWhole, bWhole, wWhole, mr, mg, mb, wt)

	var dir momentDir

	if (maxr >= maxg) && (maxr >= maxb) {
		dir = dirR
		if cutr < 0 {
			return false // can't split the box
		}
	} else {
		if (maxg >= maxr) && (maxg >= maxb) {
			dir = dirG
		} else {
			dir = dirB
		}
	}

	set2.rmax = set1.rmax
	set2.gmax = set1.gmax
	set2.bmax = set1.bmax

	switch dir {
	case dirR:
		set2.rmin = cutr
		set1.rmax = cutr
		set2.gmin = set1.gmin
		set2.bmin = set1.bmin

	case dirG:
		set2.gmin = cutg
		set1.gmax = cutg
		set2.rmin = set1.rmin
		set2.bmin = set1.bmin

	case dirB:
		set2.bmin = cutb
		set1.bmax = cutb
		set2.rmin = set1.rmin
		set2.gmin = set1.gmin

	}
	set1.vol = (set1.rmax - set1.rmin) * (set1.gmax - set1.gmin) * (set1.bmax - set1.bmin)
	set2.vol = (set2.rmax - set2.rmin) * (set2.gmax - set2.gmin) * (set2.bmax - set2.bmin)
	return true
}

func mark(cube *box, label paletteIndex, tag *tags) {
	for r := cube.rmin + 1; r <= cube.rmax; r++ {
		for g := cube.gmin + 1; g <= cube.gmax; g++ {
			for b := cube.bmin + 1; b <= cube.bmax; b++ {
				tag[(r<<10)+(r<<6)+r+(g<<5)+g+b] = label
			}
		}
	}
}

type quantizedColors struct {
	// lut_r, lut_g, lut_b as color look-up table contents
	rLut, gLut, bLut [maxColors]uint8
	paletteSize      paletteIndex
}
