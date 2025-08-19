package util

import "github.com/amureki/metabase-explorer/pkg/api"

func ExtractSchemas(tables []api.Table) []api.Schema {
	schemaMap := make(map[string]int)
	for _, table := range tables {
		schema := table.Schema
		if schema == "" {
			schema = "default"
		}
		schemaMap[schema]++
	}

	var schemas []api.Schema
	for name, count := range schemaMap {
		schemas = append(schemas, api.Schema{
			Name:       name,
			TableCount: count,
		})
	}

	// Sort schemas by name for consistent display
	for i := 0; i < len(schemas)-1; i++ {
		for j := i + 1; j < len(schemas); j++ {
			if schemas[i].Name > schemas[j].Name {
				schemas[i], schemas[j] = schemas[j], schemas[i]
			}
		}
	}

	return schemas
}
