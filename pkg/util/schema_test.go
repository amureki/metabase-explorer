package util

import (
	"reflect"
	"testing"

	"github.com/amureki/metabase-explorer/pkg/api"
)

func TestExtractSchemas(t *testing.T) {
	tests := []struct {
		name     string
		tables   []api.Table
		expected []api.Schema
	}{
		{
			name:     "empty tables",
			tables:   []api.Table{},
			expected: []api.Schema{},
		},
		{
			name: "single schema with multiple tables",
			tables: []api.Table{
				{ID: 1, Name: "users", Schema: "public"},
				{ID: 2, Name: "orders", Schema: "public"},
				{ID: 3, Name: "products", Schema: "public"},
			},
			expected: []api.Schema{
				{Name: "public", TableCount: 3},
			},
		},
		{
			name: "multiple schemas",
			tables: []api.Table{
				{ID: 1, Name: "users", Schema: "public"},
				{ID: 2, Name: "analytics", Schema: "analytics"},
				{ID: 3, Name: "orders", Schema: "public"},
				{ID: 4, Name: "events", Schema: "analytics"},
			},
			expected: []api.Schema{
				{Name: "analytics", TableCount: 2},
				{Name: "public", TableCount: 2},
			},
		},
		{
			name: "empty schema defaults to 'default'",
			tables: []api.Table{
				{ID: 1, Name: "table1", Schema: ""},
				{ID: 2, Name: "table2", Schema: ""},
			},
			expected: []api.Schema{
				{Name: "default", TableCount: 2},
			},
		},
		{
			name: "mixed schemas with empty",
			tables: []api.Table{
				{ID: 1, Name: "users", Schema: "public"},
				{ID: 2, Name: "table1", Schema: ""},
				{ID: 3, Name: "events", Schema: "analytics"},
			},
			expected: []api.Schema{
				{Name: "analytics", TableCount: 1},
				{Name: "default", TableCount: 1},
				{Name: "public", TableCount: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSchemas(tt.tables)

			if len(result) == 0 && len(tt.expected) == 0 {
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExtractSchemas() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractSchemas_Sorting(t *testing.T) {
	tables := []api.Table{
		{ID: 1, Name: "table1", Schema: "z_schema"},
		{ID: 2, Name: "table2", Schema: "a_schema"},
		{ID: 3, Name: "table3", Schema: "m_schema"},
	}

	result := ExtractSchemas(tables)
	expected := []api.Schema{
		{Name: "a_schema", TableCount: 1},
		{Name: "m_schema", TableCount: 1},
		{Name: "z_schema", TableCount: 1},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ExtractSchemas() sorting failed = %v, want %v", result, expected)
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		latest   string
		expected bool
	}{
		{
			name:     "same versions",
			current:  "v1.0.0",
			latest:   "v1.0.0",
			expected: true,
		},
		{
			name:     "same versions without v prefix",
			current:  "1.0.0",
			latest:   "1.0.0",
			expected: true,
		},
		{
			name:     "different versions",
			current:  "v1.0.0",
			latest:   "v1.0.1",
			expected: false,
		},
		{
			name:     "dev version should allow update",
			current:  "dev",
			latest:   "v1.0.0",
			expected: false,
		},
		{
			name:     "mixed prefixes",
			current:  "v1.0.0",
			latest:   "1.0.0",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.current, tt.latest)
			if result != tt.expected {
				t.Errorf("compareVersions(%s, %s) = %v, want %v", tt.current, tt.latest, result, tt.expected)
			}
		})
	}
}
