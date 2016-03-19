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

func (n *Node) addNodesForComponents(components []routeComponent, route Route) {
	firstComponent := components[0]
	componentValue := firstComponent.Value
	node := nodeForPart(n.nodes, componentValue)
	if node == nil {
		node = NewNode(componentValue)
		if strings.HasPrefix(componentValue, "{") {
			n.isWildcard = true
			if strings.Index(componentValue, ":") != -1 {
				// anything after the first colon is expected to be a regular expression
				partSplit := strings.SplitAfterN(componentValue, ":", 2)
				if len(partSplit) == 2 {
					regexpString := partSplit[1]
					regexp, regerr := regexp.Compile(regexpString)
					if regerr == nil {
						n.regexp = regexp
					}
					n.part = n.part[0 : len(n.part)-1]
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

func nodeForPart(nodes []*Node, part string) *Node {
	for _, node := range nodes {
		if node.part == part || node.isWildcard {
			return node
		}
	}
	return nil
}

func findNodeForPathComponents(nodes []*Node, pathComponents []string) *Node {
	checkNodes := nodes
	var matchedNode *Node
	for _, componentString := range pathComponents {
		node := nodeForPart(checkNodes, componentString)
		if node == nil {
			return nil
		}
		matchedNode = node
		checkNodes = node.nodes
	}
	return matchedNode
}

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
