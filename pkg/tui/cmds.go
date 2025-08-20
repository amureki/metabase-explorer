package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/amureki/metabase-explorer/pkg/api"
	"github.com/amureki/metabase-explorer/pkg/util"
	tea "github.com/charmbracelet/bubbletea"
)

func checkLatestVersion() tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get("https://api.github.com/repos/amureki/metabase-explorer/releases/latest")
		if err != nil {
			return versionChecked{err: err}
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return versionChecked{err: fmt.Errorf("GitHub API returned status %d", resp.StatusCode)}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return versionChecked{err: err}
		}

		var release struct {
			TagName string `json:"tag_name"`
		}

		if err := json.Unmarshal(body, &release); err != nil {
			return versionChecked{err: err}
		}

		return versionChecked{latestVersion: release.TagName}
	}
}

func loadDatabases(client *api.MetabaseClient) tea.Cmd {
	return func() tea.Msg {
		databases, err := client.GetDatabases()
		return databasesLoaded{databases: databases, err: err}
	}
}

func loadSchemas(client *api.MetabaseClient, databaseID int) tea.Cmd {
	return func() tea.Msg {
		tables, err := client.GetTables(databaseID)
		if err != nil {
			return schemasLoaded{err: err}
		}
		schemas := util.ExtractSchemas(tables)
		return schemasLoaded{schemas: schemas, err: nil}
	}
}

func loadTablesForSchema(client *api.MetabaseClient, databaseID int, schemaName string) tea.Cmd {
	return func() tea.Msg {
		allTables, err := client.GetTables(databaseID)
		if err != nil {
			return tablesLoaded{err: err}
		}

		var filteredTables []api.Table
		for _, table := range allTables {
			tableSchema := table.Schema
			if tableSchema == "" {
				tableSchema = "default"
			}
			if tableSchema == schemaName {
				filteredTables = append(filteredTables, table)
			}
		}

		return tablesLoaded{tables: filteredTables, err: nil}
	}
}

func loadFields(client *api.MetabaseClient, tableID int) tea.Cmd {
	return func() tea.Msg {
		fields, err := client.GetTableFields(tableID)
		return fieldsLoaded{fields: fields, err: err}
	}
}

func loadCollections(client *api.MetabaseClient) tea.Cmd {
	return func() tea.Msg {
		collections, err := client.GetCollections()
		return collectionsLoaded{collections: collections, err: err}
	}
}

func loadCollectionItems(client *api.MetabaseClient, collectionID interface{}) tea.Cmd {
	return func() tea.Msg {
		items, err := client.GetCollectionItems(collectionID)
		return collectionItemsLoaded{items: items, err: err}
	}
}

func loadGlobalSearch(client *api.MetabaseClient, query string) tea.Cmd {
	return func() tea.Msg {
		results, err := client.Search(query)
		return globalSearchLoaded{results: results, err: err}
	}
}

func tickSpinner() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return spinnerTick{}
	})
}
