package gcaptcha

import (
	"bytes"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
)

// renderImages draws a textured background, marks the gap, and returns the
// background and slider-piece PNGs.
func renderImages(width, height, piece, gapX, gapY int, rng *rand.Rand) ([]byte, []byte, error) {
	bg := image.NewRGBA(image.Rect(0, 0, width, height))
	base := color.RGBA{
		R: uint8(rng.Intn(80) + 60),
		G: uint8(rng.Intn(80) + 70),
		B: uint8(rng.Intn(80) + 90),
		A: 255,
	}
	for y := 0; y < height; y++ {
		fade := float64(y) / float64(height)
		c := color.RGBA{
			R: uint8(float64(base.R) * (0.75 + 0.25*fade)),
			G: uint8(float64(base.G) * (0.75 + 0.25*fade)),
			B: uint8(float64(base.B) * (0.75 + 0.25*fade)),
			A: 255,
		}
		draw.Draw(bg, image.Rect(0, y, width, y+1), &image.Uniform{C: c}, image.Point{}, draw.Src)
	}
	// Texture: translucent circles and lines make naive image matching harder.
	for i := 0; i < 14; i++ {
		cx, cy := rng.Intn(width), rng.Intn(height)
		r := rng.Intn(30) + 8
		c := color.RGBA{R: uint8(rng.Intn(256)), G: uint8(rng.Intn(256)), B: uint8(rng.Intn(256)), A: 90}
		drawCircle(bg, cx, cy, r, c)
	}
	for i := 0; i < 6; i++ {
		c := color.RGBA{R: uint8(rng.Intn(256)), G: uint8(rng.Intn(256)), B: uint8(rng.Intn(256)), A: 70}
		draw.Draw(bg, image.Rect(rng.Intn(width), rng.Intn(height), rng.Intn(width)+40, rng.Intn(height)+10), &image.Uniform{C: c}, image.Point{}, draw.Over)
	}

	// Piece: crop the original area and add a light border.
	pieceImg := image.NewRGBA(image.Rect(0, 0, piece, piece))
	draw.Draw(pieceImg, pieceImg.Bounds(), bg, image.Pt(gapX, gapY), draw.Src)
	border := color.RGBA{R: 255, G: 255, B: 255, A: 220}
	draw.Draw(pieceImg, image.Rect(0, 0, piece, 2), &image.Uniform{C: border}, image.Point{}, draw.Src)
	draw.Draw(pieceImg, image.Rect(0, 0, 2, piece), &image.Uniform{C: border}, image.Point{}, draw.Src)
	draw.Draw(pieceImg, image.Rect(0, piece-2, piece, piece), &image.Uniform{C: border}, image.Point{}, draw.Src)
	draw.Draw(pieceImg, image.Rect(piece-2, 0, piece, piece), &image.Uniform{C: border}, image.Point{}, draw.Src)

	// Background output: darken the gap so the slider target is visible.
	gap := color.RGBA{R: 0, G: 0, B: 0, A: 110}
	draw.Draw(bg, image.Rect(gapX, gapY, gapX+piece, gapY+piece), &image.Uniform{C: gap}, image.Point{}, draw.Over)

	var bgBuf, pieceBuf bytes.Buffer
	if err := png.Encode(&bgBuf, bg); err != nil {
		return nil, nil, err
	}
	if err := png.Encode(&pieceBuf, pieceImg); err != nil {
		return nil, nil, err
	}
	return bgBuf.Bytes(), pieceBuf.Bytes(), nil
}

func drawCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y > r*r {
				continue
			}
			px, py := cx+x, cy+y
			if image.Pt(px, py).In(img.Bounds()) {
				img.SetRGBA(px, py, c)
			}
		}
	}
}

func dataURI(pngBytes []byte) string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)
}

func newUUID() (string, error) {
	var b [16]byte
	if _, err := cryptorand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32], nil
}
