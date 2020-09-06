package goheader

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Calculable interface {
	Calculate(map[string]Value) error
	Get() string
}

type Value interface {
	Calculable
	Read(Reader) Issue
}

func calculateValue(calculable Calculable, values map[string]Value) (string, error) {
	sb := strings.Builder{}
	r := calculable.Get()
	var endIndex int
	var startIndex int
	for startIndex = strings.Index(r, "{{"); startIndex >= 0; startIndex = strings.Index(r, "{{") {
		_, _ = sb.WriteString(r[:startIndex])
		endIndex = strings.Index(r, "}}")
		if endIndex < 0 {
			return "", errors.New("missed value ending")
		}
		subVal := strings.ToLower(strings.TrimSpace(r[startIndex+2 : endIndex]))
		if val := values[subVal]; val != nil {
			if err := val.Calculate(values); err != nil {
				return "", err
			}
			sb.WriteString(val.Get())
		} else {
			return "", fmt.Errorf("unknown value name %v", subVal)
		}
		endIndex += 2
		r = r[endIndex:]
	}
	_, _ = sb.WriteString(r)
	return sb.String(), nil
}

type ConstValue struct {
	RawValue string
}

func (c *ConstValue) Calculate(values map[string]Value) error {
	v, err := calculateValue(c, values)
	if err != nil {
		return err
	}
	c.RawValue = v
	return nil
}

func (c *ConstValue) Get() string {
	return c.RawValue
}

func (c *ConstValue) Read(s Reader) Issue {
	l := s.Location()
	p := s.Position()
	for _, ch := range c.Get() {
		if ch != s.Peek() {
			s.SetPosition(p)
			f := s.ReadWhile(func(r rune) bool {
				return r != '\n'
			})
			return NewIssueWithLocation(fmt.Sprintf("Expected:%v, Actual: %v", c.Get(), f), l)
		}
		s.Next()
	}
	return nil
}

type RegexpValue struct {
	RawValue string
}

func (r *RegexpValue) Calculate(values map[string]Value) error {
	v, err := calculateValue(r, values)
	if err != nil {
		return err
	}
	r.RawValue = v
	return nil
}

func (r *RegexpValue) Get() string {
	return r.RawValue
}

func (r *RegexpValue) Read(s Reader) Issue {
	l := s.Location()
	p := regexp.MustCompile(r.Get())
	pos := s.Position()
	str := s.Finish()
	s.SetPosition(pos)
	indexes := p.FindAllIndex([]byte(str), -1)
	if len(indexes) == 0 {
		return NewIssueWithLocation(fmt.Sprintf("Pattern %v doesn't match.", p.String()), l)
	}
	s.SetPosition(pos + indexes[0][1])
	return nil
}

var _ Value = &ConstValue{}
var _ Value = &RegexpValue{}
