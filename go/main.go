package main

import (
	"fmt"
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
var maxr float64

func isinsphere(x, y, z float64) bool {
	return math.Sqrt(x*x+y*y+z*z) < sphere_radius
}

type Point struct {
	X, Y, Z float64
}

const gap float64 = 2

func domain(p Point) bool {
	return math.Abs(math.Cos(p.X)+math.Cos(p.Y)+math.Cos(p.Z)) < 0.5
	// y1 := 0.0
	// if p.Y < 4 {
	// 	y1 = 1000.0
	// } else {
	// 	y1 = math.Mod(p.Y, gap) - gap/2
	// }
	// x1 := p.X //math.Mod(p.X, gap) - gap/2
	// return math.Sqrt(x1*x1+y1*y1+(math.Mod(p.Z, gap)-gap/2)*(math.Mod(p.Z, gap)-gap/2)) < sphere_radius

}

func probe(start Point, direction Point, domain func(Point) bool) float64 {
	stepSize := 0.05 // Set a small step size for marching
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

func initialize(this js.Value, p []js.Value) interface{} {
	return nil
}

func generateImage(this js.Value, p []js.Value) interface{} {

	zpos += 0.1
	// xpos += 0.1

	// imageBuffer := js.Global().Get("imageData")

	canvas := js.Global().Get("document").Call("getElementById", "myCanvas")
	width := int(canvas.Get("width").Int())
	height := int(canvas.Get("height").Int())

	imageData := make([]byte, width*height*4)

	sw = 1 //math.Tan(FOV/2) * 2 * FL
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

	fmt.Println(xpos)
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

				val := 255.0 * (math.Exp(-r / 5))

				imageData[4*pos+0] = uint8(val)
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
	js.Global().Set("initialize", js.FuncOf(initialize))

	<-c
}
