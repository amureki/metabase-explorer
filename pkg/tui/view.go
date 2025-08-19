package tui

import (
	"fmt"
	"strings"

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
	}

	return baseURL
}

func (m Model) View() string {
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
	case viewMainMenu:
		m.renderMainMenu(&output, blue, gray, white)
	case viewDatabases:
		m.renderDatabases(&output, blue, gray, white)
	case viewCollections:
		m.renderCollections(&output, blue, gray, white)
	case viewCollectionItems:
		m.renderCollectionItems(&output, blue, gray, white)
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

func (m Model) getHelpText() string {
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
			updateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
			help.WriteString(updateStyle.Render("⚠ Update available: "))
			help.WriteString(updateStyle.Render(m.latestVersion))
			help.WriteString(descStyle.Render(" - Run: "))
			help.WriteString(keyStyle.Render("mbx update"))
		}

		return help.String()
	}
}

func (m Model) renderDatabases(output *strings.Builder, blue, gray, white lipgloss.Color) {
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

func (m Model) renderSchemas(output *strings.Builder, blue, gray, white lipgloss.Color) {
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

func (m Model) renderTables(output *strings.Builder, blue, gray, white lipgloss.Color) {
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

func (m Model) renderFields(output *strings.Builder, blue, gray, white lipgloss.Color) {
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

func (m Model) renderHelpOverlay(output *strings.Builder, blue, gray, white lipgloss.Color) string {
	// Title and copyright
	output.WriteString(lipgloss.NewStyle().Bold(true).Foreground(blue).Render(fmt.Sprintf("Metabase Explorer %s | About", m.Version)))
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

func (m Model) renderMainMenu(output *strings.Builder, blue, gray, white lipgloss.Color) {
	options := []string{"Collections", "Databases"}

	for i, option := range options {
		var numberPrefix string
		numberPrefix = lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("%d ", i+1))

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("▶ " + option))
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + option)
		}
		output.WriteString("\n")
	}
}

func (m Model) renderCollections(output *strings.Builder, blue, gray, white lipgloss.Color) {
	if len(m.collections) == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("No collections found"))
		return
	}

	// Show filtered or all collections
	var itemsToShow []int

	if m.searchMode && m.searchQuery != "" && len(m.filteredIndices) > 0 {
		itemsToShow = m.filteredIndices
	} else if m.searchMode && m.searchQuery != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("No matches found"))
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
			numberPrefix = lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("%d ", i+1))
		} else {
			numberPrefix = lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("%02d ", i+1))
		}

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("▶ " + collection.Name))
			if collection.Description != "" {
				output.WriteString(" ")
				output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("(" + collection.Description + ")"))
			}
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + collection.Name)
			if collection.Description != "" {
				output.WriteString(" ")
				output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("(" + collection.Description + ")"))
			}
		}
		output.WriteString("\n")
	}
}

func (m Model) renderCollectionItems(output *strings.Builder, blue, gray, white lipgloss.Color) {
	if len(m.collectionItems) == 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("No items found in this collection"))
		return
	}

	// Show filtered or all collection items
	var itemsToShow []int

	if m.searchMode && m.searchQuery != "" && len(m.filteredIndices) > 0 {
		itemsToShow = m.filteredIndices
	} else if m.searchMode && m.searchQuery != "" {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("No matches found"))
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
	
	// Show scroll indicators if there are items outside viewport
	if m.viewportStart > 0 {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("   ↑ ... (showing items " + fmt.Sprintf("%d-%d of %d)", m.viewportStart+1, viewportEnd, len(itemsToShow)) + ")"))
		output.WriteString("\n")
	}

	for i := m.viewportStart; i < viewportEnd; i++ {
		itemIndex := itemsToShow[i]
		item := m.collectionItems[itemIndex]
		var numberPrefix string
		if len(m.collectionItems) < 10 {
			numberPrefix = lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("%d ", i+1))
		} else {
			numberPrefix = lipgloss.NewStyle().Foreground(gray).Render(fmt.Sprintf("%02d ", i+1))
		}

		if i == m.cursor {
			output.WriteString(numberPrefix)
			output.WriteString(lipgloss.NewStyle().Foreground(blue).Bold(true).Render("▶ " + item.Name))
		} else {
			output.WriteString(numberPrefix)
			output.WriteString("  " + item.Name)
		}

		// Add type info
		if item.Model != "" {
			output.WriteString(" ")
			output.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("[" + item.Model + "]"))
		}

		// Add description if available
		if item.Description != "" {
			output.WriteString(" ")
			output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("(" + item.Description + ")"))
		}

		output.WriteString("\n")
	}
	
	// Show bottom scroll indicator if there are more items below
	if viewportEnd < len(itemsToShow) {
		output.WriteString(lipgloss.NewStyle().Foreground(gray).Render("   ↓ ... (showing items " + fmt.Sprintf("%d-%d of %d)", m.viewportStart+1, viewportEnd, len(itemsToShow)) + ")"))
		output.WriteString("\n")
	}
}
