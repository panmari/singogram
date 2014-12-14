package singogram


import (
	"image"
	"image/color"
)

type ImageData struct {
	Pix []float32
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
	v_max float32
}

func (p *ImageData) At(x, y int) float32 {
	if !(image.Point{x, y}.In(p.Rect)) {
		return 0
	}
	i := p.PixOffset(x, y)
	return p.Pix[i]
}

func (p *ImageData) AtNormalized(x, y int) float32 {
	i := p.At(x, y)
	return i / p.v_max
}

func (p *ImageData) Set(x, y int, v float32) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	if (v > p.v_max) {
		p.v_max = v
	}
	i := p.PixOffset(x, y)
	p.Pix[i] = v
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *ImageData) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*1
}

func NewImageData(r image.Rectangle) *ImageData {
	w, h := r.Dx(), r.Dy()
	pix := make([]float32, 1*w*h)
	return &ImageData{pix, 1 * w, r, 0}
}


func NewImageDataFromImage(src image.Image) *ImageData {
	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	data := NewImageData(bounds)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			oldColor := src.At(x, y)
			gray_color := color.GrayModel.Convert(oldColor).(color.Gray).Y
			data.Set(x, y, float32(gray_color)/255)
		}
	}
	return data
}
