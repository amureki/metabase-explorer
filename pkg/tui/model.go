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
	viewMainMenu viewState = iota
	viewDatabases
	viewSchemas
	viewTables
	viewFields
	viewCollections
	viewCollectionItems
	viewGlobalSearch
)

type Model struct {
	databases          []api.Database
	schemas            []api.Schema
	tables             []api.Table
	fields             []api.Field
	collections        []api.Collection
	collectionItems    []api.CollectionItem
	searchResults      []api.SearchResult
	globalSearchQuery  string
	cursor             int
	loading            bool
	error              string
	client             *api.MetabaseClient
	currentView        viewState
	selectedDatabase   *api.Database
	selectedSchema     *api.Schema
	selectedTable      *api.Table
	selectedCollection *api.Collection
	collectionStack    []*api.Collection // Track collection hierarchy for proper back navigation
	viewportStart      int               // Starting index for viewport scrolling
	viewportHeight     int               // Number of items that can be displayed at once
	spinnerIndex       int
	numberInput        string
	helpMode           bool
	helpCursor         int
	latestVersion      string
	updateAvailable    bool
	Version            string
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
		loading:     false,
		client:      client,
		currentView: viewMainMenu,
		Version:     version,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			err := m.client.TestConnection()
			if err != nil {
				return connectionTested{err: err}
			}
			return connectionTested{err: nil}
		},
		checkLatestVersion(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
	
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
			// Enter global search mode
			m.currentView = viewGlobalSearch
			m.globalSearchQuery = ""
			m.cursor = 0
			m.loading = false
			m.error = ""
			m.searchResults = nil // Start with empty results
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
			case viewMainMenu:
				itemCount = 2 // Collections and Databases
			case viewDatabases:
				itemCount = len(m.databases)
			case viewCollections:
				itemCount = len(m.collections)
			case viewCollectionItems:
				itemCount = len(m.collectionItems)
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
				// Update viewport for collections and other views that might have many items
				if m.currentView == viewCollectionItems && len(m.collectionItems) > 0 {
					m.updateViewport(len(m.collectionItems))
				}
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
			if m.currentView == viewMainMenu && m.cursor < 1 {
				m.cursor++
			} else if m.currentView == viewDatabases && m.cursor < len(m.databases)-1 {
				m.cursor++
			} else if m.currentView == viewCollections && m.cursor < len(m.collections)-1 {
				m.cursor++
			} else if m.currentView == viewCollectionItems && m.cursor < len(m.collectionItems)-1 {
				m.cursor++
				m.updateViewport(len(m.collectionItems))
			} else if m.currentView == viewSchemas && m.cursor < len(m.schemas)-1 {
				m.cursor++
			} else if m.currentView == viewTables && m.cursor < len(m.tables)-1 {
				m.cursor++
			} else if m.currentView == viewFields && m.cursor < len(m.fields)-1 {
				m.cursor++
			} else if m.currentView == viewGlobalSearch && m.cursor < len(m.searchResults)-1 {
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
			} else if m.currentView == viewDatabases || m.currentView == viewCollections {
				m.currentView = viewMainMenu
				m.cursor = 0
				m.selectedDatabase = nil
				m.databases = nil
				m.collections = nil
			} else if m.currentView == viewCollectionItems {
				if len(m.collectionStack) > 0 {
					// Pop from stack to go to parent collection
					m.selectedCollection = m.collectionStack[len(m.collectionStack)-1]
					m.collectionStack = m.collectionStack[:len(m.collectionStack)-1]
					m.cursor = 0
					m.loading = true
					m.error = ""
					return m, tea.Batch(loadCollectionItems(m.client, m.selectedCollection.ID), tickSpinner())
				} else {
					// Go back to root collections
					m.currentView = viewCollections
					m.cursor = 0
					m.selectedCollection = nil
					m.collectionItems = nil
				}
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
			} else if m.currentView == viewGlobalSearch {
				m.currentView = viewMainMenu
				m.cursor = 0
				m.searchResults = nil
				m.globalSearchQuery = ""
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

			if m.currentView == viewMainMenu {
				if m.cursor == 0 {
					// Navigate to Collections
					m.currentView = viewCollections
					m.cursor = 0
					m.loading = true
					m.error = ""
					return m, tea.Batch(loadCollections(m.client), tickSpinner())
				} else if m.cursor == 1 {
					// Navigate to Databases
					m.currentView = viewDatabases
					m.cursor = 0
					m.loading = true
					m.error = ""
					return m, tea.Batch(loadDatabases(m.client), tickSpinner())
				}
			} else if m.currentView == viewDatabases && len(m.databases) > 0 {
				m.selectedDatabase = &m.databases[m.cursor]
				m.currentView = viewSchemas
				m.cursor = 0
				m.loading = true
				m.error = ""
				return m, tea.Batch(loadSchemas(m.client, m.selectedDatabase.ID), tickSpinner())
			} else if m.currentView == viewCollections && len(m.collections) > 0 {
				m.selectedCollection = &m.collections[m.cursor]
				m.collectionStack = nil // Clear stack when entering from root collections
				m.currentView = viewCollectionItems
				m.cursor = 0
				m.loading = true
				m.error = ""
				return m, tea.Batch(loadCollectionItems(m.client, m.selectedCollection.ID), tickSpinner())
			} else if m.currentView == viewCollectionItems && len(m.collectionItems) > 0 {
				item := m.collectionItems[m.cursor]
				if item.Model == "collection" {
					// Push current collection to stack before drilling into sub-collection
					m.collectionStack = append(m.collectionStack, m.selectedCollection)
					m.selectedCollection = &api.Collection{
						ID:   item.ID,
						Name: item.Name,
					}
					m.currentView = viewCollectionItems
					m.cursor = 0
					m.loading = true
					m.error = ""
					return m, tea.Batch(loadCollectionItems(m.client, item.ID), tickSpinner())
				}
				// For non-collection items (cards, dashboards), do nothing or could open in web
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

			if m.currentView == viewMainMenu {
				if m.cursor == 0 {
					// Navigate to Collections
					m.currentView = viewCollections
					m.cursor = 0
					m.loading = true
					m.error = ""
					return m, tea.Batch(loadCollections(m.client), tickSpinner())
				} else if m.cursor == 1 {
					// Navigate to Databases
					m.currentView = viewDatabases
					m.cursor = 0
					m.loading = true
					m.error = ""
					return m, tea.Batch(loadDatabases(m.client), tickSpinner())
				}
			} else if m.currentView == viewDatabases && len(m.databases) > 0 {
				m.selectedDatabase = &m.databases[m.cursor]
				m.currentView = viewSchemas
				m.cursor = 0
				m.loading = true
				m.error = ""
				return m, tea.Batch(loadSchemas(m.client, m.selectedDatabase.ID), tickSpinner())
			} else if m.currentView == viewCollections && len(m.collections) > 0 {
				m.selectedCollection = &m.collections[m.cursor]
				m.collectionStack = nil // Clear stack when entering from root collections
				m.currentView = viewCollectionItems
				m.cursor = 0
				m.loading = true
				m.error = ""
				return m, tea.Batch(loadCollectionItems(m.client, m.selectedCollection.ID), tickSpinner())
			} else if m.currentView == viewCollectionItems && len(m.collectionItems) > 0 {
				item := m.collectionItems[m.cursor]
				if item.Model == "collection" {
					// Push current collection to stack before drilling into sub-collection
					m.collectionStack = append(m.collectionStack, m.selectedCollection)
					m.selectedCollection = &api.Collection{
						ID:   item.ID,
						Name: item.Name,
					}
					m.currentView = viewCollectionItems
					m.cursor = 0
					m.loading = true
					m.error = ""
					return m, tea.Batch(loadCollectionItems(m.client, item.ID), tickSpinner())
				}
				// For non-collection items (cards, dashboards), do nothing or could open in web
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
			// Handle backspace in global search
			if m.currentView == viewGlobalSearch && len(m.globalSearchQuery) > 0 {
				m.globalSearchQuery = m.globalSearchQuery[:len(m.globalSearchQuery)-1]
				m.cursor = 0
				
				// Only search if query has at least 2 characters, otherwise clear results
				if len(m.globalSearchQuery) >= 2 {
					m.loading = true
					m.error = ""
					return m, tea.Batch(loadGlobalSearch(m.client, m.globalSearchQuery), tickSpinner())
				} else {
					m.searchResults = nil // Clear results for short queries
				}
			}
			// Keep backspace as alternative to left arrow
			if m.numberInput != "" {
				// Clear number input
				m.numberInput = ""
			} else if m.currentView == viewDatabases || m.currentView == viewCollections {
				m.currentView = viewMainMenu
				m.cursor = 0
				m.selectedDatabase = nil
				m.databases = nil
				m.collections = nil
			} else if m.currentView == viewCollectionItems {
				if len(m.collectionStack) > 0 {
					// Pop from stack to go to parent collection
					m.selectedCollection = m.collectionStack[len(m.collectionStack)-1]
					m.collectionStack = m.collectionStack[:len(m.collectionStack)-1]
					m.cursor = 0
					m.loading = true
					m.error = ""
					return m, tea.Batch(loadCollectionItems(m.client, m.selectedCollection.ID), tickSpinner())
				} else {
					// Go back to root collections
					m.currentView = viewCollections
					m.cursor = 0
					m.selectedCollection = nil
					m.collectionItems = nil
				}
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
			} else if m.currentView == viewDatabases || m.currentView == viewCollections {
				m.currentView = viewMainMenu
				m.cursor = 0
				m.selectedDatabase = nil
				m.databases = nil
				m.collections = nil
			} else if m.currentView == viewCollectionItems {
				if len(m.collectionStack) > 0 {
					// Pop from stack to go to parent collection
					m.selectedCollection = m.collectionStack[len(m.collectionStack)-1]
					m.collectionStack = m.collectionStack[:len(m.collectionStack)-1]
					m.cursor = 0
					m.loading = true
					m.error = ""
					return m, tea.Batch(loadCollectionItems(m.client, m.selectedCollection.ID), tickSpinner())
				} else {
					// Go back to root collections
					m.currentView = viewCollections
					m.cursor = 0
					m.selectedCollection = nil
					m.collectionItems = nil
				}
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
		default:
			// Handle typing in global search view
			if m.currentView == viewGlobalSearch && !m.loading {
				if len(msg.String()) == 1 {
					// Add character to global search query
					m.globalSearchQuery += msg.String()
					m.cursor = 0
					
					// Only search if query has at least 2 characters
					if len(m.globalSearchQuery) >= 2 {
						m.loading = true
						m.error = ""
						return m, tea.Batch(loadGlobalSearch(m.client, m.globalSearchQuery), tickSpinner())
					}
				}
			}
		}

	case connectionTested:
		if msg.err != nil {
			m.error = msg.err.Error()
		}

	case databasesLoaded:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err.Error()
		} else {
			m.databases = msg.databases
		}

	case collectionsLoaded:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err.Error()
		} else {
			m.collections = msg.collections
		}

	case collectionItemsLoaded:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err.Error()
		} else {
			m.collectionItems = msg.items
			m.viewportStart = 0 // Reset viewport when loading new items
			if len(m.collectionItems) > 0 {
				m.updateViewport(len(m.collectionItems))
			}
		}

	case globalSearchLoaded:
		m.loading = false
		if msg.err != nil {
			m.error = msg.err.Error()
		} else {
			m.searchResults = msg.results
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

// updateViewport adjusts the viewport to keep the cursor visible
func (m *Model) updateViewport(itemCount int) {
	// Reserve space for header (title + path + search), help text, and some padding
	// Rough estimate: 6 lines for UI elements
	terminalHeight := 25 // Conservative estimate - in real implementation could use tea.WindowSizeMsg
	m.viewportHeight = terminalHeight - 8 // Reserve 8 lines for UI elements
	
	if m.viewportHeight < 5 {
		m.viewportHeight = 5 // Minimum viewport
	}
	
	// Adjust viewport to keep cursor visible
	if m.cursor < m.viewportStart {
		m.viewportStart = m.cursor
	} else if m.cursor >= m.viewportStart+m.viewportHeight {
		m.viewportStart = m.cursor - m.viewportHeight + 1
	}
	
	// Ensure viewport doesn't go beyond bounds
	if m.viewportStart < 0 {
		m.viewportStart = 0
	}
	maxStart := itemCount - m.viewportHeight
	if maxStart < 0 {
		maxStart = 0
	}
	if m.viewportStart > maxStart {
		m.viewportStart = maxStart
	}
}
