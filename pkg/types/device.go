package types

// Device represents an Android device
type Device struct {
	ID         string   `json:"id"`
	Serial     string   `json:"serial"`
	State      string   `json:"state"`
	Model      string   `json:"model"`
	Brand      string   `json:"brand"`
	Type       string   `json:"type"` // "wired", "wireless", or "both"
	IDs        []string `json:"ids"`
	WifiAddr   string   `json:"wifiAddr"`
	LastActive int64    `json:"lastActive"`
	IsPinned   bool     `json:"isPinned"`
}

// DeviceInfo contains detailed device information
type DeviceInfo struct {
	Model        string            `json:"model"`
	Brand        string            `json:"brand"`
	Manufacturer string            `json:"manufacturer"`
	AndroidVer   string            `json:"androidVer"`
	SDK          string            `json:"sdk"`
	ABI          string            `json:"abi"`
	Serial       string            `json:"serial"`
	Resolution   string            `json:"resolution"`
	Density      string            `json:"density"`
	CPU          string            `json:"cpu"`
	Memory       string            `json:"memory"`
	Props        map[string]string `json:"props"`
}

// AppPackage represents an installed application
type AppPackage struct {
	Name                 string   `json:"name"`
	Label                string   `json:"label"`
	Icon                 string   `json:"icon"`
	Type                 string   `json:"type"`  // "system" or "user"
	State                string   `json:"state"` // "enabled" or "disabled"
	VersionName          string   `json:"versionName"`
	VersionCode          string   `json:"versionCode"`
	MinSdkVersion        string   `json:"minSdkVersion"`
	TargetSdkVersion     string   `json:"targetSdkVersion"`
	Permissions          []string `json:"permissions"`
	Activities           []string `json:"activities"`
	LaunchableActivities []string `json:"launchableActivities"`
}

// ScrcpyConfig contains screen recording configuration
type ScrcpyConfig struct {
	MaxSize     int  `json:"maxSize"`
	BitRate     int  `json:"bitRate"`
	MaxFps      int  `json:"maxFps"`
	ShowTouches bool `json:"showTouches"`
}

// UIHierarchyResult contains UI hierarchy data
type UIHierarchyResult struct {
	Root   interface{} `json:"root"`
	RawXML string      `json:"rawXml"`
}

// EventQuery defines parameters for querying events
type EventQuery struct {
	SessionID  string   `json:"sessionId,omitempty"`
	DeviceID   string   `json:"deviceId,omitempty"`
	Types      []string `json:"types,omitempty"`
	Sources    []string `json:"sources,omitempty"` // Will be converted to EventSource
	Levels     []string `json:"levels,omitempty"`  // Will be converted to EventLevel
	StartTime  int64    `json:"startTime,omitempty"`
	EndTime    int64    `json:"endTime,omitempty"`
	SearchText string   `json:"searchText,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Offset     int      `json:"offset,omitempty"`
}

// EventQueryResult contains query results
type EventQueryResult struct {
	Events  []interface{} `json:"events"`
	Total   int           `json:"total"`
	HasMore bool          `json:"hasMore"`
}

// DeviceSession represents a tracking session
type DeviceSession struct {
	ID            string `json:"id"`
	DeviceID      string `json:"deviceId"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	StartTime     int64  `json:"startTime"`
	EndTime       int64  `json:"endTime,omitempty"`
	EventCount    int    `json:"eventCount"`
	VideoPath     string `json:"videoPath,omitempty"`
	VideoDuration int64  `json:"videoDuration,omitempty"`
}

// VideoMetadata contains video file metadata
type VideoMetadata struct {
	Path        string  `json:"path"`
	Duration    float64 `json:"duration"`   // Duration in seconds
	DurationMs  int64   `json:"durationMs"` // Duration in milliseconds
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	FrameRate   float64 `json:"frameRate"`
	Codec       string  `json:"codec"`
	BitRate     int64   `json:"bitRate"`
	TotalFrames int64   `json:"totalFrames"`
}
