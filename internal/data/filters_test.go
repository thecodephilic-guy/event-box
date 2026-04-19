package data

import (
	"testing"

	"thecodephilic-guy/eventbox/internal/validator"
)

func TestSortColumn(t *testing.T) {
	tests := []struct {
		name     string
		sort     string
		safeList []string
		expected string
	}{
		{"ascending", "title", []string{"id", "title", "-id", "-title"}, "title"},
		{"descending prefix stripped", "-title", []string{"id", "title", "-id", "-title"}, "title"},
		{"id", "id", []string{"id", "-id"}, "id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filters{Sort: tt.sort, SortSafeList: tt.safeList}
			if got := f.sortColumn(); got != tt.expected {
				t.Errorf("sortColumn() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSortColumnPanicsOnUnsafe(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unsafe sort parameter")
		}
	}()

	f := Filters{Sort: "malicious", SortSafeList: []string{"id", "-id"}}
	f.sortColumn()
}

func TestSortDirection(t *testing.T) {
	tests := []struct {
		sort     string
		expected string
	}{
		{"title", "ASC"},
		{"-title", "DESC"},
		{"id", "ASC"},
		{"-id", "DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.sort, func(t *testing.T) {
			f := Filters{Sort: tt.sort}
			if got := f.sortDirection(); got != tt.expected {
				t.Errorf("sortDirection() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestLimitAndOffset(t *testing.T) {
	tests := []struct {
		name           string
		page, pageSize int
		wantLimit      int
		wantOffset     int
	}{
		{"page 1", 1, 20, 20, 0},
		{"page 2", 2, 20, 20, 20},
		{"page 3 size 10", 3, 10, 10, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filters{Page: tt.page, PageSize: tt.pageSize}
			if got := f.limit(); got != tt.wantLimit {
				t.Errorf("limit() = %d, want %d", got, tt.wantLimit)
			}
			if got := f.offset(); got != tt.wantOffset {
				t.Errorf("offset() = %d, want %d", got, tt.wantOffset)
			}
		})
	}
}

func TestCalculateMetaData(t *testing.T) {
	t.Run("no records", func(t *testing.T) {
		m := calculateMetaData(0, 1, 20)
		if m.TotalRecords != 0 {
			t.Errorf("expected 0 total records, got %d", m.TotalRecords)
		}
	})

	t.Run("with records", func(t *testing.T) {
		m := calculateMetaData(50, 2, 20)
		if m.CurrentPage != 2 {
			t.Errorf("expected current page 2, got %d", m.CurrentPage)
		}
		if m.PageSize != 20 {
			t.Errorf("expected page size 20, got %d", m.PageSize)
		}
		if m.FirstPage != 1 {
			t.Errorf("expected first page 1, got %d", m.FirstPage)
		}
		if m.LastPage != 3 {
			t.Errorf("expected last page 3, got %d", m.LastPage)
		}
		if m.TotalRecords != 50 {
			t.Errorf("expected 50 total records, got %d", m.TotalRecords)
		}
	})

	t.Run("exact division", func(t *testing.T) {
		m := calculateMetaData(40, 1, 20)
		if m.LastPage != 2 {
			t.Errorf("expected last page 2, got %d", m.LastPage)
		}
	})
}

func TestValidateFilters(t *testing.T) {
	tests := []struct {
		name    string
		filters Filters
		valid   bool
	}{
		{
			"valid",
			Filters{Page: 1, PageSize: 20, Sort: "id", SortSafeList: []string{"id", "-id"}},
			true,
		},
		{
			"page zero",
			Filters{Page: 0, PageSize: 20, Sort: "id", SortSafeList: []string{"id", "-id"}},
			false,
		},
		{
			"page too large",
			Filters{Page: 10_000_001, PageSize: 20, Sort: "id", SortSafeList: []string{"id", "-id"}},
			false,
		},
		{
			"page_size zero",
			Filters{Page: 1, PageSize: 0, Sort: "id", SortSafeList: []string{"id", "-id"}},
			false,
		},
		{
			"page_size too large",
			Filters{Page: 1, PageSize: 101, Sort: "id", SortSafeList: []string{"id", "-id"}},
			false,
		},
		{
			"unsafe sort",
			Filters{Page: 1, PageSize: 20, Sort: "malicious", SortSafeList: []string{"id", "-id"}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateFilters(v, tt.filters)
			if v.Valid() != tt.valid {
				t.Errorf("ValidateFilters() valid = %v, want %v (errors: %v)", v.Valid(), tt.valid, v.Errors)
			}
		})
	}
}
