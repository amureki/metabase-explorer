package api

type DetailInfo interface {
	GetCreator() *UserInfo
	GetLastEditInfo() *LastEditInfo
	GetCreatedAt() string
	GetUpdatedAt() string
}

type UserInfo struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LastEditInfo struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Timestamp string `json:"timestamp"`
}

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

type CardDetail struct {
	ID               int           `json:"id"`
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	CollectionID     int           `json:"collection_id"`
	DatabaseID       *int          `json:"database_id"`
	Archived         bool          `json:"archived"`
	CreatorID        int           `json:"creator_id"`
	CreatedAt        string        `json:"created_at"`
	UpdatedAt        string        `json:"updated_at"`
	LastEditInfo     *LastEditInfo `json:"last-edit-info"`
	Creator          *UserInfo     `json:"creator"`
}

func (c *CardDetail) GetCreator() *UserInfo        { return c.Creator }
func (c *CardDetail) GetLastEditInfo() *LastEditInfo { return c.LastEditInfo }
func (c *CardDetail) GetCreatedAt() string         { return c.CreatedAt }
func (c *CardDetail) GetUpdatedAt() string         { return c.UpdatedAt }

type DashboardDetail struct {
	ID               int           `json:"id"`
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	CollectionID     int           `json:"collection_id"`
	Archived         bool          `json:"archived"`
	CreatorID        int           `json:"creator_id"`
	CreatedAt        string        `json:"created_at"`
	UpdatedAt        string        `json:"updated_at"`
	LastEditInfo     *LastEditInfo `json:"last-edit-info"`
	Creator          *UserInfo     `json:"creator"`
}

func (d *DashboardDetail) GetCreator() *UserInfo        { return d.Creator }
func (d *DashboardDetail) GetLastEditInfo() *LastEditInfo { return d.LastEditInfo }
func (d *DashboardDetail) GetCreatedAt() string         { return d.CreatedAt }
func (d *DashboardDetail) GetUpdatedAt() string         { return d.UpdatedAt }

type MetricDetail struct {
	ID               int           `json:"id"`
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	CollectionID     int           `json:"collection_id"`
	DatabaseID       *int          `json:"database_id"`
	Archived         bool          `json:"archived"`
	CreatorID        int           `json:"creator_id"`
	CreatedAt        string        `json:"created_at"`
	UpdatedAt        string        `json:"updated_at"`
	LastEditInfo     *LastEditInfo `json:"last-edit-info"`
	Creator          *UserInfo     `json:"creator"`
}

func (m *MetricDetail) GetCreator() *UserInfo        { return m.Creator }
func (m *MetricDetail) GetLastEditInfo() *LastEditInfo { return m.LastEditInfo }
func (m *MetricDetail) GetCreatedAt() string         { return m.CreatedAt }
func (m *MetricDetail) GetUpdatedAt() string         { return m.UpdatedAt }
