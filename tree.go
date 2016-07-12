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
	part     string
	nodeType RouteComponentType
	regexp   *regexp.Regexp // Not used currently
	route    *Route
	nodes    []*Node
	parent   *Node
}

// Tree - storage for routes
type Tree struct {
	nodes []*Node
}

// NewTree creates a new Tree instance
func NewTree() *Tree {
	return &Tree{
		nodes: []*Node{},
	}
}

// NewNode creates a new Node instance and appends it to the tree
func (t *Tree) NewNode(part string, parent *Node) *Node {
	nodeType := ComponentTypeFixed
	if strings.HasPrefix(part, "*") {
		nodeType = ComponentTypeCatchAll
	} else if strings.HasPrefix(part, ":") {
		nodeType = ComponentTypeWildcard
	}
	node := &Node{
		part:     part,
		nodeType: nodeType,
		regexp:   nil,
		nodes:    []*Node{},
		parent:   parent,
		route:    nil,
	}
	if parent == nil {
		t.nodes = append(t.nodes, node)
	} else {
		parent.nodes = append(parent.nodes, node)
	}
	return node
}

// AddRouteToTree splits the route into Nodes and adds them to the tree
func (t *Tree) AddRouteToTree(route *Route, variables map[string]string) {
	path := route.PathFormat
	deslashedPath := path[1:len(path)]
	split := strings.Split(deslashedPath, "/")
	if route.IsRoot() {
		node := t.NewNode("/", nil)
		node.route = route
	} else {
		var parentNode *Node
		var node *Node
		for _, component := range split {
			componentToUse := component
			// check to see if we need to do any variable substitution before parsing
			if isVariablePart(component) {
				componentToUse = variables[component]
				if componentToUse == "" {
					// we couldn't substitute the requested variable as there is no value definition
					panic(fmt.Sprintf("Missing variable substitution for '%s'. route='%s'", component, path))
				}
			}
			// get an existing node for this segment or attach to the tree
			node = t.nodeForExactPart(componentToUse, parentNode)
			if node == nil {
				node = t.NewNode(componentToUse, parentNode)
			}
			parentNode = node
		}
		node.route = route
	}
}

// nodeForExactPart - finds a Node either in the top-level Tree nodes or in the
// children of a parent node (if supplied) that has an exact string match for the
// supplied 'part' value
func (t *Tree) nodeForExactPart(part string, parentNode *Node) *Node {
	var nodesToCheck []*Node
	if parentNode == nil {
		nodesToCheck = t.nodes
	} else {
		nodesToCheck = parentNode.nodes
	}
	for _, node := range nodesToCheck {
		if node.part == part {
			return node
		}
	}
	return nil
}

// HasChildren returns true if the Node has 1 or more sub-Nodes
func (node *Node) HasChildren() bool {
	return node.nodes != nil && len(node.nodes) > 0
}

// part type helper functions
func isVariablePart(part string) bool {
	return strings.HasPrefix(part, "$")
}

func isWildcardPart(part string) bool {
	return strings.HasPrefix(part, ":")
}

func isCatchAllPart(part string) bool {
	return strings.HasPrefix(part, "*")
}

// String is the string representation of the object when printing
func (node *Node) String() string {
	return fmt.Sprintf("goro.Node # type=%d, part=%s, children=%d, route=%v",
		node.nodeType, node.part, len(node.nodes), node.route)
}
