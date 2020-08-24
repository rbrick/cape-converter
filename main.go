package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	inputDirectory    = flag.String("in-dir", "capes", "sets the directory used for input")
	oldToNewDirectory = flag.String("old-dir", "convert-old", "sets the directory used for input")
	outputDirectory   = flag.String("out-dir", "capes-output", "sets the directory used for output")
	fixedDirectory    = flag.String("fixed-dir", "fixed-capes", "sets the directory used to output fixed capes")
)

func init() {
	flag.Parse()

	if err := os.MkdirAll(*outputDirectory, os.ModePerm); err != nil {
		log.Fatalln(err)
	}

	if err := os.MkdirAll(*fixedDirectory, os.ModePerm); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	var wg sync.WaitGroup
	count := 0
	progressChan := make(chan int)
	failures, completed := 0, 0

	err := filepath.Walk(*inputDirectory, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), "png") {
			wg.Add(1)
			count++
			go scaleImage(path, info, false, progressChan)
		}
		return nil
	})

	err = filepath.Walk(*oldToNewDirectory, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(strings.ToLower(info.Name()), "png") {
			wg.Add(1)
			count++
			go scaleImage(path, info, true, progressChan)
		}
		return nil
	})

	if err != nil {
		fmt.Println("failed to start?")
		return
	}

	log.Println("waiting for", count, "image(s) to be converted.")

	start := time.Now()
	go func() {
		for v := range progressChan {
			wg.Done()
			if v <= -1 {
				failures++
			}
			completed++
		}
	}()

	go func() {
		for {
			elapsed := time.Now().Sub(start)
			fmt.Printf("\rTime Elapsed: %s, Completed: %d, Failures: %d", elapsed.String(), completed, failures)
		}
	}()

	wg.Wait()
}

func scaleImage(path string, info os.FileInfo, old bool, progressChan chan int) {
	f, err := os.Open(path)

	if err != nil {
		progressChan <- -1 // failed
		return
	}

	img, err := png.Decode(f)

	if err != nil {
		progressChan <- -1
		return
	}

	bounds := img.Bounds()
	scaleX := bounds.Dx() / 64
	scaleY := bounds.Dy() / 32

	if !old {
		if bounds.Dx()%64 != 0 || bounds.Dx()%32 != 0 {
			progressChan <- -2

			fixedImg := image.NewNRGBA(image.Rect(0, 0, bounds.Dx()-(bounds.Dx()%64), bounds.Dy()-(bounds.Dy()%32)))
			nImage := img.(*image.NRGBA)

			if fixedImg.Bounds().Dx() > nImage.Bounds().Dx() || fixedImg.Bounds().Dy() > nImage.Bounds().Dy() {
				draw.Draw(fixedImg, nImage.Bounds(), nImage, image.Pt(0, 0), draw.Src)
			} else {
				draw.Draw(fixedImg, fixedImg.Bounds(), nImage, image.Pt(0, 0), draw.Src)
			}
			newFile, err := os.Create(filepath.Join(*fixedDirectory, info.Name()))

			if err != nil {
				progressChan <- -1
				return
			}

			err = png.Encode(newFile, fixedImg)

			return
		}
	}

	fixedWidth := 22 * (bounds.Dx() / 64)
	fixedHeight := 17 * (bounds.Dy() / 32)
	scaledY := 32

	if old {
		scaledY = 17
		scaleX = bounds.Dx() / 22
		scaleY = bounds.Dy() / 17

		fixedWidth = 64 * (bounds.Dx() / 22)
		fixedHeight = 32 * (bounds.Dy() / 17)
	}

	newImg := image.NewRGBA(image.Rect(0, 0, fixedWidth, fixedHeight))

	needsScaling := fixedWidth/22 > 1 && fixedHeight/17 > 1

	if needsScaling {
		for i := 0; i < max(scaleY/scaledY, 1); i++ {
			for y := 0; y < fixedHeight/(max(scaleY/scaledY, 1)); y++ {
				for x := 0; x < fixedWidth; x++ {
					if scaleX != scaleY {
						newImg.Set(x, y+(i*fixedHeight/(max(scaleY/scaledY, 1))), img.At(x, y+(i*(bounds.Dy()/max(scaleY/scaledY, 1)))))
					} else {
						newImg.Set(x, y*(i+1), img.At(x, y*(i+1)))
					}

				}
			}
		}
	} else {
		for i := 0; i < fixedHeight/17; i++ {
			draw.Draw(newImg, image.Rect(0, i*17, 22, (i+1)*17), img, image.Pt(0, i*32), draw.Src)
		}
	}

	name := info.Name()
	od := *outputDirectory
	if old {
		trimmed := strings.TrimSuffix(name, ".png")
		name = trimmed + "_of.png"

		od = *fixedDirectory
	}

	newFile, err := os.Create(filepath.Join(od, name))

	if err != nil {
		progressChan <- -1
		fmt.Println(err)
		return
	}

	err = png.Encode(newFile, newImg)

	if err != nil {
		progressChan <- -1
		return
	}

	progressChan <- 1
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func nextPow2(i int) int {
	x := float64(i)

	if i == 0 {
		return 0
	}

	return int(math.Pow(2.0, math.Ceil(math.Log(x)/math.Log(2))))
}

func prevPow2(x int) int {
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	return x - (x >> 1)
}
