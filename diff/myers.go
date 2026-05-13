package diff

import (
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
	Value string
}
type Myers struct {
	a     []string
	b     []string
	edits []Edit
}

func NewMyersDiff(a, b string) *Myers {
	return &Myers{
		a:     lines(a),
		b:     lines(b),
		edits: make([]Edit, 0),
	}
}
func (ms *Myers) Diff() []Edit {
	trace := ms.shortestEdit()
	ms.backtrack(trace, func(prev_x, prev_y, x, y int) {
		if prev_x == x {
			ms.edits = append(ms.edits, Edit{INS, ms.b[y-1]})
		} else if prev_y == y {
			ms.edits = append(ms.edits, Edit{DEL, ms.a[x-1]})
		} else {
			ms.edits = append(ms.edits, Edit{EQU, ms.a[x-1]})
		}
	})

	slices.Reverse(ms.edits)
	return ms.edits
}
func (ms *Myers) shortestEdit() [][]int {
	n, m := len(ms.a), len(ms.b)
	max := n + m
	v := make([]int, 2*max+1)

	trace := make([][]int, 0)

	for d := range max + 1 {
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
	x, y := len(ms.a), len(ms.b)
	max := x + y

	for d := len(trace) - 1; d >= 0; d-- {
		k := x - y
		idx := max + k

		v := trace[d]

		var previous_k int
		if k == -d || (k != d && v[idx-1] < v[idx+1]) {
			previous_k = k + 1
		} else {
			previous_k = k - 1
		}

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
