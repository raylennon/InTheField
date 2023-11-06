package main

import (
	"math"
	"syscall/js"

	"gonum.org/v1/gonum/mat"
)

// or "image/png" for PNG

const FL float64 = 0.2 * 0.75   //math.Pi / 6  // Focal length
const FOV float64 = math.Pi / 2 // Field-of-View: radians
const maxdistance float64 = 20

var sw float64 // sensor width
var sh float64 // sensor height

const sphere_radius = 0.5

var xpos float64 = 0
var ypos float64 = -0.7
var zpos float64 = 0

var phi = 0
var theta = math.Pi / 2

var paused bool = false

func isinsphere(x, y, z float64) bool {
	return math.Sqrt(x*x+y*y+z*z) < sphere_radius
}

type Point struct {
	X, Y, Z float64
}

const gap float64 = 2

func domain(p Point) bool {
	// return math.Abs(math.Cos(p.X)+math.Cos(p.Y)+math.Cos(p.Z)) < 0.2 // P-surface
	// return math.Abs(math.Cos(p.X)*math.Cos(p.Y)*math.Cos(p.Z)-math.Sin(p.X)*math.Sin(p.Y)*math.Sin(p.Z)) < 0.2 // D-surface
	if p.Y < 5 {
		return false
	}
	midway := math.Sin(p.X)*math.Cos(p.Y) + math.Sin(p.Y)*math.Cos(p.Z)*math.Sin(p.Z)*math.Cos(p.X)
	return math.Abs(midway) < 0.2 // Gyroid
}

func probe(start Point, direction Point, domain func(Point) bool) float64 {
	stepSize := 0.1 // Set a small step size for marching
	current := start
	distance := 0.0

	mag := math.Sqrt(math.Pow(direction.X, 2) + math.Pow(direction.Y, 2) + math.Pow(direction.Z, 2))

	for distance < maxdistance {
		value := domain(current)
		if value {
			return distance
		}

		current.X += direction.X * stepSize / mag
		current.Y += direction.Y * stepSize / mag
		current.Z += direction.Z * stepSize / mag

		distance += stepSize
	}

	return -1
}

var canvas = js.Global().Get("document").Call("getElementById", "myCanvas")
var width = int(canvas.Get("width").Int())
var height = int(canvas.Get("height").Int())

var fwidth = float64(js.Global().Get("window").Get("innerWidth").Float())
var fheight = float64(js.Global().Get("window").Get("innerHeight").Float())

func updateGamestate(this js.Value, p []js.Value) interface{} {

	if len(p) > 2 && (p[2].String() == "Pause!") {
		paused = !paused
	}

	if paused {
		return nil
	}

	cursorx := 0.0 //fwidth/2.0
	cursory := 0.0 //fheight/2.0

	cursorx = p[0].Float()
	cursory = p[1].Float()

	xpos += ((cursorx - fwidth/2.0) / fwidth) * 1
	zpos += ((cursory - fheight/2.0) / fheight) * 1

	return nil
}

func generateImage(this js.Value, p []js.Value) interface{} {

	if paused == true {
		return nil
	}

	ypos += 0.2

	imageData := make([]byte, width*height*4)

	sw = 0.8 //math.Tan(FOV/2) * 2 * FL
	sh = sw * (float64(height) / float64(width))

	render := mat.NewDense(width, height, nil)
	x := mat.DenseCopyOf(render)
	y := mat.DenseCopyOf(render)

	// fmt.Println(width)
	// fmt.Println(height)

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			x.Set(i, j, sw*float64(i-width/2)/float64(width))
			y.Set(i, j, sh*float64(j-height/2)/float64(height))
		}
	}

	render.Apply(func(i, j int, v float64) float64 {
		dist := probe(Point{xpos, ypos, zpos}, Point{x.At(i, j), FL, y.At(i, j)}, domain)
		return dist
	}, render)

	// fmt.Println(xpos)
	// dmax := 4.0
	// dmin := 0.0
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {

			pos := width*j + i

			imageData[4*pos+3] = 255
			r := render.At(i, j)
			if r < 0 {
				// imageData[4*pos] = 100
				continue
			} else {
				// fmt.Println("Bonk!")
				// val := 255.0 * math.Pow((1-(math.Max(math.Min(dmax, r), dmin)-dmin)/(dmax-dmin)), 3)

				val := 255.0 * (math.Exp(-r / 1.5))

				imageData[4*pos+0] = uint8(val) // * (math.Sin(1*(ypos+r)) + 1) / 2.0)
				imageData[4*pos+1] = uint8(val)
				imageData[4*pos+2] = uint8(val)
			}
		}
	}

	jsArray := js.Global().Get("Uint8Array").New(len(imageData))
	js.CopyBytesToJS(jsArray, imageData)

	js.Global().Call("updateFromBuffer", jsArray)
	return nil
}

func main() {
	c := make(chan struct{}, 0)

	// Generate the image
	js.Global().Set("generateImage", js.FuncOf(generateImage))
	js.Global().Set("updateGamestate", js.FuncOf(updateGamestate))

	<-c
}
