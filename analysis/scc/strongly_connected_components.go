package scc

type (
	// State is contains all the state for one run of the strongly connected components (SCC) graph algorithm.
	State interface {
		aState() // Don't allow interface to be implemented outside of this package

		// Visit has to be called when visiting a node,
		// before any recursion happens.
		Visit(node interface{})

		// EndVisit has to be called before ending a visit to a node,
		// after all recursions have happened.
		EndVisit(node interface{}) (isComponentRoot bool, component Component)

		// HandleEdge is to be called to before recursing into a neighbour.
		// If the result (shouldRecurse) is false: don't recurse!
		HandleEdge(from, to interface{}) (shouldRecurse bool)

		// AfterRecurse is to be called after recursion into a neighbour is done.
		// It should not be called if HandleEdge(from, to) returned false.
		AfterRecurse(from, to interface{})
	}
	state struct {
		index    int
		vertices map[interface{}]*vertex
		stack    []*vertex
	}
	vertex struct {
		index, lowindex int
		isOnStack       bool
		node            interface{}
	}
	Component []interface{}
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func StartSCC() State {
	return &state{
		index:    0,
		vertices: map[interface{}]*vertex{},
		stack:    nil,
	}
}

func (s *state) aState() {}

func (s *state) Visit(node interface{}) {
	if _, ok := s.vertices[node]; ok {
		panic("invalid call to Visit(node): given node was already visited.")
	}

	v := &vertex{
		index:     s.index,
		lowindex:  s.index,
		isOnStack: true,
		node:      node,
	}
	s.vertices[node] = v
	s.stack = append(s.stack, v)
	s.index++
}

func (s *state) EndVisit(node interface{}) (isComponentRoot bool, component Component) {
	v := s.vertices[node]
	if v.index != v.lowindex {
		return false, nil
	}

	// Remove all vertices up to and including v from the stack.
	// The removed vertices form a strongly connected component.
	v.isOnStack = false
	i := len(s.stack) - 1
	for ; v != s.stack[i]; i-- {
		s.stack[i].isOnStack = false
	}

	// At this point the following holds: v == s.stack[i] && 0 <= i && i < len(s.stack)
	result := make([]interface{}, len(s.stack)-i)
	for j, w := range s.stack[i:] {
		result[j] = w.node
	}
	s.stack = s.stack[:i]

	return true, result
}

func (s *state) HandleEdge(from, to interface{}) (shouldRecurse bool) {
	toVector, ok := s.vertices[to]
	if ok && toVector.isOnStack {
		fromVector := s.vertices[from]
		// Note: toVector.index instead of toVector.lowindex is correct here
		fromVector.lowindex = min(fromVector.lowindex, toVector.index)
	}
	return !ok
}

func (s *state) AfterRecurse(from, to interface{}) {
	fromVector := s.vertices[from]
	toVector := s.vertices[to]
	fromVector.lowindex = min(fromVector.lowindex, toVector.lowindex)
}
