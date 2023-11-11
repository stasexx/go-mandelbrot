package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	width   = 800
	height  = 800
	maxIter = 200
)

type ComplexNumber struct {
	Real, Imag float64
}

func mandelbrot(c ComplexNumber) color.Color {
	z := ComplexNumber{}
	for i := 0; i < maxIter; i++ {
		z = ComplexNumber{z.Real*z.Real - z.Imag*z.Imag + c.Real, 2*z.Real*z.Imag + c.Imag}
		if z.Real*z.Real+z.Imag*z.Imag > 4 {
			return color.Gray{uint8(i % 256)}
		}
	}
	return color.Black
}

func generateMandelbrotSequential(img *image.RGBA, existingImg image.Image) {
	bounds := img.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c := ComplexNumber{
				Real: float64(x-width/2) / (width / 4),
				Imag: float64(y-height/2) / (height / 4),
			}
			img.Set(x, y, mandelbrot(c))
			r, g, b, _ := existingImg.At(x, y).RGBA()
			img.Set(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255})
		}
	}
}

func generateMandelbrotParallel(img *image.RGBA, existingImg image.Image) {
	var wg sync.WaitGroup
	wg.Add(height)

	for x := 0; x < width; x++ {
		go func(x int) {
			defer wg.Done()
			for y := 0; y < height; y++ {
				c := ComplexNumber{
					Real: float64(x-width/2) / (width / 4),
					Imag: float64(y-height/2) / (height / 4),
				}
				img.Set(x, y, mandelbrot(c))
				r, g, b, _ := existingImg.At(x, y).RGBA()
				img.Set(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255})
			}
		}(x)
	}

	wg.Wait()
}

func processImage(file string, level string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Processing %s image...\n", level)

	inputFile, err := os.Open(file)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer inputFile.Close()

	existingImg, _, err := image.Decode(inputFile)
	if err != nil {
		fmt.Println("Error decoding image:", err)
		return
	}

	sequentialImg := image.NewRGBA(existingImg.Bounds())
	parallelImg := image.NewRGBA(existingImg.Bounds())

	copyImage(sequentialImg, existingImg)
	copyImage(parallelImg, existingImg)

	startTimeSequential := time.Now()
	generateMandelbrotSequential(sequentialImg, existingImg)
	elapsedTimeSequential := time.Since(startTimeSequential)

	sequentialOutputPath := filepath.Join("photos", "result", level, "mandelbrot_sequential.png")
	saveImage(sequentialImg, sequentialOutputPath)
	fmt.Printf("%s Sequential: Elapsed time: %s\n", level, elapsedTimeSequential)

	startTimeParallel := time.Now()
	generateMandelbrotParallel(parallelImg, existingImg)
	elapsedTimeParallel := time.Since(startTimeParallel)

	parallelOutputPath := filepath.Join("photos", "result", level, "mandelbrot_parallel.png")
	saveImage(parallelImg, parallelOutputPath)
	fmt.Printf("%s Parallel: Elapsed time: %s\n", level, elapsedTimeParallel)
}

func copyImage(destImg *image.RGBA, srcImg image.Image) {
	b := destImg.Bounds()
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			destImg.Set(x, y, srcImg.At(x, y))
		}
	}
}

func saveImage(img *image.RGBA, fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		fmt.Println("Error encoding PNG:", err)
	}
}

func main() {
	levelFiles := map[string]string{
		"easy":   "photos/easy.png",
		"normal": "photos/normal.png",
		"hard":   "photos/hard.png",
	}

	var wg sync.WaitGroup
	for level, file := range levelFiles {
		wg.Add(1)
		go processImage(file, level, &wg)
	}

	wg.Wait()
}
