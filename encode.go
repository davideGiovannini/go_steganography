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

func encode(im image.Image, terminator byte, full_byte bool, align rune, data chan byte) {

	var has_terminated = false
	var last_byte byte

	rect := im.Bounds()

	//Create new image

	rgbaImage := image.NewNRGBA(rect)

	// Encoding part

	for x := 0; x < rect.Dx(); x++ {
		for y := 0; y < rect.Dy(); y++ {
			R, G, B, A := get_pixel(im.At(x, y))

			if full_byte {
				R, G, B, A, has_terminated, last_byte =
					encode_full_byte(R, G, B, A, data, has_terminated,
						last_byte, terminator)
			} else {
				R, G, B, A, has_terminated, last_byte =
					encode_aligned(R, G, B, A, data, has_terminated,
						last_byte, terminator, align)
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

func get_pixel(in_pixel color.Color) (R, G, B, A uint8) {
	var ok bool
	var pixel color.NRGBA
	// must do type assertion to get real values from NRGBA
	switch t := in_pixel.(type) {
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
		panic("Assertion color.NRGBA error")
	}
	R, G, B, A = pixel.R, pixel.G, pixel.B, pixel.A
	return
}

func encode_full_byte(R, G, B, A uint8, data chan byte, has_terminated bool, last_byte, terminator byte) (nR, nG, nB, nA uint8, new_has_terminated bool, new_last_byte byte) {
	// FIRST BYTE --------------

	bits, still_open := <-data

	// if it has terminated it's printing 0 and its fine,
	// if it has not term it's printing right values
	//
	// ELSE
	bits, has_terminated = terminatorCheck(still_open, has_terminated, bits, last_byte, terminator)

	last_byte = bits
	nR, nG = byteWrite(R, G, bits)

	// SECOND BYTE --------------

	bits, still_open = <-data

	// if it has terminated it's printing 0 and its fine,
	// if it has not term it's printing right values
	//
	// ELSE
	bits, new_has_terminated = terminatorCheck(still_open, has_terminated, bits, last_byte, terminator)
	new_last_byte = bits
	nB, nA = byteWrite(B, A, bits)
	return
}

func encode_aligned(R, G, B, A uint8, data chan byte, has_terminated bool, last_byte, terminator byte, align rune) (nR, nG, nB, nA uint8, new_has_terminated bool, new_last_byte byte) {
	bits, still_open := <-data

	var mask, offsett uint8

	// if it has terminated it's printing 0 and its fine,
	// if it has not term it's printing right values
	//
	// ELSE
	bits, new_has_terminated = terminatorCheck(still_open, has_terminated, bits, last_byte, terminator)

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

	new_last_byte = bits

	nR, nG, nB, nA = alignedWrite(R, G, B, A, mask, offsett, bits)
	return
}

func byteWrite(x, y uint8, bits byte) (X, Y uint8) {
	X = x & 0xF0
	Y = y & 0xF0

	X = X | bits>>4
	Y = Y | bits&0x0F
	return
}

func alignedWrite(r, g, b, a, mask, offset uint8, bits byte) (R, G, B, A uint8) {
	R = r & mask
	G = g & mask
	B = b & mask
	A = a & mask

	br, bg, bb, ba := bits>>6, (bits>>4)&0x3, (bits>>2)&0x3, bits&0x3

	R = R | br<<offset
	G = G | bg<<offset
	B = B | bb<<offset
	A = A | ba<<offset
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
