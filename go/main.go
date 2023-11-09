package main

import (
	"fmt"
	"math"
	"syscall/js"
	"github.com/khezen/rootfinding"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/num/quat"
)

// IMPORTED JAVASCRIPT INFO
var canvas = js.Global().Get("document").Call("getElementById", "myCanvas")
var width = int(canvas.Get("width").Int())
var height = int(canvas.Get("height").Int())

// CAMERA SETTINGS
const FL float64 = 0.2 * 0.75  // Focal length
const maxdistance float64 = 20 // Distance at which we give up
const stepsize = 0.1           // affects "smoothness", qualitatively

var sw float64 = 0.8                                     // sensor width
var sh float64 = sw * (float64(height) / float64(width)) // sensor height

// PLAYER VARIABLES
var pos = mat.NewVecDense(3, []float64{0, 0, 0})
var dir = quat.Number{Real: 0.0, Imag: 0.0, Jmag: 1.0, Kmag: 0.0}
var paused bool = false
var movespeed float64 = 0.1
var turnspeed float64 = 0.3

var R = mat.NewDense(3, 3, nil)

// TO-DO: TRACK VELOCITY

// Domain: the environment field
func domain(p [3]float64) bool {
	if p[1] < 2 { // Entrance
		return false
	}
	midway := math.Sin(p[0])*math.Cos(p[1]) + math.Sin(p[1])*math.Cos(p[2])*math.Sin(p[2])*math.Cos(p[0])
	return math.Abs(midway) < 0.2 // Gyroid Approximation
}

func domain_bb(p [3]float64) float64 {
	if p[1]<2 {
		return 1
	}
	return math.Sin(p[0])*math.Cos(p[1]) + math.Sin(p[1])*math.Cos(p[2])*math.Sin(p[2])*math.Cos(p[0])
}


func probe2(start [3]float64, direction *mat.VecDense) float64 {
	fac := math.Sqrt(direction.At(0,0)*direction.At(0,0) + direction.At(1,0)*direction.At(1,0) + direction.At(2,0)*direction.At(2,0))
	conv := func(t float64) float64{
		return domain_bb([3]float64{
			start[0]+t*direction.At(0,0), 
			start[1]+t*direction.At(1,0), 
			start[2]+t*direction.At(2,0), 
		})
	}
	result, err := rootfinding.Brent(conv, 0.01, 5/fac, 3)
	// fmt.Println(err)
	if err != nil {
        return -1
    }
	if false {
		fmt.Println("false.")
	}
	return math.Abs(result)*fac
}

// TODO: Replace this function with something smarter/ faster. Root-finding?
// https://eprints.whiterose.ac.uk/3762/1/gamito_newraycastvisual.pdf
func probe(start [3]float64, direction *mat.VecDense, domain func([3]float64) bool) float64 {
	current := start
	distance := 0.0
	mag := math.Sqrt(math.Pow(direction.At(0, 0), 2) + math.Pow(direction.At(1, 0), 2) + math.Pow(direction.At(2, 0), 2))
	for distance < maxdistance {
		value := domain(current)
		if value {
			return distance
		}
		current[0] += direction.At(0, 0) * stepsize / mag
		current[1] += direction.At(1, 0) * stepsize / mag
		current[2] += direction.At(2, 0) * stepsize / mag
		distance += stepsize
	}
	return -1
}

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

			forward := mat.NewVecDense(3, []float64{0, 1, 0})
			forward.MulVec(R, forward)

			pos.SetVec(0, pos.At(0, 0)+forward.At(0, 0)*movespeed)
			pos.SetVec(1, pos.At(1, 0)+forward.At(1, 0)*movespeed)
			pos.SetVec(2, pos.At(2, 0)+forward.At(2, 0)*movespeed)
			// TODO: Extend to WASD for strafing & backward
		}
	}
	if paused {
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

		// fmt.Println(axis)
		axis.ScaleVec(1.0/math.Sqrt(axis.At(0, 0)*axis.At(0, 0)+axis.At(1, 0)*axis.At(1, 0)+axis.At(2, 0)*axis.At(2, 0)), axis)
		// fmt.Println(axis)
		axis.ScaleVec(cursormag, axis)
		axis.MulVec(R, axis)

		rot := quat.Number{Real: math.Cos(cursorx / 2.0), Imag: axis.At(0, 0), Jmag: axis.At(1, 0), Kmag: axis.At(2, 0)}
		// rot := quat.Number{Real: math.Cos(cursory / 2.0), Imag: math.Sin(cursory / 2), Jmag: 0.0, Kmag: 0.0}
		// rot = quat.Mul(rot1, rot)

		// rot := quat.Number{Real: math.Cos(cursory / 2.0), Imag: (cursory / cursormag) * math.Sin(cursormag/2), Jmag: 0.0, Kmag: (cursorx / cursormag) * math.Sin(cursormag/2)}
		// yrot := quat.Number{Real: math.Cos(cursory / 2.0), Imag: math.Sin(cursory / 2), Jmag: 0.0, Kmag: 0.0}

		rot = quat.Scale(1.0/quat.Abs(rot), rot) // maybe unnecessary
		// yrot = quat.Scale(1.0/quat.Abs(yrot), yrot) // maybe unnecessary

		dir = quat.Mul(rot, dir)
		dir = quat.Scale(1.0/quat.Abs(dir), dir) // maybe unnecessary
	}

	return nil
}

func generateImage(this js.Value, p []js.Value) interface{} {

	if paused == true {
		return nil
	}

	imageData := make([]byte, width*height*4)

	render := mat.NewDense(width, height, nil)
	x := mat.DenseCopyOf(render)
	y := mat.DenseCopyOf(render)

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			x.Set(i, j, sw*float64(i-width/2)/float64(width))
			y.Set(i, j, sh*float64(j-height/2)/float64(height))
		}
	}

	screenloc := mat.NewVecDense(3, []float64{0, 0, 0})

	s := 1.0 / math.Pow(quat.Abs(dir), 2)

	R = mat.NewDense(3, 3, []float64{
		1 - 2*s*(dir.Jmag*dir.Jmag+dir.Kmag*dir.Kmag), 2 * s * (dir.Imag*dir.Jmag - dir.Kmag*dir.Real), 2 * s * (dir.Imag*dir.Kmag + dir.Jmag*dir.Real),
		2 * s * (dir.Imag*dir.Jmag + dir.Kmag*dir.Real), 1 - 2*s*(dir.Imag*dir.Imag+dir.Kmag*dir.Kmag), 2 * s * (dir.Jmag*dir.Kmag - dir.Imag*dir.Real),
		2 * s * (dir.Imag*dir.Kmag - dir.Jmag*dir.Real), 2 * s * (dir.Jmag*dir.Kmag + dir.Imag*dir.Real), 1 - 2*s*(dir.Imag*dir.Imag+dir.Jmag*dir.Jmag),
	})

	render.Apply(func(i, j int, v float64) float64 {

		screenloc.SetVec(0, x.At(i, j))
		screenloc.SetVec(1, FL)
		screenloc.SetVec(2, y.At(i, j))

		screenloc.MulVec(R, screenloc)

		dist := probe2(
			[3]float64(pos.RawVector().Data),
			screenloc,
			// domain,
		)
		return dist
	}, render)

	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {

			k := width*j + i

			imageData[4*k+3] = 255
			r := render.At(i, j)
			if r ==-1 {
				imageData[4*k+0] = uint8(255) // * (math.Sin(1*(ypos+r)) + 1) / 2.0)
				continue // the terrifying void

			} else {
				val := 255.0 * (math.Exp(-r / 1.5))
				imageData[4*k+0] = uint8(val * 0.8) // * (math.Sin(1*(ypos+r)) + 1) / 2.0)
				imageData[4*k+1] = uint8(val)
				imageData[4*k+2] = uint8(val * 0.8)
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

	js.Global().Set("generateImage", js.FuncOf(generateImage))
	js.Global().Set("updateGamestate", js.FuncOf(updateGamestate))

	<-c
}
