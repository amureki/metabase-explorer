package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

func (m *Model) updateSearch() {
	// Only filter if we have actual search query content
	if !m.searchMode || m.searchQuery == "" {
		m.filteredIndices = nil
		return
	}

	m.filteredIndices = nil

	switch m.currentView {
	case viewMainMenu:
		// No search for main menu
		return
	case viewDatabases:
		var names []string
		for _, db := range m.databases {
			names = append(names, db.Name)
		}
		matches := fuzzy.Find(m.searchQuery, names)
		for _, match := range matches {
			m.filteredIndices = append(m.filteredIndices, match.Index)
		}
	case viewCollections:
		var names []string
		for _, collection := range m.collections {
			names = append(names, collection.Name)
		}
		matches := fuzzy.Find(m.searchQuery, names)
		for _, match := range matches {
			m.filteredIndices = append(m.filteredIndices, match.Index)
		}
	case viewCollectionItems:
		var names []string
		for _, item := range m.collectionItems {
			names = append(names, item.Name)
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

func (m Model) getWebURL() string {
	baseURL := strings.TrimSuffix(m.client.BaseURL, "/")

	switch m.currentView {
	case viewMainMenu:
		return baseURL
	case viewDatabases:
		if len(m.databases) > 0 && m.cursor < len(m.databases) {
			db := m.databases[m.cursor]
			return fmt.Sprintf("%s/browse/databases/%d", baseURL, db.ID)
		}
	case viewCollections:
		if len(m.collections) > 0 && m.cursor < len(m.collections) {
			collection := m.collections[m.cursor]
			return fmt.Sprintf("%s/collection/%v", baseURL, collection.ID)
		}
	case viewCollectionItems:
		if len(m.collectionItems) > 0 && m.cursor < len(m.collectionItems) {
			item := m.collectionItems[m.cursor]
			switch item.Model {
			case "card":
				return fmt.Sprintf("%s/question/%d", baseURL, item.ID)
			case "dashboard":
				return fmt.Sprintf("%s/dashboard/%d", baseURL, item.ID)
			case "collection":
				return fmt.Sprintf("%s/collection/%d", baseURL, item.ID)
			default:
				return fmt.Sprintf("%s/collection/%v", baseURL, m.selectedCollection.ID)
			}
		} else if m.selectedCollection != nil {
			return fmt.Sprintf("%s/collection/%v", baseURL, m.selectedCollection.ID)
		}
	case viewSchemas:
		if len(m.schemas) > 0 && m.cursor < len(m.schemas) && m.selectedDatabase != nil {
			// Open the specific schema browse page
			schema := m.schemas[m.cursor]
			return fmt.Sprintf("%s/browse/databases/%d/schema/%s", baseURL, m.selectedDatabase.ID, schema.Name)
		} else if m.selectedDatabase != nil {
			db := m.selectedDatabase
			return fmt.Sprintf("%s/browse/databases/%d", baseURL, db.ID)
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
	case viewItemDetail:
		if m.selectedItem != nil {
			switch m.selectedItem.Model {
			case "card":
				return fmt.Sprintf("%s/question/%d", baseURL, m.selectedItem.ID)
			case "dashboard":
				return fmt.Sprintf("%s/dashboard/%d", baseURL, m.selectedItem.ID)
			case "collection":
				return fmt.Sprintf("%s/collection/%d", baseURL, m.selectedItem.ID)
			default:
				// Fallback to the current collection
				if m.selectedCollection != nil {
					return fmt.Sprintf("%s/collection/%v", baseURL, m.selectedCollection.ID)
				}
			}
		}
	}

	return baseURL
}

func (m Model) View() string {
	var output strings.Builder

	// Handle help mode first - return immediately without showing main content
	if m.helpMode {
		return m.renderHelpOverlay(&output)
	}

	// Header
	title := ""
	path := ""

	switch m.currentView {
	case viewMainMenu:
		title = fmt.Sprintf("Metabase Explorer %s", m.Version)
		path = "Main Menu"
	case viewDatabases:
		title = fmt.Sprintf("Metabase Explorer %s | Databases", m.Version)
		if len(m.databases) > 0 {
			path = fmt.Sprintf("Databases (%d)", len(m.databases))
		} else {
			path = "Databases"
		}
	case viewCollections:
		title = fmt.Sprintf("Metabase Explorer %s | Collections", m.Version)
		if len(m.collections) > 0 {
			path = fmt.Sprintf("Collections (%d)", len(m.collections))
		} else {
			path = "Collections"
		}
	case viewCollectionItems:
		title = fmt.Sprintf("Metabase Explorer %s | Collection items", m.Version)
		// Build breadcrumb path showing collection hierarchy
		var pathParts []string
		pathParts = append(pathParts, "Collections")
		for _, collection := range m.collectionStack {
			pathParts = append(pathParts, collection.Name)
		}
		pathParts = append(pathParts, m.selectedCollection.Name)
		
		if len(m.collectionItems) > 0 {
			path = fmt.Sprintf("%s (%d)", strings.Join(pathParts, " > "), len(m.collectionItems))
		} else {
			path = strings.Join(pathParts, " > ")
		}
	case viewItemDetail:
		title = fmt.Sprintf("Metabase Explorer %s | Item Details", m.Version)
		// Build breadcrumb path showing collection hierarchy with item name
		var pathParts []string
		pathParts = append(pathParts, "Collections")
		for _, collection := range m.collectionStack {
			pathParts = append(pathParts, collection.Name)
		}
		pathParts = append(pathParts, m.selectedCollection.Name)
		pathParts = append(pathParts, m.selectedItem.Name)
		path = strings.Join(pathParts, " > ")
	case viewSchemas:
		title = fmt.Sprintf("Metabase Explorer %s | Database schemas", m.Version)
		if len(m.schemas) > 0 {
			path = fmt.Sprintf("Databases > %s (%d)", m.selectedDatabase.Name, len(m.schemas))
		} else {
			path = fmt.Sprintf("Databases > %s", m.selectedDatabase.Name)
		}
	case viewTables:
		title = fmt.Sprintf("Metabase Explorer %s | Schema tables", m.Version)
		if len(m.tables) > 0 {
			path = fmt.Sprintf("Databases > %s > %s (%d)", m.selectedDatabase.Name, m.selectedSchema.Name, len(m.tables))
		} else {
			path = fmt.Sprintf("Databases > %s > %s", m.selectedDatabase.Name, m.selectedSchema.Name)
		}
	case viewFields:
		title = fmt.Sprintf("Metabase Explorer %s | Table fields", m.Version)
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

	output.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(title))
	output.WriteString("\n")
	output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(path))

	// Always reserve a line for search bar to prevent jumping
	output.WriteString("\n")
	if m.searchMode {
		searchPrompt := "/" + m.searchQuery + "_"
		output.WriteString(lipgloss.NewStyle().Foreground(ColorInfo).Render("Search: " + searchPrompt))
		if len(m.filteredIndices) > 0 {
			output.WriteString(" ")
			output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("(%d matches)", len(m.filteredIndices))))
		}
	} else if m.numberInput != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorInfo).Render("Select: " + m.numberInput + "_"))
	}

	output.WriteString("\n")

	// Handle loading
	if m.loading {
		spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinner := spinnerChars[m.spinnerIndex%len(spinnerChars)]
		loadingMsg := spinner + " Loading..."
		output.WriteString(lipgloss.NewStyle().Foreground(ColorInfo).Render(loadingMsg))
		output.WriteString("\n\n")
		output.WriteString(m.getHelpText())
		return output.String()
	}

	// Handle errors
	if m.error != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorError).Render("Error: " + m.error))
		output.WriteString("\n\n")
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("Press 'q' to quit"))
		return output.String()
	}

	// Render content based on view
	switch m.currentView {
	case viewMainMenu:
		m.renderMainMenu(&output)
	case viewDatabases:
		m.renderDatabases(&output)
	case viewCollections:
		m.renderCollections(&output)
	case viewCollectionItems:
		m.renderCollectionItems(&output)
	case viewItemDetail:
		m.renderItemDetail(&output)
	case viewSchemas:
		m.renderSchemas(&output)
	case viewTables:
		m.renderTables(&output)
	case viewFields:
		m.renderFields(&output)
	}

	output.WriteString("\n")
	output.WriteString(m.getHelpText())

	return output.String()
}

func (m Model) getHelpText() string {
	keyStyle := lipgloss.NewStyle().Foreground(ColorHighlight)
	descStyle := lipgloss.NewStyle().Foreground(ColorMuted)

	if m.searchMode {
		return keyStyle.Render("esc") + descStyle.Render(" cancel  ") +
			keyStyle.Render("enter") + descStyle.Render(" select  ") +
			keyStyle.Render("↑↓") + descStyle.Render(" navigate")
	} else {
		var help strings.Builder

		// Navigation section - combine all arrows
		var navigation strings.Builder
		if m.currentView == viewMainMenu {
			navigation.WriteString(keyStyle.Render("↑↓→"))
			navigation.WriteString(descStyle.Render(" navigate  "))
		} else if m.currentView == viewDatabases || m.currentView == viewCollections {
			navigation.WriteString(keyStyle.Render("↑↓←→"))
			navigation.WriteString(descStyle.Render(" navigate  "))
		} else {
			navigation.WriteString(keyStyle.Render("↑↓←→"))
			navigation.WriteString(descStyle.Render(" navigate  "))
		}

		// Quick select (context-aware)
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
			updateStyle := lipgloss.NewStyle().Foreground(ColorWarning)
			help.WriteString(updateStyle.Render("⚠ Update available: "))
			help.WriteString(updateStyle.Render(m.latestVersion))
			help.WriteString(descStyle.Render(" - Run: "))
			help.WriteString(keyStyle.Render("mbx update"))
		}

		return help.String()
	}
}

func (m Model) renderDatabases(output *strings.Builder) {
	if len(m.databases) == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No databases found"))
		return
	}

	// Show filtered or all databases
	var itemsToShow []int

	if m.searchMode && m.searchQuery != "" && len(m.filteredIndices) > 0 {
		itemsToShow = m.filteredIndices
	} else if m.searchMode && m.searchQuery != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No matches found"))
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
			numberPrefix = lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("%d ", i+1))
		} else {
			numberPrefix = lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("%02d ", i+1))
		}

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("▶ " + db.Name))
			output.WriteString(" ")
			output.WriteString(lipgloss.NewStyle().Foreground(getItemTypeColor("database")).Render("(" + db.Engine + ")"))
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + db.Name + " ")
			output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("(" + db.Engine + ")"))
		}
		output.WriteString("\n")
	}
}

func (m Model) renderSchemas(output *strings.Builder) {
	if len(m.schemas) == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No schemas found"))
		return
	}

	// Show filtered or all schemas
	var itemsToShow []int

	if m.searchMode && m.searchQuery != "" && len(m.filteredIndices) > 0 {
		itemsToShow = m.filteredIndices
	} else if m.searchMode && m.searchQuery != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No matches found"))
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
			numberPrefix = lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("%d ", i+1))
		} else {
			numberPrefix = lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("%02d ", i+1))
		}

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("▶ " + schema.Name))
			output.WriteString(" ")
			output.WriteString(lipgloss.NewStyle().Foreground(ColorInfo).Render(fmt.Sprintf("(%d tables)", schema.TableCount)))
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + schema.Name + " ")
			output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("(%d tables)", schema.TableCount)))
		}
		output.WriteString("\n")
	}
}

func (m Model) renderTables(output *strings.Builder) {
	if len(m.tables) == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No tables found"))
		return
	}

	// Show filtered or all tables
	var itemsToShow []int

	if m.searchMode && m.searchQuery != "" && len(m.filteredIndices) > 0 {
		itemsToShow = m.filteredIndices
	} else if m.searchMode && m.searchQuery != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No matches found"))
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
			numberPrefix = lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("%d ", i+1))
		} else {
			numberPrefix = lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("%02d ", i+1))
		}

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("▶ " + name))
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + name)
		}

		output.WriteString("\n")
	}

}

func (m Model) renderFields(output *strings.Builder) {
	if len(m.fields) == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No fields found"))
		return
	}

	// Show filtered or all fields
	var itemsToShow []int

	if m.searchMode && m.searchQuery != "" && len(m.filteredIndices) > 0 {
		itemsToShow = m.filteredIndices
	} else if m.searchMode && m.searchQuery != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No matches found"))
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

		numberPrefix := lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("%02d ", i+1))

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("▶ " + name))
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + name)
		}

		// Add type info
		if field.DatabaseType != "" {
			output.WriteString(" ")
			output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(field.DatabaseType))
		}

		if field.SemanticType != "" {
			output.WriteString(" ")
			color := getSemanticTypeColor(field.SemanticType)
			output.WriteString(lipgloss.NewStyle().Foreground(color).Render("[" + field.SemanticType + "]"))
		}

		output.WriteString("\n")
	}

}

func (m Model) renderHelpOverlay(output *strings.Builder) string {
	// Title and copyright
	output.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(fmt.Sprintf("Metabase Explorer %s | About", m.Version)))
	output.WriteString("\n")
	output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("Copyright 2025 Rust Saiargaliev"))
	output.WriteString("\n\n")

	// Repository info
	output.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("Links"))
	output.WriteString("\n")

	// Repository link
	if m.helpCursor == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("▶ Repository: "))
		output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("https://github.com/amureki/metabase-explorer"))
	} else {
		output.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Render("  Repository: "))
		output.WriteString(lipgloss.NewStyle().Foreground(ColorInfo).Render("https://github.com/amureki/metabase-explorer"))
	}
	output.WriteString("\n")

	// Issues link
	if m.helpCursor == 1 {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("▶ Issues:     "))
		output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("https://github.com/amureki/metabase-explorer/issues"))
	} else {
		output.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Render("  Issues:     "))
		output.WriteString(lipgloss.NewStyle().Foreground(ColorInfo).Render("https://github.com/amureki/metabase-explorer/issues"))
	}
	output.WriteString("\n")

	// Sponsor link
	if m.helpCursor == 2 {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("▶ Sponsor:    "))
		output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("https://github.com/sponsors/amureki"))
	} else {
		output.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Render("  Sponsor:    "))
		output.WriteString(lipgloss.NewStyle().Foreground(ColorInfo).Render("https://github.com/sponsors/amureki"))
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
	output.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Render(logo))
	output.WriteString("\n\n")

	output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("Use ↑↓ to navigate, Enter to open link, ? or esc to close"))

	return output.String()
}

func (m Model) renderMainMenu(output *strings.Builder) {
	options := []string{"Collections", "Databases"}

	for i, option := range options {
		var numberPrefix string
		numberPrefix = lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("%d ", i+1))

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("▶ " + option))
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + option)
		}
		output.WriteString("\n")
	}
}

func (m Model) renderCollections(output *strings.Builder) {
	if len(m.collections) == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No collections found"))
		return
	}

	// Show filtered or all collections
	var itemsToShow []int

	if m.searchMode && m.searchQuery != "" && len(m.filteredIndices) > 0 {
		itemsToShow = m.filteredIndices
	} else if m.searchMode && m.searchQuery != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No matches found"))
		return
	} else {
		for i := range m.collections {
			itemsToShow = append(itemsToShow, i)
		}
	}

	for i, collectionIndex := range itemsToShow {
		collection := m.collections[collectionIndex]
		var numberPrefix string
		if len(m.collections) < 10 {
			numberPrefix = lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("%d ", i+1))
		} else {
			numberPrefix = lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("%02d ", i+1))
		}

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("▶ " + collection.Name))
			if collection.Description != "" {
				output.WriteString(" ")
				output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("(" + collection.Description + ")"))
			}
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + collection.Name)
			if collection.Description != "" {
				output.WriteString(" ")
				output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("(" + collection.Description + ")"))
			}
		}
		output.WriteString("\n")
	}
}

func (m Model) renderCollectionItems(output *strings.Builder) {
	if len(m.collectionItems) == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No items found in this collection"))
		return
	}

	// Show filtered or all collection items
	var itemsToShow []int

	if m.searchMode && m.searchQuery != "" && len(m.filteredIndices) > 0 {
		itemsToShow = m.filteredIndices
	} else if m.searchMode && m.searchQuery != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No matches found"))
		return
	} else {
		for i := range m.collectionItems {
			itemsToShow = append(itemsToShow, i)
		}
	}

	// Apply viewport limiting for large lists
	viewportEnd := m.viewportStart + m.viewportHeight
	if viewportEnd > len(itemsToShow) {
		viewportEnd = len(itemsToShow)
	}
	// Show top pagination indicator when pagination is needed
	if len(itemsToShow) > m.viewportHeight {
		var prefix string
		if len(m.collectionItems) < 10 {
			prefix = "  "  // 2 chars for single digits
		} else {
			prefix = "   " // 3 chars for double digits  
		}
		prefix += "  " // 2 more chars to align with item names (after ▶ or spaces)
		
		if m.viewportStart > 0 {
			output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("↑" + prefix[1:] + "... " + fmt.Sprintf("%d-%d of %d items", m.viewportStart+1, viewportEnd, len(itemsToShow))))
		} else {
			output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(prefix + "... " + fmt.Sprintf("%d-%d of %d items", m.viewportStart+1, viewportEnd, len(itemsToShow))))
		}
		output.WriteString("\n")
	}

	for i := m.viewportStart; i < viewportEnd; i++ {
		itemIndex := itemsToShow[i]
		item := m.collectionItems[itemIndex]
		var numberPrefix string
		if len(m.collectionItems) < 10 {
			numberPrefix = lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("%d ", i+1))
		} else {
			numberPrefix = lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf("%02d ", i+1))
		}

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(ColorSelected).Bold(true).Render("▶ " + item.Name))
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + item.Name)
		}

		// Add type info
		if item.Model != "" {
			output.WriteString(" ")
			typeColor := getItemTypeColor(item.Model)
			output.WriteString(lipgloss.NewStyle().Foreground(typeColor).Render("[" + item.Model + "]"))
		}


		output.WriteString("\n")
	}
	// Show bottom pagination indicator when pagination is needed
	if len(itemsToShow) > m.viewportHeight {
		var prefix string
		if len(m.collectionItems) < 10 {
			prefix = "  "  // 2 chars for single digits
		} else {
			prefix = "   " // 3 chars for double digits  
		}
		prefix += "  " // 2 more chars to align with item names (after ▶ or spaces)
		
		if viewportEnd < len(itemsToShow) {
			output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("↓" + prefix[1:] + "... " + fmt.Sprintf("%d-%d of %d items", m.viewportStart+1, viewportEnd, len(itemsToShow))))
		} else {
			output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(prefix + "... " + fmt.Sprintf("%d-%d of %d items", m.viewportStart+1, viewportEnd, len(itemsToShow))))
		}
		output.WriteString("\n")
	}
}

func (m Model) renderItemDetail(output *strings.Builder) {
	if m.selectedItem == nil {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No item selected"))
		return
	}

	item := m.selectedItem

	// Item Name (title)
	output.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(item.Name))
	output.WriteString("\n\n")

	// Item Description
	if item.Description != "" {
		output.WriteString(lipgloss.NewStyle().Bold(true).Render("Description:"))
		output.WriteString("\n")
		// Wrap description text to fit terminal width (conservative width with margin)
		wrappedDesc := lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Width(80).
			Render(item.Description)
		output.WriteString(wrappedDesc)
		output.WriteString("\n\n")
	} else {
		output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("No description available"))
		output.WriteString("\n\n")
	}

	// Show detailed metadata if available (from detail API)
	if m.itemDetail != nil {
		if creator := m.itemDetail.GetCreator(); creator != nil {
			output.WriteString(lipgloss.NewStyle().Bold(true).Render("Created by: "))
			creatorName := fmt.Sprintf("%s %s", creator.FirstName, creator.LastName)
			if creatorName == " " {
				creatorName = creator.Email
			}
			output.WriteString(lipgloss.NewStyle().Foreground(ColorInfo).Render(creatorName))
			output.WriteString("\n")
		}

		if lastEditInfo := m.itemDetail.GetLastEditInfo(); lastEditInfo != nil {
			output.WriteString(lipgloss.NewStyle().Bold(true).Render("Last edited by: "))
			editorName := fmt.Sprintf("%s %s", lastEditInfo.FirstName, lastEditInfo.LastName)
			if editorName == " " {
				editorName = lastEditInfo.Email
			}
			output.WriteString(lipgloss.NewStyle().Foreground(ColorInfo).Render(editorName))
			output.WriteString("\n")
		}

		if createdAt := m.itemDetail.GetCreatedAt(); createdAt != "" {
			output.WriteString(lipgloss.NewStyle().Bold(true).Render("Created: "))
			output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(m.formatTimestamp(createdAt)))
			output.WriteString("\n")
		}

		if updatedAt := m.itemDetail.GetUpdatedAt(); updatedAt != "" {
			output.WriteString(lipgloss.NewStyle().Bold(true).Render("Updated: "))
			output.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(m.formatTimestamp(updatedAt)))
			output.WriteString("\n")
		}

		output.WriteString("\n")
	}

	// Archived status
	if item.Archived {
		output.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("⚠ This item is archived"))
	}
}

func (m Model) formatTimestamp(timestamp string) string {
	if timestamp == "" {
		return ""
	}

	// Parse the timestamp (assuming ISO 8601 format)
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		// Try alternative format if RFC3339 fails
		t, err = time.Parse("2006-01-02T15:04:05.000000Z", timestamp)
		if err != nil {
			return timestamp // Return as-is if parsing fails
		}
	}

	// Format as a human-readable date
	return t.Format("Jan 2, 2006 at 3:04 PM")
}
