package goro

import ()

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
func NewMatcher(stringValue string, startDelim string, endDelim string) Matcher {
	m := Matcher{
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
// 			   returns a match with .Range == StringRangeNotFound()
func (m *Matcher) NextMatch() Match {
	// out of bounds .. we are done
	if m.startIndex > len(m.stringValue)-1 {
		return NotFoundMatch()
	}

	startIdx, str := 0, m.stringValue[m.startIndex:]
	for cidx, c := range str {
		if !m.inMatch && string(c) == m.startDelim {
			m.inMatch = true
			startIdx = cidx
		} else if m.inMatch && string(c) == m.endDelim {
			m.inMatch = false
			m.startIndex = cidx + 1
			matchType := "wildcard"
			match := Match{
				Type:          matchType,
				OriginalValue: str,
				Value:         str[startIdx : cidx+1],
				Range:         NewStringRange(startIdx, cidx-startIdx),
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
		Range:         NotFoundStringRange(),
	}
}

// string range functions
// --

// NewStringRange - helper to generate new string range
func NewStringRange(start int, length int) StringRange {
	return StringRange{Start: start, Length: length}
}

// StringRangeNotFound - not found value
func NotFoundStringRange() StringRange {
	return StringRange{Start: -1, Length: 0}
}
