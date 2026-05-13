package diff

import (
	"JIT/utils"
	"fmt"
)

const (
	HUNK_CONTEXT = 3
)

type Hunk struct {
	a_start int
	b_start int
	edits   []*edit
}

func newHunk(a_start, b_start int, edits []*edit) *Hunk {
	return &Hunk{a_start: a_start, b_start: b_start, edits: edits}
}
func (h *Hunk) Header() string {
	a_start, a_size := h.offsetFor("a_line", h.a_start)
	b_start, b_size := h.offsetFor("b_line", h.b_start)

	return utils.Format("cyan", fmt.Sprintf("@@ -%d,%d +%d,%d @@", a_start, a_size, b_start, b_size))
}

func (h *Hunk) offsetFor(lineType string, start int) (int, int) {
	lines := make([]*line, 0)
	for _, edit := range h.edits {
		switch lineType {
		case "a_line":
			if edit.aLine != nil {
				lines = append(lines, edit.aLine)
			}
		case "b_line":
			if edit.bLine != nil {
				lines = append(lines, edit.bLine)
			}
		}
	}

	if len(lines) > 0 {
		start = lines[0].lineNumber
	}

	return start, len(lines)
}
func (h *Hunk) ToString() string {
	message := h.Header() + "\n"

	for _, edit := range h.edits {
		message += edit.toString()
	}

	return message
}

func hunkFilter(edits []*edit) []*Hunk {

	//We want to know for each hunk, what is the beginning of it ?/

	a_start, b_start := 0, 0
	hunks := make([]*Hunk, 0)

	offset := 0

	/*
		we will iterate over our edits till we found our first change (ins or del)
		once we find it, we move back HUNK_CONTEXT + 1 as we want to be on the first right before the
			beginning of our hunk to build from it
	*/
	for {
		for offset < len(edits) && edits[offset].Type == EQU {
			offset += 1
		}
		if offset >= len(edits) {
			break
		}

		// we got a change, lets move back
		offset -= HUNK_CONTEXT + 1

		// the current offset is our a_start and b_start
		if offset >= 0 {
			a_start = edits[offset].aLine.lineNumber
			b_start = edits[offset].bLine.lineNumber
		}

		// now lets build our hunk for the current offset
		hunk := newHunk(a_start, b_start, make([]*edit, 0))
		hunks = append(hunks, hunk)

		// we get the last position for the prev hunk and use it to begin the next one
		offset = build(hunk, edits, offset)
	}

	return hunks
}

func build(hunk *Hunk, edits []*edit, offset int) int {
	// this counter tells us that we can still add edits to our current hunk

	counter := -1
	for counter != 0 {
		if counter > 0 && offset >= 0 && offset < len(edits) {
			hunk.edits = append(hunk.edits, edits[offset])
		}

		offset += 1
		if offset >= len(edits) {
			break
		}

		// if we found a change 3 steps away from us, we will reset the counter
		lookAhead := offset + HUNK_CONTEXT

		if lookAhead < len(edits) {
			if edits[lookAhead].Type != EQU {
				counter = 2*HUNK_CONTEXT + 1
			} else {
				counter -= 1
			}
		} else {
			counter -= 1
		}
	}
	return offset
}
