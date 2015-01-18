package singogram

import (
	math "github.com/barnex/fmath"
	"github.com/cheggaaa/pb"
	"github.com/ungerik/go3d/vec2"
	"image"
	"image/color"
//	"log"
	"runtime"
	"sync"
)

var (
	delta_s    float32 = 0.05
	angle_step float32 = 1
)

type Axis int

const (
	X Axis = iota
	Y
)

type Sinegogram struct {
	data          *ImageData
	FCD_mm        float32
	DCD_mm        float32
	n_dexel       int
	dexel_size_mm float32
	pixel_size_mm float32
}

func NewSinegogram(data *ImageData, FCD_mm float32, DCD_mm float32, n_dexel int, dexel_size_mm float32, pixel_size_mm float32) *Sinegogram {
	s := Sinegogram{data, FCD_mm, DCD_mm, n_dexel, dexel_size_mm, pixel_size_mm}
	return &s
}

// Returns a new vector that is equal to the given rotated CLOCKWISE by angle_deg degrees.
func rotate(v *vec2.T, angle_deg float32) vec2.T {
	angle_rad := angle_deg * math.Pi / 180
	return v.Rotated(-angle_rad)
}

func (s *Sinegogram) xy_to_rc(xy *vec2.T) *vec2.T {
	// TODO: +1 necessary?
	trans := vec2.T{float32(s.data.Rect.Dx() + 1), float32(s.data.Rect.Dy() + 1)}
	trans.Scale(1.0 / 2.0)
	cr := xy.Scaled(1.0 / s.pixel_size_mm)
	// Change sign of y. wat?
	cr[1] = -cr[1]
	cr.Add(&trans)
	// Some switcheroo
	cr[0], cr[1] = cr[1], cr[0]
	return &cr
}

func (s *Sinegogram) detector_positions(angle_deg float32) []vec2.T {
	dexels := make([]vec2.T, s.n_dexel)

	trans := float32(s.n_dexel-1) * s.dexel_size_mm / 2
	for i := 0; i < s.n_dexel; i++ {
		x := s.dexel_size_mm*float32(i) - trans
		y := -s.DCD_mm
		d := vec2.T{x, y}
		dexels[i] = rotate(&d, angle_deg)
	}
	return dexels
}

// Return p per pixel, using a ray marching algorithm.
func (s *Sinegogram) line_integral_rc(source *vec2.T, dexel *vec2.T) float32 {
	dir := vec2.Sub(source, dexel)
	dir_length := dir.Length()
	dir.Scale(1 / dir_length)

	sum_p := float32(0.0)
	min, max, does_intersect := s.data.Intersections(dexel, &dir)
	for scale := min; does_intersect && scale <= max; scale += delta_s {
		p := dir.Scaled(scale)
		p.Add(dexel)
		x := Round(p[0])
		y := Round(p[1])
		mu_p := s.data.MatlabAt(x, y)
		sum_p += mu_p
	}
	return sum_p * delta_s
}

// Return p per pixel, using a pixel intersection algorithm.
func (s *Sinegogram) line_integral_rc_intersection(source *vec2.T, dexel *vec2.T) float32 {
	dir := vec2.Sub(source, dexel)
	dir_length := dir.Length()
	dir.Scale(1 / dir_length)

	sum_p := float32(0.0)
	min, max, does_intersect := s.data.Intersections(dexel, &dir)
	// Immediately return if there is no intersection.
	if !does_intersect {
		return 0
	}
	p_min := dir.Scaled(min)
	p_min.Add(dexel)

	var dir_gridified [2]int
	var current_grid_pos [2]int
	var next_grid_pos [2]int
	var next_t [2]float32
	for i := range dir_gridified {
		if dir[i] != 0 {
			// sign(dir[i])
			dir_gridified[i] = int(dir[i] / math.Abs(dir[i]))
		} else {
			// zero, if was zero before (ray parallel to axis).
			dir_gridified[i] = 0
		}
		// On grid (round numbers!)
		current_grid_pos[i] = int(p_min[i])
		next_grid_pos[i] = int(p_min[i]) + dir_gridified[i]
		if dir[i] != 0 {
			next_t[i] = (float32(next_grid_pos[i]) - dexel[i]) / dir[i]
		} else {
			next_t[i] = math.MaxFloat32 // Ray parallel => it never crosses 
		}
	}
	
	lastT := min
	for lastT < max {
		var dist_in_pixel float32
		var axis Axis
		if next_t[X] < next_t[Y] {
			axis = X
		} else {
			axis = Y
		}
		mu_p := s.data.MatlabAt(current_grid_pos[X], current_grid_pos[Y])

		dist_in_pixel = next_t[axis] - lastT
		lastT = next_t[axis]
		current_grid_pos[axis] += dir_gridified[axis]
		next_grid_pos[axis] += dir_gridified[axis]
		next_t[axis] = (float32(next_grid_pos[axis]) - dexel[axis]) / dir[axis]
		sum_p += mu_p * dist_in_pixel
	}
	return sum_p
}

// Returns p per centimeter
func (s *Sinegogram) line_integral_xy(source_xy *vec2.T, dexel_xy *vec2.T) float32 {
	source_rc := s.xy_to_rc(source_xy)
	dexel_rc := s.xy_to_rc(dexel_xy)
	p := s.line_integral_rc_intersection(source_rc, dexel_rc)
	return p * s.pixel_size_mm / 10
}

func (s *Sinegogram) tube_position(angle_deg float32) vec2.T {
	tube := vec2.T{0, s.FCD_mm}
	return rotate(&tube, angle_deg)
}

func (s *Sinegogram) view(angle_deg float32) []float32 {
	tube := s.tube_position(angle_deg)

	dexels := s.detector_positions(angle_deg)

	p := make([]float32, len(dexels))
	for i := range dexels {
		p[i] = s.line_integral_xy(&tube, &dexels[i])
	}
	return p
}

// Does the simulation for a given range of angles (start_ang:step:stop_ang)
func (s *Sinegogram) SimulationForRange(start_ang, stop_ang, step float32,
	pb *pb.ProgressBar, wg *sync.WaitGroup, sinogram *ImageData) {
	defer wg.Done()
	// assumes first was at 0
	y := int(start_ang / step)
	for angle_deg := start_ang; angle_deg < stop_ang; angle_deg += step {
		values := s.view(angle_deg)
		// Write stuff to image.
		for i := range values {
			sinogram.Set(i, y, values[i])
		}
		pb.Increment()
		y++
	}
}

func (s *Sinegogram) Simulation() *image.Gray {
	nr_steps := 360 / angle_step
	line_count := int(nr_steps)
	bounds := image.Rect(0, 0, s.n_dexel, line_count)
	sinogram := NewImageData(bounds)

	var wg sync.WaitGroup

	pb := pb.StartNew(line_count)

	angle_per_task := nr_steps / float32(runtime.GOMAXPROCS(-1))
	for start_angle := float32(0); start_angle < 360; start_angle += angle_per_task {
		stop_angle := start_angle + angle_per_task
		go s.SimulationForRange(start_angle, stop_angle, angle_step, pb, &wg, sinogram)
		wg.Add(1)
	}
	wg.Wait()
	pb.Finish()

	sinogram_gray := image.NewGray(bounds)
	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			sinogram_gray.Set(x, y, color.Gray{uint8(sinogram.AtNormalized(x, y) * 255)})
		}
	}
	return sinogram_gray
}

// Writes the length the ray traveled through the pixel to ImageData.
func (s *Sinegogram) backproject_cr(source *vec2.T, dexel *vec2.T, c_i float32) *ImageData {
	dir := vec2.Sub(source, dexel)
	dir_length := dir.Length()
	dir.Scale(1 / dir_length)
	min, max, does_intersect := s.data.Intersections(dexel, &dir)

	correction_image := NewImageData(s.data.Rect)
	corrected_delta_s := delta_s * c_i
	for scale := min; does_intersect && scale <= max; scale += delta_s {
		p := dir.Scaled(scale)
		p.Add(dexel)
		x := Round(p[0])
		y := Round(p[1])
		correction_image.MatlabAddAt(x, y, corrected_delta_s)
	}
	return correction_image
}

func Round(f float32) int {
	return int(f + 0.5)
}
