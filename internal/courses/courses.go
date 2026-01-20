// Package courses handles course tracking functionality.
package courses

import (
	"github.com/BioWare/lazyobsidian/pkg/types"
)

// Manager handles course operations.
type Manager struct {
	courses []*types.Course
}

// NewManager creates a new course manager.
func NewManager() *Manager {
	return &Manager{
		courses: make([]*types.Course, 0),
	}
}

// Add adds a course.
func (m *Manager) Add(course *types.Course) {
	m.courses = append(m.courses, course)
}

// All returns all courses.
func (m *Manager) All() []*types.Course {
	return m.courses
}

// Active returns courses that are in progress.
func (m *Manager) Active() []*types.Course {
	var active []*types.Course
	for _, c := range m.courses {
		if c.Completed < c.TotalLessons && c.Completed > 0 {
			active = append(active, c)
		}
	}
	return active
}

// CalculateProgress calculates progress percentage for a course.
func CalculateProgress(course *types.Course) float64 {
	if course.TotalLessons == 0 {
		return 0
	}
	return float64(course.Completed) / float64(course.TotalLessons)
}

// CalculateSectionProgress calculates progress for a section.
func CalculateSectionProgress(section *types.CourseSection) float64 {
	if len(section.Lessons) == 0 {
		return 0
	}

	completed := 0
	for _, lesson := range section.Lessons {
		if lesson.Status == "x" || lesson.Status == "done" {
			completed++
		}
	}
	return float64(completed) / float64(len(section.Lessons))
}
