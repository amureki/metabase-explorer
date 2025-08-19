package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/amureki/metabase-explorer/pkg/api"
	"github.com/amureki/metabase-explorer/pkg/config"
	"github.com/amureki/metabase-explorer/pkg/util"
	tea "github.com/charmbracelet/bubbletea"
)

type viewState int

const (
	viewDatabases viewState = iota
	viewSchemas
	viewTables
	viewFields
)

type Model struct {
	databases        []api.Database
	schemas          []api.Schema
	tables           []api.Table
	fields           []api.Field
	cursor           int
	loading          bool
	error            string
	client           *api.MetabaseClient
	currentView      viewState
	selectedDatabase *api.Database
	selectedSchema   *api.Schema
	selectedTable    *api.Table
	searchMode       bool
	searchQuery      string
	filteredIndices  []int
	spinnerIndex     int
	numberInput      string
	helpMode         bool
	helpCursor       int
	latestVersion    string
	updateAvailable  bool
	Version          string
}

func InitialModel(flagURL, flagToken, flagProfile, version string) Model {
	metabaseURL, apiToken, err := config.ResolveConfiguration(flagURL, flagToken, flagProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, `Error: %v

Configuration sources (in priority order):
1. CLI flags: --url and --token
2. Config file: ~/.config/mbx/config.yaml (or --config <path>)

To get started, run: mbx init

See https://www.metabase.com/docs/latest/people-and-groups/api-keys for API token setup.
Run 'mbx --help' for more information.
`, err)
		os.Exit(1)
	}

	client := api.NewMetabaseClient(metabaseURL, apiToken)
	return Model{
		loading:     true,
		client:      client,
		currentView: viewDatabases,
		Version:     version,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			err := m.client.TestConnection()
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				if err := util.OpenInBrowser(url); err != nil {
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
				if err := util.OpenInBrowser(url); err != nil {
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
			if err := util.OpenInBrowser(webURL); err != nil {
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
			currentVersion := m.Version
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
