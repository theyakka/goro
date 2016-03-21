// tree.go
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
	"regexp"
	"strings"
)

// Node - tree node to store route information
type Node struct {
	part       string
	isWildcard bool
	regexp     *regexp.Regexp
	route      *Route
	nodes      []*Node
	parent     *Node
}

// Tree - storage for routes
type Tree struct {
	nodes []*Node
}

// NewNode - creates a new node
func NewNode(part string) *Node {
	return &Node{
		part:       part,
		isWildcard: false,
		regexp:     nil,
		nodes:      []*Node{},
		parent:     nil,
	}
}

// Node find / search functions

// nodesForPart - given a part of a path, find any matching nodes
func nodesForPart(nodes []*Node, part string) []*Node {
	matchingNodes := []*Node{}
	for _, node := range nodes {
		if node.part == part || node.isWildcard {
			if node.regexp != nil {
				if !node.regexp.MatchString(part) {
					// if there is an assigned regular expression and it does not match
					// then this node is invalid as a match
					continue
				}
			}
			matchingNodes = append(matchingNodes, node)
		}
	}
	return matchingNodes
}

// nodesForPart - given a part of a path, find a matching node (only checks exact string match)
func nodeForPart(nodes []*Node, part string) *Node {
	for _, node := range nodes {
		if node.part == part {
			return node
		}
	}
	return nil
}

// findNodeForPathComponents - helper function to recursively check the tree for
//														 a node that matches the pathComponents slice
func findNodeForPathComponents(nodes []*Node, pathComponents []string) *Node {
	checkNodes := nodes
	var matchedNode *Node
	for idx, componentString := range pathComponents {
		matchedNodes := nodesForPart(checkNodes, componentString)
		nodeCount := len(matchedNodes)
		if nodeCount == 1 {
			// only 1 node matches the criteria, so we can use it now
			matchedNode = matchedNodes[0]
			checkNodes = matchedNode.nodes
		} else if nodeCount > 1 {
			if idx < (len(pathComponents) - 1) {
				subComponents := pathComponents[idx+1 : len(pathComponents)]
				subCheckNodes := []*Node{}
				for _, node := range matchedNodes {
					subCheckNodes = append(subCheckNodes, node.nodes...)
				}
				return findNodeForPathComponents(subCheckNodes, subComponents)
			}
			// just check matched node for first option with a route attached
			for _, node := range matchedNodes {
				if node.route != nil {
					return node
				}
			}
		}
	}
	return matchedNode
}

// RouteForPath - find the assigned route matching the given path (if it exists)
func (t *Tree) RouteForPath(path string) (route *Route, params map[string]interface{}) {
	var pathComponents []string
	var matchedParams = map[string]interface{}{}
	if path != "/" {
		pathComponents = strings.Split(path, "/")
		pathComponents = pathComponents[1:len(pathComponents)]
	} else {
		pathComponents = []string{"/"}
	}
	if len(t.nodes) > 0 {
		foundNode := findNodeForPathComponents(t.nodes, pathComponents)
		// extract the parameters
		checkNode := foundNode
		compIndex := len(pathComponents) - 1
		for checkNode != nil {
			if checkNode.isWildcard {
				paramName := stripTokenDelimiters(checkNode.part)
				matchedParams[paramName] = pathComponents[compIndex]
			}
			checkNode = checkNode.parent
			compIndex--
		}
		if foundNode != nil {
			return foundNode.route, matchedParams
		}
	}
	return nil, map[string]interface{}{}
}

// Parameter parsing functions

func parameterForNode(node *Node, part string) (paramKey string, paramValue string) {
	strippedPart := stripTokenDelimiters(node.part)
	return strippedPart, part
}

// Node creation functions

// addNodesForComponents - given an array of pathComponents, create the relevant
// 												 tree nodes and attach the Route
func (n *Node) addNodesForComponents(components []routeComponent, route *Route) {
	firstComponent := components[0]
	componentValue := firstComponent.Value
	node := nodeForPart(n.nodes, componentValue)
	if node == nil {
		node = NewNode(componentValue)
		node.parent = n
		if strings.HasPrefix(componentValue, "{") {
			node.isWildcard = true
			if strings.Index(componentValue, ":") != -1 {
				// anything after the first colon is expected to be a regular expression
				partSplit := strings.SplitAfterN(componentValue, ":", 2)
				if len(partSplit) == 2 {
					regexpString := partSplit[1]
					regexp, regerr := regexp.Compile(regexpString)
					if regerr == nil {
						// NOTE: only add the regular expression if it is valid. if it isn't,
						// we will assume this is a regular wildcard. This is important to
						// understand.
						node.regexp = regexp
					}
					node.part = node.part[0 : len(node.part)-1]
				}
			}
		}
		n.nodes = append(n.nodes, node)
	}

	if len(components) > 1 {
		node.addNodesForComponents(components[1:len(components)], route)
	} else {
		node.route = route
	}
}

// AddRoute - add the route to the tree for the given path
func (t *Tree) AddRoute(path string, route *Route) {
	isSingleComponent := (len(route.pathComponents) == 1)
	firstComponent := route.pathComponents[0]
	nodePart := firstComponent.Value
	if path == "/" {
		nodePart = path
	}
	node := nodeForPart(t.nodes, nodePart)
	if node == nil {
		node = NewNode(nodePart)
		t.nodes = append(t.nodes, node)
	}
	if isSingleComponent {
		node.route = route
	} else {
		slicedComponents := route.pathComponents[1:len(route.pathComponents)]
		node.addNodesForComponents(slicedComponents, route)
	}
}
