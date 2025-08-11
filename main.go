package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/sahilm/fuzzy"
)

var version = "dev" // Will be overridden by ldflags during release builds

type Database struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Engine string `json:"engine"`
}

type Schema struct {
	Name       string
	TableCount int
}

type Table struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Schema      string  `json:"schema"`
	Description string  `json:"description"`
	Fields      []Field `json:"fields"`
}

type Field struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	Description    string `json:"description"`
	BaseType       string `json:"base_type"`
	EffectiveType  string `json:"effective_type"`
	SemanticType   string `json:"semantic_type"`
	DatabaseType   string `json:"database_type"`
	TableID        int    `json:"table_id"`
	Position       int    `json:"position"`
	Active         bool   `json:"active"`
	PreviewDisplay bool   `json:"preview_display"`
	Visibility     string `json:"visibility_type"`
}

type MetabaseClient struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

func NewMetabaseClient(baseURL, apiToken string) *MetabaseClient {
	return &MetabaseClient{
		baseURL:    baseURL,
		apiToken:   apiToken,
		httpClient: &http.Client{},
	}
}

func (c *MetabaseClient) testConnection() error {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %v", err)
	}

	apiURL, err := baseURL.Parse("/api/user/current")
	if err != nil {
		return fmt.Errorf("failed to construct API URL: %v", err)
	}

	req, _ := http.NewRequest("GET", apiURL.String(), nil)
	req.Header.Set("X-API-Key", c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API token authentication failed with status: %d - %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *MetabaseClient) getDatabases() ([]Database, error) {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	apiURL, err := baseURL.Parse("/api/database")
	if err != nil {
		return nil, fmt.Errorf("failed to construct API URL: %v", err)
	}

	req, _ := http.NewRequest("GET", apiURL.String(), nil)
	req.Header.Set("X-API-Key", c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get databases: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string][]Database
	json.NewDecoder(resp.Body).Decode(&result)
	return result["data"], nil
}

func (c *MetabaseClient) getTables(databaseID int) ([]Table, error) {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	apiURL, err := baseURL.Parse(fmt.Sprintf("/api/database/%d/metadata", databaseID))
	if err != nil {
		return nil, fmt.Errorf("failed to construct API URL: %v", err)
	}

	req, _ := http.NewRequest("GET", apiURL.String(), nil)
	req.Header.Set("X-API-Key", c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get tables: %d - %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var metadata struct {
		Tables []Table `json:"tables"`
	}

	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return metadata.Tables, nil
}

func (c *MetabaseClient) getTableFields(tableID int) ([]Field, error) {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	apiURL, err := baseURL.Parse(fmt.Sprintf("/api/table/%d/query_metadata", tableID))
	if err != nil {
		return nil, fmt.Errorf("failed to construct API URL: %v", err)
	}

	req, _ := http.NewRequest("GET", apiURL.String(), nil)
	req.Header.Set("X-API-Key", c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get table fields: %d - %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var queryMeta struct {
		Fields []Field `json:"fields"`
	}

	if err := json.Unmarshal(body, &queryMeta); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return queryMeta.Fields, nil
}

func openInBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}

func toSlug(name string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	// Remove other special characters that might cause issues
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (m model) getWebURL() string {
	baseURL := strings.TrimSuffix(m.client.baseURL, "/")

	switch m.currentView {
	case viewDatabases:
		if len(m.databases) > 0 && m.cursor < len(m.databases) {
			db := m.databases[m.cursor]
			slug := toSlug(db.Name)
			return fmt.Sprintf("%s/browse/databases/%d-%s", baseURL, db.ID, slug)
		}
	case viewSchemas:
		if len(m.schemas) > 0 && m.cursor < len(m.schemas) && m.selectedDatabase != nil {
			// Open the specific schema browse page
			schema := m.schemas[m.cursor]
			return fmt.Sprintf("%s/browse/databases/%d/schema/%s", baseURL, m.selectedDatabase.ID, schema.Name)
		} else if m.selectedDatabase != nil {
			db := m.selectedDatabase
			slug := toSlug(db.Name)
			return fmt.Sprintf("%s/browse/databases/%d-%s", baseURL, db.ID, slug)
		}
	case viewTables:
		if len(m.tables) > 0 && m.cursor < len(m.tables) && m.selectedDatabase != nil {
			// Open the specific table's reference page
			return fmt.Sprintf("%s/reference/databases/%d/tables/%d", baseURL, m.selectedDatabase.ID, m.tables[m.cursor].ID)
		} else if m.selectedDatabase != nil {
			return fmt.Sprintf("%s/admin/databases/%d", baseURL, m.selectedDatabase.ID)
		}
	case viewFields:
		if len(m.fields) > 0 && m.cursor < len(m.fields) && m.selectedTable != nil && m.selectedDatabase != nil {
			// Open the specific field's reference page
			field := m.fields[m.cursor]
			return fmt.Sprintf("%s/reference/databases/%d/tables/%d/fields/%d", baseURL, m.selectedDatabase.ID, m.selectedTable.ID, field.ID)
		} else if m.selectedTable != nil && m.selectedDatabase != nil {
			// Fallback to table reference page
			return fmt.Sprintf("%s/reference/databases/%d/tables/%d", baseURL, m.selectedDatabase.ID, m.selectedTable.ID)
		}
	}

	return baseURL
}

type viewState int

const (
	viewDatabases viewState = iota
	viewSchemas
	viewTables
	viewFields
)

type model struct {
	databases        []Database
	schemas          []Schema
	tables           []Table
	fields           []Field
	cursor           int
	loading          bool
	error            string
	client           *MetabaseClient
	currentView      viewState
	selectedDatabase *Database
	selectedSchema   *Schema
	selectedTable    *Table
	searchMode       bool
	searchQuery      string
	filteredIndices  []int
	spinnerIndex     int
	numberInput      string
	helpMode         bool
	helpCursor       int
	latestVersion    string
	updateAvailable  bool
}

type databasesLoaded struct {
	databases []Database
	err       error
}

type schemasLoaded struct {
	schemas []Schema
	err     error
}

type tablesLoaded struct {
	tables []Table
	err    error
}

type fieldsLoaded struct {
	fields []Field
	err    error
}

type versionChecked struct {
	latestVersion string
	err           error
}

type spinnerTick struct{}

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

func loadDatabases(client *MetabaseClient) tea.Cmd {
	return func() tea.Msg {
		databases, err := client.getDatabases()
		return databasesLoaded{databases: databases, err: err}
	}
}

func loadTables(client *MetabaseClient, databaseID int) tea.Cmd {
	return func() tea.Msg {
		tables, err := client.getTables(databaseID)
		return tablesLoaded{tables: tables, err: err}
	}
}

func loadSchemas(client *MetabaseClient, databaseID int) tea.Cmd {
	return func() tea.Msg {
		tables, err := client.getTables(databaseID)
		if err != nil {
			return schemasLoaded{err: err}
		}
		schemas := extractSchemas(tables)
		return schemasLoaded{schemas: schemas, err: nil}
	}
}

func extractSchemas(tables []Table) []Schema {
	schemaMap := make(map[string]int)
	for _, table := range tables {
		schema := table.Schema
		if schema == "" {
			schema = "default"
		}
		schemaMap[schema]++
	}

	var schemas []Schema
	for name, count := range schemaMap {
		schemas = append(schemas, Schema{
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

func loadTablesForSchema(client *MetabaseClient, databaseID int, schemaName string) tea.Cmd {
	return func() tea.Msg {
		allTables, err := client.getTables(databaseID)
		if err != nil {
			return tablesLoaded{err: err}
		}

		var filteredTables []Table
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

func loadFields(client *MetabaseClient, tableID int) tea.Cmd {
	return func() tea.Msg {
		fields, err := client.getTableFields(tableID)
		return fieldsLoaded{fields: fields, err: err}
	}
}

func tickSpinner() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return spinnerTick{}
	})
}

func initialModel() model {
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, `Error: .env file not found

To get started:
1. Create a .env file in the current directory:
   cp .env.example .env  # if available, or create manually

2. Add your Metabase configuration:
   METABASE_URL="https://your-metabase-instance.com/"
   METABASE_API_TOKEN="your-api-token-here"

3. Get your API token from Metabase Admin Settings → API Keys

Run 'mbx --help' for more information.
`)
		os.Exit(1)
	}

	metabaseURL := os.Getenv("METABASE_URL")
	apiToken := os.Getenv("METABASE_API_TOKEN")

	if metabaseURL == "" || apiToken == "" {
		fmt.Fprintf(os.Stderr, `Error: Missing required configuration

Please set the following in your .env file:
- METABASE_URL="https://your-metabase-instance.com/"
- METABASE_API_TOKEN="your-api-token-here"

Get your API token from Metabase Admin Settings → API Keys

Run 'mbx --help' for more information.
`)
		os.Exit(1)
	}

	client := NewMetabaseClient(metabaseURL, apiToken)
	return model{
		loading:     true,
		client:      client,
		currentView: viewDatabases,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			err := m.client.testConnection()
			if err != nil {
				return databasesLoaded{err: err}
			}
			return nil
		},
		loadDatabases(m.client),
		tickSpinner(),
		checkLatestVersion(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle search mode
		if m.searchMode {
			switch msg.String() {
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				m.filteredIndices = nil
				m.cursor = 0
			case "enter":
				// Select from filtered results
				if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
					actualIndex := m.filteredIndices[m.cursor]
					m.cursor = actualIndex
					m.searchMode = false
					m.searchQuery = ""
					m.filteredIndices = nil

					// Trigger selection
					if m.currentView == viewDatabases && len(m.databases) > 0 {
						m.selectedDatabase = &m.databases[actualIndex]
						m.currentView = viewSchemas
						m.cursor = 0
						m.loading = true
						m.error = ""
						return m, tea.Batch(loadSchemas(m.client, m.selectedDatabase.ID), tickSpinner())
					} else if m.currentView == viewSchemas && len(m.schemas) > 0 {
						m.selectedSchema = &m.schemas[actualIndex]
						m.currentView = viewTables
						m.cursor = 0
						m.loading = true
						m.error = ""
						return m, tea.Batch(loadTablesForSchema(m.client, m.selectedDatabase.ID, m.selectedSchema.Name), tickSpinner())
					} else if m.currentView == viewTables && len(m.tables) > 0 {
						m.selectedTable = &m.tables[actualIndex]
						m.currentView = viewFields
						m.cursor = 0
						m.loading = true
						m.error = ""
						return m, tea.Batch(loadFields(m.client, m.selectedTable.ID), tickSpinner())
					}
				}
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.updateSearch()
				}
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices)-1 {
					m.cursor++
				}
			default:
				// Add character to search query
				if len(msg.String()) == 1 {
					m.searchQuery += msg.String()
					m.updateSearch()
				}
			}
			return m, nil
		}

		// Normal navigation mode
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "?":
			m.helpMode = !m.helpMode
			if m.helpMode {
				m.helpCursor = 0
			}
			return m, nil
		case "/":
			if m.helpMode {
				return m, nil
			}
			m.searchMode = true
			m.searchQuery = ""
			m.cursor = 0
			return m, nil
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if m.helpMode {
				return m, nil
			}
			// Build up number input
			m.numberInput += msg.String()

			// Get current item count for validation
			var itemCount int
			switch m.currentView {
			case viewDatabases:
				itemCount = len(m.databases)
			case viewSchemas:
				itemCount = len(m.schemas)
			case viewTables:
				itemCount = len(m.tables)
			case viewFields:
				itemCount = len(m.fields)
			}

			// Try to parse the number and hover over the item if valid
			if num, err := strconv.Atoi(m.numberInput); err == nil && num >= 1 && num <= itemCount {
				// Just hover over the item, don't navigate yet
				m.cursor = num - 1 // Convert to 0-based index
			} else if len(m.numberInput) >= 3 || (len(m.numberInput) == 2 && m.numberInput[0] != '0') {
				// Clear input if it's too long or invalid
				// Allow leading zeros for 2-digit numbers (01, 02, etc)
				if num, err := strconv.Atoi(m.numberInput); err != nil || num > itemCount {
					m.numberInput = ""
				}
			}
		case "up", "k":
			if m.helpMode {
				if m.helpCursor > 0 {
					m.helpCursor--
				}
				return m, nil
			}
			m.numberInput = "" // Clear number input when using arrow keys
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.helpMode {
				// We have 3 links: Repository, Issues, Sponsor
				if m.helpCursor < 2 {
					m.helpCursor++
				}
				return m, nil
			}
			m.numberInput = "" // Clear number input when using arrow keys
			if m.currentView == viewDatabases && m.cursor < len(m.databases)-1 {
				m.cursor++
			} else if m.currentView == viewSchemas && m.cursor < len(m.schemas)-1 {
				m.cursor++
			} else if m.currentView == viewTables && m.cursor < len(m.tables)-1 {
				m.cursor++
			} else if m.currentView == viewFields && m.cursor < len(m.fields)-1 {
				m.cursor++
			}
		case "left", "h":
			if m.helpMode {
				// Exit help mode
				m.helpMode = false
				return m, nil
			}
			if m.numberInput != "" {
				// Clear number input
				m.numberInput = ""
			} else if m.currentView == viewSchemas {
				m.currentView = viewDatabases
				m.cursor = 0
				m.selectedDatabase = nil
				m.schemas = nil
			} else if m.currentView == viewTables {
				m.currentView = viewSchemas
				m.cursor = 0
				m.selectedSchema = nil
				m.tables = nil
			} else if m.currentView == viewFields {
				m.currentView = viewTables
				m.cursor = 0
				m.selectedTable = nil
				m.fields = nil
			}
		case "right", "l":
			if m.helpMode {
				// Open selected link in browser (same as Enter)
				var url string
				switch m.helpCursor {
				case 0:
					url = "https://github.com/amureki/metabase-explorer"
				case 1:
					url = "https://github.com/amureki/metabase-explorer/issues"
				case 2:
					url = "https://github.com/sponsors/amureki"
				}
				if err := openInBrowser(url); err != nil {
					m.error = fmt.Sprintf("Failed to open browser: %v", err)
				}
				return m, nil
			}
			// Clear number input after navigation
			m.numberInput = ""

			if m.currentView == viewDatabases && len(m.databases) > 0 {
				m.selectedDatabase = &m.databases[m.cursor]
				m.currentView = viewSchemas
				m.cursor = 0
				m.loading = true
				m.error = ""
				return m, tea.Batch(loadSchemas(m.client, m.selectedDatabase.ID), tickSpinner())
			} else if m.currentView == viewSchemas && len(m.schemas) > 0 {
				m.selectedSchema = &m.schemas[m.cursor]
				m.currentView = viewTables
				m.cursor = 0
				m.loading = true
				m.error = ""
				return m, tea.Batch(loadTablesForSchema(m.client, m.selectedDatabase.ID, m.selectedSchema.Name), tickSpinner())
			} else if m.currentView == viewTables && len(m.tables) > 0 {
				m.selectedTable = &m.tables[m.cursor]
				m.currentView = viewFields
				m.cursor = 0
				m.loading = true
				m.error = ""
				return m, tea.Batch(loadFields(m.client, m.selectedTable.ID), tickSpinner())
			}
		case "enter":
			if m.helpMode {
				// Open selected link in browser
				var url string
				switch m.helpCursor {
				case 0:
					url = "https://github.com/amureki/metabase-explorer"
				case 1:
					url = "https://github.com/amureki/metabase-explorer/issues"
				case 2:
					url = "https://github.com/sponsors/amureki"
				}
				if err := openInBrowser(url); err != nil {
					m.error = fmt.Sprintf("Failed to open browser: %v", err)
				}
				return m, nil
			}

			// Keep Enter as alternative to right arrow
			m.numberInput = ""

			if m.currentView == viewDatabases && len(m.databases) > 0 {
				m.selectedDatabase = &m.databases[m.cursor]
				m.currentView = viewSchemas
				m.cursor = 0
				m.loading = true
				m.error = ""
				return m, tea.Batch(loadSchemas(m.client, m.selectedDatabase.ID), tickSpinner())
			} else if m.currentView == viewSchemas && len(m.schemas) > 0 {
				m.selectedSchema = &m.schemas[m.cursor]
				m.currentView = viewTables
				m.cursor = 0
				m.loading = true
				m.error = ""
				return m, tea.Batch(loadTablesForSchema(m.client, m.selectedDatabase.ID, m.selectedSchema.Name), tickSpinner())
			} else if m.currentView == viewTables && len(m.tables) > 0 {
				m.selectedTable = &m.tables[m.cursor]
				m.currentView = viewFields
				m.cursor = 0
				m.loading = true
				m.error = ""
				return m, tea.Batch(loadFields(m.client, m.selectedTable.ID), tickSpinner())
			}
		case "w":
			webURL := m.getWebURL()
			if err := openInBrowser(webURL); err != nil {
				m.error = fmt.Sprintf("Failed to open browser: %v", err)
			}
		case "backspace":
			// Keep backspace as alternative to left arrow
			if m.numberInput != "" {
				// Clear number input
				m.numberInput = ""
			} else if m.currentView == viewSchemas {
				m.currentView = viewDatabases
				m.cursor = 0
				m.selectedDatabase = nil
				m.schemas = nil
			} else if m.currentView == viewTables {
				m.currentView = viewSchemas
				m.cursor = 0
				m.selectedSchema = nil
				m.tables = nil
			} else if m.currentView == viewFields {
				m.currentView = viewTables
				m.cursor = 0
				m.selectedTable = nil
				m.fields = nil
			}
		case "esc":
			if m.helpMode {
				m.helpMode = false
				return m, nil
			} else if m.numberInput != "" {
				// Clear number input
				m.numberInput = ""
			} else if m.currentView == viewSchemas {
				m.currentView = viewDatabases
				m.cursor = 0
				m.selectedDatabase = nil
				m.schemas = nil
			} else if m.currentView == viewTables {
				m.currentView = viewSchemas
				m.cursor = 0
				m.selectedSchema = nil
				m.tables = nil
			} else if m.currentView == viewFields {
				m.currentView = viewTables
				m.cursor = 0
				m.selectedTable = nil
				m.fields = nil
			}
		}

	case databasesLoaded:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err.Error()
		} else {
			m.databases = msg.databases
		}

	case schemasLoaded:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err.Error()
		} else {
			m.schemas = msg.schemas
			// Auto-skip schema view if only one schema
			if len(m.schemas) == 1 {
				m.selectedSchema = &m.schemas[0]
				m.currentView = viewTables
				m.cursor = 0
				m.loading = true
				return m, tea.Batch(loadTablesForSchema(m.client, m.selectedDatabase.ID, m.selectedSchema.Name), tickSpinner())
			}
		}

	case tablesLoaded:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err.Error()
		} else {
			m.tables = msg.tables
		}

	case fieldsLoaded:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err.Error()
		} else {
			m.fields = msg.fields
		}

	case versionChecked:
		if msg.err == nil && msg.latestVersion != "" {
			m.latestVersion = msg.latestVersion
			// Compare versions (handle v prefix)
			currentVersion := version
			if currentVersion != "dev" {
				// Normalize versions by removing v prefix
				normalizedCurrent := strings.TrimPrefix(currentVersion, "v")
				normalizedLatest := strings.TrimPrefix(msg.latestVersion, "v")
				if normalizedLatest != normalizedCurrent {
					m.updateAvailable = true
				}
			}
		}

	case spinnerTick:
		if m.loading {
			m.spinnerIndex = (m.spinnerIndex + 1) % 10
			return m, tickSpinner()
		}
	}

	return m, nil
}

func (m model) View() string {
	var output strings.Builder

	// Colors
	blue := lipgloss.Color("12")
	gray := lipgloss.Color("240")
	white := lipgloss.Color("15")
	red := lipgloss.Color("9")

	// Handle help mode first - return immediately without showing main content
	if m.helpMode {
		return m.renderHelpOverlay(&output, blue, gray, white)
	}

	// Header
	title := ""
	path := ""

	switch m.currentView {
	case viewDatabases:
		title = fmt.Sprintf("Metabase Explorer %s", version)
		if len(m.databases) > 0 {
			path = fmt.Sprintf("Databases (%d)", len(m.databases))
		} else {
			path = "Databases"
		}
	case viewSchemas:
		title = "Database Schemas"
		if len(m.schemas) > 0 {
			path = fmt.Sprintf("Databases > %s (%d)", m.selectedDatabase.Name, len(m.schemas))
		} else {
			path = fmt.Sprintf("Databases > %s", m.selectedDatabase.Name)
		}
	case viewTables:
		title = "Schema Tables"
		if len(m.tables) > 0 {
			path = fmt.Sprintf("Databases > %s > %s (%d)", m.selectedDatabase.Name, m.selectedSchema.Name, len(m.tables))
		} else {
			path = fmt.Sprintf("Databases > %s > %s", m.selectedDatabase.Name, m.selectedSchema.Name)
		}
	case viewFields:
		title = "Table Schema"
		tableName := m.selectedTable.DisplayName
		if tableName == "" {
			tableName = m.selectedTable.Name
		}
		if len(m.fields) > 0 {
			path = fmt.Sprintf("Databases > %s > %s > %s (%d)", m.selectedDatabase.Name, m.selectedSchema.Name, tableName, len(m.fields))
		} else {
			path = fmt.Sprintf("Databases > %s > %s > %s", m.selectedDatabase.Name, m.selectedSchema.Name, tableName)
		}
	}

	output.WriteString(lipgloss.NewStyle().Bold(true).Foreground(blue).Render(title))
	output.WriteString("\n")
	output.WriteString(lipgloss.NewStyle().Foreground(gray).Render(path))

	// Always reserve a line for search bar to prevent jumping
	output.WriteString("\n")
	if m.searchMode {
		searchPrompt := "/" + m.searchQuery + "_"
		output.WriteString(lipgloss.NewStyle().Foreground(blue).Render("Search: " + searchPrompt))
		if len(m.filteredIndices) > 0 {
			output.WriteString(" ")
			output.WriteString(lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("(%d matches)", len(m.filteredIndices))))
		}
	} else if m.numberInput != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(blue).Render("Select: " + m.numberInput + "_"))
	}

	output.WriteString("\n")

	// Handle loading
	if m.loading {
		spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinner := spinnerChars[m.spinnerIndex%len(spinnerChars)]
		loadingMsg := spinner + " Loading..."
		output.WriteString(lipgloss.NewStyle().Foreground(blue).Render(loadingMsg))
		output.WriteString("\n\n")
		output.WriteString(m.getHelpText())
		return output.String()
	}

	// Handle errors
	if m.error != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(red).Render("Error: " + m.error))
		output.WriteString("\n\n")
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("Press 'q' to quit"))
		return output.String()
	}

	// Render content based on view
	switch m.currentView {
	case viewDatabases:
		m.renderDatabases(&output, blue, gray, white)
	case viewSchemas:
		m.renderSchemas(&output, blue, gray, white)
	case viewTables:
		m.renderTables(&output, blue, gray, white)
	case viewFields:
		m.renderFields(&output, blue, gray, white)
	}

	output.WriteString("\n")
	output.WriteString(m.getHelpText())

	return output.String()
}

func (m *model) updateSearch() {
	// Only filter if we have actual search query content
	if !m.searchMode || m.searchQuery == "" {
		m.filteredIndices = nil
		return
	}

	m.filteredIndices = nil

	switch m.currentView {
	case viewDatabases:
		var names []string
		for _, db := range m.databases {
			names = append(names, db.Name)
		}
		matches := fuzzy.Find(m.searchQuery, names)
		for _, match := range matches {
			m.filteredIndices = append(m.filteredIndices, match.Index)
		}
	case viewSchemas:
		var names []string
		for _, schema := range m.schemas {
			names = append(names, schema.Name)
		}
		matches := fuzzy.Find(m.searchQuery, names)
		for _, match := range matches {
			m.filteredIndices = append(m.filteredIndices, match.Index)
		}
	case viewTables:
		var names []string
		for _, table := range m.tables {
			name := table.DisplayName
			if name == "" {
				name = table.Name
			}
			names = append(names, name)
		}
		matches := fuzzy.Find(m.searchQuery, names)
		for _, match := range matches {
			m.filteredIndices = append(m.filteredIndices, match.Index)
		}
	case viewFields:
		var names []string
		for _, field := range m.fields {
			name := field.DisplayName
			if name == "" {
				name = field.Name
			}
			names = append(names, name)
		}
		matches := fuzzy.Find(m.searchQuery, names)
		for _, match := range matches {
			m.filteredIndices = append(m.filteredIndices, match.Index)
		}
	}

	// Reset cursor when search results change
	m.cursor = 0
}

func (m model) getHelpText() string {
	gray := lipgloss.Color("240")
	blue := lipgloss.Color("12")

	keyStyle := lipgloss.NewStyle().Foreground(blue)
	descStyle := lipgloss.NewStyle().Foreground(gray)

	if m.searchMode {
		return keyStyle.Render("esc") + descStyle.Render(" cancel  ") +
			keyStyle.Render("enter") + descStyle.Render(" select  ") +
			keyStyle.Render("↑↓") + descStyle.Render(" navigate")
	} else {
		var help strings.Builder

		// Navigation section - combine all arrows
		var navigation strings.Builder
		if m.currentView != viewDatabases {
			navigation.WriteString(keyStyle.Render("↑↓←→"))
			navigation.WriteString(descStyle.Render(" navigate  "))
		} else {
			navigation.WriteString(keyStyle.Render("↑↓→"))
			navigation.WriteString(descStyle.Render(" navigate  "))
		}

		// Quick select (context-aware)
		var itemCount int
		switch m.currentView {
		case viewDatabases:
			itemCount = len(m.databases)
		case viewSchemas:
			itemCount = len(m.schemas)
		case viewTables:
			itemCount = len(m.tables)
		case viewFields:
			itemCount = len(m.fields)
		}

		if m.currentView != viewFields && itemCount > 0 {
			if itemCount < 10 {
				navigation.WriteString(keyStyle.Render("1-9"))
			} else {
				navigation.WriteString(keyStyle.Render("01-99"))
			}
			navigation.WriteString(descStyle.Render(" select"))
		}

		// Actions section
		var actions strings.Builder
		actions.WriteString(keyStyle.Render("w"))
		actions.WriteString(descStyle.Render(" web  "))
		actions.WriteString(keyStyle.Render("/"))
		actions.WriteString(descStyle.Render(" search  "))
		actions.WriteString(keyStyle.Render("?"))
		actions.WriteString(descStyle.Render(" help  "))
		actions.WriteString(keyStyle.Render("q"))
		actions.WriteString(descStyle.Render(" quit"))

		// Combine sections on separate lines
		help.WriteString(navigation.String())
		help.WriteString("\n")
		help.WriteString(actions.String())

		// Add update notification if available
		if m.updateAvailable {
			help.WriteString("\n")
			updateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
			help.WriteString(updateStyle.Render("⚠ Update available: "))
			help.WriteString(updateStyle.Render(m.latestVersion))
			help.WriteString(descStyle.Render(" - Run: "))
			help.WriteString(keyStyle.Render("curl -sSL https://raw.githubusercontent.com/amureki/metabase-explorer/main/install.sh | bash"))
		}

		return help.String()
	}
}

func (m model) renderDatabases(output *strings.Builder, blue, gray, white lipgloss.Color) {
	if len(m.databases) == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("No databases found"))
		return
	}

	// Show filtered or all databases
	var itemsToShow []int

	if m.searchMode && m.searchQuery != "" && len(m.filteredIndices) > 0 {
		itemsToShow = m.filteredIndices
	} else if m.searchMode && m.searchQuery != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("No matches found"))
		return
	} else {
		for i := range m.databases {
			itemsToShow = append(itemsToShow, i)
		}
	}

	for i, dbIndex := range itemsToShow {
		db := m.databases[dbIndex]
		var numberPrefix string
		if len(m.databases) < 10 {
			numberPrefix = lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("%d ", i+1))
		} else {
			numberPrefix = lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("%02d ", i+1))
		}

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("▶ " + db.Name))
			output.WriteString(" ")
			output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("(" + db.Engine + ")"))
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + db.Name + " ")
			output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("(" + db.Engine + ")"))
		}
		output.WriteString("\n")
	}
}

func (m model) renderSchemas(output *strings.Builder, blue, gray, white lipgloss.Color) {
	if len(m.schemas) == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("No schemas found"))
		return
	}

	// Show filtered or all schemas
	var itemsToShow []int

	if m.searchMode && m.searchQuery != "" && len(m.filteredIndices) > 0 {
		itemsToShow = m.filteredIndices
	} else if m.searchMode && m.searchQuery != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("No matches found"))
		return
	} else {
		for i := range m.schemas {
			itemsToShow = append(itemsToShow, i)
		}
	}

	for i, schemaIndex := range itemsToShow {
		schema := m.schemas[schemaIndex]
		var numberPrefix string
		if len(m.schemas) < 10 {
			numberPrefix = lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("%d ", i+1))
		} else {
			numberPrefix = lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("%02d ", i+1))
		}

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("▶ " + schema.Name))
			output.WriteString(" ")
			output.WriteString(lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("(%d tables)", schema.TableCount)))
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + schema.Name + " ")
			output.WriteString(lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("(%d tables)", schema.TableCount)))
		}
		output.WriteString("\n")
	}
}

func (m model) renderTables(output *strings.Builder, blue, gray, white lipgloss.Color) {
	if len(m.tables) == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("No tables found"))
		return
	}

	// Show filtered or all tables
	var itemsToShow []int

	if m.searchMode && m.searchQuery != "" && len(m.filteredIndices) > 0 {
		itemsToShow = m.filteredIndices
	} else if m.searchMode && m.searchQuery != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("No matches found"))
		return
	} else {
		for i := range m.tables {
			itemsToShow = append(itemsToShow, i)
		}
	}

	for i, tableIndex := range itemsToShow {
		table := m.tables[tableIndex]
		name := table.DisplayName
		if name == "" {
			name = table.Name
		}

		var numberPrefix string
		if len(m.tables) < 10 {
			numberPrefix = lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("%d ", i+1))
		} else {
			numberPrefix = lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("%02d ", i+1))
		}

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("▶ " + name))
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + name)
		}

		output.WriteString("\n")
	}

}

func (m model) renderFields(output *strings.Builder, blue, gray, white lipgloss.Color) {
	if len(m.fields) == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("No fields found"))
		return
	}

	// Show filtered or all fields
	var itemsToShow []int

	if m.searchMode && m.searchQuery != "" && len(m.filteredIndices) > 0 {
		itemsToShow = m.filteredIndices
	} else if m.searchMode && m.searchQuery != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("No matches found"))
		return
	} else {
		for i := range m.fields {
			itemsToShow = append(itemsToShow, i)
		}
	}

	for i, fieldIndex := range itemsToShow {
		field := m.fields[fieldIndex]
		name := field.DisplayName
		if name == "" {
			name = field.Name
		}

		numberPrefix := lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("%02d ", i+1))

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("▶ " + name))
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + name)
		}

		// Add type info
		if field.DatabaseType != "" {
			output.WriteString(" ")
			output.WriteString(lipgloss.NewStyle().Foreground(gray).Render(field.DatabaseType))
		}

		if field.SemanticType != "" {
			output.WriteString(" ")
			output.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("[" + field.SemanticType + "]"))
		}

		output.WriteString("\n")
	}

}

func (m model) renderHelpOverlay(output *strings.Builder, blue, gray, white lipgloss.Color) string {
	// Title and copyright
	output.WriteString(lipgloss.NewStyle().Bold(true).Foreground(blue).Render("Metabase Explorer"))
	output.WriteString("\n")
	output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("Copyright 2025 Rust Saiargaliev"))
	output.WriteString("\n\n")

	// Repository info
	output.WriteString(lipgloss.NewStyle().Bold(true).Foreground(blue).Render("Links"))
	output.WriteString("\n")

	// Repository link
	if m.helpCursor == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("▶ Repository: "))
		output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("https://github.com/amureki/metabase-explorer"))
	} else {
		output.WriteString(lipgloss.NewStyle().Foreground(white).Render("  Repository: "))
		output.WriteString(lipgloss.NewStyle().Foreground(blue).Render("https://github.com/amureki/metabase-explorer"))
	}
	output.WriteString("\n")

	// Issues link
	if m.helpCursor == 1 {
		output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("▶ Issues:     "))
		output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("https://github.com/amureki/metabase-explorer/issues"))
	} else {
		output.WriteString(lipgloss.NewStyle().Foreground(white).Render("  Issues:     "))
		output.WriteString(lipgloss.NewStyle().Foreground(blue).Render("https://github.com/amureki/metabase-explorer/issues"))
	}
	output.WriteString("\n")

	// Sponsor link
	if m.helpCursor == 2 {
		output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("▶ Sponsor:    "))
		output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("https://github.com/sponsors/amureki"))
	} else {
		output.WriteString(lipgloss.NewStyle().Foreground(white).Render("  Sponsor:    "))
		output.WriteString(lipgloss.NewStyle().Foreground(blue).Render("https://github.com/sponsors/amureki"))
	}
	output.WriteString("\n\n")

	// ASCII text logo
	logo := " __    __     ______     ______   ______     ______     ______     ______     ______    \n" +
		"/\\ \"-./  \\   /\\  ___\\   /\\__  _\\ /\\  __ \\   /\\  == \\   /\\  __ \\   /\\  ___\\   /\\  ___\\   \n" +
		"\\ \\ \\-./\\ \\  \\ \\  __\\   \\/_/\\ \\/ \\ \\  __ \\  \\ \\  __<   \\ \\  __ \\  \\ \\___  \\  \\ \\  __\\   \n" +
		" \\ \\_\\ \\ \\_\\  \\ \\_____\\    \\ \\_\\  \\ \\_\\ \\_\\  \\ \\_____\\  \\ \\_\\ \\_\\  \\/\\_____\\  \\ \\_____\\ \n" +
		"  \\/_/  \\/_/   \\/_____/     \\/_/   \\/_/\\/_/   \\/_____/   \\/_/\\/_/   \\/_____/   \\/_____/ \n" +
		"                                                                                        \n" +
		" ______     __  __     ______   __         ______     ______     ______     ______      \n" +
		"/\\  ___\\   /\\_\\_\\_\\   /\\  == \\ /\\ \\       /\\  __ \\   /\\  == \\   /\\  ___\\   /\\  == \\     \n" +
		"\\ \\  __\\   \\/_/\\_\\/_  \\ \\  _-/ \\ \\ \\____  \\ \\ \\/\\ \\  \\ \\  __<   \\ \\  __\\   \\ \\  __<     \n" +
		" \\ \\_____\\   /\\_\\/\\_\\  \\ \\_\\    \\ \\_____\\  \\ \\_____\\  \\ \\_\\ \\_\\  \\ \\_____\\  \\ \\_\\ \\_\\   \n" +
		"  \\/_____/   \\/_/\\/_/   \\/_/     \\/_____/   \\/_____/   \\/_/ /_/   \\/_____/   \\/_/ /_/   \n" +
		"                                                                                        "
	output.WriteString(lipgloss.NewStyle().Foreground(blue).Render(logo))
	output.WriteString("\n\n")

	output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("Use ↑↓ to navigate, Enter to open link, ? or esc to close"))

	return output.String()
}

func printHelp() {
	fmt.Printf(`mbx - Metabase Explorer %s

A Terminal User Interface for exploring Metabase database metadata.

USAGE:
    mbx [OPTIONS]

OPTIONS:
    -h, --help     Show this help message
    -v, --version  Show version information

CONFIGURATION:
    Create a .env file in your current directory with:

    METABASE_URL="https://your-metabase-instance.com/"
    METABASE_API_TOKEN="your-api-token-here"

    Get your API token from Metabase Admin Settings → API Keys

For more information, visit: https://github.com/amureki/metabase-explorer
`, version)
}

func main() {
	var showVersion = flag.Bool("version", false, "Show version information")
	var showVersionShort = flag.Bool("v", false, "Show version information")
	var showHelp = flag.Bool("help", false, "Show help information")
	var showHelpShort = flag.Bool("h", false, "Show help information")

	flag.Parse()

	if *showVersion || *showVersionShort {
		fmt.Printf("mbx version %s\n", version)
		return
	}

	if *showHelp || *showHelpShort {
		printHelp()
		return
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
