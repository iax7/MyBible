package tui

import "github.com/iax7/mybible/internal/module"

// recordHistory appends the newly reached Reference to the visit trail. A jump
// made after going back drops the entries that would be revisited, and the
// current Reference is never repeated.
func (m Model) recordHistory(previous module.Reference) Model {
	current, ok := m.currentReference()
	if !ok {
		return m
	}
	if len(m.history) == 0 {
		m.history, m.position = []module.Reference{previous}, 0
	}
	m.history = m.history[:m.position+1]
	if m.history[len(m.history)-1] != current {
		m.history = append(m.history, current)
		m.position = len(m.history) - 1
	}
	return m
}

// back moves to the previous visited Reference, adding no History entry.
func (m Model) back() Model { return m.stepHistory(-1) }

// forward moves to the next visited Reference, adding no History entry.
func (m Model) forward() Model { return m.stepHistory(1) }

// stepHistory jumps to the Reference a delta away in History without recording
// the move, staying put at either end.
func (m Model) stepHistory(delta int) Model {
	position := m.position + delta
	if position < 0 || position >= len(m.history) {
		return m
	}
	next, ok := m.jumpTo(m.history[position])
	if !ok {
		return m
	}
	next.position = position
	return next
}

// jumpToHistory jumps to a Reference chosen from the History modal, recording it
// as a jump the way any other deliberate move is.
func (m Model) jumpToHistory(position int) Model {
	if position < 0 || position >= len(m.history) {
		return m
	}
	next, _ := m.hop(m.history[position])
	return next
}
