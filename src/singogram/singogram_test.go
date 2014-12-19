package singogram

import (
	math "github.com/barnex/fmath"
	"github.com/ungerik/go3d/vec2"
	"image"
	_ "image/jpeg" // register the JPEG format with the image package
	"os"
	"testing"
)

var kEpsilon = float32(1e-5)

func TestRotate(t *testing.T) {
	v := vec2.T{5, 0}
	r := rotate(&v, 90)

	expected := vec2.T{0, -5}

	if diff := vec2.Sub(&r, &expected); diff.Length() > kEpsilon {
		t.Error(r, expected)
	}
}

func TestXyToCr(t *testing.T) {
	data := NewImageData(image.Rect(0, 0, 5, 5))
	pixel_size_mm := float32(3.0)
	s := NewSinegogram(data, 0, 0, 0, 0, pixel_size_mm)

	xy := vec2.T{6, 0}
	cr := s.xy_to_cr(&xy)
	if cr[0] != 5 {
		t.Error(cr)
	}

	xy = vec2.T{0, 0}
	cr = s.xy_to_cr(&xy)
	if cr[0] != 3 {
		t.Error(cr)
	}

	xy = vec2.T{-7.5, 0}
	cr = s.xy_to_cr(&xy)
	if cr[0] != 0.5 {
		t.Error(cr)
	}

	pixel_size_mm = float32(6.0)
	s = NewSinegogram(data, 0, 0, 0, 0, pixel_size_mm)
	
	xy = vec2.T{0, 0}
	cr = s.xy_to_cr(&xy)
	if cr[1] != 3 {
		t.Error(cr)
	}

	xy = vec2.T{0, 15}
	cr = s.xy_to_cr(&xy)
	if cr[1] != 0.5 {
		t.Error(cr)
	}
}

func generateTestImageData() (*ImageData, image.Rectangle) {
	infile, err := os.Open("../../CTLab-Introduction2.jpg")
	if err != nil {
		// replace this with real error handling
		panic(err)
	}
	defer infile.Close()

	// Decode will figure out what type of image is in the file on its own.
	// We just have to be sure all the image packages we want are imported.
	src, _, err := image.Decode(infile)
	if err != nil {
		// replace this with real error handling
		panic(err)
	}
	return NewImageDataFromImage(src), src.Bounds()
}

func TestImageData(t *testing.T) {
	data, _ := generateTestImageData()
	center_value := data.At(100, 100)
	if center_value != 0.4117647 {
		t.Error()
	}
}
/*
// TODO: fix tests
func TestDetectorPositions(t *testing.T) {
	// These seemed alright.
	vertical := s.detector_positions( 100, 90, 10, 10)
	t.Log(vertical)
	
	horizontal := s.detector_positions( 100, 0, 10, 10)
	t.Log(horizontal)
}
*/

func TestLineIntegralCr(t *testing.T) {
	bounds := image.Rect(0, 0, 4, 4)
	data := NewImageData(bounds)
	data.Pix = []float32{0, 0, 0, 0, 0, 5, 2, 0, 0, 1, 3, 0, 0, 0, 0, 0}

	s := NewSinegogram(data, 0, 0, 0, 0, 0)
	
	kAllowedError := float32(0.16)
	
	source := vec2.T{4.5, 1.7}
	dexel := vec2.T{0.5, 1.7} 
	p := s.line_integral_cr(&source, &dexel)
	if math.Abs(p - 7.15) > kAllowedError {
		t.Error(p)
	}
	
	source = vec2.T{0.5, 3.0}
	dexel = vec2.T{4.5, 3.0} 
	p = s.line_integral_cr(&source, &dexel)
	if math.Abs(p - 4.15) > kAllowedError {
		t.Error(p)
	}
	
	source = vec2.T{0.5, 4.5}
	dexel = vec2.T{4.5, 0.5} 
	p = s.line_integral_cr(&source, &dexel)
	if math.Abs(p - 4.20) > kAllowedError {
		t.Error(p)
	}
	
	source = vec2.T{3.5, 4.5}
	dexel = vec2.T{1.5, 0.5} 
	p = s.line_integral_cr(&source, &dexel)
	kAllowedError = float32(0.3)
	if math.Abs(p - 9.05) > kAllowedError {
		t.Error(p)
	}
}

func TestView(t *testing.T) {
	if testing.Short() {
        t.Skip("skipping test in short mode.")
    }
	// Convert to imagedata.
	data, bounds := generateTestImageData()

	FCD_mm := float32(300)
	DCD_mm := float32(300)

	n_dexel := 200

	image_size_mm := float32(300)

	detector_size_mm := float32(600)

	dexel_size_mm := detector_size_mm / float32(n_dexel)
	pixel_size_mm := image_size_mm / float32(bounds.Dx())

	angle_deg := float32(45)
	
	s := NewSinegogram(data, FCD_mm, DCD_mm, n_dexel, dexel_size_mm, pixel_size_mm)
	
	tube := s.tube_position(angle_deg)
	expected_tube := vec2.T{212.13, 212.13}
	if (expected_tube.Sub(&tube).Length() > 0.01) {
		t.Error(tube)
	}
	
	dexels := s.detector_positions(angle_deg)
	expected_first_dexel := vec2.T{-423.2, -1.06}
	if (expected_first_dexel.Sub(&dexels[0]).Length() > 0.01) {
		t.Error(dexels[0])
	}
	
	source_cr := s.xy_to_cr(&tube)
	first_dexel_cr := s.xy_to_cr(&dexels[0])
	t.Logf("Source cr: %v", source_cr)
	t.Logf("First dexel cr: %v", first_dexel_cr)

	result := s.view(angle_deg)

	t.Log(result)
}