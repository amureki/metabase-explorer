package api

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

type Collection struct {
	ID          interface{} `json:"id"` // Can be int or string ("root")
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Slug        string      `json:"slug"`
	Color       string      `json:"color"`
	Archived    bool        `json:"archived"`
	Location    string      `json:"location"`
	IsPersonal  bool        `json:"is_personal"`
}

type CollectionItem struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Model        string `json:"model"` // "card", "dashboard", "collection", etc.
	CollectionID int    `json:"collection_id"`
	DatabaseID   *int   `json:"database_id"` // Nullable for non-database items
	Archived     bool   `json:"archived"`
}

type SearchResult struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Model       string `json:"model"` // "card", "dashboard", "collection", "table", "database", etc.
	Archived    bool   `json:"archived"`
	Collection  struct {
		ID   interface{} `json:"id"` // Can be int or string ("root")
		Name string      `json:"name"`
	} `json:"collection"`
}
