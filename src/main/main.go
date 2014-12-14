package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg" // register the JPEG format with the image package
	"image/png"    // register the PNG format with the image package
	"os"
	"runtime"
	"runtime/pprof"
	"singogram"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var maxProcs = flag.Int("procs", runtime.NumCPU(), "set the number of processors to use")

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
	runtime.GOMAXPROCS(*maxProcs)

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
	data := singogram.NewImageDataFromImage(src)

	FCD_mm := float32(300)
	DCD_mm := float32(300)

	n_dexel := 200

	image_size_mm := float32(300)

	detector_size_mm := float32(600)

	dexel_size_mm := detector_size_mm / float32(n_dexel)
	bounds := src.Bounds()
	pixel_size_mm := image_size_mm / float32(bounds.Dx())
	s := singogram.NewSinegogram(data, FCD_mm, DCD_mm, n_dexel, dexel_size_mm, pixel_size_mm)

	start := time.Now()
	sinogram := s.Simulation()
	duration := time.Since(start)
	fmt.Println(duration.String())

	// Encode the grayscale image to the output file
	outfile, err := os.Create(os.Args[2])
	if err != nil {
		// replace this with real error handling
		panic(err)
	}
	defer outfile.Close()
	png.Encode(outfile, sinogram)
}
