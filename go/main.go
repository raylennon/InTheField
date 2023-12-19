package main

// package vectorops

import (
	"fmt"
	"math"
	"syscall/js"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/num/quat"
)

// IMPORTED JAVASCRIPT INFO
var canvas = js.Global().Get("document").Call("getElementById", "myCanvas")
var width = int(canvas.Get("width").Int())
var height = int(canvas.Get("height").Int())

// CAMERA SETTINGS
const maxdistance float64 = 4 * math.Pi // Distance at which we give up
const stepsize = 0.08                   // affects "smoothness", qualitatively

const FL float64 = 0.2 * 0.1                             // Focal length
var sw float64 = 0.8 * 0.1                               // sensor width
var sh float64 = sw * (float64(height) / float64(width)) // sensor height

// PLAYER VARIABLES
var pos = mat.NewVecDense(3, []float64{0, 0, 0})
var viewdir = quat.Number{Real: 0.0, Imag: 0.0, Jmag: 1.0, Kmag: 0.0}
var movedir = mat.NewVecDense(3, []float64{0, 0, 0})

var movesMap = map[string]*mat.VecDense{
	"U":  mat.NewVecDense(3, []float64{0, 1, 0}),
	"D":  mat.NewVecDense(3, []float64{0, -1, 0}),
	"L":  mat.NewVecDense(3, []float64{-2, 0, 0}),
	"R":  mat.NewVecDense(3, []float64{2, 0, 0}),
	"LU": mat.NewVecDense(3, []float64{-1, 1, 0}),
	"RU": mat.NewVecDense(3, []float64{1, 1, 0}),
	"LD": mat.NewVecDense(3, []float64{-1, -1, 0}),
	"RD": mat.NewVecDense(3, []float64{1, -1, 0}),
	"":   mat.NewVecDense(3, []float64{0, 0, 0}),
	// Add more mappings if needed
}

var status = ""
var impact = mat.NewVecDense(3, nil)

var paused bool = false
var movespeed float64 = 0.1
var turnspeed float64 = 0.2

var R = mat.NewDense(3, 3, nil)

// TO-DO: TRACK VELOCITY

var scandist = 2 * maxdistance

// Domain: the environment field
func domain(p [3]float64) bool {
	if p[1] < 2 { // Entrance
		return false
	}
	midway := domain_bb(p[0], p[1], p[2])
	return math.Abs(midway) < 0.2 // Gyroid Approximation
}

func domain_bb(x, y, z float64) float64 {
	res := math.Abs(math.Sin(x/2)*math.Cos(y/2)+math.Sin(y/2)*math.Cos(z/2)*math.Sin(z/2)*math.Cos(x/2)) - 0.2
	if y < 2 {
		return res - y + 2
	}
	return res
	// return z + 0.3*math.Sin(5*x) + 0.3*math.Sin(4*y)
}

func probe(start [3]float64, direction *mat.VecDense) float64 {
	fac := 1.0 //math.Sqrt(direction.At(0, 0)*direction.At(0, 0) + direction.At(1, 0)*direction.At(1, 0) + direction.At(2, 0)*direction.At(2, 0))
	conv := func(t float64) float64 {
		return domain_bb(
			start[0]+t*direction.At(0, 0)/fac,
			start[1]+t*direction.At(1, 0)/fac,
			start[2]+t*direction.At(2, 0)/fac,
		)
	}

	distance := 0.0
	// init := conv(0)
	val := maxdistance
	count := 0

	for count < 50 {
		count += 1
		val = conv(distance) *0.8
		if val < 0.01 {
			return distance
			// ret, _ := rootfinding.Brent(conv, distance-stepsize, distance, 2)
			// return ret
		}
		distance += math.Abs(val)
	}
	return distance
}

var fwidth = float64(js.Global().Get("window").Get("innerWidth").Float())
var fheight = float64(js.Global().Get("window").Get("innerHeight").Float())

func Cross(a, b [3]float64) [3]float64 {
	return [3]float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}
func grad(f func(x, y, z float64) float64, p [3]float64) []float64 {
	t := 0.01
	out := mat.NewVecDense(3, []float64{
		(f(p[0]+t/2, p[1], p[2]) - f(p[0]-t/2, p[1], p[2])) / 2,
		(f(p[0], p[1]+t/2, p[2]) - f(p[0], p[1]-t/2, p[2])) / 2,
		(f(p[0], p[1], p[2]+t/2) - f(p[0], p[1], p[2]-t/2)) / 2})
	out.ScaleVec(1/mat.Norm(out, 2), out)
	return out.RawVector().Data
}

func updateGamestate(this js.Value, p []js.Value) interface{} {

	if len(p) > 2 {
		if p[2].String() == "Pause!" {
			paused = !paused
		} else if p[2].String() == "Blast!" {
			scandist = 0.3
		} else if p[2].String() == "STOP" {
			// movespeed = 0
			status = ""
		} else {
			// movespeed = 0.3
			// ok := false
			status = p[2].String()
		}
	}
	// fmt.Println(status)
	if !(status == "") {

		temp, ok := movesMap[status]

		fmt.Println(temp)
		fmt.Println(mat.Norm(temp, 2))

		temp.ScaleVec(1/mat.Norm(temp, 2), temp)
		// movespeed += 0.01 * mat.Dot(temp, movedir)
		temp.ScaleVec(0.1, temp)
		// movedir.ScaleVec(movespeed, movedir)
		// temp.MulVec(R, temp)
		movedir.AddVec(temp, movedir)
		// movedir.ScaleVec(1/mat.Norm(movedir, 2), movedir)

		if !ok {
			movedir = mat.NewVecDense(3, []float64{0, 1, 0})
		}
	}

	if paused {
		return nil
	}
	if len(p) > 2 {
		return nil
	}
	cursorx := turnspeed * (p[0].Float() - fwidth/2.0) / fwidth
	cursory := 2 * turnspeed * (p[1].Float() - fheight/2.0) / fwidth
	cursormag := math.Sqrt(cursorx*cursorx + cursory*cursory)

	if cursormag > 0 {
		a := mat.NewVecDense(3, []float64{0, FL, 0})
		b := mat.NewVecDense(3, []float64{cursorx * fwidth, FL, cursory * fwidth})

		axis := mat.NewVecDense(3, []float64{
			a.At(1, 0)*b.At(2, 0) - a.At(2, 0)*b.At(1, 0),
			a.At(2, 0)*b.At(0, 0) - a.At(0, 0)*b.At(2, 0),
			a.At(0, 0)*b.At(1, 0) - a.At(1, 0)*b.At(0, 0),
		})

		axis.ScaleVec(1.0/math.Sqrt(axis.At(0, 0)*axis.At(0, 0)+axis.At(1, 0)*axis.At(1, 0)+axis.At(2, 0)*axis.At(2, 0)), axis)
		axis.ScaleVec(cursormag, axis)
		axis.MulVec(R, axis)

		rot := quat.Number{Real: math.Cos(cursorx / 2.0), Imag: axis.At(0, 0), Jmag: axis.At(1, 0), Kmag: axis.At(2, 0)}
		rot = quat.Scale(1.0/quat.Abs(rot), rot) // maybe unnecessary

		viewdir = quat.Mul(rot, viewdir)
		viewdir = quat.Scale(1.0/quat.Abs(viewdir), viewdir) // maybe unnecessary
	}
	t2 := mat.NewVecDense(3, nil)
	t2.MulVec(R, movedir)

	if probe([3]float64(pos.RawVector().Data), t2) > movespeed {
		pos.AddScaledVec(pos, movespeed, t2)
	} else {
		pos.AddScaledVec(pos, movespeed, t2)
	}

	if scandist < 2*maxdistance {
		scandist += 0.75
	}

	return nil
}

var render = mat.NewDense(width, height, nil)
var screenloc = mat.NewVecDense(3, []float64{0, 0, 0})
var x = mat.DenseCopyOf(render)
var y = mat.DenseCopyOf(render)

func generateImage(this js.Value, p []js.Value) interface{} {

	if paused == true {
		return nil
	}

	imageData := make([]byte, width*height*4)

	s := 1.0 / math.Pow(quat.Abs(viewdir), 2)

	R = mat.NewDense(3, 3, []float64{
		1 - 2*s*(viewdir.Jmag*viewdir.Jmag+viewdir.Kmag*viewdir.Kmag), 2 * s * (viewdir.Imag*viewdir.Jmag - viewdir.Kmag*viewdir.Real), 2 * s * (viewdir.Imag*viewdir.Kmag + viewdir.Jmag*viewdir.Real),
		2 * s * (viewdir.Imag*viewdir.Jmag + viewdir.Kmag*viewdir.Real), 1 - 2*s*(viewdir.Imag*viewdir.Imag+viewdir.Kmag*viewdir.Kmag), 2 * s * (viewdir.Jmag*viewdir.Kmag - viewdir.Imag*viewdir.Real),
		2 * s * (viewdir.Imag*viewdir.Kmag - viewdir.Jmag*viewdir.Real), 2 * s * (viewdir.Jmag*viewdir.Kmag + viewdir.Imag*viewdir.Real), 1 - 2*s*(viewdir.Imag*viewdir.Imag+viewdir.Jmag*viewdir.Jmag),
	})

	render.Apply(func(i, j int, v float64) float64 {

		screenloc.SetVec(0, x.At(i, j))
		screenloc.SetVec(1, FL)
		screenloc.SetVec(2, y.At(i, j))
		screenloc.MulVec(R, screenloc)
		screenloc.ScaleVec(1/mat.Norm(screenloc, 2), screenloc)
		dist := probe(
			[3]float64(pos.RawVector().Data),
			screenloc,
		)
		impact.AddScaledVec(pos, dist, screenloc)
		// if pso

		screenloc.ScaleVec(1/mat.Norm(screenloc, 2), screenloc)
		g := mat.NewVecDense(3, grad(domain_bb, [3]float64{
			pos.At(0, 0) + screenloc.At(0, 0)*dist,
			pos.At(1, 0) + screenloc.At(1, 0)*dist,
			pos.At(2, 0) + screenloc.At(2, 0)*dist}))
		// fmt.Println(g)
		g.ScaleVec(1/mat.Norm(g, 2), g)
		k := width*j + i

		imageData[4*k+3] = 255
		if dist == -1 {
			return dist
			// imageData[4*k+0] = uint8(255) // * (math.Sin(1*(ypos+r)) + 1) / 2.0)

		} else {
			val := (math.Exp(-(dist) * dist / 15))
			val1 := 1.0 //((grad(domain_bb, [3]float64(impact.RawVector().Data))[0]) + 2) / 3.0
			val2 := ((grad(domain_bb, [3]float64(impact.RawVector().Data))[0]) + 1) / 2.0
			val3 := ((grad(domain_bb, [3]float64(impact.RawVector().Data))[0]) + 1) / 2.0
			// if scandist < dist && dist < scandist+0.3 {
			// 	val = math.Min(1, val*3)
			// }
			if scandist < 2*maxdistance {

				val = math.Min(1, val+math.Exp(-math.Abs(scandist-dist))*math.Exp(-2*scandist/maxdistance))
			}
			// gx := math.Mod(impact.At(0, 0), math.Pi/2)
			// gy := math.Mod(impact.At(1, 0), math.Pi/2)
			// gz := math.Mod(impact.At(2, 0), math.Pi/2)
			// bt := 0.12
			// if !((-bt < gx && bt > gx) || (-bt < gy && bt > gy) || (-bt < gz && bt > gz)) {
			// 	val *= 0.75
			// }
			imageData[4*k+0] = uint8(val * val1 * 255) //(math.Sin(math.Pi*g.At(0, 0)) + 1) *
			imageData[4*k+1] = uint8(val * val2 * 255)
			imageData[4*k+2] = uint8(val * val3 * 255)
		}

		return dist
	}, render)

	jsArray := js.Global().Get("Uint8Array").New(len(imageData))
	js.CopyBytesToJS(jsArray, imageData)

	js.Global().Call("updateFromBuffer", jsArray)
	return nil
}

// func getImagePixel(this js.Value, p []js.Value) interface{} {
// 	// Access the canvas and load the image
// 	canvas := js.Global().Get("document").Call("createElement", "canvas")
// 	document := js.Global().Get("document")
// 	imageElement := document.Call("createElement", "img")
// 	imageElement.Set("src", "image.png") // Replace with your image path

// 	// Wait for the image to load
// 	imageElement.Call("addEventListener", "load", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
// 		// Set canvas dimensions to match image
// 		canvas.Set("width", imageElement.Get("width"))
// 		canvas.Set("height", imageElement.Get("height"))

// 		// Draw image onto canvas
// 		context := canvas.Call("getContext", "2d")
// 		context.Call("drawImage", imageElement, 0, 0)

// 		// Get pixel data
// 		imageData := context.Call("getImageData", 100, 150, 1, 1)
// 		pixelData := imageData.Get("data")

// 		// Extract RGB values from pixel data
// 		r := pixelData.Index(0).Int()
// 		g := pixelData.Index(1).Int()
// 		b := pixelData.Index(2).Int()
// 		a := pixelData.Index(3).Int()

// 		// Log pixel information
// 		fmt.Printf("Pixel at (100, 150): R=%d, G=%d, B=%d, A=%d\n", r, g, b, a)

// 		return nil
// 	}))
// 	return nil
// }

func main() {

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			x.Set(i, j, sw*float64(i-width/2)/float64(width))
			y.Set(i, j, sh*float64(j-height/2)/float64(height))
		}
	}

	c := make(chan struct{}, 0)

	// js.Global().Set("getImagePixel", js.FuncOf(getImagePixel))
	js.Global().Set("generateImage", js.FuncOf(generateImage))
	js.Global().Set("updateGamestate", js.FuncOf(updateGamestate))

	<-c
}
