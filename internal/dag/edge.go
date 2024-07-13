// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dag

// Edge represents an edge in the graph, with a source and target vertex.
type Edge interface {
	Source() Vertex
	Target() Vertex
}

// BasicEdge returns an Edge implementation that simply tracks the source
// and target given as-is.
func BasicEdge(source, target Vertex) Edge {
	return &basicEdge{source, target}
}

// basicEdge is a basic implementation of Edge that has the source and
// target vertex.

type basicEdge [2]interface{}

func (e *basicEdge) Source() Vertex {
	return e[0]
}

func (e *basicEdge) Target() Vertex {
	return e[1]
}
