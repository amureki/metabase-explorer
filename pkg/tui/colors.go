package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("4")
	ColorSecondary = lipgloss.Color("5")
	ColorMuted     = lipgloss.Color("8")
	
	ColorSuccess   = lipgloss.Color("2")
	ColorWarning   = lipgloss.Color("3")
	ColorError     = lipgloss.Color("1")
	ColorInfo      = lipgloss.Color("6")
	
	ColorHighlight = lipgloss.Color("12")
	ColorSelected  = lipgloss.Color("15")
	
	ColorString    = lipgloss.Color("10")
	ColorNumber    = lipgloss.Color("11")
	ColorBoolean   = lipgloss.Color("13")
	ColorDate      = lipgloss.Color("14")
)

func getItemTypeColor(itemType string) lipgloss.Color {
	switch itemType {
	case "database":
		return ColorSuccess
	case "collection":
		return ColorSecondary
	case "card":
		return ColorPrimary
	case "dashboard":
		return ColorWarning
	default:
		return ColorInfo
	}
}

func getSemanticTypeColor(semanticType string) lipgloss.Color {
	if semanticType == "" {
		return ColorMuted
	}
	return ColorInfo
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr ||
		      indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}