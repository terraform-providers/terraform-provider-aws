package goheader

type Reader interface {
	Peek() rune
	Next() rune
	Done() bool
	Finish() string
	Position() int
	Location() Location
	SetPosition(int)
	ReadWhile(func(rune) bool) string
}

func NewReader(text string) Reader {
	return &reader{source: text}
}

type reader struct {
	source   string
	position int
	location Location
}

func (r *reader) Position() int {
	return r.position
}

func (r *reader) Location() Location {
	return r.location
}

func (r *reader) Peek() rune {
	if r.Done() {
		return rune(0)
	}
	return rune(r.source[r.position])
}

func (r *reader) Done() bool {
	return r.position >= len(r.source)
}

func (r *reader) Next() rune {
	if r.Done() {
		return rune(0)
	}
	reuslt := r.Peek()
	if reuslt == '\n' {
		r.location.Line++
		r.location.Position = 0
	} else {
		r.location.Position++
	}
	r.position++
	return reuslt
}

func (r *reader) Finish() string {
	if r.position >= len(r.source) {
		return ""
	}
	defer r.till()
	return r.source[r.position:]
}

func (r *reader) SetPosition(pos int) {
	if pos < 0 {
		r.position = 0
	}
	r.position = pos
	r.location = r.calculateLocation()
}

func (r *reader) ReadWhile(match func(rune) bool) string {
	if match == nil {
		return ""
	}
	start := r.position
	for !r.Done() && match(r.Peek()) {
		r.Next()
	}
	return r.source[start:r.position]
}

func (r *reader) till() {
	r.position = len(r.source)
	r.location = r.calculateLocation()
}

func (r *reader) calculateLocation() Location {
	min := len(r.source)
	if min > r.position {
		min = r.position
	}
	x, y := 0, 0
	for i := 0; i < min; i++ {
		if r.source[i] == '\n' {
			y++
			x = 0
		} else {
			x++
		}
	}
	return Location{Line: y, Position: x}
}
