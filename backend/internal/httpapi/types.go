package httpapi

import "time"

type Config struct {
	ID              string       `json:"id"`
	Namespace       string       `json:"namespace"`
	Path            string       `json:"path"`
	Format          ConfigFormat `json:"format"`
	LatestVersionID *string      `json:"latest_version_id,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type ConfigVersion struct {
	ID            string    `json:"id"`
	Version       int       `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     *string   `json:"created_by,omitempty"`
	Comment       *string   `json:"comment,omitempty"`
	ContentSHA256 *string   `json:"content_sha256,omitempty"`
	BodyRaw       string    `json:"body_raw"`
	BodyJSON      any       `json:"body_json,omitempty"`
}

type ConfigVersionMeta struct {
	ID            string    `json:"id"`
	Version       int       `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     *string   `json:"created_by,omitempty"`
	Comment       *string   `json:"comment,omitempty"`
	ContentSHA256 *string   `json:"content_sha256,omitempty"`
}

type GetConfigResponse struct {
	Config Config        `json:"config"`
	Latest ConfigVersion `json:"latest"`
}

type GetVersionResponse struct {
	Config  Config        `json:"config"`
	Version ConfigVersion `json:"version"`
}

type VersionListResponse struct {
	Items      []ConfigVersionMeta `json:"items"`
	NextCursor *string             `json:"next_cursor,omitempty"`
}

type ConfigListItem struct {
	Config     Config            `json:"config"`
	LatestMeta ConfigVersionMeta `json:"latest_meta"`
}

type ConfigListResponse struct {
	Items      []ConfigListItem `json:"items"`
	NextCursor *string          `json:"next_cursor,omitempty"`
}

type Namespace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NamespaceWithCount struct {
	Namespace
	ConfigCount int `json:"config_count"`
}

type NamespaceListResponse struct {
	Items      []NamespaceWithCount `json:"items"`
	NextCursor *string              `json:"next_cursor,omitempty"`
}

type BrowseEntryFolder struct {
	Type     string `json:"type"` // folder
	Name     string `json:"name"`
	FullPath string `json:"full_path"` // ends with /
}

type BrowseEntryConfig struct {
	Type          string       `json:"type"` // config
	Name          string       `json:"name"`
	FullPath      string       `json:"full_path"` // no trailing /
	Format        ConfigFormat `json:"format"`
	LatestVersion int          `json:"latest_version"`
}

type BrowseResponse struct {
	Items      []any   `json:"items"`
	NextCursor *string `json:"next_cursor,omitempty"`
}
