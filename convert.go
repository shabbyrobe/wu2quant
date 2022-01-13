package wu2quant

import (
	"image"
	"image/color"
)

type rgbaAtImage interface {
	image.Image
	RGBAAt(x, y int) color.RGBA
}

var _ rgbaAtImage = &image.RGBA{}

func convertToRGBA(img image.Image) *image.RGBA {
	switch img := img.(type) {
	case *image.RGBA:
		return img

	case rgbaAtImage:
		return convertRGBAAtToRGBA(img)

	default:
		return convertImageToRGBA(img)
	}
}

// convertRGBAAtToRGBA is hopefully a less grim fallback slow-path than the
// CPU-warmer convertImageToRGBA.
func convertRGBAAtToRGBA(img rgbaAtImage) *image.RGBA {
	bounds := img.Bounds()
	size := bounds.Size()
	pix := make([]uint8, size.X*size.Y*4)

	var idx int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			pix[idx] = c.R
			pix[idx+1] = c.G
			pix[idx+2] = c.B
			pix[idx+3] = c.A
			idx += 4
		}
	}

	return &image.RGBA{Rect: bounds, Stride: size.X * 4, Pix: pix}
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
