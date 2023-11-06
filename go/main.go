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

var pitch float64 = 0.0 // TO BE REMOVED
var yaw float64 = 0.0   //math.Pi / 2
var bearing [4]float64 = [4]float64{0, 0, 1, 0}

var paused bool = false

const gap float64 = 2

func domain(p [3]float64) bool {
	// return math.Abs(math.Cos(p.X)+math.Cos(p.Y)+math.Cos(p.Z)) < 0.2 // P-surface
	// return math.Abs(math.Cos(p.X)*math.Cos(p.Y)*math.Cos(p.Z)-math.Sin(p.X)*math.Sin(p.Y)*math.Sin(p.Z)) < 0.2 // D-surface
	if p[1] < 2 {
		return false
	}
	midway := math.Sin(p[0])*math.Cos(p[1]) + math.Sin(p[1])*math.Cos(p[2])*math.Sin(p[2])*math.Cos(p[0])
	return math.Abs(midway) < 0.2 // Gyroid
}

func probe(start [3]float64, direction *mat.VecDense, domain func([3]float64) bool) float64 {
	stepSize := 0.1 // Set a small step size for marching
	current := start
	distance := 0.0

	mag := math.Sqrt(math.Pow(direction.At(0, 0), 2) + math.Pow(direction.At(1, 0), 2) + math.Pow(direction.At(2, 0), 2))

	for distance < maxdistance {
		value := domain(current)
		if value {
			return distance
		}

		current[0] += direction.At(0, 0) * stepSize / mag
		current[1] += direction.At(1, 0) * stepSize / mag
		current[2] += direction.At(2, 0) * stepSize / mag

		distance += stepSize
	}

	return -1
}

var canvas = js.Global().Get("document").Call("getElementById", "myCanvas")
var width = int(canvas.Get("width").Int())
var height = int(canvas.Get("height").Int())

var fwidth = float64(js.Global().Get("window").Get("innerWidth").Float())
var fheight = float64(js.Global().Get("window").Get("innerHeight").Float())

func Cross(a, b [3]float64) [3]float64 {
	return [3]float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}

func updateGamestate(this js.Value, p []js.Value) interface{} {

	if len(p) > 2 {
		if p[2].String() == "Pause!" {
			paused = !paused
		} else if p[2].String() == "Go!" {
			xpos += 0.1 * math.Sin(pitch-math.Pi/2) * math.Sin(yaw)
			ypos -= 0.1 * math.Sin(pitch-math.Pi/2) * math.Cos(yaw)
			zpos += 0.1 * math.Sin(pitch)
		}
	}

	if paused {
		return nil
	}

	cursorx := p[0].Float()
	cursory := p[1].Float()

	// ax := Cross([3]float64{bearing[1], bearing[2], bearing[3]}, [3]float64{bearing[1], bearing[2], bearing[3]})

	// bearing.Apply(func(i, v float64) float64 {

	// }, bearing)

	yaw = math.Mod(yaw-((cursorx-fwidth/2.0)/fwidth)*0.4, 2*math.Pi)
	pitch = math.Mod(pitch+((cursory-fheight/2.0)/fheight)*0.4, 2*math.Pi)

	return nil
}

func generateImage(this js.Value, p []js.Value) interface{} {

	if paused == true {
		return nil
	}

	// ypos += 0.2

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

	screenloc := mat.NewVecDense(3, []float64{0, 0, 0})
	Rx := mat.NewDense(3, 3, []float64{1, 0, 0, 0, math.Cos(pitch), -math.Sin(pitch), 0.0, math.Sin(pitch), math.Cos(pitch)})
	Rz := mat.NewDense(3, 3, []float64{math.Cos(yaw), -math.Sin(yaw), 0, math.Sin(yaw), math.Cos(yaw), 0, 0, 0, 1})

	render.Apply(func(i, j int, v float64) float64 {

		screenloc.SetVec(0, x.At(i, j))
		screenloc.SetVec(1, FL)
		screenloc.SetVec(2, y.At(i, j))

		screenloc.MulVec(Rx, screenloc)
		screenloc.MulVec(Rz, screenloc)

		dist := probe(
			[3]float64{xpos, ypos, zpos},
			screenloc,
			domain)
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
