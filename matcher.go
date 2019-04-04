// Goro
//
// Created by Yakka
// http://theyakka.com
//
// Copyright (c) 2019 Yakka LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Matcher is the global matching engine
type Matcher struct {
	router             *Router
	LogMatchTime       bool
	FallbackToCatchAll bool
}

// Match represents a matched node in the tree
type Match struct {
	Node          *Node
	Params        map[string][]string
	CatchAllValue string
	ParentMatch   *Match
}

// NewMatch creates a new Match instance
func NewMatch(node *Node) *Match {
	return &Match{
		Node:          node,
		Params:        map[string][]string{},
		CatchAllValue: "",
	}
}

// NewMatchWithParent creates a new instance of a match that passes down the values from
// the parent match
func NewMatchWithParent(node *Node, parentMatch *Match) *Match {
	match := &Match{
		Node:          node,
		Params:        map[string][]string{},
		CatchAllValue: "",
	}
	match.ParentMatch = parentMatch
	if parentMatch != nil {
		match.Params = map[string][]string{}
		// need to copy
		for key, value := range parentMatch.Params {
			match.Params[key] = value
		}
		match.CatchAllValue = parentMatch.CatchAllValue
	}
	return match
}

func (match *Match) String() string {
	return fmt.Sprintf("goro.Match # part=%s, wc=%v, ca=%v",
		match.Node.part, match.Params, match.CatchAllValue)
}

// NewMatcher creates a new instance of the Matcher
func NewMatcher(router *Router) *Matcher {
	return &Matcher{
		router:             router,
		LogMatchTime:       false,
		FallbackToCatchAll: true,
	}
}

// MatchPathToRoute attempts to match the given path to a registered Route
func (m *Matcher) MatchPathToRoute(method string, path string, req *http.Request) *Match {

	startTime := time.Now()
	tree := m.router.routes
	if tree == nil {
		return nil // no routes registered for this method
	}

	nodesToCheck := tree.nodes
	var currentMatches []*Match
	candidate := NewMatchCandidate(path)
	if path == RootPath {
		candidate.part = RootPath
	}
	finalMatches := []*Match{}
	catchAlls := []*Match{}
	// loop until no more match candidates
	for candidate != NoMatchCandidate() {
		var matches []*Match
		var catchAllMatches []*Match
		if currentMatches == nil {
			matches, catchAllMatches, _ = m.matchNodesForCandidate(method, candidate, nodesToCheck)
			// Log("++", candidate.part)
			// Log("++", matches, catchAllMatches)
		} else {
			matches, catchAllMatches, _ = m.matchCurrentMatchesForCandidate(method, candidate, currentMatches)
			// Log("--", candidate.part)
			// Log("--", matches)
		}
		if len(catchAllMatches) > 0 {
			// append the old catch alls to the new ones so the deeper matches take precedent
			catchAlls = append(catchAllMatches, catchAlls...)
		}
		if len(matches) == 0 {
			break
		}
		currentMatches = matches
		if !candidate.HasRemainingCandidates() {
			finalMatches = append(finalMatches, matches...)
			break
		} else {
			candidate = candidate.NextCandidate()
		}
	}

	// fmt.Println("")
	// fmt.Println("")
	// Log("-----")
	// Log("final matches: ", finalMatches)
	// Log("catch alls:    ", catchAlls)

	var matchToUse *Match
	if len(finalMatches) > 0 {
		if !m.FallbackToCatchAll {
			matchToUse = finalMatches[0]
		} else {
			// check to see if the match contains a route that matches the requested method.
			// if not, fallback to a catch-all route (if one is matched)
			for _, match := range finalMatches {
				if match.Node.RouteForMethod(method) != nil {
					matchToUse = match
					break
				}
			}
		}
	}

	if matchToUse == nil && len(catchAlls) > 0 {
		matchToUse = catchAlls[0]
	}

	// fmt.Println("")
	// Log("-----")
	// Log("final match:   ", matchToUse)

	if m.LogMatchTime {
		endTime := time.Now()
		Log("Matched in", endTime.Sub(startTime))
	}
	return matchToUse
}

func (m *Matcher) matchNodesForCandidate(method string, candidate MatchCandidate, nodes []*Node) (matches []*Match, catchalls []*Match, errorCode int) {
	if candidate != NoMatchCandidate() {
		return m.checkNodesForMatches(method, candidate, nodes, nil)
	}
	return []*Match{}, []*Match{}, 0
}

func (m *Matcher) matchCurrentMatchesForCandidate(method string, candidate MatchCandidate, currentMatches []*Match) (matches []*Match, catchalls []*Match, errorCode int) {
	matchedNodes := []*Match{}
	catchAllNodes := []*Match{}
	errCode := 0
	for _, match := range currentMatches {
		node := match.Node
		if node.HasChildren() {
			nodeMatches, nodeCatchAlls, matchErrCode := m.checkNodesForMatches(method, candidate, node.nodes, match)
			matchedNodes = append(matchedNodes, nodeMatches...)
			catchAllNodes = append(catchAllNodes, nodeCatchAlls...)
			if errCode != 0 {
				errCode = matchErrCode
			}
		}
	}
	return matchedNodes, catchAllNodes, errCode
}

func (m *Matcher) checkNodesForMatches(method string, candidate MatchCandidate, nodes []*Node, parentMatch *Match) (matches []*Match, catchalls []*Match, errorCode int) {
	matchedNodes := []*Match{}
	catchAllNodes := []*Match{}
	errCode := 0

	for _, node := range nodes {
		isWildcard := isWildcardPart(node.part)
		if (node.nodeType == ComponentTypeFixed && strings.ToLower(node.part) == strings.ToLower(candidate.part)) ||
			isWildcard {
			match := NewMatchWithParent(node, parentMatch)
			if isWildcard {
				paramKey := strings.ToLower(node.part[1:len(node.part)])
				arr := match.Params[paramKey]
				if arr == nil {
					arr = []string{}
				}
				arr = append(arr, candidate.part)
				match.Params[paramKey] = arr
			}
			matchedNodes = append(matchedNodes, match)
			if m.FallbackToCatchAll == false && candidate.HasRemainingCandidates() == false {
				break // break early, we found a match
			}
		} else if isCatchAllPart(node.part) {
			match := NewMatchWithParent(node, parentMatch)
			match.CatchAllValue = candidate.currentPath
			catchAllNodes = append(catchAllNodes, match)
		}
	}

	return matchedNodes, catchAllNodes, errCode
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
