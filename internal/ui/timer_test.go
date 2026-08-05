package ui

import (
	"testing"
	"time"
)

func TestComputeElapsed(t *testing.T) {
	m := testModel()
	if m.computeElapsed() != 0 {
		t.Error("zero start should be 0")
	}

	m.timerStart = time.Now().Add(-5 * time.Second)
	if m.computeElapsed() != 5 {
		t.Errorf("elapsed = %d, want 5", m.computeElapsed())
	}

	m.timerStart = time.Now().Add(-2000 * time.Second)
	if m.computeElapsed() != 999 {
		t.Error("should cap at 999")
	}
}

func TestStopTimer(t *testing.T) {
	m := testModel()
	m.timerActive = true
	m.timerStart = time.Now().Add(-3 * time.Second)

	m = m.stopTimer()
	if m.timerActive {
		t.Error("timer should be inactive")
	}
	if m.elapsed != 3 {
		t.Errorf("elapsed = %d, want 3", m.elapsed)
	}
}

func TestTickDoesNotRescheduleWhenInactive(t *testing.T) {
	m := testModel()
	m.timerActive = false
	_, cmd := m.Update(tickMsg{})
	if cmd != nil {
		t.Error("inactive timer should not reschedule tick")
	}
}
