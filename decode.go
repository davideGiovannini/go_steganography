package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"os"
)

func writeToFile(w *os.File, data chan byte, signal chan bool) {
	defer w.Close()
	//Using a buffered writer
	//for performance
	//(writing 1 byte at time is slow otherwise)
	bufWriter := bufio.NewWriter(w)

	for b := range data {

		err := bufWriter.WriteByte(b)

		if err != nil {
			fmt.Println("Error while writing to file:", err)
			signal <- false
			return
		}
	}

	//Make sure to write all the bytes
	err := bufWriter.Flush()

	if err != nil {
		fmt.Println("Error while flushing to file:", err)
		signal <- false
		return
	}

	//After writing
	//send a  signal to the main thread,
	//so that it can end and terminate the program
	signal <- true

}

/*
   Read the image from the bottom up, when it finds something different than 0 it assumes it is a terminator,
   and after that it begins sending bytes to writing
*/

func decode(im image.Image, full_byte bool, align rune, data chan byte) {
	defer close(data)

	notInFile := true // true as long as I have still not found terminator

	rect := im.Bounds()

	for x := rect.Dx() - 1; x >= 0; x-- {
		for y := rect.Dy() - 1; y >= 0; y-- {

			pixel, ok := im.At(x, y).(color.NRGBA) // must do type assertion to get real values from NRGBA

			if !ok {
				fmt.Println("Assertion color.NRGBA error at pixel", x, y)
			}

			R, G, B, A := pixel.R, pixel.G, pixel.B, pixel.A

			// Remove unnecessary part
			R = R & 0xF
			G = G & 0xF
			B = B & 0xF
			A = A & 0xF

			if full_byte { //Full byte ---------------------
				b1, b2 := extractFullBytes(R, G, B, A)

				if notInFile {
					if b2 != 0 {
						notInFile = false
						data <- b1
					} else {
						if b1 != 0 {
							notInFile = false
						}
					}

				} else {
					data <- b2
					data <- b1
				}

			} else { // If ALIGNED ---------------------

				byt := extractFromAlignment(R, G, B, A, align)

				if notInFile {
					if byt != 0 {
						notInFile = false
					}

				} else {
					data <- byt
				}
			}

		}
	}
}

func extractFromAlignment(R, G, B, A uint8, align rune) (b byte) {
	var mask, shift uint8
	switch {
	case align == 'r':
		mask = 3 // 0011
		shift = 0
	case align == 'c':
		mask = 6 // 0110
		shift = 1
	case align == 'l':
		mask = 12 // 1100
		shift = 2
	default:
		fmt.Println("Horrible Error! Wrong alignment")
		return
	}

	R = (R & mask) >> shift
	G = (G & mask) >> shift
	B = (B & mask) >> shift
	A = (A & mask) >> shift

	b = (R << 6) | (G << 4) | (B << 2) | A

	return
}

func extractFullBytes(r, g, b, a uint8) (b1, b2 byte) {
	b1 = r<<4 | g
	b2 = b<<4 | a
	return
}
