package v1

import (
	"adventure-engine/internal/engine"
	"encoding/json"
)

// --- engine state ---

type FightingEnemy struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	HP          int    `json:"hp"`
}

type EngineStateInfo struct {
	LevelCompletionState string         `json:"level_completion"`
	Mode                 string         `json:"mode"`
	CurrentFloor         string         `json:"current_floor"`
	CurrentRoom          string         `json:"current_room"`
	PlayerHealth         string         `json:"player_health"`
	FightingEnemy        *FightingEnemy `json:"fighting_enemy,omitempty"`
	Notification         string         `json:"notification,omitempty"`
}

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

type ListSessionsResponse struct {
	Sessions []Session `json:"sessions"`
}

type GetSessionResponse struct {
	Session         `json:"session"`
	EngineStateInfo `json:"engine_state"`
}

type DeleteSessionResponse struct {
	SessionID string `json:"session_id"`
}

type DebugResponse struct {
	Session Session         `json:"session"`
	Debug   json.RawMessage `json:"debug"`
}

// --- game actions ---

type ObserveRequest struct{}

type ObserveResponse struct {
	EngineStateInfo `json:"engine_state"`
	RoomInfo        RoomInfo `json:"room_info,omitempty"`
}

type InspectRequest struct {
	TargetName string `json:"target_name" binding:"required"`
}

type InspectResponse struct {
	EngineStateInfo `json:"engine_state"`
	ItemInfo        *ItemInfo `json:"item_info,omitempty"`
	DoorInfo        *DoorInfo `json:"door_info,omitempty"`
}

type UncoverRequest struct {
	TargetName string `json:"target_name" binding:"required"`
}

type UncoverResponse struct {
	EngineStateInfo `json:"engine_state"`
	RevealedItem    ItemInfo `json:"revealed_item"`
}

type UnlockRequest struct {
	KeyOrCode  string `json:"key_or_code" binding:"required"`
	TargetName string `json:"target_name" binding:"required"`
}

type UnlockResponse struct {
	EngineStateInfo `json:"engine_state"`
	Unlocked        bool `json:"unlocked"`
}

type SearchRequest struct {
	TargetName string `json:"target_name" binding:"required"`
}

type SearchResponse struct {
	EngineStateInfo `json:"engine_state"`
	ContainedItem   *ItemInfo `json:"contained_item,omitempty"`
	IsEmpty         bool      `json:"is_empty,omitempty"`
}

type TakeRequest struct {
	TargetName string `json:"target_name" binding:"required"`
}

type TakeResponse struct {
	EngineStateInfo `json:"engine_state"`
	TakenItem       ItemInfo `json:"added_to_inventory"`
}

type InventoryRequest struct{}

type InventoryResponse struct {
	EngineStateInfo `json:"engine_state"`
	Inventory       []ItemInfo  `json:"inventory"`
	Ammo            []AmmoCount `json:"ammo"`
}

type HealRequest struct {
	HealthItemName string `json:"health_item_name" binding:"required"`
}

type HealResponse struct {
	EngineStateInfo `json:"engine_state"`
	HealthState     string `json:"player_health"`
}

type TraverseRequest struct {
	Destination string `json:"door_or_direction" binding:"required"`
}

type TraverseResponse struct {
	EngineStateInfo `json:"engine_state"`
	EnteredRoom     RoomInfo   `json:"entered_room"`
	ChangedFloor    *FloorInfo `json:"changed_floor,omitempty"`
	Unlatched       bool       `json:"unlatched_door,omitempty"`
}

type BattleRequest struct {
	WeaponName string `json:"weapon_name" binding:"required"`
}

type BattleResponse struct {
	EngineStateInfo `json:"engine_state"`
	EnemyName       string `json:"enemy_name"`
	WonRound        bool   `json:"won_round"`
	EnemyAlive      bool   `json:"enemy_alive"`
	PlayerAlive     bool   `json:"player_alive"`
}

type CombineRequest struct {
	InputItemAName string `json:"item_a_name" binding:"required"`
	InputItemBName string `json:"item_b_name" binding:"required"`
}

type CombineResponse struct {
	EngineStateInfo `json:"engine_state"`
	CraftedItem     ItemInfo `json:"crafted_item"`
}

type UseRequest struct {
	ItemName   string `json:"item_name" binding:"required"`
	TargetName string `json:"target_name" binding:"required"`
}

type UseResponse struct {
	EngineStateInfo `json:"engine_state"`
	AcceptedItem    bool      `json:"accepted_item"`
	ProducedItem    *ItemInfo `json:"produced_item,omitempty"`
}

type ContextRequest struct{}

type ContextResponse struct {
	EngineStateInfo `json:"engine_state"`
	RoomInfo        RoomInfo   `json:"room_info"`
	Inventory       []ItemInfo `json:"inventory"`
}

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
	Details      string `json:"details,omitempty"`
	IsFixture    bool   `json:"is_fixture,omitempty"`
}

// DoorInfo is a door as seen from a specific room, a "materialized" door.
// The location is relative to the room from which it is observed, and comes from the
// "connections" field in the room object in the level definition.
type DoorInfo struct {
	Name        string `json:"name"`
	Location    string `json:"location,omitempty"`
	IsLocked    bool   `json:"is_locked,omitempty"`
	HasKeyLock  bool   `json:"has_key_lock,omitempty"`
	HasCodeLock bool   `json:"has_code_lock,omitempty"`
	RoomName    string `json:"room_name,omitempty"`
	IsStairwell bool   `json:"is_stairwell,omitempty"`
	IsLatched   bool   `json:"is_locked_from_the_other_side,omitempty"`
	LeadsTo     string `json:"leads_to,omitempty"`
}

type RoomInfo struct {
	RoomName        string     `json:"name"`
	RoomDescription string     `json:"description"`
	VisibleItems    []ItemInfo `json:"visible_items"`
	Doors           []DoorInfo `json:"connections"`
}

type FloorInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AmmoCount struct {
	WeaponName string `json:"weapon_name"`
	AmmoCount  int    `json:"ammo_count"`
}

// --- helpers to translate engine results to API responses ---

// engineResultToResponseObserve translates an engine.ObserveResult to an ObserveResponse
func EngineResultToResponseObserve(result *engine.ObserveResult) *ObserveResponse {
	items := make([]ItemInfo, len(result.Result.VisibleItems))
	for i, item := range result.Result.VisibleItems {
		items[i] = *getResponseItemInfo(&item)
	}
	doors := make([]DoorInfo, len(result.Result.Doors))
	for i, door := range result.Result.Doors {
		doors[i] = *getResponseDoorInfo(&door)
	}

	observeResponse := &ObserveResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&result.EngineStateInfo),
	}
	if result.EngineStateInfo.Mode == engine.Investigation {
		observeResponse.RoomInfo = RoomInfo{
			RoomName:        result.Result.RoomName,
			RoomDescription: result.Result.RoomDescription,
			VisibleItems:    items,
			Doors:           doors,
		}
	}
	return observeResponse
}

// engineResultToResponseInspect translates an engine.InspectResult to an InspectResponse
func EngineResultToResponseInspect(result *engine.InspectResult) *InspectResponse {
	response := &InspectResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&result.EngineStateInfo),
	}
	if result.Result.ItemInspection != nil {
		response.ItemInfo = getResponseItemInfo(&result.Result.ItemInspection.ItemInfo)
		response.ItemInfo.Details = result.Result.ItemInspection.Detail
	}
	if result.Result.DoorInspection != nil {
		response.DoorInfo = getResponseDoorInfo(&result.Result.DoorInspection.DoorInfo)
	}
	return response
}

// engineResultToResponseUncover translates an engine.UncoverResult to an UncoverResponse
func EngineResultToResponseUncover(result *engine.UncoverResult) *UncoverResponse {
	return &UncoverResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&result.EngineStateInfo),
		RevealedItem:    *getResponseItemInfo(&result.Result.RevealedItem),
	}
}

// engineResultToResponseUnlock translates an engine.UnlockResult to an UnlockResponse
func EngineResultToResponseUnlock(result *engine.UnlockResult) *UnlockResponse {
	return &UnlockResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&result.EngineStateInfo),
		Unlocked:        result.Result.Unlocked,
	}
}

// engineResultToResponseSearch translates an engine.SearchResult to a SearchResponse
func EngineResultToResponseSearch(result *engine.SearchResult) *SearchResponse {
	searchResponse := &SearchResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&result.EngineStateInfo),
	}
	if result.Result.ContainedItemInfo != nil {
		searchResponse.ContainedItem = getResponseItemInfo(result.Result.ContainedItemInfo)
	} else {
		searchResponse.IsEmpty = true
	}
	return searchResponse
}

// engineResultToResponseTake translates an engine.TakeResult to a TakeResponse
func EngineResultToResponseTake(result *engine.TakeResult) *TakeResponse {
	return &TakeResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&result.EngineStateInfo),
		TakenItem:       *getResponseItemInfo(&result.Result.ItemInfo),
	}
}

// engineResultToResponseInventory translates an engine.InventoryResult to an InventoryResponse
func EngineResultToResponseInventory(result *engine.InventoryResult) *InventoryResponse {
	inventory := make([]ItemInfo, len(result.Result.Items))
	for i, item := range result.Result.Items {
		inventory[i] = ItemInfo{
			Name:        item.Name,
			Description: item.Description,
		}
	}
	ammo := make([]AmmoCount, len(result.Result.Ammo))
	for i, ammoCount := range result.Result.Ammo {
		ammo[i] = AmmoCount{
			WeaponName: ammoCount.WeaponName,
			AmmoCount:  ammoCount.AmmoCount,
		}
	}
	return &InventoryResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&result.EngineStateInfo),
		Inventory:       inventory,
		Ammo:            ammo,
	}
}

// engineResultToResponseHeal translates an engine.HealResult to a HealResponse
func EngineResultToResponseHeal(result *engine.HealResult) *HealResponse {
	return &HealResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&result.EngineStateInfo),
		HealthState:     string(result.Result.Health),
	}
}

// engineResultToResponseTraverse translates an engine.TraverseResult to a TraverseResponse
func EngineResultToResponseTraverse(result *engine.TraverseResult) *TraverseResponse {
	items := make([]ItemInfo, len(result.Result.EnteredRoom.VisibleItems))
	for i, item := range result.Result.EnteredRoom.VisibleItems {
		items[i] = *getResponseItemInfo(&item)
	}
	doors := make([]DoorInfo, len(result.Result.EnteredRoom.Doors))
	for i, door := range result.Result.EnteredRoom.Doors {
		doors[i] = *getResponseDoorInfo(&door)
	}

	traverseResponse := &TraverseResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&result.EngineStateInfo),
		EnteredRoom: RoomInfo{
			RoomName:        result.Result.EnteredRoom.RoomName,
			RoomDescription: result.Result.EnteredRoom.RoomDescription,
			VisibleItems:    items,
			Doors:           doors,
		},
		Unlatched: result.Result.Unlatched,
	}
	if result.Result.ChangedFloor != nil {
		traverseResponse.ChangedFloor = &FloorInfo{
			Name:        result.Result.ChangedFloor.Name,
			Description: result.Result.ChangedFloor.Description,
		}
	}
	return traverseResponse
}

// engineResultToResponseBattle translates an engine.BattleResult to a BattleResponse
func EngineResultToResponseBattle(result *engine.BattleResult) *BattleResponse {
	return &BattleResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&result.EngineStateInfo),
		EnemyName:       result.Result.EnemyName,
		WonRound:        result.Result.WonRound,
		EnemyAlive:      result.Result.EnemyAlive,
		PlayerAlive:     result.Result.PlayerAlive,
	}
}

// engineResultToResponseCombine translates an engine.CombineResult to a CombineResponse
func EngineResultToResponseCombine(result *engine.CombineResult) *CombineResponse {
	return &CombineResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&result.EngineStateInfo),
		CraftedItem:     *getResponseItemInfo(&result.Result.CraftedItem),
	}
}

// engineResultToResponseUse translates an engine.UseResult to a UseResponse
func EngineResultToResponseUse(result *engine.UseResult) *UseResponse {
	useResponse := &UseResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&result.EngineStateInfo),
		AcceptedItem:    true,
	}
	if result.Result.ProducedItem != nil {
		useResponse.ProducedItem = getResponseItemInfo(result.Result.ProducedItem)
	}
	return useResponse
}

func EngineResultToResponseContext(observeResult *engine.ObserveResult, inventoryResult *engine.InventoryResult) *ContextResponse {
	return &ContextResponse{
		EngineStateInfo: *getResponseEngineStateInfo(&observeResult.EngineStateInfo),
		RoomInfo:        EngineResultToResponseObserve(observeResult).RoomInfo,
		Inventory:       EngineResultToResponseInventory(inventoryResult).Inventory,
	}
}

// --- private helpers ---

func getResponseItemInfo(item *engine.ItemInfo) *ItemInfo {
	itemInfo := &ItemInfo{
		Name:         item.Name,
		Description:  item.Description,
		Location:     item.Location,
		IsPortable:   item.IsPortable,
		IsKey:        item.IsKey,
		IsWeapon:     item.IsWeapon,
		IsContainer:  item.IsContainer,
		IsConcealer:  item.IsConcealer,
		IsAmmoBox:    item.IsAmmoBox,
		IsHealthItem: item.IsHealthItem,
		HasKeyLock:   item.HasKeyLock,
		HasCodeLock:  item.HasCodeLock,
		IsLocked:     item.IsLocked,
		Contains:     item.Contains,
		IsFixture:    item.IsFixture,
	}

	// Suppress irrelevant information in final response
	if !item.IsLocked {
		itemInfo.HasKeyLock = false
		itemInfo.HasCodeLock = false
	}
	if item.IsSearched {
		itemInfo.IsContainer = false
		if item.Contains == "" {
			itemInfo.Contains = "empty"
		}
	}
	if item.IsConcealer && item.IsUncovered {
		itemInfo.IsConcealer = false
	}

	return itemInfo
}

func getResponseDoorInfo(door *engine.DoorInfo) *DoorInfo {
	doorInfo := &DoorInfo{
		Name:        door.Name,
		Location:    door.Location,
		IsLocked:    door.IsLocked,
		HasKeyLock:  door.HasKeyLock,
		HasCodeLock: door.HasCodeLock,
		IsStairwell: door.IsStairwell,
		IsLatched:   door.IsLatched,
		LeadsTo:     door.LeadsTo,
	}

	// Suppress irrelevant information in final response
	if !door.IsLocked {
		doorInfo.HasKeyLock = false
		doorInfo.HasCodeLock = false
	}
	return doorInfo
}

func getResponseEngineStateInfo(engineState *engine.EngineStateInfo) *EngineStateInfo {
	engineStateInfo := &EngineStateInfo{
		LevelCompletionState: string(engineState.LevelCompletionState),
		Mode:                 string(engineState.Mode),
		PlayerHealth:         string(engineState.PlayerHealth),
		CurrentFloor:         engineState.CurrentFloor.Name,
		CurrentRoom:          engineState.CurrentRoom.Name,
	}
	if engineState.FightingEnemy != nil {
		engineStateInfo.FightingEnemy = &FightingEnemy{
			Name:        engineState.FightingEnemy.BaseEntity.Name,
			Description: engineState.FightingEnemy.BaseEntity.Description,
			HP:          engineState.FightingEnemy.HP,
		}
	}
	if engineState.EngineStateChangeNotification != nil {
		engineStateInfo.Notification = string(*engineState.EngineStateChangeNotification)
	}
	return engineStateInfo
}
