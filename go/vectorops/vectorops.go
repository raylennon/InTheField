package vectorops

import "fmt"

func gradient(f func(x, y, z float64) float64, p [3]float64) []float64 {
	t := 0.01
	return []float64{
		(f(p[0]+t/2, p[1], p[2]) - f(p[0]-t/2, p[1], p[2])) / 2,
		(f(p[0], p[1]+t/2, p[2]) - f(p[0], p[1]-t/2, p[2])) / 2,
		(f(p[0], p[1], p[2]+t/2) - f(p[0], p[1], p[2]-t/2)) / 2,
	}
}

// Function in secondary.go
func SecondaryFunction() {
	fmt.Println("This is a function from secondary.go")
}
