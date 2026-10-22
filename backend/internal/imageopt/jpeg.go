// Package imageopt resizes and compresses backend-managed images before they
// are stored or sent to clients.
package imageopt

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"

	xdraw "golang.org/x/image/draw"
)

const (
	maxImageBytes = 200 * 1024
	maxDimension  = 1200
)

// JPEG returns a white-background JPEG no larger than 200 KiB. Images are
// resized only when needed, then quality and dimensions are reduced gradually
// until the encoded result fits the transfer budget.
func JPEG(source image.Image) ([]byte, error) {
	if source == nil || source.Bounds().Empty() {
		return nil, fmt.Errorf("image is empty")
	}

	flattened := flatten(resize(source, maxDimension))
	for dimension := maxDimension; dimension >= 480; dimension = dimension * 4 / 5 {
		resized := resize(flattened, dimension)
		for quality := 82; quality >= 50; quality -= 4 {
			var encoded bytes.Buffer
			if err := jpeg.Encode(&encoded, resized, &jpeg.Options{Quality: quality}); err != nil {
				return nil, err
			}
			if encoded.Len() <= maxImageBytes {
				return encoded.Bytes(), nil
			}
		}
	}
	return nil, fmt.Errorf("image cannot be compressed below %d bytes", maxImageBytes)
}

func flatten(source image.Image) image.Image {
	area := source.Bounds()
	result := image.NewRGBA(image.Rect(0, 0, area.Dx(), area.Dy()))
	draw.Draw(result, result.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(result, result.Bounds(), source, area.Min, draw.Over)
	return result
}

func resize(source image.Image, maxSide int) image.Image {
	bounds := source.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= maxSide && h <= maxSide {
		return source
	}
	if w >= h {
		h = max(1, h*maxSide/w)
		w = maxSide
	} else {
		w = max(1, w*maxSide/h)
		h = maxSide
	}
	result := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(result, result.Bounds(), source, bounds, xdraw.Src, nil)
	return result
}
