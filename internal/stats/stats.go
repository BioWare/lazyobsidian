// Package stats handles statistics calculation and aggregation.
package stats

import (
	"time"

	"github.com/BioWare/lazyobsidian/pkg/types"
)

// Calculator handles statistics calculations.
type Calculator struct {
	sessions []types.PomodoroSession
}

// NewCalculator creates a new statistics calculator.
func NewCalculator() *Calculator {
	return &Calculator{
		sessions: make([]types.PomodoroSession, 0),
	}
}

// AddSession adds a pomodoro session.
func (c *Calculator) AddSession(session types.PomodoroSession) {
	c.sessions = append(c.sessions, session)
}

// Calculate calculates overall statistics.
func (c *Calculator) Calculate() types.Stats {
	stats := types.Stats{
		ByCategory: make(map[string]time.Duration),
		ByDay:      make(map[string]int),
	}

	for _, session := range c.sessions {
		if session.Type != types.PomodoroTypeWork {
			continue
		}

		duration := time.Duration(session.Duration) * time.Minute
		stats.TotalFocusTime += duration
		stats.TotalPomodoros++

		// Group by day
		day := session.StartedAt.Format("2006-01-02")
		stats.ByDay[day]++

		// Group by category
		if session.Context != "" {
			stats.ByCategory[session.Context] += duration
		}
	}

	// Calculate streaks
	stats.CurrentStreak, stats.LongestStreak = c.calculateStreaks()

	return stats
}

// ForPeriod returns statistics for a specific period.
func (c *Calculator) ForPeriod(start, end time.Time) types.Stats {
	filtered := &Calculator{sessions: make([]types.PomodoroSession, 0)}

	for _, session := range c.sessions {
		if session.StartedAt.After(start) && session.StartedAt.Before(end) {
			filtered.AddSession(session)
		}
	}

	return filtered.Calculate()
}

// Today returns today's statistics.
func (c *Calculator) Today() types.Stats {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	return c.ForPeriod(start, end)
}

// ThisWeek returns this week's statistics.
func (c *Calculator) ThisWeek() types.Stats {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	end := start.Add(7 * 24 * time.Hour)
	return c.ForPeriod(start, end)
}

func (c *Calculator) calculateStreaks() (current, longest int) {
	if len(c.sessions) == 0 {
		return 0, 0
	}

	// Build set of days with work sessions
	days := make(map[string]bool)
	for _, session := range c.sessions {
		if session.Type == types.PomodoroTypeWork {
			day := session.StartedAt.Format("2006-01-02")
			days[day] = true
		}
	}

	// Calculate current streak
	today := time.Now()
	for i := 0; ; i++ {
		day := today.AddDate(0, 0, -i).Format("2006-01-02")
		if days[day] {
			current++
		} else if i > 0 {
			break
		}
	}

	// Calculate longest streak (simplified - could be optimized)
	var streak int
	for day := range days {
		s := 1
		d, _ := time.Parse("2006-01-02", day)
		for {
			next := d.AddDate(0, 0, 1).Format("2006-01-02")
			if days[next] {
				s++
				d = d.AddDate(0, 0, 1)
			} else {
				break
			}
		}
		if s > longest {
			longest = s
		}
		streak = s
	}
	_ = streak

	return current, longest
}
