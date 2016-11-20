# Go_Steganography

This is a simple command line Go program that lets you embed a file inside a png image.

It was inspired by how the game [Monaco](https://en.wikipedia.org/wiki/Monaco:_What's_Yours_Is_Mine) stores its level data inside the level thumbnail itself.

~~~
Usage of steganography:
  -a string
        alignement if any - can be nothing, left(l), right(r) or center (c)
  -d    Decode from image
  -i    To get only information about the max possible size.
  -image string
        the image to use
  -t uint
        Terminator to use - must be an integer in range 1-255 (default 24)
  -target string
        the target file to encode or the decode output
~~~

## Examples

### Encoding
`steganography -image <container.png> -target <payload.file>`

### Decoding
`steganography -image <input.png> -target <extracted.file>`

### Info
Passing the `-i` parameter will just print size information about the container and the payload.

`steganography -i -image <container.png> -target <payload.file>`

`901 KB\ 916 KB`
