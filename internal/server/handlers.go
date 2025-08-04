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

// createSession creates a new game session, loading the level from the request body
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
		EngineStateInfo: v1.EngineStateInfo{
			LevelCompletionState: string(s.Engine.LevelCompletionState),
			Mode:                 string(s.Engine.Mode),
		},
	}
	s.mu.RUnlock()
	c.JSON(http.StatusOK, resp)
}

// deleteSession deletes a game session
// Note: this just deletes the reference -- it should be GC'd eventually
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

// getDebug returns detailed debug information for a game session
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

// observe handles observe action requests
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

	c.JSON(http.StatusOK, v1.EngineResultToResponseObserve(result))
}

// inspect handles inspect action requests
func inspect(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	var requestBody v1.InspectRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid InspectRequest", "details": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.Engine.Inspect(requestBody.TargetName)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v1.EngineResultToResponseInspect(result))
}

// uncover handles uncover action requests
func uncover(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	var requestBody v1.InspectRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UncoverRequest", "details": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.Engine.Uncover(requestBody.TargetName)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v1.EngineResultToResponseUncover(result))
}

// unlock handles unlock action requests
func unlock(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	var requestBody v1.UnlockRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UnlockRequest", "details": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.Engine.Unlock(requestBody.KeyOrCode, requestBody.TargetName)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v1.EngineResultToResponseUnlock(result))
}

// search handles search action requests
func search(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	var requestBody v1.SearchRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SearchRequest", "details": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.Engine.Search(requestBody.TargetName)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v1.EngineResultToResponseSearch(result))
}

// take handles take action requests
func take(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	var requestBody v1.TakeRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid TakeRequest", "details": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.Engine.Take(requestBody.TargetName)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v1.EngineResultToResponseTake(result))
}

// inventory handles inventory action requests
func inventory(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.Engine.Inventory()
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v1.EngineResultToResponseInventory(result))
}

// heal handles heal action requests
func heal(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		return
	}

	var requestBody v1.HealRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid HealRequest", "details": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.Engine.Heal(requestBody.HealthItemName)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v1.EngineResultToResponseHeal(result))
}

// traverse handles traverse action requests
func traverse(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	var requestBody v1.TraverseRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid TraverseRequest", "details": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	traverseResult, err := s.Engine.Traverse(requestBody.Destination)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// Observe the room after entering and use this as the response
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v1.EngineResultToResponseTraverse(traverseResult))
}

// battle handles battle action requests
func battle(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	var requestBody v1.BattleRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid BattleRequest", "details": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.Engine.Battle(requestBody.WeaponName)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v1.EngineResultToResponseBattle(result))
}

// combine handles combine action requests
func combine(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	var requestBody v1.CombineRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CombineRequest", "details": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.Engine.Combine(requestBody.InputItemAName, requestBody.InputItemBName)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v1.EngineResultToResponseCombine(result))
}

// use handles use action requests
func use(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	var requestBody v1.UseRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UseRequest", "details": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.Engine.Use(requestBody.ItemName, requestBody.TargetName)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v1.EngineResultToResponseUse(result))
}

// context handles context requests -- this is a special method used to obtain game
// state information necessary for LLM action mapping. It could be optimized.
func context(c *gin.Context) {
	sid := c.Param("sid")
	s := safeGetSessionFromStore(sid, c)
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Skip validation for context requests
	s.Engine.DisableValidation()

	observeResult, err := s.Engine.Observe()
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	inventoryResult, err := s.Engine.Inventory()
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	s.Engine.EnableValidation()

	c.JSON(http.StatusOK, v1.EngineResultToResponseContext(observeResult, inventoryResult))
}
