package main

import (
	"flag"
	"image"
	_ "image/jpeg" // register the JPEG format with the image package
	"image/png"    // register the PNG format with the image package
	"os"
	"runtime/pprof"
	"singogram"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	infile, err := os.Open(os.Args[1])
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

	// Convert to imagedata.
	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	data := singogram.NewImageDataFromImage(src)

	n_pixel := w * h

	FCD_mm := float32(300)
	DCD_mm := float32(300)

	n_dexel := 200

	image_size_mm := float32(300)

	detector_size_mm := float32(600)

	dexel_size_mm := detector_size_mm / float32(n_dexel)
	pixel_size_mm := image_size_mm / float32(n_pixel)
	s := singogram.NewSinegogram(data, FCD_mm, DCD_mm, n_dexel, dexel_size_mm, pixel_size_mm)
	sinogram := s.Simulation()

	// Encode the grayscale image to the output file
	outfile, err := os.Create(os.Args[2])
	if err != nil {
		// replace this with real error handling
		panic(err)
	}
	defer outfile.Close()
	png.Encode(outfile, sinogram)
}
