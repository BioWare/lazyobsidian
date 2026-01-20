// Package books handles book tracking functionality.
package books

import (
	"github.com/BioWare/lazyobsidian/pkg/types"
)

// Manager handles book operations.
type Manager struct {
	books []*types.Book
}

// NewManager creates a new book manager.
func NewManager() *Manager {
	return &Manager{
		books: make([]*types.Book, 0),
	}
}

// Add adds a book.
func (m *Manager) Add(book *types.Book) {
	m.books = append(m.books, book)
}

// All returns all books.
func (m *Manager) All() []*types.Book {
	return m.books
}

// CurrentlyReading returns books that are being read.
func (m *Manager) CurrentlyReading() []*types.Book {
	var reading []*types.Book
	for _, b := range m.books {
		if b.CurrentPage > 0 && b.CurrentPage < b.TotalPages {
			reading = append(reading, b)
		}
	}
	return reading
}

// CalculateProgress calculates progress percentage for a book.
func CalculateProgress(book *types.Book) float64 {
	if book.TotalPages == 0 {
		return 0
	}
	return float64(book.CurrentPage) / float64(book.TotalPages)
}

// EstimateTimeLeft estimates reading time remaining based on pages and reading speed.
func EstimateTimeLeft(book *types.Book, pagesPerHour int) float64 {
	if pagesPerHour == 0 {
		pagesPerHour = 30 // default
	}
	pagesLeft := book.TotalPages - book.CurrentPage
	return float64(pagesLeft) / float64(pagesPerHour)
}
