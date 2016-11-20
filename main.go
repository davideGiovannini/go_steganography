/*
   Quando finisce di scrivere il file, per marcare la posizione il programma scrive un byte speciale di terminazione, poi scrive solo 0

   In decoding l'ultimo terminatore prima degli 0 e' scartato
*/
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
)

const (
	output_path = "output.png"
)

func main() {

	//FLAG READING
	var full_byte, onlyInfo, wants_decode bool
	var align rune
	var str_var, image_path, target_path string
	var terminator uint
	flag.StringVar(&str_var, "a", "", "alignement if any - can be nothing, left(l), right(r) or center (c)")

	flag.StringVar(&image_path, "image", "", "the image to use")
	flag.StringVar(&target_path, "target", "", "the target file to encode or the decode output")

	flag.BoolVar(&onlyInfo, "i", false, "To get only information about the max possible size. ")
	flag.BoolVar(&wants_decode, "d", false, "Decode from image")

	flag.UintVar(&terminator, "t", 0x18, "Terminator to use - must be an integer in range 1-255")

	flag.Parse()

	full_byte = false

	switch {
	case str_var == "":
		full_byte = true
	case str_var == "r":
		align = 'r'
	case str_var == "c":
		align = 'c'
	case str_var == "l":
		align = 'l'
	default:
		fmt.Println("Wrong align parameter")
		return
	}

	if image_path == "" || target_path == "" {
		fmt.Println("You must specify a valid path for both image and target")
		return
	}

	if terminator > 255 || terminator == 0 {
		fmt.Println("Wrong terminator, must be in range 1-255")
		return
	}

	//

	data := make(chan byte)

	//Read the Png
	image_file, err := os.Open(image_path)
	if err != nil {

		return
	}
	defer image_file.Close()

	image, err := png.Decode(image_file)
	if err != nil {
		fmt.Println("Error while decoding png!")
		return
	}

	if !wants_decode {
		//ENCODE or get INFO
		//
		//Size check -------------------------

		//Open the target file

		file, err := os.Open(target_path)
		if err != nil {
			fmt.Println("Error while opening", target_path)
			return
		}
		defer file.Close()

		// get the file size
		stat, err := file.Stat()
		if err != nil {
			return
		}

		rect := image.Bounds()

		var isOk bool
		var container_space int64
		if full_byte {
			container_space = int64(rect.Dx()*rect.Dy()) * 2

		} else {
			container_space = int64(rect.Dx() * rect.Dy())
		}

		isOk = stat.Size() < container_space

		fmt.Println(stat.Size()/1000, "KB\\", container_space/1000, "KB")

		if onlyInfo {
			return
		}

		if isOk {
			fmt.Println("Encoding")

			go scan(file, stat.Size(), data)

			encapsulate(image, uint8(terminator), full_byte, align, data)

		}
	} else {
		// DECODE

		target_file, err := os.Create(target_path)
		if err != nil {
			fmt.Println("Error while opening", target_path, "for writing.")
			return
		}

		fmt.Println("Decode")

		signal := make(chan bool)

		go writeToFile(target_file, data, signal)

		decode(image, full_byte, align, data)

		terminated := <-signal

		if !terminated {
			fmt.Println("There were errors writing the output file")
		}

	}
}
