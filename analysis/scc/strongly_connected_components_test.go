package scc_test

import (
	"testing"

	"github.com/warpfork/go-ree/analysis/scc"
)

type test struct {
	// graph is an adjacency list of a graph with string nodes.
	graph map[string][]string

	// result maps node strings to component numbers.
	// The only thing that matters is, that nodes of the same component have the same number.
	result map[string]int

	// visited tracks which nodes have been visited by the dfs implemented for testing the SCC.
	visited map[string]struct{}

	// found is a set that allows tracking which nodes have been found in components.
	// This is to make sure all nodes are assigned to exactly one component.
	found map[string]struct{}
}

func TestSCC(t *testing.T) {
	tests := []test{
		{
			graph: map[string][]string{
				"A": {"B", "C"},
				"B": {"C", "D"},
				"C": {"E"},
				"D": {},
				"E": {"B"},
			},
			result: map[string]int{
				"A": 0,
				"B": 1,
				"C": 1,
				"E": 1,
				"D": 2,
			},
		},
	}

	for _, test := range tests {
		runSCCTest(test, t)
	}
}

func runSCCTest(test test, t *testing.T) {
	test.found = map[string]struct{}{}
	test.visited = map[string]struct{}{}
	scc := scc.StartSCC()
	for node := range test.graph {
		if _, ok := test.visited[node]; !ok {
			visit(test, t, scc, node)
		}
	}

	if len(test.found) != len(test.graph) {
		t.Error("there is a mismatch between the number of found and actual nodes")
	}
}

func visit(test test, t *testing.T, scc scc.State, node string) {
	test.visited[node] = struct{}{}
	scc.Visit(node)

	for _, neighbour := range test.graph[node] {
		shouldRecurse := scc.HandleEdge(node, neighbour)
		if shouldRecurse {
			visit(test, t, scc, neighbour)
			scc.AfterRecurse(node, neighbour)
		}
	}

	isComponentRoot, component := scc.EndVisit(node)
	if isComponentRoot {
		checkComponent(test, t, component)
	}
}

func checkComponent(test test, t *testing.T, component []interface{}) {
	if len(component) < 1 {
		t.Error("returned component is empty: components should always have at least one element")
	}

	// Check for duplicate nodes.
	nodes := map[string]struct{}{}
	for _, c := range component {
		if node, ok := c.(string); ok {
			nodes[node] = struct{}{}
			// Add node to found set, so we can later check if all nodes were found in a component.
			test.found[node] = struct{}{}
		} else {
			t.Error("returned component node has wrong type")
		}
	}

	if len(nodes) != len(component) {
		t.Errorf("found duplicate nodes in returned component: %v", component)
	}

	componentNumber := test.result[component[0].(string)]
	for node := range nodes {
		if test.result[node] != componentNumber {
			t.Errorf("nodes %q and %q were returned in the same component but should not be", component[0], node)
		}
	}

	// Check that the returned component is of the correct size.
	expectedSize := 0
	for _, number := range test.result {
		if componentNumber == number {
			expectedSize++
		}
	}

	if expectedSize != len(component) {
		t.Errorf("returned component was expected to contain %d nodes but contained %d nodes (namely: %v)", expectedSize, len(component), component)
	}
}
