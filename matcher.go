package goro

import (
	"strings"
)

type routeMatcher struct {
	methodKeyedRoutes map[string][]Route
	registeredRoutes  []Route
	variables         map[string]interface{}
}

func NewRouteMatcher() *routeMatcher {
	return &routeMatcher{
		methodKeyedRoutes: make(map[string][]Route),
		registeredRoutes:  make([]Route, 0),
		variables:         make(map[string]interface{}),
	}
}

type StringRange struct {
	Start  int
	Length int
}

type Match struct {
	Type          string
	Value         string
	OriginalValue string
	Range         StringRange
}

type Matcher struct {
	inMatch     bool
	startIndex  int
	stringValue string
	startDelim  string
	endDelim    string
	matches     []Match
}

// matcher functions
// --

// NewMatcher - creates a new default Matcher instance
func NewMatcher(stringValue string, startDelim string, endDelim string) *Matcher {
	m := &Matcher{
		inMatch:     false,
		startIndex:  0,
		stringValue: stringValue,
		startDelim:  startDelim,
		endDelim:    endDelim,
		matches:     make([]Match, 0),
	}
	return m
}

// NextMatch - find the next string match, if no additional match is found,
// 			   		 returns a match with .Range == StringRangeNotFound()
func (m *Matcher) NextMatch() Match {
	// out of bounds .. we are done
	if m.startIndex > len(m.stringValue)-1 {
		return NotFoundMatch()
	}

	startIdx, str := 0, m.stringValue[m.startIndex:]
	rangeStart := 0
	for cidx, c := range str {
		if !m.inMatch && string(c) == m.startDelim {
			m.inMatch = true
			startIdx = cidx
			rangeStart = cidx + m.startIndex
		} else if m.inMatch && string(c) == m.endDelim {
			nextIndex := cidx + 1
			m.inMatch = false
			m.startIndex = m.startIndex + nextIndex
			val := str[startIdx:nextIndex]
			matchType := "wildcard"
			if strings.HasPrefix(val, "{$") {
				matchType = "variable"
			}
			match := Match{
				Type:          matchType,
				OriginalValue: m.stringValue,
				Value:         val,
				Range:         newStringRange(rangeStart, nextIndex-startIdx),
			}
			return match
		}
	}
	return NotFoundMatch()
}

func NotFoundMatch() Match {
	return Match{
		Type:          "notfound",
		Value:         "",
		OriginalValue: "",
		Range:         notFoundStringRange(),
	}
}

// string range functions
// --

// NewStringRange - helper to generate new string range
func newStringRange(start int, length int) StringRange {
	return StringRange{Start: start, Length: length}
}

// StringRangeNotFound - not found value
func notFoundStringRange() StringRange {
	return StringRange{Start: -1, Length: 0}
}
