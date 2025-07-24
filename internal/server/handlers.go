package server

import (
	"net/http"
	"time"

	v1 "adventure-engine/api/v1"
	"adventure-engine/internal/engine"
	"adventure-engine/internal/loader"

	"encoding/json"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GameSession represents a single game session
// Per-session mutexes synchronizes access to live game state
type GameSession struct {
	ID        string
	LevelName string
	CreatedAt time.Time
	Engine    *engine.Engine
	mu        sync.RWMutex
}

// SessionStore holds all active game sessions
// Global mutex synchronizes access to sessions map
type SessionStore struct {
	sessions map[string]*GameSession
	mu       sync.RWMutex
}

// Global session store
var sessionStore = &SessionStore{
	sessions: make(map[string]*GameSession),
}

func safeGetSessionFromStore(sid string, c *gin.Context) *GameSession {
	sessionStore.mu.RLock()
	s, ok := sessionStore.sessions[sid]
	sessionStore.mu.RUnlock()
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return nil
	}
	return s
}

// --- session management ---

func createSession(c *gin.Context) {
	var req v1.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	level, err := loader.LoadGame(req.Level)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load level", "details": err.Error()})
		return
	}

	sid := uuid.New().String()
	session := &GameSession{
		ID:        sid,
		LevelName: level.Name,
		CreatedAt: time.Now(),
		Engine:    engine.NewEngine(level),
	}

	sessionStore.mu.Lock()
	sessionStore.sessions[sid] = session
	sessionStore.mu.Unlock()

	c.JSON(http.StatusOK, v1.CreateSessionResponse{SessionID: sid})
}

// listSessions returns metadata about all active sessions
func listSessions(c *gin.Context) {
	sessionStore.mu.RLock()
	sessions := make([]v1.Session, 0, len(sessionStore.sessions))
	for _, s := range sessionStore.sessions {
		s.mu.RLock()
		sessions = append(sessions, v1.Session{
			ID:        s.ID,
			LevelName: s.LevelName,
			CreatedAt: s.CreatedAt.Format(time.RFC3339),
		})
		s.mu.RUnlock()
	}
	sessionStore.mu.RUnlock()
	c.JSON(http.StatusOK, v1.ListSessionsResponse{Sessions: sessions})
}

// getSession returns metadata and live engine state for a session
func getSession(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	s.mu.RLock()
	resp := v1.GetSessionResponse{
		Session: v1.Session{
			ID:        s.ID,
			LevelName: s.LevelName,
			CreatedAt: s.CreatedAt.Format(time.RFC3339),
		},
		EngineState: v1.EngineState{
			LevelCompletionState: string(s.Engine.LevelCompletionState),
		},
	}
	s.mu.RUnlock()
	c.JSON(http.StatusOK, resp)
}

func deleteSession(c *gin.Context) {
	sid := c.Param("sid")
	sessionStore.mu.Lock()
	_, ok := sessionStore.sessions[sid]
	if !ok {
		sessionStore.mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	delete(sessionStore.sessions, sid)
	sessionStore.mu.Unlock()
	c.JSON(http.StatusOK, v1.DeleteSessionResponse{SessionID: sid})
}

func getDebug(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}
	s.mu.RLock()
	debugResult, err := s.Engine.Debug()
	s.mu.RUnlock()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get debug info", "details": err.Error()})
		return
	}
	debugJSON, err := json.Marshal(debugResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal debug info", "details": err.Error()})
		return
	}
	resp := v1.DebugResponse{
		Session: v1.Session{
			ID:        s.ID,
			LevelName: s.LevelName,
			CreatedAt: s.CreatedAt.Format(time.RFC3339),
		},
		Debug: debugJSON,
	}
	c.JSON(http.StatusOK, resp)
}

// --- game actions ---
//
// Note on return codes for game actions:
// Failed validation in handler: 400 bad request
// Engine returned error: 422 unprocessable entity
// Engine did not return error: 200 ok

func observe(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.Engine.Observe()
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Map engine.EngineStateInfo to v1.EngineStateInfo
	var notification string
	if result.EngineStateInfo.EngineStateChangeNotification != nil {
		notification = string(*result.EngineStateInfo.EngineStateChangeNotification)
	}
	engineState := v1.EngineStateInfo{
		LevelCompletionState: string(result.EngineStateInfo.LevelCompletionState),
		Mode:                 string(result.EngineStateInfo.Mode),
		Notification:         notification,
	}

	// Map items and doors
	items := make([]v1.ItemInfo, len(result.Result.VisibleItems))
	for i, item := range result.Result.VisibleItems {
		items[i] = v1.ItemInfo{
			Name:        item.Name,
			Description: item.Description,
			Location:    item.Location,
			IsPortable:  item.IsPortable,
			IsKey:       item.IsKey,
			IsWeapon:    item.IsWeapon,
			IsContainer: item.IsContainer,
			IsConcealer: item.IsConcealer,
			IsAmmoBox:   item.IsAmmoBox,
			HasKeyLock:  item.HasKeyLock,
			HasCodeLock: item.HasCodeLock,
			IsLocked:    item.IsLocked,
			Contains:    item.Contains,
		}
	}

	doors := make([]v1.DoorInfo, len(result.Result.Doors))
	for i, door := range result.Result.Doors {
		// RoomName is not present in engine.DoorInfo, so leave blank
		doors[i] = v1.DoorInfo{
			Name:        door.Name,
			Description: door.Description,
			Direction:   door.Direction,
			IsLocked:    door.IsLocked,
			HasKeyLock:  door.HasKeyLock,
			HasCodeLock: door.HasCodeLock,
			RoomName:    "",
		}
	}

	resp := v1.ObserveResponse{
		EngineStateInfo: engineState,
		RoomName:        result.Result.RoomName,
		RoomDescription: result.Result.RoomDescription,
		VisibleItems:    items,
		Doors:           doors,
	}
	c.JSON(http.StatusOK, resp)
}

func inspect(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// TODO: Implement inspect action
	c.JSON(http.StatusOK, gin.H{"message": "inspect - TODO", "sid": sid})
}

func uncover(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// TODO: Implement uncover action
	c.JSON(http.StatusOK, gin.H{"message": "uncover - TODO", "sid": sid})
}

func unlock(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// TODO: Implement unlock action
	c.JSON(http.StatusOK, gin.H{"message": "unlock - TODO", "sid": sid})
}

func search(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// TODO: Implement search action
	c.JSON(http.StatusOK, gin.H{"message": "search - TODO", "sid": sid})
}

func take(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// TODO: Implement take action
	c.JSON(http.StatusOK, gin.H{"message": "take - TODO", "sid": sid})
}

func inventory(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// TODO: Implement inventory retrieval
	c.JSON(http.StatusOK, gin.H{"message": "inventory - TODO", "sid": sid})
}

func heal(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// TODO: Implement heal action
	c.JSON(http.StatusOK, gin.H{"message": "heal - TODO", "sid": sid})
}

func traverse(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// TODO: Implement traverse action
	c.JSON(http.StatusOK, gin.H{"message": "traverse - TODO", "sid": sid})
}

func battle(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// TODO: Implement battle action
	c.JSON(http.StatusOK, gin.H{"message": "battle - TODO", "sid": sid})
}
