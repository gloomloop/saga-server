package v1

import "encoding/json"

// --- session management ---

type CreateSessionRequest struct {
	Level json.RawMessage `json:"level"`
}

type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
}

type Session struct {
	ID        string `json:"id"`
	LevelName string `json:"level_name"`
	CreatedAt string `json:"created_at"`
}

type EngineState struct {
	LevelCompletionState string `json:"level_completion_state"`
}

type ListSessionsResponse struct {
	Sessions []Session `json:"sessions"`
}

type GetSessionResponse struct {
	Session     `json:"session"`
	EngineState `json:"engine_state"`
}

type DeleteSessionResponse struct {
	SessionID string `json:"session_id"`
}

// --- game actions ---

type EngineStateInfo struct {
	LevelCompletionState string `json:"level_completion"`
	Mode                 string `json:"mode"`
	Notification         string `json:"notification,omitempty"`
}

type ObserveRequest struct{}

type ItemInfo struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Location     string `json:"location,omitempty"`
	IsPortable   bool   `json:"is_portable,omitempty"`
	IsKey        bool   `json:"is_key,omitempty"`
	IsWeapon     bool   `json:"is_weapon,omitempty"`
	IsContainer  bool   `json:"is_container,omitempty"`
	IsConcealer  bool   `json:"coneals_something,omitempty"`
	IsAmmoBox    bool   `json:"is_ammo_box,omitempty"`
	IsHealthItem bool   `json:"is_health_item,omitempty"`
	HasKeyLock   bool   `json:"has_key_lock,omitempty"`
	HasCodeLock  bool   `json:"has_code_lock,omitempty"`
	IsLocked     bool   `json:"is_locked,omitempty"`
	Contains     string `json:"contains,omitempty"`
}

type DoorInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Direction   string `json:"direction"`
	IsLocked    bool   `json:"is_locked"`
	HasKeyLock  bool   `json:"has_key_lock"`
	HasCodeLock bool   `json:"has_code_lock"`
	RoomName    string `json:"room_name"`
}

type ObserveResponse struct {
	EngineStateInfo EngineStateInfo `json:"engine_state"`
	RoomName        string          `json:"room_name"`
	RoomDescription string          `json:"room_description"`
	VisibleItems    []ItemInfo      `json:"visible_items"`
	Doors           []DoorInfo      `json:"doors"`
}

type DebugResponse struct {
	Session Session         `json:"session"`
	Debug   json.RawMessage `json:"debug"`
}
