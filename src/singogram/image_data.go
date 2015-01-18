package singogram

import (
	"github.com/ungerik/go3d/vec2"
	"image"
	"image/color"
)

// Gets about 1/5 faster if we are certain that At(x, y) never tries
// to access an invalid position.
var (
	kSafeAccess = true
)

// An image like structure consisting of floats that keeps
// track of its maximum value (for normalization purposes)
// over all Set calls.
type ImageData struct {
	Pix []float32
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect   image.Rectangle
	bounds [2]vec2.T
	v_max  float32
}

func (p *ImageData) At(x, y int) float32 {
	if kSafeAccess && !(image.Point{x, y}.In(p.Rect)) {
		return 0
	}
	i := p.PixOffset(x, y)
	return p.Pix[i]
}

// Adaption for matlab vs go arrays -> start with (0 0) instead of (1 1).
// Additionally swap x and y, bc matlab matrix access does first row index, then column index.
// Inlining this gives 10 % better performance?!?
func (p *ImageData) MatlabAt(x, y int) float32 {
	return p.At(y-1, x-1)
}

func (p *ImageData) AddAt(x, y int, v float32) {
	if kSafeAccess && !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	p.Pix[i] += v
}

// See above.
func (p *ImageData) MatlabAddAt(x, y int, v float32) {
	p.AddAt(y-1, x-1, v)
}

func (p *ImageData) AtNormalized(x, y int) float32 {
	i := p.At(x, y)
	return i / p.v_max
}

func (p *ImageData) Set(x, y int, v float32) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	if v > p.v_max {
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

func (p *ImageData) Intersections(orig *vec2.T, dir *vec2.T) (float32, float32, bool) {
	invdir := vec2.T{1 / dir[0], 1 / dir[1]}
	// sign_x is 0 if -----> (pointing to right).
	// sign_x is 1 if <----- (pointing to left).
	sign_x := 0
	if invdir[0] < 0 {
		sign_x = 1
	}
	// sign_y is 0 if pointing up.
	// sign_y is 1 if pointing down.
	sign_y := 0
	if invdir[1] < 0 {
		sign_y = 1
	}
	// p.bounds contains the borders of the image (m = [min_x  max_x;
	//												    min_y  max_y;])
	tmin := (p.bounds[sign_x][0] - orig[0]) * invdir[0]
	tmax := (p.bounds[1-sign_x][0] - orig[0]) * invdir[0]
	tymin := (p.bounds[sign_y][1] - orig[1]) * invdir[1]
	tymax := (p.bounds[1-sign_y][1] - orig[1]) * invdir[1]
	if tmin > tymax || tymin > tmax {
		return 0, 0, false
	}
	if tymin > tmin {
		tmin = tymin
	}
	if tymax < tmax {
		tmax = tymax
	}
	return tmin, tmax, true
}

func NewImageData(r image.Rectangle) *ImageData {
	w, h := r.Dx(), r.Dy()
	pix := make([]float32, 1*w*h)
	// Bounds are chosen as rectangle from [1, 1] to [w, h] to
	// to simulate matlab arrays.
	bounds := [2]vec2.T{vec2.T{1, 1}, vec2.T{float32(w), float32(h)}}
	return &ImageData{pix, 1 * w, r, bounds, 0}
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
