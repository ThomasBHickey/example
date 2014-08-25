// Copyright 2010 The Freetype-Go Authors. All rights reserved.
// Use of this source code is governed by your choice of either the
// FreeType License or the GNU General Public License version 2 (or
// any later version), both of which can be found in the LICENSE file.
// Modified by Thom Hickey, August 2014

package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"code.google.com/p/freetype-go/freetype/raster"
	"code.google.com/p/freetype-go/freetype/truetype"
)

var fontfile = flag.String("fontfile", "../testdata/luxisr.ttf", "filename of the ttf font")

func printBounds(b truetype.Bounds) {
	fmt.Printf("XMin:%d YMin:%d XMax:%d YMax:%d\n", b.XMin, b.YMin, b.XMax, b.YMax)
}

func printGlyph(g *truetype.GlyphBuf) {
	printBounds(g.B)
	fmt.Print("Points:\n---\n")
	e := 0
	for i, p := range g.Point {
		fmt.Printf("%4d, %4d", p.X, p.Y)
		if p.Flags&0x01 != 0 {
			fmt.Print("  on\n")
		} else {
			fmt.Print("  off\n")
		}
		if i+1 == int(g.End[e]) {
			fmt.Print("---\n")
			e++
		}
	}
}

func drawContour(r *raster.Rasterizer, ps []truetype.Point) {
	if len(ps) == 0 {
		return
	}
	start := rp(ps[0])
	r.Start(start)
	q0, on0 := start, true
	for _, p := range ps[1:] {
		q := rp(p)
		on := p.Flags&0x01 != 0
		if on {
			if on0 {
				r.Add1(q)
			} else {
				r.Add2(q0, q)
			}
		} else {
			if on0 { // No-op
			} else {
				mid := raster.Point{
					X: (q0.X + q.X) / 2,
					Y: (q0.Y + q.Y) / 2,
				}
				r.Add2(q0, mid)
			}
		}
		q0, on0 = q, on
	}
	// Close the curve
	if on0 {
		r.Add1(start)
	} else {
		r.Add2(q0, start)
	}
}

func addGlyph(r *raster.Rasterizer, g *truetype.GlyphBuf) {
	r.Clear()
	e0 := 0
	for _, e1 := range g.End {
		drawContour(r, g.Point[e0:e1])
		e0 = e1
	}
}

func showNodes(m *image.RGBA, g *truetype.GlyphBuf) {
	for _, p := range g.Point {
		mrp := rp(p)
		x, y := int(mrp.X)/256, int(mrp.Y)/256
		if !(image.Point{x, y}).In(m.Bounds()) {
			continue
		}
		var c color.Color
		switch p.Flags & 0x01 {
		case 0:
			c = color.RGBA{0, 255, 255, 255}
		case 1:
			c = color.RGBA{255, 0, 0, 255}
		}
		if c != nil {
			m.Set(x, y, c)
		}
	}
}

type node struct {
	x, y, degree int
}

func rp(p truetype.Point) raster.Point {
	x, y := 20+p.X/4, 380-p.Y/4
	return raster.Point{
		X: raster.Fix32(x * 256),
		Y: raster.Fix32(y * 256),
	}
}

func main() {
	const (
		w = 400
		h = 400
	)

	flag.Parse()
	fmt.Printf("Loading fontfile %q\n", *fontfile)
	b, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return
	}
	font, err := truetype.Parse(b)
	if err != nil {
		log.Println(err)
		return
	}
	fupe := font.FUnitsPerEm()
	printBounds(font.Bounds(fupe))
	fmt.Printf("FUnitsPerEm:%d\n\n", fupe)

	c0, c1 := 'a', 'V'

	i0 := font.Index(c0)
	hm := font.HMetric(fupe, i0)
	g := truetype.NewGlyphBuf()
	err = g.Load(font, fupe, i0, truetype.NoHinting) // truetype.FullHinting
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("'%c' glyph\n", c0)
	fmt.Printf("AdvanceWidth:%d LeftSideBearing:%d\n", hm.AdvanceWidth, hm.LeftSideBearing)

	r := raster.NewRasterizer(w, h)

	//printGlyph(g)
	addGlyph(r, g)
	mask := image.NewAlpha(image.Rect(0, 0, w, h))
	paint := raster.NewAlphaSrcPainter(mask)
	r.Rasterize(paint)

	// Draw the mask image (in gray) onto an RGBA image.
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	fg, bg := image.NewUniform(color.RGBA{0,0,0,128}), image.White
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	draw.DrawMask(rgba, rgba.Bounds(), fg, image.ZP, mask, image.ZP, draw.Over)
	showNodes(rgba, g)
	// Save that RGBA image to disk.
	f, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	bufw := bufio.NewWriter(f)
	err = png.Encode(bufw, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = bufw.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")

	i1 := font.Index(c1)
	fmt.Printf("\n'%c', '%c' Kerning:%d\n", c0, c1, font.Kerning(fupe, i0, i1))

}
