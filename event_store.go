package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ========================================
// EventStore - SQLite 事件存储
// ========================================

type EventStore struct {
	db     *sql.DB
	dbPath string

	// 写入缓冲
	writeBuffer    []UnifiedEvent
	writeBufferMu  sync.Mutex
	flushInterval  time.Duration
	flushThreshold int
	flushTicker    *time.Ticker
	stopChan       chan struct{}

	// 预编译语句
	stmtInsertEvent        *sql.Stmt
	stmtInsertEventData    *sql.Stmt
	stmtInsertSession      *sql.Stmt
	stmtUpdateSession      *sql.Stmt
	stmtUpsertTimeIndex    *sql.Stmt
	stmtInsertAssertion    *sql.Stmt
	stmtUpdateAssertion    *sql.Stmt
	stmtInsertAssertResult *sql.Stmt
}

// SQL Schema
const schemaSQL = `
-- 启用 WAL 模式提升并发写入性能
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -64000;
PRAGMA temp_store = MEMORY;
PRAGMA mmap_size = 268435456;

-- ==================== Sessions 表 ====================
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    start_time INTEGER NOT NULL,
    end_time INTEGER DEFAULT 0,
    status TEXT DEFAULT 'active',
    event_count INTEGER DEFAULT 0,
    video_path TEXT,
    video_duration INTEGER,
    video_offset INTEGER DEFAULT 0,
    metadata TEXT DEFAULT '{}',
    created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000),
    updated_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
);

CREATE INDEX IF NOT EXISTS idx_sessions_device ON sessions(device_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_time ON sessions(start_time DESC);

-- ==================== Events 表 ====================
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    device_id TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    relative_time INTEGER NOT NULL,
    duration INTEGER DEFAULT 0,
    source TEXT NOT NULL,
    category TEXT NOT NULL,
    type TEXT NOT NULL,
    level TEXT NOT NULL,
    title TEXT NOT NULL,
    summary TEXT,
    parent_id TEXT,
    step_id TEXT,
    trace_id TEXT,
    aggregate_count INTEGER DEFAULT 0,
    aggregate_first INTEGER DEFAULT 0,
    aggregate_last INTEGER DEFAULT 0,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- 核心索引
CREATE INDEX IF NOT EXISTS idx_events_session_time ON events(session_id, relative_time);
CREATE INDEX IF NOT EXISTS idx_events_session_timestamp ON events(session_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_events_device_time ON events(device_id, timestamp);

-- 筛选索引
CREATE INDEX IF NOT EXISTS idx_events_source ON events(session_id, source);
CREATE INDEX IF NOT EXISTS idx_events_category ON events(session_id, category);
CREATE INDEX IF NOT EXISTS idx_events_level ON events(session_id, level);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(session_id, type);

-- 关联索引
CREATE INDEX IF NOT EXISTS idx_events_parent ON events(parent_id);
CREATE INDEX IF NOT EXISTS idx_events_step ON events(step_id);
CREATE INDEX IF NOT EXISTS idx_events_trace ON events(trace_id);

-- ==================== Event Data 表 (大数据分离存储) ====================
CREATE TABLE IF NOT EXISTS event_data (
    event_id TEXT PRIMARY KEY,
    data TEXT NOT NULL,
    data_size INTEGER,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
);

-- ==================== Time Index 表 ====================
CREATE TABLE IF NOT EXISTS time_index (
    session_id TEXT NOT NULL,
    second INTEGER NOT NULL,
    event_count INTEGER NOT NULL,
    first_event_id TEXT NOT NULL,
    has_error INTEGER DEFAULT 0,
    PRIMARY KEY (session_id, second),
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- ==================== Bookmarks 表 ====================
CREATE TABLE IF NOT EXISTS bookmarks (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    relative_time INTEGER NOT NULL,
    label TEXT NOT NULL,
    color TEXT,
    type TEXT DEFAULT 'user',
    created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000),
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_bookmarks_session ON bookmarks(session_id, relative_time);

-- ==================== Assertions 表 (断言定义) ====================
CREATE TABLE IF NOT EXISTS assertions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL,
    session_id TEXT,
    device_id TEXT,
    time_range_start INTEGER,
    time_range_end INTEGER,
    criteria TEXT NOT NULL,
    expected TEXT NOT NULL,
    timeout INTEGER DEFAULT 0,
    tags TEXT,
    metadata TEXT,
    is_template INTEGER DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000),
    updated_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
);

CREATE INDEX IF NOT EXISTS idx_assertions_session ON assertions(session_id);
CREATE INDEX IF NOT EXISTS idx_assertions_device ON assertions(device_id);
CREATE INDEX IF NOT EXISTS idx_assertions_template ON assertions(is_template);

-- ==================== Assertion Results 表 (断言执行结果) ====================
CREATE TABLE IF NOT EXISTS assertion_results (
    id TEXT PRIMARY KEY,
    assertion_id TEXT NOT NULL,
    assertion_name TEXT NOT NULL,
    session_id TEXT,
    passed INTEGER NOT NULL,
    message TEXT,
    matched_events TEXT,
    actual_value TEXT,
    expected_value TEXT,
    executed_at INTEGER NOT NULL,
    duration INTEGER NOT NULL,
    details TEXT
);

CREATE INDEX IF NOT EXISTS idx_assertion_results_session ON assertion_results(session_id, executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_assertion_results_assertion ON assertion_results(assertion_id);
`

// FTS Schema (单独创建，因为可能需要检查是否存在)
const ftsSchemaSQL = `
-- 全文搜索 (FTS5)
CREATE VIRTUAL TABLE IF NOT EXISTS events_fts USING fts5(
    id,
    title,
    summary,
    content='events',
    content_rowid='rowid'
);
`

const ftsTriggerSQL = `
-- FTS 触发器
CREATE TRIGGER IF NOT EXISTS events_ai AFTER INSERT ON events BEGIN
    INSERT INTO events_fts(rowid, id, title, summary)
    VALUES (new.rowid, new.id, new.title, new.summary);
END;

CREATE TRIGGER IF NOT EXISTS events_ad AFTER DELETE ON events BEGIN
    INSERT INTO events_fts(events_fts, rowid, id, title, summary)
    VALUES('delete', old.rowid, old.id, old.title, old.summary);
END;

CREATE TRIGGER IF NOT EXISTS events_au AFTER UPDATE ON events BEGIN
    INSERT INTO events_fts(events_fts, rowid, id, title, summary)
    VALUES('delete', old.rowid, old.id, old.title, old.summary);
    INSERT INTO events_fts(rowid, id, title, summary)
    VALUES (new.rowid, new.id, new.title, new.summary);
END;
`

// NewEventStore 创建事件存储
func NewEventStore(dataDir string) (*EventStore, error) {
	// 确保目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "events.db")

	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-64000&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池
	db.SetMaxOpenConns(1) // SQLite 单写入
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	store := &EventStore{
		db:             db,
		dbPath:         dbPath,
		writeBuffer:    make([]UnifiedEvent, 0, 1000),
		flushInterval:  500 * time.Millisecond,
		flushThreshold: 500,
		stopChan:       make(chan struct{}),
	}

	// 初始化 schema
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	// 预编译语句
	if err := store.prepareStatements(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	// 启动后台写入
	store.startBackgroundWriter()

	return store, nil
}

// initSchema 初始化数据库 schema
func (s *EventStore) initSchema() error {
	// 执行主 schema
	if _, err := s.db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// 尝试创建 FTS 表 (可能失败如果 SQLite 没有 FTS5 支持)
	if _, err := s.db.Exec(ftsSchemaSQL); err != nil {
		// FTS5 不可用，跳过
		fmt.Printf("FTS5 not available, full-text search disabled: %v\n", err)
	} else {
		// 创建 FTS 触发器
		if _, err := s.db.Exec(ftsTriggerSQL); err != nil {
			fmt.Printf("Failed to create FTS triggers: %v\n", err)
		}
	}

	return nil
}

// prepareStatements 预编译 SQL 语句
func (s *EventStore) prepareStatements() error {
	var err error

	s.stmtInsertEvent, err = s.db.Prepare(`
		INSERT INTO events (
			id, session_id, device_id, timestamp, relative_time, duration,
			source, category, type, level, title, summary,
			parent_id, step_id, trace_id,
			aggregate_count, aggregate_first, aggregate_last
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert event: %w", err)
	}

	s.stmtInsertEventData, err = s.db.Prepare(`
		INSERT OR REPLACE INTO event_data (event_id, data, data_size) VALUES (?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert event data: %w", err)
	}

	s.stmtInsertSession, err = s.db.Prepare(`
		INSERT INTO sessions (
			id, device_id, type, name, start_time, end_time, status,
			event_count, video_path, video_duration, video_offset, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert session: %w", err)
	}

	s.stmtUpdateSession, err = s.db.Prepare(`
		UPDATE sessions SET
			end_time = ?, status = ?, event_count = ?,
			video_path = ?, video_duration = ?, video_offset = ?,
			metadata = ?, updated_at = ?
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare update session: %w", err)
	}

	s.stmtUpsertTimeIndex, err = s.db.Prepare(`
		INSERT INTO time_index (session_id, second, event_count, first_event_id, has_error)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(session_id, second) DO UPDATE SET
			event_count = event_count + excluded.event_count,
			has_error = has_error OR excluded.has_error
	`)
	if err != nil {
		return fmt.Errorf("prepare upsert time index: %w", err)
	}

	s.stmtInsertAssertion, err = s.db.Prepare(`
		INSERT INTO assertions (
			id, name, description, type, session_id, device_id,
			time_range_start, time_range_end, criteria, expected,
			timeout, tags, metadata, is_template, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert assertion: %w", err)
	}

	s.stmtUpdateAssertion, err = s.db.Prepare(`
		UPDATE assertions SET
			name = ?, description = ?, type = ?, session_id = ?, device_id = ?,
			time_range_start = ?, time_range_end = ?, criteria = ?, expected = ?,
			timeout = ?, tags = ?, metadata = ?, is_template = ?, updated_at = ?
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("prepare update assertion: %w", err)
	}

	s.stmtInsertAssertResult, err = s.db.Prepare(`
		INSERT INTO assertion_results (
			id, assertion_id, assertion_name, session_id, passed, message,
			matched_events, actual_value, expected_value, executed_at, duration, details
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert assertion result: %w", err)
	}

	return nil
}

// startBackgroundWriter 启动后台写入
func (s *EventStore) startBackgroundWriter() {
	s.flushTicker = time.NewTicker(s.flushInterval)

	go func() {
		for {
			select {
			case <-s.flushTicker.C:
				s.Flush()
			case <-s.stopChan:
				s.flushTicker.Stop()
				s.Flush() // 最后一次刷新
				return
			}
		}
	}()
}

// Close 关闭存储
func (s *EventStore) Close() error {
	close(s.stopChan)

	// 等待最后的写入完成
	time.Sleep(100 * time.Millisecond)

	// 关闭预编译语句
	if s.stmtInsertEvent != nil {
		s.stmtInsertEvent.Close()
	}
	if s.stmtInsertEventData != nil {
		s.stmtInsertEventData.Close()
	}
	if s.stmtInsertSession != nil {
		s.stmtInsertSession.Close()
	}
	if s.stmtUpdateSession != nil {
		s.stmtUpdateSession.Close()
	}
	if s.stmtUpsertTimeIndex != nil {
		s.stmtUpsertTimeIndex.Close()
	}
	if s.stmtInsertAssertion != nil {
		s.stmtInsertAssertion.Close()
	}
	if s.stmtUpdateAssertion != nil {
		s.stmtUpdateAssertion.Close()
	}
	if s.stmtInsertAssertResult != nil {
		s.stmtInsertAssertResult.Close()
	}

	return s.db.Close()
}

// ========================================
// Event 写入
// ========================================

// WriteEvent 写入单个事件 (缓冲)
func (s *EventStore) WriteEvent(event UnifiedEvent) {
	s.writeBufferMu.Lock()
	s.writeBuffer = append(s.writeBuffer, event)
	shouldFlush := len(s.writeBuffer) >= s.flushThreshold
	s.writeBufferMu.Unlock()

	if shouldFlush {
		go s.Flush()
	}
}

// WriteEventDirect 直接写入事件 (不缓冲)
func (s *EventStore) WriteEventDirect(event UnifiedEvent) error {
	return s.writeEventsBatch([]UnifiedEvent{event})
}

// Flush 刷新缓冲区
func (s *EventStore) Flush() {
	s.writeBufferMu.Lock()
	if len(s.writeBuffer) == 0 {
		s.writeBufferMu.Unlock()
		return
	}

	events := s.writeBuffer
	s.writeBuffer = make([]UnifiedEvent, 0, 1000)
	s.writeBufferMu.Unlock()

	if err := s.writeEventsBatch(events); err != nil {
		fmt.Printf("Failed to flush events: %v\n", err)
	}
}

// writeEventsBatch 批量写入事件
func (s *EventStore) writeEventsBatch(events []UnifiedEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmtEvent := tx.Stmt(s.stmtInsertEvent)
	stmtData := tx.Stmt(s.stmtInsertEventData)

	for _, event := range events {
		// 插入基础事件
		_, err := stmtEvent.Exec(
			event.ID, event.SessionID, event.DeviceID,
			event.Timestamp, event.RelativeTime, event.Duration,
			string(event.Source), string(event.Category), event.Type, string(event.Level),
			event.Title, nullString(event.Summary),
			nullString(event.ParentID), nullString(event.StepID), nullString(event.TraceID),
			event.AggregateCount, event.AggregateFirst, event.AggregateLast,
		)
		if err != nil {
			return fmt.Errorf("insert event %s: %w", event.ID, err)
		}

		// 插入扩展数据
		if len(event.Data) > 0 {
			_, err = stmtData.Exec(event.ID, string(event.Data), len(event.Data))
			if err != nil {
				return fmt.Errorf("insert event data %s: %w", event.ID, err)
			}
		}
	}

	return tx.Commit()
}

// ========================================
// Session 操作
// ========================================

// CreateSession 创建 Session
func (s *EventStore) CreateSession(session *DeviceSession) error {
	metadata, _ := json.Marshal(session.Metadata)
	_, err := s.stmtInsertSession.Exec(
		session.ID, session.DeviceID, session.Type, session.Name,
		session.StartTime, session.EndTime, session.Status, session.EventCount,
		nullString(session.VideoPath), session.VideoDuration, session.VideoOffset,
		string(metadata),
	)
	return err
}

// UpdateSession 更新 Session
func (s *EventStore) UpdateSession(session *DeviceSession) error {
	metadata, _ := json.Marshal(session.Metadata)
	_, err := s.stmtUpdateSession.Exec(
		session.EndTime, session.Status, session.EventCount,
		nullString(session.VideoPath), session.VideoDuration, session.VideoOffset,
		string(metadata), time.Now().UnixMilli(),
		session.ID,
	)
	return err
}

// RenameSession 重命名 Session
func (s *EventStore) RenameSession(id, newName string) error {
	_, err := s.db.Exec(`UPDATE sessions SET name = ?, updated_at = ? WHERE id = ?`,
		newName, time.Now().UnixMilli(), id)
	return err
}

// GetSession 获取 Session
func (s *EventStore) GetSession(id string) (*DeviceSession, error) {
	row := s.db.QueryRow(`
		SELECT id, device_id, type, name, start_time, end_time, status, event_count,
			video_path, video_duration, video_offset, metadata
		FROM sessions WHERE id = ?
	`, id)
	return s.scanSession(row)
}

// GetActiveSession 获取设备的活跃 Session
func (s *EventStore) GetActiveSession(deviceID string) (*DeviceSession, error) {
	row := s.db.QueryRow(`
		SELECT id, device_id, type, name, start_time, end_time, status, event_count,
			video_path, video_duration, video_offset, metadata
		FROM sessions
		WHERE device_id = ? AND status = 'active'
		ORDER BY start_time DESC
		LIMIT 1
	`, deviceID)
	return s.scanSession(row)
}

// ListSessions 列出 Sessions
func (s *EventStore) ListSessions(deviceID string, limit int) ([]DeviceSession, error) {
	query := `
		SELECT id, device_id, type, name, start_time, end_time, status, event_count,
			video_path, video_duration, video_offset, metadata
		FROM sessions
	`
	var args []interface{}

	if deviceID != "" {
		query += ` WHERE device_id = ?`
		args = append(args, deviceID)
	}
	query += ` ORDER BY start_time DESC`
	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []DeviceSession
	for rows.Next() {
		session, err := s.scanSessionRow(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, *session)
	}
	return sessions, rows.Err()
}

// DeleteSession 删除 Session
func (s *EventStore) DeleteSession(id string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// scanSession 扫描单行 Session
func (s *EventStore) scanSession(row *sql.Row) (*DeviceSession, error) {
	var session DeviceSession
	var videoPath, metadata sql.NullString
	var videoDuration, videoOffset sql.NullInt64

	err := row.Scan(
		&session.ID, &session.DeviceID, &session.Type, &session.Name,
		&session.StartTime, &session.EndTime, &session.Status, &session.EventCount,
		&videoPath, &videoDuration, &videoOffset, &metadata,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	session.VideoPath = videoPath.String
	session.VideoDuration = videoDuration.Int64
	session.VideoOffset = videoOffset.Int64

	if metadata.Valid {
		json.Unmarshal([]byte(metadata.String), &session.Metadata)
	}

	return &session, nil
}

// scanSessionRow 扫描 Session 行
func (s *EventStore) scanSessionRow(rows *sql.Rows) (*DeviceSession, error) {
	var session DeviceSession
	var videoPath, metadata sql.NullString
	var videoDuration, videoOffset sql.NullInt64

	err := rows.Scan(
		&session.ID, &session.DeviceID, &session.Type, &session.Name,
		&session.StartTime, &session.EndTime, &session.Status, &session.EventCount,
		&videoPath, &videoDuration, &videoOffset, &metadata,
	)
	if err != nil {
		return nil, err
	}

	session.VideoPath = videoPath.String
	session.VideoDuration = videoDuration.Int64
	session.VideoOffset = videoOffset.Int64

	if metadata.Valid {
		json.Unmarshal([]byte(metadata.String), &session.Metadata)
	}

	return &session, nil
}

// ========================================
// Event 查询
// ========================================

// EventQuery 事件查询参数
type EventQuery struct {
	SessionID  string         `json:"sessionId,omitempty"`
	DeviceID   string         `json:"deviceId,omitempty"`
	Sources    []EventSource  `json:"sources,omitempty"`
	Categories []EventCategory `json:"categories,omitempty"`
	Types      []string       `json:"types,omitempty"`
	Levels     []EventLevel   `json:"levels,omitempty"`
	StartTime  int64          `json:"startTime,omitempty"`  // 相对时间 (ms)
	EndTime    int64          `json:"endTime,omitempty"`
	SearchText string         `json:"searchText,omitempty"`
	ParentID   string         `json:"parentId,omitempty"`
	StepID     string         `json:"stepId,omitempty"`
	TraceID    string         `json:"traceId,omitempty"`
	Limit      int            `json:"limit,omitempty"`
	Offset     int            `json:"offset,omitempty"`
	OrderDesc  bool           `json:"orderDesc,omitempty"` // true = 时间倒序
}

// EventQueryResult 查询结果
type EventQueryResult struct {
	Events  []UnifiedEvent `json:"events"`
	Total   int            `json:"total"`
	HasMore bool           `json:"hasMore"`
}

// QueryEvents 查询事件 (优化版：不加载 event_data，只加载列表需要的字段)
func (s *EventStore) QueryEvents(q EventQuery) (*EventQueryResult, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}

	if q.SessionID != "" {
		conditions = append(conditions, "session_id = ?")
		args = append(args, q.SessionID)
	}
	if q.DeviceID != "" {
		conditions = append(conditions, "device_id = ?")
		args = append(args, q.DeviceID)
	}
	if len(q.Sources) > 0 {
		placeholders := make([]string, len(q.Sources))
		for i, src := range q.Sources {
			placeholders[i] = "?"
			args = append(args, string(src))
		}
		conditions = append(conditions, fmt.Sprintf("source IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(q.Categories) > 0 {
		placeholders := make([]string, len(q.Categories))
		for i, cat := range q.Categories {
			placeholders[i] = "?"
			args = append(args, string(cat))
		}
		conditions = append(conditions, fmt.Sprintf("category IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(q.Types) > 0 {
		placeholders := make([]string, len(q.Types))
		for i, t := range q.Types {
			placeholders[i] = "?"
			args = append(args, t)
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(q.Levels) > 0 {
		placeholders := make([]string, len(q.Levels))
		for i, lvl := range q.Levels {
			placeholders[i] = "?"
			args = append(args, string(lvl))
		}
		conditions = append(conditions, fmt.Sprintf("level IN (%s)", strings.Join(placeholders, ",")))
	}
	if q.StartTime > 0 {
		conditions = append(conditions, "relative_time >= ?")
		args = append(args, q.StartTime)
	}
	if q.EndTime > 0 {
		conditions = append(conditions, "relative_time <= ?")
		args = append(args, q.EndTime)
	}
	if q.ParentID != "" {
		conditions = append(conditions, "parent_id = ?")
		args = append(args, q.ParentID)
	}
	if q.StepID != "" {
		conditions = append(conditions, "step_id = ?")
		args = append(args, q.StepID)
	}
	if q.TraceID != "" {
		conditions = append(conditions, "trace_id = ?")
		args = append(args, q.TraceID)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// 全文搜索
	if q.SearchText != "" {
		// 检查是否有 FTS 表
		var ftsExists int
		s.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='events_fts'").Scan(&ftsExists)
		if ftsExists > 0 {
			if whereClause == "" {
				whereClause = " WHERE "
			} else {
				whereClause += " AND "
			}
			whereClause += "id IN (SELECT id FROM events_fts WHERE events_fts MATCH ?)"
			args = append(args, q.SearchText)
		} else {
			// 降级到 LIKE 搜索
			if whereClause == "" {
				whereClause = " WHERE "
			} else {
				whereClause += " AND "
			}
			whereClause += "(title LIKE ? OR summary LIKE ?)"
			searchPattern := "%" + q.SearchText + "%"
			args = append(args, searchPattern, searchPattern)
		}
	}

	// 获取总数 - 使用带 LIMIT 的估算以加速
	var total int
	if q.Limit > 0 && q.Limit < 10000 {
		// 快速估算：如果查询有 limit，只检查是否超过 limit
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (SELECT 1 FROM events %s LIMIT %d)", whereClause, q.Limit+1)
		if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
			// 降级到普通 count
			countQuery = "SELECT COUNT(*) FROM events" + whereClause
			s.db.QueryRow(countQuery, args...).Scan(&total)
		}
	} else {
		countQuery := "SELECT COUNT(*) FROM events" + whereClause
		if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
			return nil, fmt.Errorf("count query: %w", err)
		}
	}

	// 构建最终查询 - 不 JOIN event_data，列表不需要完整数据
	order := "ASC"
	if q.OrderDesc {
		order = "DESC"
	}
	query := fmt.Sprintf(`
		SELECT id, session_id, device_id, timestamp, relative_time, duration,
			source, category, type, level, title, summary,
			parent_id, step_id, trace_id,
			aggregate_count, aggregate_first, aggregate_last
		FROM events
		%s
		ORDER BY relative_time %s
	`, whereClause, order)

	if q.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", q.Limit)
	}
	if q.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", q.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []UnifiedEvent
	for rows.Next() {
		event, err := s.scanEventRowWithoutData(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}

	hasMore := false
	if q.Limit > 0 {
		hasMore = q.Offset+len(events) < total
	}

	return &EventQueryResult{
		Events:  events,
		Total:   total,
		HasMore: hasMore,
	}, nil
}

// GetEvent 获取单个事件
func (s *EventStore) GetEvent(id string) (*UnifiedEvent, error) {
	row := s.db.QueryRow(`
		SELECT e.id, e.session_id, e.device_id, e.timestamp, e.relative_time, e.duration,
			e.source, e.category, e.type, e.level, e.title, e.summary,
			e.parent_id, e.step_id, e.trace_id,
			e.aggregate_count, e.aggregate_first, e.aggregate_last,
			ed.data
		FROM events e
		LEFT JOIN event_data ed ON e.id = ed.event_id
		WHERE e.id = ?
	`, id)

	return s.scanEventSingle(row)
}

// scanEventRow 扫描事件行 (包含 data)
func (s *EventStore) scanEventRow(rows *sql.Rows) (*UnifiedEvent, error) {
	var event UnifiedEvent
	var summary, parentID, stepID, traceID, data sql.NullString
	var source, category, level string

	err := rows.Scan(
		&event.ID, &event.SessionID, &event.DeviceID,
		&event.Timestamp, &event.RelativeTime, &event.Duration,
		&source, &category, &event.Type, &level,
		&event.Title, &summary,
		&parentID, &stepID, &traceID,
		&event.AggregateCount, &event.AggregateFirst, &event.AggregateLast,
		&data,
	)
	if err != nil {
		return nil, err
	}

	event.Source = EventSource(source)
	event.Category = EventCategory(category)
	event.Level = EventLevel(level)
	event.Summary = summary.String
	event.ParentID = parentID.String
	event.StepID = stepID.String
	event.TraceID = traceID.String

	if data.Valid && data.String != "" {
		event.Data = json.RawMessage(data.String)
	}

	return &event, nil
}

// scanEventRowWithoutData 扫描事件行 (不包含 data，用于列表查询优化)
func (s *EventStore) scanEventRowWithoutData(rows *sql.Rows) (*UnifiedEvent, error) {
	var event UnifiedEvent
	var summary, parentID, stepID, traceID sql.NullString
	var source, category, level string

	err := rows.Scan(
		&event.ID, &event.SessionID, &event.DeviceID,
		&event.Timestamp, &event.RelativeTime, &event.Duration,
		&source, &category, &event.Type, &level,
		&event.Title, &summary,
		&parentID, &stepID, &traceID,
		&event.AggregateCount, &event.AggregateFirst, &event.AggregateLast,
	)
	if err != nil {
		return nil, err
	}

	event.Source = EventSource(source)
	event.Category = EventCategory(category)
	event.Level = EventLevel(level)
	event.Summary = summary.String
	event.ParentID = parentID.String
	event.StepID = stepID.String
	event.TraceID = traceID.String

	return &event, nil
}

// scanEventSingle 扫描单个事件
func (s *EventStore) scanEventSingle(row *sql.Row) (*UnifiedEvent, error) {
	var event UnifiedEvent
	var summary, parentID, stepID, traceID, data sql.NullString
	var source, category, level string

	err := row.Scan(
		&event.ID, &event.SessionID, &event.DeviceID,
		&event.Timestamp, &event.RelativeTime, &event.Duration,
		&source, &category, &event.Type, &level,
		&event.Title, &summary,
		&parentID, &stepID, &traceID,
		&event.AggregateCount, &event.AggregateFirst, &event.AggregateLast,
		&data,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	event.Source = EventSource(source)
	event.Category = EventCategory(category)
	event.Level = EventLevel(level)
	event.Summary = summary.String
	event.ParentID = parentID.String
	event.StepID = stepID.String
	event.TraceID = traceID.String

	if data.Valid && data.String != "" {
		event.Data = json.RawMessage(data.String)
	}

	return &event, nil
}

// ========================================
// Time Index 操作
// ========================================

// UpsertTimeIndex 更新时间索引
func (s *EventStore) UpsertTimeIndex(sessionID string, entry TimeIndexEntry) error {
	hasError := 0
	if entry.HasError {
		hasError = 1
	}
	_, err := s.stmtUpsertTimeIndex.Exec(
		sessionID, entry.Second, entry.EventCount, entry.FirstEventID, hasError,
	)
	return err
}

// GetTimeIndex 直接从事件数据生成时间索引（更可靠）
func (s *EventStore) GetTimeIndex(sessionID string) ([]TimeIndexEntry, error) {
	// 先检查事件的时间分布
	var minTime, maxTime int64
	var totalCount int
	err := s.db.QueryRow(`
		SELECT MIN(relative_time), MAX(relative_time), COUNT(*)
		FROM events WHERE session_id = ?
	`, sessionID).Scan(&minTime, &maxTime, &totalCount)
	if err != nil {
		log.Printf("[GetTimeIndex] Failed to get time range: %v", err)
	} else {
		log.Printf("[GetTimeIndex] Session %s: minTime=%d, maxTime=%d, totalEvents=%d", sessionID, minTime, maxTime, totalCount)
	}

	// 直接从 events 表聚合生成，确保数据准确
	rows, err := s.db.Query(`
		SELECT
			relative_time / 1000 as second,
			COUNT(*) as event_count,
			MIN(id) as first_event_id,
			MAX(CASE WHEN level IN ('error', 'fatal') THEN 1 ELSE 0 END) as has_error
		FROM events
		WHERE session_id = ?
		GROUP BY relative_time / 1000
		ORDER BY second
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []TimeIndexEntry
	for rows.Next() {
		var e TimeIndexEntry
		var hasError int
		if err := rows.Scan(&e.Second, &e.EventCount, &e.FirstEventID, &hasError); err != nil {
			return nil, err
		}
		e.HasError = hasError != 0
		entries = append(entries, e)
	}

	log.Printf("[GetTimeIndex] Session %s: returned %d time index entries", sessionID, len(entries))
	if len(entries) > 0 {
		log.Printf("[GetTimeIndex] First entry: second=%d, count=%d; Last entry: second=%d, count=%d",
			entries[0].Second, entries[0].EventCount,
			entries[len(entries)-1].Second, entries[len(entries)-1].EventCount)
	}

	return entries, rows.Err()
}

// ========================================
// Bookmark 操作
// ========================================

// CreateBookmark 创建书签
func (s *EventStore) CreateBookmark(bookmark *Bookmark) error {
	_, err := s.db.Exec(`
		INSERT INTO bookmarks (id, session_id, relative_time, label, color, type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, bookmark.ID, bookmark.SessionID, bookmark.RelativeTime,
		bookmark.Label, nullString(bookmark.Color), bookmark.Type, bookmark.CreatedAt)
	return err
}

// GetBookmarks 获取 Session 的书签
func (s *EventStore) GetBookmarks(sessionID string) ([]Bookmark, error) {
	rows, err := s.db.Query(`
		SELECT id, session_id, relative_time, label, color, type, created_at
		FROM bookmarks
		WHERE session_id = ?
		ORDER BY relative_time
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var b Bookmark
		var color sql.NullString
		if err := rows.Scan(&b.ID, &b.SessionID, &b.RelativeTime, &b.Label, &color, &b.Type, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.Color = color.String
		bookmarks = append(bookmarks, b)
	}
	return bookmarks, rows.Err()
}

// DeleteBookmark 删除书签
func (s *EventStore) DeleteBookmark(id string) error {
	_, err := s.db.Exec(`DELETE FROM bookmarks WHERE id = ?`, id)
	return err
}

// ========================================
// 统计和维护
// ========================================

// GetSessionStats 获取 Session 统计信息
func (s *EventStore) GetSessionStats(sessionID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 事件总数
	var totalEvents int
	s.db.QueryRow(`SELECT COUNT(*) FROM events WHERE session_id = ?`, sessionID).Scan(&totalEvents)
	stats["totalEvents"] = totalEvents

	// 按来源统计
	rows, err := s.db.Query(`
		SELECT source, COUNT(*) as count
		FROM events WHERE session_id = ?
		GROUP BY source
	`, sessionID)
	if err == nil {
		sourceStats := make(map[string]int)
		for rows.Next() {
			var source string
			var count int
			rows.Scan(&source, &count)
			sourceStats[source] = count
		}
		rows.Close()
		stats["bySource"] = sourceStats
	}

	// 按级别统计
	rows, err = s.db.Query(`
		SELECT level, COUNT(*) as count
		FROM events WHERE session_id = ?
		GROUP BY level
	`, sessionID)
	if err == nil {
		levelStats := make(map[string]int)
		for rows.Next() {
			var level string
			var count int
			rows.Scan(&level, &count)
			levelStats[level] = count
		}
		rows.Close()
		stats["byLevel"] = levelStats
	}

	// 错误数量
	var errorCount int
	s.db.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE session_id = ? AND level IN ('error', 'fatal')
	`, sessionID).Scan(&errorCount)
	stats["errorCount"] = errorCount

	return stats, nil
}

// GetEventTypes 获取 Session 中所有事件类型
func (s *EventStore) GetEventTypes(sessionID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT type FROM events
		WHERE session_id = ? AND type != ''
		ORDER BY type
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil && t != "" {
			types = append(types, t)
		}
	}
	return types, nil
}

// GetEventSources 获取 Session 中所有事件来源
func (s *EventStore) GetEventSources(sessionID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT source FROM events
		WHERE session_id = ?
		ORDER BY source
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err == nil && s != "" {
			sources = append(sources, s)
		}
	}
	return sources, nil
}

// GetEventLevels 获取 Session 中所有事件级别
func (s *EventStore) GetEventLevels(sessionID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT level FROM events
		WHERE session_id = ?
		ORDER BY level
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []string
	for rows.Next() {
		var l string
		if err := rows.Scan(&l); err == nil && l != "" {
			levels = append(levels, l)
		}
	}
	return levels, nil
}

// PreviewAssertionMatch 预览断言匹配的事件数量
func (s *EventStore) PreviewAssertionMatch(sessionID string, types []string, titleMatch string) (int, error) {
	query := `SELECT COUNT(*) FROM events WHERE session_id = ?`
	args := []interface{}{sessionID}

	if len(types) > 0 {
		placeholders := make([]string, len(types))
		for i := range types {
			placeholders[i] = "?"
			args = append(args, types[i])
		}
		query += ` AND type IN (` + strings.Join(placeholders, ",") + `)`
	}

	if titleMatch != "" {
		query += ` AND title REGEXP ?`
		args = append(args, titleMatch)
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// CleanupOldSessions 清理旧 Session
func (s *EventStore) CleanupOldSessions(maxAge time.Duration) (int, error) {
	cutoff := time.Now().Add(-maxAge).UnixMilli()
	result, err := s.db.Exec(`
		DELETE FROM sessions
		WHERE end_time > 0 AND end_time < ?
	`, cutoff)
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// VacuumDatabase 压缩数据库
func (s *EventStore) VacuumDatabase() error {
	_, err := s.db.Exec("VACUUM")
	return err
}

// ========================================
// 辅助函数
// ========================================

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// ========================================
// Assertion 操作
// ========================================

// StoredAssertion 数据库中存储的断言结构
type StoredAssertion struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
	SessionID   string `json:"sessionId,omitempty"`
	DeviceID    string `json:"deviceId,omitempty"`
	TimeRange   *struct {
		Start int64 `json:"start"`
		End   int64 `json:"end"`
	} `json:"timeRange,omitempty"`
	Criteria   json.RawMessage `json:"criteria"`
	Expected   json.RawMessage `json:"expected"`
	Timeout    int64           `json:"timeout,omitempty"`
	Tags       []string        `json:"tags,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	IsTemplate bool            `json:"isTemplate"`
	CreatedAt  int64           `json:"createdAt"`
	UpdatedAt  int64           `json:"updatedAt"`
}

// CreateAssertion 创建断言
func (s *EventStore) CreateAssertion(assertion *StoredAssertion) error {
	var timeRangeStart, timeRangeEnd sql.NullInt64
	if assertion.TimeRange != nil {
		timeRangeStart = sql.NullInt64{Int64: assertion.TimeRange.Start, Valid: true}
		timeRangeEnd = sql.NullInt64{Int64: assertion.TimeRange.End, Valid: true}
	}

	tagsJSON, _ := json.Marshal(assertion.Tags)
	isTemplate := 0
	if assertion.IsTemplate {
		isTemplate = 1
	}

	_, err := s.stmtInsertAssertion.Exec(
		assertion.ID, assertion.Name, nullString(assertion.Description),
		assertion.Type, nullString(assertion.SessionID), nullString(assertion.DeviceID),
		timeRangeStart, timeRangeEnd,
		string(assertion.Criteria), string(assertion.Expected),
		assertion.Timeout, string(tagsJSON), string(assertion.Metadata),
		isTemplate, assertion.CreatedAt, assertion.UpdatedAt,
	)
	return err
}

// UpdateAssertion 更新断言
func (s *EventStore) UpdateAssertion(assertion *StoredAssertion) error {
	var timeRangeStart, timeRangeEnd sql.NullInt64
	if assertion.TimeRange != nil {
		timeRangeStart = sql.NullInt64{Int64: assertion.TimeRange.Start, Valid: true}
		timeRangeEnd = sql.NullInt64{Int64: assertion.TimeRange.End, Valid: true}
	}

	tagsJSON, _ := json.Marshal(assertion.Tags)
	isTemplate := 0
	if assertion.IsTemplate {
		isTemplate = 1
	}

	_, err := s.stmtUpdateAssertion.Exec(
		assertion.Name, nullString(assertion.Description),
		assertion.Type, nullString(assertion.SessionID), nullString(assertion.DeviceID),
		timeRangeStart, timeRangeEnd,
		string(assertion.Criteria), string(assertion.Expected),
		assertion.Timeout, string(tagsJSON), string(assertion.Metadata),
		isTemplate, time.Now().UnixMilli(),
		assertion.ID,
	)
	return err
}

// GetAssertion 获取单个断言
func (s *EventStore) GetAssertion(id string) (*StoredAssertion, error) {
	row := s.db.QueryRow(`
		SELECT id, name, description, type, session_id, device_id,
			time_range_start, time_range_end, criteria, expected,
			timeout, tags, metadata, is_template, created_at, updated_at
		FROM assertions WHERE id = ?
	`, id)
	return s.scanAssertion(row)
}

// ListAssertions 列出断言
func (s *EventStore) ListAssertions(sessionID string, deviceID string, templatesOnly bool, limit int) ([]StoredAssertion, error) {
	query := `
		SELECT id, name, description, type, session_id, device_id,
			time_range_start, time_range_end, criteria, expected,
			timeout, tags, metadata, is_template, created_at, updated_at
		FROM assertions WHERE 1=1
	`
	var args []interface{}

	if sessionID != "" {
		query += ` AND session_id = ?`
		args = append(args, sessionID)
	}
	if deviceID != "" {
		query += ` AND device_id = ?`
		args = append(args, deviceID)
	}
	if templatesOnly {
		query += ` AND is_template = 1`
	}
	query += ` ORDER BY created_at DESC`
	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assertions []StoredAssertion
	for rows.Next() {
		a, err := s.scanAssertionRow(rows)
		if err != nil {
			return nil, err
		}
		assertions = append(assertions, *a)
	}
	return assertions, rows.Err()
}

// DeleteAssertion 删除断言
func (s *EventStore) DeleteAssertion(id string) error {
	_, err := s.db.Exec(`DELETE FROM assertions WHERE id = ?`, id)
	return err
}

// scanAssertion 扫描单个断言
func (s *EventStore) scanAssertion(row *sql.Row) (*StoredAssertion, error) {
	var a StoredAssertion
	var description, sessionID, deviceID sql.NullString
	var timeRangeStart, timeRangeEnd sql.NullInt64
	var criteria, expected, tags, metadata string
	var isTemplate int

	err := row.Scan(
		&a.ID, &a.Name, &description, &a.Type, &sessionID, &deviceID,
		&timeRangeStart, &timeRangeEnd, &criteria, &expected,
		&a.Timeout, &tags, &metadata, &isTemplate, &a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	a.Description = description.String
	a.SessionID = sessionID.String
	a.DeviceID = deviceID.String
	a.Criteria = json.RawMessage(criteria)
	a.Expected = json.RawMessage(expected)
	a.IsTemplate = isTemplate != 0

	if timeRangeStart.Valid && timeRangeEnd.Valid {
		a.TimeRange = &struct {
			Start int64 `json:"start"`
			End   int64 `json:"end"`
		}{Start: timeRangeStart.Int64, End: timeRangeEnd.Int64}
	}

	if tags != "" && tags != "null" {
		json.Unmarshal([]byte(tags), &a.Tags)
	}
	if metadata != "" && metadata != "null" {
		a.Metadata = json.RawMessage(metadata)
	}

	return &a, nil
}

// scanAssertionRow 扫描断言行
func (s *EventStore) scanAssertionRow(rows *sql.Rows) (*StoredAssertion, error) {
	var a StoredAssertion
	var description, sessionID, deviceID sql.NullString
	var timeRangeStart, timeRangeEnd sql.NullInt64
	var criteria, expected, tags, metadata string
	var isTemplate int

	err := rows.Scan(
		&a.ID, &a.Name, &description, &a.Type, &sessionID, &deviceID,
		&timeRangeStart, &timeRangeEnd, &criteria, &expected,
		&a.Timeout, &tags, &metadata, &isTemplate, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	a.Description = description.String
	a.SessionID = sessionID.String
	a.DeviceID = deviceID.String
	a.Criteria = json.RawMessage(criteria)
	a.Expected = json.RawMessage(expected)
	a.IsTemplate = isTemplate != 0

	if timeRangeStart.Valid && timeRangeEnd.Valid {
		a.TimeRange = &struct {
			Start int64 `json:"start"`
			End   int64 `json:"end"`
		}{Start: timeRangeStart.Int64, End: timeRangeEnd.Int64}
	}

	if tags != "" && tags != "null" {
		json.Unmarshal([]byte(tags), &a.Tags)
	}
	if metadata != "" && metadata != "null" {
		a.Metadata = json.RawMessage(metadata)
	}

	return &a, nil
}

// ========================================
// Assertion Result 操作
// ========================================

// StoredAssertionResult 数据库中存储的断言结果
type StoredAssertionResult struct {
	ID            string          `json:"id"`
	AssertionID   string          `json:"assertionId"`
	AssertionName string          `json:"assertionName"`
	SessionID     string          `json:"sessionId"`
	Passed        bool            `json:"passed"`
	Message       string          `json:"message"`
	MatchedEvents []string        `json:"matchedEvents,omitempty"`
	ActualValue   json.RawMessage `json:"actualValue,omitempty"`
	ExpectedValue json.RawMessage `json:"expectedValue,omitempty"`
	ExecutedAt    int64           `json:"executedAt"`
	Duration      int64           `json:"duration"`
	Details       json.RawMessage `json:"details,omitempty"`
}

// SaveAssertionResult 保存断言结果
func (s *EventStore) SaveAssertionResult(result *StoredAssertionResult) error {
	matchedEventsJSON, _ := json.Marshal(result.MatchedEvents)
	passed := 0
	if result.Passed {
		passed = 1
	}

	_, err := s.stmtInsertAssertResult.Exec(
		result.ID, result.AssertionID, result.AssertionName, nullString(result.SessionID),
		passed, nullString(result.Message),
		string(matchedEventsJSON), string(result.ActualValue), string(result.ExpectedValue),
		result.ExecutedAt, result.Duration, string(result.Details),
	)
	return err
}

// GetAssertionResult 获取单个断言结果
func (s *EventStore) GetAssertionResult(id string) (*StoredAssertionResult, error) {
	row := s.db.QueryRow(`
		SELECT id, assertion_id, assertion_name, session_id, passed, message,
			matched_events, actual_value, expected_value, executed_at, duration, details
		FROM assertion_results WHERE id = ?
	`, id)
	return s.scanAssertionResult(row)
}

// ListAssertionResults 列出断言结果
func (s *EventStore) ListAssertionResults(sessionID string, assertionID string, limit int) ([]StoredAssertionResult, error) {
	query := `
		SELECT id, assertion_id, assertion_name, session_id, passed, message,
			matched_events, actual_value, expected_value, executed_at, duration, details
		FROM assertion_results WHERE 1=1
	`
	var args []interface{}

	if sessionID != "" {
		query += ` AND session_id = ?`
		args = append(args, sessionID)
	}
	if assertionID != "" {
		query += ` AND assertion_id = ?`
		args = append(args, assertionID)
	}
	query += ` ORDER BY executed_at DESC`
	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []StoredAssertionResult
	for rows.Next() {
		r, err := s.scanAssertionResultRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, *r)
	}
	return results, rows.Err()
}

// DeleteAssertionResult 删除断言结果
func (s *EventStore) DeleteAssertionResult(id string) error {
	_, err := s.db.Exec(`DELETE FROM assertion_results WHERE id = ?`, id)
	return err
}

// DeleteAssertionResultsBySession 删除 Session 的所有断言结果
func (s *EventStore) DeleteAssertionResultsBySession(sessionID string) error {
	_, err := s.db.Exec(`DELETE FROM assertion_results WHERE session_id = ?`, sessionID)
	return err
}

// scanAssertionResult 扫描单个断言结果
func (s *EventStore) scanAssertionResult(row *sql.Row) (*StoredAssertionResult, error) {
	var r StoredAssertionResult
	var sessionID, message sql.NullString
	var matchedEvents, actualValue, expectedValue, details string
	var passed int

	err := row.Scan(
		&r.ID, &r.AssertionID, &r.AssertionName, &sessionID, &passed, &message,
		&matchedEvents, &actualValue, &expectedValue, &r.ExecutedAt, &r.Duration, &details,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	r.SessionID = sessionID.String
	r.Message = message.String
	r.Passed = passed != 0

	if matchedEvents != "" && matchedEvents != "null" {
		json.Unmarshal([]byte(matchedEvents), &r.MatchedEvents)
	}
	if actualValue != "" && actualValue != "null" {
		r.ActualValue = json.RawMessage(actualValue)
	}
	if expectedValue != "" && expectedValue != "null" {
		r.ExpectedValue = json.RawMessage(expectedValue)
	}
	if details != "" && details != "null" {
		r.Details = json.RawMessage(details)
	}

	return &r, nil
}

// scanAssertionResultRow 扫描断言结果行
func (s *EventStore) scanAssertionResultRow(rows *sql.Rows) (*StoredAssertionResult, error) {
	var r StoredAssertionResult
	var sessionID, message sql.NullString
	var matchedEvents, actualValue, expectedValue, details string
	var passed int

	err := rows.Scan(
		&r.ID, &r.AssertionID, &r.AssertionName, &sessionID, &passed, &message,
		&matchedEvents, &actualValue, &expectedValue, &r.ExecutedAt, &r.Duration, &details,
	)
	if err != nil {
		return nil, err
	}

	r.SessionID = sessionID.String
	r.Message = message.String
	r.Passed = passed != 0

	if matchedEvents != "" && matchedEvents != "null" {
		json.Unmarshal([]byte(matchedEvents), &r.MatchedEvents)
	}
	if actualValue != "" && actualValue != "null" {
		r.ActualValue = json.RawMessage(actualValue)
	}
	if expectedValue != "" && expectedValue != "null" {
		r.ExpectedValue = json.RawMessage(expectedValue)
	}
	if details != "" && details != "null" {
		r.Details = json.RawMessage(details)
	}

	return &r, nil
}
