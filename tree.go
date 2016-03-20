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
	"fmt"
	"regexp"
	"strings"
)

// Node - tree node to store route information
type Node struct {
	part       string
	isWildcard bool
	regexp     *regexp.Regexp
	route      Route
	nodes      []*Node
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
	}
}

// Node find / search functions

// nodeForPart - given a part of a path, find the first matching node
func nodeForPart(nodes []*Node, part string) *Node {
	for _, node := range nodes {
		fmt.Println("part =", node.part, ", is WC=", node.isWildcard)
		if node.part == part || node.isWildcard {
			if node.regexp != nil {
				fmt.Println("check regexp")
				if !node.regexp.MatchString(part) {
					// if there is an assigned regular expression and it does not match
					// then this node is invalid as a match
					return nil
				}
			}
			fmt.Println("matches")
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
	for _, componentString := range pathComponents {
		node := nodeForPart(checkNodes, componentString)
		if node != nil {
			matchedNode = node
			checkNodes = node.nodes
		}
	}
	return matchedNode
}

// RouteForPath - find the assigned route matching the given path (if it exists)
func (t *Tree) RouteForPath(path string) Route {
	if path != "/" {
		pathComponents := strings.Split(path, "/")
		pathComponents = pathComponents[1:len(pathComponents)]
		if len(t.nodes) > 0 {
			foundNode := findNodeForPathComponents(t.nodes, pathComponents)
			if foundNode != nil {
				return foundNode.route
			}
		}
	}
	return NotFoundRoute()
}

// Node creation functions

// addNodesForComponents - given an array of pathComponents, create the relevant
// 												 tree nodes and attach the Route
func (n *Node) addNodesForComponents(components []routeComponent, route Route) {
	firstComponent := components[0]
	componentValue := firstComponent.Value
	node := nodeForPart(n.nodes, componentValue)
	if node == nil {
		node = NewNode(componentValue)
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
func (t *Tree) AddRoute(path string, route Route) {
	isSingleComponent := (len(route.pathComponents) == 1)
	firstComponent := route.pathComponents[0]
	node := nodeForPart(t.nodes, firstComponent.Value)
	if node == nil {
		node = NewNode(firstComponent.Value)
		t.nodes = append(t.nodes, node)
	}
	if isSingleComponent {
		node.route = route
	} else {
		slicedComponents := route.pathComponents[1:len(route.pathComponents)]
		node.addNodesForComponents(slicedComponents, route)
	}
}
