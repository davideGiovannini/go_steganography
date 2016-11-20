/*
   Quando finisce di scrivere il file, per marcare la posizione il programma scrive un byte speciale di terminazione, poi scrive solo 0

   In decoding l'ultimo terminatore prima degli 0 e' scartato
*/
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
)

const (
	output_path = "output.png"
)

func main() {
	args := parse_command_args()

	//
	data := make(chan byte)

	//Read the Png
	image_file, err := os.Open(args.image_path)
	if err != nil {
		panic("Error while opening image!")
	}
	defer image_file.Close()

	image, err := png.Decode(image_file)
	if err != nil {
		panic("Error while decoding png!")
	}

	if args.wants_decode {
		call_decode(args, image, data)
	} else {
		call_encode(args, image, data)
	}
}

func call_encode(args Arguments, image image.Image, data chan byte) {
	//ENCODE or get INFO
	//
	//Size check -------------------------

	//Open the target file
	file, err := os.Open(args.target_path)
	if err != nil {
		panic(fmt.Sprintf("Error while opening %s", args.target_path))
	}
	defer file.Close()

	// get the file size
	stat, err := file.Stat()
	if err != nil {
		panic(fmt.Sprintf("Error while getting stats of file: %s", args.target_path))
	}

	rect := image.Bounds()

	var isOk bool
	var container_space int64
	if args.use_full_byte {
		container_space = int64(rect.Dx()*rect.Dy()) * 2

	} else {
		container_space = int64(rect.Dx() * rect.Dy())
	}

	isOk = stat.Size() < container_space

	fmt.Println(stat.Size()/1000, "KB\\", container_space/1000, "KB")

	if !isOk {
		panic(fmt.Sprintf("Target file: %s does not fit into image file: %s", args.target_path, args.image_path))
	}

	if args.only_info {
		return
	}

	if isOk {
		fmt.Println("Encoding")

		go scan(file, stat.Size(), data)

		encapsulate(image, uint8(args.terminator), args.use_full_byte, args.align, data)
	}
}

func call_decode(args Arguments, image image.Image, data chan byte) {
	target_file, err := os.Create(args.target_path)
	if err != nil {
		panic(fmt.Sprintf("Error while opening %s for writing.", args.target_path))
	}

	fmt.Println("Decode")

	signal := make(chan bool)

	go writeToFile(target_file, data, signal)

	decode(image, args.use_full_byte, args.align, data)

	terminated := <-signal

	if !terminated {
		panic("There were errors writing the output file")
	}
}

type Arguments struct {
	image_path    string
	target_path   string
	wants_decode  bool
	use_full_byte bool
	align         rune
	terminator    uint
	only_info     bool
}

func parse_command_args() Arguments {
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
		panic("Wrong align parameter")
	}

	if image_path == "" || target_path == "" {
		panic("You must specify a valid path for both image and target")
	}

	if terminator > 255 || terminator == 0 {
		panic("Wrong terminator, must be in range 1-255")
	}

	return Arguments{image_path, target_path, wants_decode, full_byte, align, terminator, onlyInfo}
}
