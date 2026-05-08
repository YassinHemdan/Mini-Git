package diff

import (
	"fmt"
	"slices"
)

type Diff interface {
	Diff()
}

const (
	EQU = ' '
	DEL = '-'
	INS = '+'
)

type Edit struct {
	Type  byte
	Value byte
}
type Myers struct {
	a      string
	b      string
	script []Edit
}

func NewMyersDiff(a, b string) *Myers {
	return &Myers{
		a:      a,
		b:      b,
		script: make([]Edit, 0),
	}
}
func (ms *Myers) Diff() {
	trace := ms.shortestEdit()
	ms.backtrack(trace, func(prev_x, prev_y, x, y int) {
		if prev_x == x {
			ms.script = append(ms.script, Edit{INS, ms.b[y-1]})
		} else if prev_y == y {
			ms.script = append(ms.script, Edit{DEL, ms.a[x-1]})
		} else {
			ms.script = append(ms.script, Edit{EQU, ms.a[x-1]})
		}
	})

	for i, j := 0, len(ms.script)-1; i < j; i, j = i+1, j-1 {
		ms.script[i], ms.script[j] = ms.script[j], ms.script[i]
	}

	for _, edit := range ms.script {
		fmt.Printf("%c %c\n", edit.Type, edit.Value)
	}
}

func (ms *Myers) shortestEdit() [][]int {
	/*
		The max possible edits is the sum of our two strings, for example, deleting all a and insert all b
		we will imagine of our string a is in our x-axis and b in our y-axis
		whenever we move to the right, it means that we are deleting from a
		and if we moveed down, it means we are inserting from b
		we will define k and d where k = x - y and d is the maximium depth we reached (could be size(a) + size(b))
	*/
	n, m := len(ms.a), len(ms.b)
	max := n + m
	v := make([]int, 2*max+1) // k could be ranged from -max to +max

	// we will save our iterations of d so that we can use it to backtrack and get our desired best path
	trace := make([][]int, 0)

	/*
		v[k] means the best x we reached for k in all d
		for the current (k, d) that we reached, did we came from a downward or a rightward from the previous d ?
		after we decide the path that we came from, we will compute the best x

		if k == -d -> must be downward
		else if k == +d -> must be rightward
		otherwise, we get the best x (highest) from both paths
	*/
	for d := range max {
		trace = append(trace, slices.Clone(v))

		for k := -d; k <= d; k += 2 {
			idx := max + k // to prevent negative indexes
			var x int
			if k == -d || (k != d && v[idx-1] < v[idx+1]) {
				x = v[idx+1]
			} else {
				x = v[idx-1] + 1
			}

			y := x - k

			// if the current x and y have the same letter, we make a diagonal move
			for x < n && y < m && ms.a[x] == ms.b[y] {
				x++
				y++
			}

			v[idx] = x
			if x >= n && y >= m {
				return trace
			}
		}
	}
	return nil
}
func (ms *Myers) backtrack(trace [][]int, fn func(prev_x, prev_y, x, y int)) {

	/*
		By definition, our target is the (len(a), len(b)), which is the right bottom corner
		the trace doesn't contain the values of the last iteration of 'd' as we only want from it the position of
		our target which is the right bottom corner and this is simply our len(a) and len(b)
		so using this info, we will backtrack from it to the previous position and we will figure it our as follows:

		We know that at a current position, we came from a rightward, downward or a diagonal from the previous d "d - 1"
		also since that we have our x, y. our last k = x - y
		and the previous k either equals to k + 1 or k - 1
		k + 1 if we came from a downward
		k - 1 if we came from a rightward
	*/
	x, y := len(ms.a), len(ms.b)
	max := x + y

	for d := len(trace) - 1; d >= 0; d-- {
		k := x - y
		idx := max + k

		v := trace[d] // this is our previous d that we will get our previous move from

		var previous_k int
		if k == -d || (k != d && v[idx-1] < v[idx+1]) {
			previous_k = k + 1
		} else {
			previous_k = k - 1
		}

		// now we need to get the previous_x and previous_y
		previous_idx := max + previous_k
		previous_x := v[previous_idx]
		previous_y := previous_x - previous_k

		// check for diagonal moves
		for x > previous_x && y > previous_y {
			fn(x-1, y-1, x, y)
			x--
			y--
		}
		if d > 0 {
			fn(previous_x, previous_y, x, y)
		}
		x, y = previous_x, previous_y
	}
}
