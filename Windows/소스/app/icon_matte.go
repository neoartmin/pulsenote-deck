package main

// recoverOpaqueBlackMatteBGRA removes only an opaque, near-black matte that
// touches an icon's outer boundary. It never keys arbitrary dark pixels in
// the interior and never removes a colored corner/background. This is a
// conservative fallback for malformed Windows icons with a baked-in black
// rectangle (e.g. a branded mark on a square black canvas). Existing alpha is
// preserved and callers keep the original image when detection is uncertain.
// The pixel layout is GDI+ PixelFormat32bppARGB memory: B, G, R, A.
func recoverOpaqueBlackMatteBGRA(data []byte, width, height, stride int) bool {
	if width < 8 || height < 8 || width > 1024 || height > 1024 || stride < width*4 || len(data) < stride*height {
		return false
	}
	black := func(x, y int) bool {
		i := y*stride + x*4
		b, g, r, a := data[i], data[i+1], data[i+2], data[i+3]
		if a < 230 {
			return false
		}
		maxc, minc := b, b
		if g > maxc {
			maxc = g
		}
		if r > maxc {
			maxc = r
		}
		if g < minc {
			minc = g
		}
		if r < minc {
			minc = r
		}
		return maxc <= 56 && int(maxc)-int(minc) <= 24
	}
	corners := [][2]int{{0, 0}, {width - 1, 0}, {0, height - 1}, {width - 1, height - 1}}
	for _, c := range corners {
		if !black(c[0], c[1]) {
			return false
		}
	}
	edge, darkEdge := 0, 0
	for x := 0; x < width; x++ {
		edge += 2
		if black(x, 0) {
			darkEdge++
		}
		if black(x, height-1) {
			darkEdge++
		}
	}
	for y := 1; y < height-1; y++ {
		edge += 2
		if black(0, y) {
			darkEdge++
		}
		if black(width-1, y) {
			darkEdge++
		}
	}
	if darkEdge*100 < edge*80 {
		return false
	}

	// Flood fill only near-black pixels connected to the perimeter. Interior
	// black details of the mark are never considered a matte.
	marked := make([]bool, width*height)
	queue := make([]int, 0, width*height/2)
	add := func(x, y int) {
		idx := y*width + x
		if !marked[idx] && black(x, y) {
			marked[idx] = true
			queue = append(queue, idx)
		}
	}
	for x := 0; x < width; x++ {
		add(x, 0)
		add(x, height-1)
	}
	for y := 1; y < height-1; y++ {
		add(0, y)
		add(width-1, y)
	}
	for k := 0; k < len(queue); k++ {
		pos := queue[k]
		x, y := pos%width, pos/width
		if x > 0 {
			add(x-1, y)
		}
		if x+1 < width {
			add(x+1, y)
		}
		if y > 0 {
			add(x, y-1)
		}
		if y+1 < height {
			add(x, y+1)
		}
	}
	// Protect nearly empty icons and near-solid black marks.
	total := width * height
	if len(queue)*100 < total*8 || len(queue)*100 > total*90 {
		return false
	}
	for _, idx := range queue {
		x, y := idx%width, idx/width
		off := y*stride + x*4
		data[off] = 0
		data[off+1] = 0
		data[off+2] = 0
		data[off+3] = 0
	}
	return true
}
