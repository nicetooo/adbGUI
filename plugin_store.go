package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// PluginStore 插件数据库存储
type PluginStore struct {
	db *sql.DB
}

// NewPluginStore 创建插件存储
func NewPluginStore(db *sql.DB) *PluginStore {
	return &PluginStore{db: db}
}

// InitSchema 初始化数据库表
func (ps *PluginStore) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS plugins (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		version TEXT NOT NULL,
		author TEXT,
		description TEXT,
		
		-- 源代码
		source_code TEXT NOT NULL,
		source_language TEXT DEFAULT 'typescript',
		compiled_code TEXT NOT NULL,
		
		-- 配置
		filters TEXT,
		config TEXT,
		
		-- 状态
		enabled INTEGER DEFAULT 1,
		
		-- 时间戳
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	`

	_, err := ps.db.Exec(schema)
	return err
}

// SavePlugin 保存插件
func (ps *PluginStore) SavePlugin(plugin *Plugin) error {
	filtersJSON, err := json.Marshal(plugin.Metadata.Filters)
	if err != nil {
		return fmt.Errorf("序列化 filters 失败: %w", err)
	}

	configJSON, err := json.Marshal(plugin.Metadata.Config)
	if err != nil {
		return fmt.Errorf("序列化 config 失败: %w", err)
	}

	now := time.Now().Unix()
	enabled := 0
	if plugin.Metadata.Enabled {
		enabled = 1
	}

	// 检查是否已存在
	var exists int
	err = ps.db.QueryRow("SELECT COUNT(*) FROM plugins WHERE id = ?", plugin.Metadata.ID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists > 0 {
		// 更新
		_, err = ps.db.Exec(`
			UPDATE plugins SET
				name = ?,
				version = ?,
				author = ?,
				description = ?,
				source_code = ?,
				source_language = ?,
				compiled_code = ?,
				filters = ?,
				config = ?,
				enabled = ?,
				updated_at = ?
			WHERE id = ?
		`,
			plugin.Metadata.Name,
			plugin.Metadata.Version,
			plugin.Metadata.Author,
			plugin.Metadata.Description,
			plugin.SourceCode,
			plugin.Language,
			plugin.CompiledCode,
			string(filtersJSON),
			string(configJSON),
			enabled,
			now,
			plugin.Metadata.ID,
		)
	} else {
		// 插入
		_, err = ps.db.Exec(`
			INSERT INTO plugins (
				id, name, version, author, description,
				source_code, source_language, compiled_code,
				filters, config, enabled,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			plugin.Metadata.ID,
			plugin.Metadata.Name,
			plugin.Metadata.Version,
			plugin.Metadata.Author,
			plugin.Metadata.Description,
			plugin.SourceCode,
			plugin.Language,
			plugin.CompiledCode,
			string(filtersJSON),
			string(configJSON),
			enabled,
			now,
			now,
		)
	}

	return err
}

// GetPlugin 获取插件
func (ps *PluginStore) GetPlugin(id string) (*Plugin, error) {
	var (
		name         string
		version      string
		author       string
		description  string
		sourceCode   string
		language     string
		compiledCode string
		filtersJSON  string
		configJSON   string
		enabled      int
		createdAt    int64
		updatedAt    int64
	)

	err := ps.db.QueryRow(`
		SELECT name, version, author, description,
			   source_code, source_language, compiled_code,
			   filters, config, enabled,
			   created_at, updated_at
		FROM plugins
		WHERE id = ?
	`, id).Scan(
		&name, &version, &author, &description,
		&sourceCode, &language, &compiledCode,
		&filtersJSON, &configJSON, &enabled,
		&createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("插件不存在: %s", id)
	}
	if err != nil {
		return nil, err
	}

	var filters PluginFilters
	if err := json.Unmarshal([]byte(filtersJSON), &filters); err != nil {
		return nil, fmt.Errorf("解析 filters 失败: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("解析 config 失败: %w", err)
	}

	plugin := &Plugin{
		Metadata: PluginMetadata{
			ID:          id,
			Name:        name,
			Version:     version,
			Author:      author,
			Description: description,
			Enabled:     enabled == 1,
			Filters:     filters,
			Config:      config,
			CreatedAt:   time.Unix(createdAt, 0),
			UpdatedAt:   time.Unix(updatedAt, 0),
		},
		SourceCode:   sourceCode,
		Language:     language,
		CompiledCode: compiledCode,
		State:        make(map[string]interface{}),
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (ps *PluginStore) ListPlugins() ([]*Plugin, error) {
	rows, err := ps.db.Query(`
		SELECT id, name, version, author, description,
			   source_code, source_language, compiled_code,
			   filters, config, enabled,
			   created_at, updated_at
		FROM plugins
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plugins []*Plugin

	for rows.Next() {
		var (
			id           string
			name         string
			version      string
			author       string
			description  string
			sourceCode   string
			language     string
			compiledCode string
			filtersJSON  string
			configJSON   string
			enabled      int
			createdAt    int64
			updatedAt    int64
		)

		err := rows.Scan(
			&id, &name, &version, &author, &description,
			&sourceCode, &language, &compiledCode,
			&filtersJSON, &configJSON, &enabled,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		var filters PluginFilters
		if err := json.Unmarshal([]byte(filtersJSON), &filters); err != nil {
			continue
		}

		var config map[string]interface{}
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			continue
		}

		plugin := &Plugin{
			Metadata: PluginMetadata{
				ID:          id,
				Name:        name,
				Version:     version,
				Author:      author,
				Description: description,
				Enabled:     enabled == 1,
				Filters:     filters,
				Config:      config,
				CreatedAt:   time.Unix(createdAt, 0),
				UpdatedAt:   time.Unix(updatedAt, 0),
			},
			SourceCode:   sourceCode,
			Language:     language,
			CompiledCode: compiledCode,
			State:        make(map[string]interface{}),
		}

		plugins = append(plugins, plugin)
	}

	return plugins, rows.Err()
}

// DeletePlugin 删除插件
func (ps *PluginStore) DeletePlugin(id string) error {
	_, err := ps.db.Exec("DELETE FROM plugins WHERE id = ?", id)
	return err
}
