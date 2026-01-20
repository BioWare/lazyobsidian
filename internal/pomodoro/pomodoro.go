// Package pomodoro implements the pomodoro timer functionality.
package pomodoro

import (
	"sync"
	"time"
)

// State represents the current state of the pomodoro timer.
type State int

const (
	StateIdle State = iota
	StateRunning
	StatePaused
	StateBreak
)

// Timer manages the pomodoro timer.
type Timer struct {
	mu            sync.RWMutex
	state         State
	remaining     time.Duration
	workDuration  time.Duration
	shortBreak    time.Duration
	longBreak     time.Duration
	sessionsCount int
	sessionsGoal  int
	dailyGoal     int
	dailyComplete int
	context       string
	ticker        *time.Ticker
	done          chan struct{}
	onTick        func(remaining time.Duration)
	onComplete    func(sessionType State)
}

// Config holds pomodoro timer configuration.
type Config struct {
	WorkMinutes        int
	ShortBreakMinutes  int
	LongBreakMinutes   int
	SessionsBeforeLong int
	DailyGoal          int
}

// NewTimer creates a new pomodoro timer.
func NewTimer(cfg Config) *Timer {
	return &Timer{
		state:         StateIdle,
		workDuration:  time.Duration(cfg.WorkMinutes) * time.Minute,
		shortBreak:    time.Duration(cfg.ShortBreakMinutes) * time.Minute,
		longBreak:     time.Duration(cfg.LongBreakMinutes) * time.Minute,
		sessionsGoal:  cfg.SessionsBeforeLong,
		dailyGoal:     cfg.DailyGoal,
		done:          make(chan struct{}),
	}
}

// Start starts a work session.
func (t *Timer) Start(context string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateRunning {
		return
	}

	t.context = context
	t.remaining = t.workDuration
	t.state = StateRunning
	t.ticker = time.NewTicker(time.Second)

	go t.run()
}

// Pause pauses the timer.
func (t *Timer) Pause() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != StateRunning {
		return
	}

	t.state = StatePaused
	if t.ticker != nil {
		t.ticker.Stop()
	}
}

// Resume resumes a paused timer.
func (t *Timer) Resume() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != StatePaused {
		return
	}

	t.state = StateRunning
	t.ticker = time.NewTicker(time.Second)
	go t.run()
}

// Stop stops the timer completely.
func (t *Timer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.state = StateIdle
	if t.ticker != nil {
		t.ticker.Stop()
	}
	close(t.done)
	t.done = make(chan struct{})
}

// State returns the current timer state.
func (t *Timer) State() State {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

// Remaining returns the remaining time.
func (t *Timer) Remaining() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.remaining
}

// Context returns the current context.
func (t *Timer) Context() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.context
}

// SessionsToday returns the number of completed sessions today.
func (t *Timer) SessionsToday() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.dailyComplete
}

// DailyGoal returns the daily goal.
func (t *Timer) DailyGoal() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.dailyGoal
}

// OnTick sets the callback for each tick.
func (t *Timer) OnTick(fn func(remaining time.Duration)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onTick = fn
}

// OnComplete sets the callback for session completion.
func (t *Timer) OnComplete(fn func(sessionType State)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onComplete = fn
}

func (t *Timer) run() {
	for {
		select {
		case <-t.done:
			return
		case <-t.ticker.C:
			t.mu.Lock()
			if t.state != StateRunning {
				t.mu.Unlock()
				return
			}

			t.remaining -= time.Second

			if t.onTick != nil {
				t.onTick(t.remaining)
			}

			if t.remaining <= 0 {
				t.handleComplete()
			}
			t.mu.Unlock()
		}
	}
}

func (t *Timer) handleComplete() {
	t.ticker.Stop()

	previousState := t.state

	if t.state == StateRunning {
		// Work session completed
		t.sessionsCount++
		t.dailyComplete++

		// Determine break type
		if t.sessionsCount >= t.sessionsGoal {
			t.remaining = t.longBreak
			t.sessionsCount = 0
		} else {
			t.remaining = t.shortBreak
		}
		t.state = StateBreak
	} else if t.state == StateBreak {
		// Break completed
		t.state = StateIdle
	}

	if t.onComplete != nil {
		t.onComplete(previousState)
	}
}

// AdjustTime adjusts the remaining time by delta minutes.
func (t *Timer) AdjustTime(deltaMinutes int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delta := time.Duration(deltaMinutes) * time.Minute
	t.remaining += delta

	if t.remaining < 0 {
		t.remaining = 0
	}
}
