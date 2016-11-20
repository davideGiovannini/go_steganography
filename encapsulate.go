package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
)

func scan(r io.Reader, size int64, data chan byte) {
	defer close(data)

	// read the file
	bs := make([]byte, size)
	_, err := r.Read(bs)
	if err != nil {
		fmt.Println("Error while reading input file!")
		return
	}

	// To simplify decoding  I'm writing bytes in reversed order
	for i := len(bs) - 1; i >= 0; i-- {
		data <- bs[i]
	}

}

func encapsulate(im image.Image, terminator byte, full_byte bool, align rune, data chan byte) {

	var has_terminated = false
	var last_byte byte

	rect := im.Bounds()

	//Create new image

	rgbaImage := image.NewNRGBA(rect)

	// Encoding part

	var pixel color.NRGBA
	var ok bool

	for x := 0; x < rect.Dx(); x++ {
		for y := 0; y < rect.Dy(); y++ {

			// must do type assertion to get real values from NRGBA
			switch t := im.At(x, y).(type) {
			case color.RGBA:
				pixel, ok = color.NRGBAModel.Convert(t).(color.NRGBA)
			case color.NRGBA:
				pixel = t
				ok = true
			case color.NRGBA64:
				fmt.Println("color.NRGBA64")
				pixel, ok = color.NRGBAModel.Convert(t).(color.NRGBA)
			default:
				fmt.Println("Other")
				pixel, ok = t.(color.NRGBA)
			}

			if !ok {
				fmt.Println("Assertion color.NRGBA error at pixel", x, y)
				return
			}

			R, G, B, A := pixel.R, pixel.G, pixel.B, pixel.A

			if full_byte {

				// FIRST BYTE --------------

				bits, still_open := <-data

				// if it has terminated it's printing 0 and its fine,
				// if it has not term it's printing right values
				//
				// ELSE
				bits, has_terminated = terminatorCheck(still_open, has_terminated, bits, last_byte, terminator)

				last_byte = bits
				R, G = byteWrite(R, G, bits)

				// SECOND BYTE --------------

				bits, still_open = <-data

				// if it has terminated it's printing 0 and its fine,
				// if it has not term it's printing right values
				//
				// ELSE
				bits, has_terminated = terminatorCheck(still_open, has_terminated, bits, last_byte, terminator)
				last_byte = bits
				B, A = byteWrite(B, A, bits)

				// -------------------------------

			} else { // -----------------------------------------------------------------------------------------
				bits, still_open := <-data

				var mask, offsett uint8

				// if it has terminated it's printing 0 and its fine,
				// if it has not term it's printing right values
				//
				// ELSE
				bits, has_terminated = terminatorCheck(still_open, has_terminated, bits, last_byte, terminator)

				switch {
				case align == 'c':
					mask = 0xF9
					offsett = 1
				case align == 'l':
					mask = 0xF3
					offsett = 2
				case align == 'r':
					mask = 0xFC
					offsett = 0
				}

				last_byte = bits

				R, G, B, A = alignedWrite(R, G, B, A, mask, offsett, bits)

				//

			}

			rgbaImage.Set(x, y, color.NRGBA{R, G, B, A})
		}
	}

	// Save new image
	image_file, err := os.Create(output_path)
	if err != nil {
		fmt.Println("Error while reopening", output_path, "for writing.")
		return
	}
	defer image_file.Close()

	err = png.Encode(image_file, rgbaImage)
	if err != nil {
		fmt.Println("Error while writing output png", err)
		return
	}
	fmt.Println("Done")
}

func byteWrite(x, y uint8, bits byte) (X, Y uint8) {
	X = x & 0xF0
	Y = y & 0xF0

	X = X | bits>>4
	Y = Y | bits&0x0F
	return
}

func alignedWrite(r, g, b, a, mask, offsett uint8, bits byte) (R, G, B, A uint8) {
	R = r & mask
	G = g & mask
	B = b & mask
	A = a & mask

	br, bg, bb, ba := bits>>6, (bits>>4)&0x3, (bits>>2)&0x3, bits&0x3

	R = R | br<<offsett
	G = G | bg<<offsett
	B = B | bb<<offsett
	A = A | ba<<offsett
	return
}

func terminatorCheck(still_open, has_terminated bool, bits, last_byte, terminator byte) (b byte, t bool) {
	if !still_open && !has_terminated {
		t = true
		b = terminator
	} else {
		t = has_terminated
		b = bits
	}
	return
}
