// Goro
//
// Created by Posse in NYC
// http://goposse.com
//
// Copyright (c) 2016 Posse Productions LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

import (
	"fmt"
	"strings"
	"time"
)

// Matcher is the global matching engine
type Matcher struct {
	router       *Router
	LogMatchTime bool
}

// Match represents a matched node in the tree
type Match struct {
	Node           *Node
	WildcardValues map[string]string
	CatchAllValue  string
	ParentMatch    *Match
}

// NewMatch creates a new Match instance
func NewMatch(node *Node) *Match {
	return &Match{
		Node:           node,
		WildcardValues: map[string]string{},
		CatchAllValue:  "",
	}
}

// NewMatchWithParent creates a new instance of a match that passes down the values from
// the parent match
func NewMatchWithParent(node *Node, parentMatch *Match) *Match {
	match := &Match{
		Node:           node,
		WildcardValues: map[string]string{},
		CatchAllValue:  "",
	}
	match.ParentMatch = parentMatch
	if parentMatch != nil {
		match.WildcardValues = map[string]string{}
		// need to copy
		for key, value := range parentMatch.WildcardValues {
			match.WildcardValues[key] = value
		}
		match.CatchAllValue = parentMatch.CatchAllValue
	}
	return match
}

func (match *Match) String() string {
	return fmt.Sprintf("goro.Match # part=%s, wc=%v, ca=%v",
		match.Node.part, match.WildcardValues, match.CatchAllValue)
}

// NewMatcher creates a new instance of the Matcher
func NewMatcher(router *Router) *Matcher {
	return &Matcher{
		router:       router,
		LogMatchTime: false,
	}
}

// MatchPathToRoute attempts to match the given path to a registered Route
func (m *Matcher) MatchPathToRoute(method string, path string) *Match {
	startTime := time.Now()
	tree := m.router.routes
	if tree == nil {
		return nil // no routes registered for this method
	}

	nodesToCheck := tree.nodes
	var currentMatches []*Match
	candidate := NewMatchCandidate(path)
	if path == "/" {
		candidate.part = "/"
	}
	finalMatches := []*Match{}
	catchAlls := []*Match{}
	// loop until no more match candidates
	for candidate != NoMatchCandidate() {
		var matches []*Match
		var catchAllMatches []*Match
		if currentMatches == nil {
			matches, catchAllMatches = matchNodesForCandidate(method, candidate, nodesToCheck)
		} else {
			matches, catchAllMatches = matchCurrentMatchesForCandidate(method, candidate, currentMatches)
		}
		if len(catchAllMatches) > 0 {
			// append the old catch alls to the new ones so the deeper matches take precedent
			catchAlls = append(catchAllMatches, catchAlls...)
		}
		if len(matches) == 0 {
			break
		}
		currentMatches = matches
		if candidate.HasRemainingCandidates() {
			candidate = candidate.NextCandidate()
		} else {
			finalMatches = append(finalMatches, matches...)
		}
	}

	// fmt.Println("")
	// fmt.Println("")
	// log.Println("-----")
	// log.Println("final matches: ", finalMatches)
	// log.Println("catch alls:    ", catchAlls)

	var matchToUse *Match
	if len(finalMatches) > 0 {
		matchToUse = finalMatches[0]
	} else if len(catchAlls) > 0 {
		matchToUse = catchAlls[0]
	}

	// fmt.Println("")
	// log.Println("-----")
	// log.Println("final match:   ", matchToUse)
	if m.LogMatchTime {
		endTime := time.Now()
		Log("Matched in", endTime.Sub(startTime))
	}

	return matchToUse
}

func matchNodesForCandidate(method string, candidate MatchCandidate, nodes []*Node) (matches []*Match, catchalls []*Match) {
	if candidate != NoMatchCandidate() {
		return checkNodesForMatches(method, candidate, nodes, nil)
	}
	return []*Match{}, []*Match{}
}

func matchCurrentMatchesForCandidate(method string, candidate MatchCandidate, currentMatches []*Match) (matches []*Match, catchalls []*Match) {
	matchedNodes := []*Match{}
	catchAllNodes := []*Match{}
	for _, match := range currentMatches {
		node := match.Node
		if node.HasChildren() {
			nodeMatches, nodeCatchAlls := checkNodesForMatches(method, candidate, node.nodes, match)
			matchedNodes = append(matchedNodes, nodeMatches...)
			catchAllNodes = append(catchAllNodes, nodeCatchAlls...)
		}
	}
	return matchedNodes, catchAllNodes
}

func checkNodesForMatches(method string, candidate MatchCandidate, nodes []*Node, parentMatch *Match) (matches []*Match, catchalls []*Match) {
	matchedNodes := []*Match{}
	catchAllNodes := []*Match{}
	for _, node := range nodes {
		isWildcard := isWildcardPart(node.part)
		if (node.nodeType == ComponentTypeFixed && node.part == candidate.part) ||
			isWildcard {
			match := NewMatchWithParent(node, parentMatch)
			if isWildcard {
				match.WildcardValues[node.part[1:len(node.part)]] = candidate.part
			}
			if !candidate.HasRemainingCandidates() {
				routes := match.Node.routes
				foundMatch := false
				if len(routes) > 0 {
					for _, route := range routes {
						if route.Method == method {
							matchedNodes = append(matchedNodes, match)
							foundMatch = true
							break
						}
					}
				}
				if foundMatch {
					break // exit early we found a match
				}
			} else {
				matchedNodes = append(matchedNodes, match)
			}
		} else if isCatchAllPart(node.part) {
			match := NewMatchWithParent(node, parentMatch)
			match.CatchAllValue = candidate.currentPath
			catchAllNodes = append(catchAllNodes, match)
			if !candidate.HasRemainingCandidates() {
				break
			}
		}
	}
	// log.Println("pass matched:")
	// log.Println(" - nodes:     ", matchedNodes)
	// log.Println(" - sub nodes:     ", matchedSubNodes)
	// log.Println(" - catch alls:", catchAllNodes)
	return matchedNodes, catchAllNodes
}

// MatchCandidate is a helper class for matching path components
type MatchCandidate struct {
	part        string
	remainder   string
	currentPath string
}

// NewMatchCandidate creates a new match candidate instance and initializes if for the first part
func NewMatchCandidate(path string) MatchCandidate {

	cleanPath := path
	if strings.HasPrefix(path, "/") {
		cleanPath = path[1:len(path)]
	}
	split := strings.SplitN(cleanPath, "/", 2)
	candidate := MatchCandidate{
		part:        split[0],
		remainder:   "",
		currentPath: cleanPath,
	}
	if len(split) == 2 {
		candidate.remainder = split[1]
	}
	return candidate
}

// NextCandidate returns the next MatchCandidate in the full path
func (mc MatchCandidate) NextCandidate() MatchCandidate {
	if mc.remainder == "" {
		return NoMatchCandidate()
	}
	return NewMatchCandidate(mc.remainder)
}

// HasRemainingCandidates returns true if the MatchCandidate has more candidate parts
func (mc MatchCandidate) HasRemainingCandidates() bool {
	return mc.remainder != ""
}

// NoMatchCandidate represents an empty MatchCandidate
func NoMatchCandidate() MatchCandidate {
	return MatchCandidate{
		part:      "",
		remainder: "",
	}
}

// IsNoMatch returns true if the MatchCandidate equals the NoMatchCandidate value
func (mc MatchCandidate) IsNoMatch() bool {
	return mc == NoMatchCandidate()
}
