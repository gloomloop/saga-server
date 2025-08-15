package engine

import (
	"adventure-engine/internal/world"
	"errors"
	"fmt"
	"math/rand/v2"
)

// Engine contains all live game state and logic for a single level.
type Engine struct {
	Level                *world.Level
	Player               *world.Player
	CurrentFloor         *world.Floor
	CurrentRoom          *world.Room
	FightingEnemy        *world.Enemy
	Rng                  Rng
	LevelCompletionState LevelCompletionState
	Mode                 Mode
	ValidationDisabled   bool
	MinimapData          map[string]*MinimapDoorInfo // door name -> minimap info
}

// NewEngine creates a new engine for a level.
func NewEngine(
	level *world.Level,
) *Engine {
	engine := &Engine{
		Level: level,
		Player: &world.Player{
			Inventory: make([]*world.Item, 0),
			Health:    world.HealthState(world.HealthFine),
			Ammo:      make(map[string]int),
		},
		CurrentFloor:         level.Floors[0],
		CurrentRoom:          level.GetRoom(level.Floors[0].Name, level.Floors[0].Rooms[0].Name),
		FightingEnemy:        nil,
		Rng:                  &DefaultRng{},
		LevelCompletionState: LevelCompletionStateInProgress,
		Mode:                 Investigation,
		ValidationDisabled:   false,
		MinimapData:          make(map[string]*MinimapDoorInfo),
	}

	engine.initializeMinimapData()

	return engine
}

// Current "mode" of the game.
// Investigation is the default mode.
// Certain events (e.g. entering a room) may trigger a switch to Combat.
type Mode string

const (
	Investigation Mode = "investigation"
	Combat        Mode = "combat"
)

// LevelCompletionState indicates the completion state of the current level.
type LevelCompletionState string

const (
	LevelCompletionStateInProgress LevelCompletionState = "in_progress"
	LevelCompletionStateComplete   LevelCompletionState = "complete"
	LevelCompletionStateFailed     LevelCompletionState = "failed"
)

// EngineStateChangeNotification indicates a state change in the engine.
type EngineStateChangeNotification string

const (
	EngineStateChangeLevelComplete EngineStateChangeNotification = "level_complete"
	EngineStateChangeLevelFailed   EngineStateChangeNotification = "level_failed"
	EngineStateChangeEnterCombat   EngineStateChangeNotification = "enter_combat"
	EngineStateChangeExitCombat    EngineStateChangeNotification = "exit_combat"
)

// runEffect runs a triggered effect.
// Returns a state change notification if applicable.
func (e *Engine) runEffect(effect *world.Effect) *EngineStateChangeNotification {
	switch effect.EffectType {
	case world.EffectEnterCombat:
		e.Mode = Combat
		e.FightingEnemy = e.Level.GetEnemy(effect.EnemyName)
		stateChange := EngineStateChangeEnterCombat
		return &stateChange
	}
	return nil
}

// processTriggers checks if an event matches a trigger.
// Returns a state change notification if applicable.
func (e *Engine) processTriggers(event *world.Event) *EngineStateChangeNotification {
	for _, trigger := range e.Level.Triggers {
		if trigger.Event.Event == event.Event {
			switch trigger.Event.Event {
			case world.EventItemTaken:
				if trigger.Event.ItemName == event.ItemName {
					stateChange := e.runEffect(&trigger.Effect)
					return stateChange
				}
			case world.EventRoomEntered:
				if trigger.Event.RoomName == event.RoomName {
					stateChange := e.runEffect(&trigger.Effect)
					return stateChange
				}
			}
		}
	}
	return nil
}

// processWinCondition checks if an event matches the win condition.
// Returns a state change notification if applicable.
func (e *Engine) processWinCondition(event *world.Event) *EngineStateChangeNotification {
	if e.Level.WinCondition == nil {
		return nil
	}

	switch e.Level.WinCondition.Event {
	case world.EventRoomEntered:
		if e.Level.WinCondition.RoomName == event.RoomName {
			e.LevelCompletionState = LevelCompletionStateComplete
			stateChange := EngineStateChangeLevelComplete
			return &stateChange
		}
	case world.EventEnemyKilled:
		if e.Level.WinCondition.EnemyName == event.EnemyName {
			e.LevelCompletionState = LevelCompletionStateComplete
			stateChange := EngineStateChangeLevelComplete
			return &stateChange
		}
	}
	return nil
}

// handleEnemyKilled handles the event when an enemy is killed.
// Returns a state change notification.
func (e *Engine) handleEnemyKilled() *EngineStateChangeNotification {
	e.Mode = Investigation
	e.FightingEnemy = nil
	stateChange := EngineStateChangeExitCombat
	return &stateChange
}

// handlePlayerKilled handles the event when the player is killed.
// Returns a state change notification.
func (e *Engine) handlePlayerKilled() *EngineStateChangeNotification {
	e.LevelCompletionState = LevelCompletionStateFailed
	stateChange := EngineStateChangeLevelFailed
	return &stateChange
}

// handleEvent handles an event.
func (e *Engine) handleEvent(event *world.Event) *EngineStateChangeNotification {
	switch event.Event {
	case world.EventEnemyKilled:
		enemyKilled := e.handleEnemyKilled()
		won := e.processWinCondition(event)
		if won != nil {
			return won
		}
		return enemyKilled
	case world.EventPlayerKilled:
		return e.handlePlayerKilled()
	case world.EventItemTaken:
		return e.processTriggers(event)
	case world.EventRoomEntered:
		if stateChange := e.processTriggers(event); stateChange != nil {
			return stateChange
		}
		return e.processWinCondition(event)
	}
	return nil
}

// EngineStateInfo contains general engine state info.
type EngineStateInfo struct {
	LevelCompletionState          LevelCompletionState
	CurrentLevel                  *world.Level
	CurrentFloor                  *world.Floor
	CurrentRoom                   *world.Room
	Mode                          Mode
	PlayerHealth                  world.HealthState
	EngineStateChangeNotification *EngineStateChangeNotification
	FightingEnemy                 *world.Enemy
	OutroNarrative                string
}

// --- public wrapper results ---

type ObserveResult struct {
	EngineStateInfo EngineStateInfo
	Result          observeResultInternal
}

type InspectResult struct {
	EngineStateInfo EngineStateInfo
	Result          inspectResultInternal
}

type UncoverResult struct {
	EngineStateInfo EngineStateInfo
	Result          uncoverResultInternal
}

type UnlockResult struct {
	EngineStateInfo EngineStateInfo
	Result          unlockResultInternal
}

type SearchResult struct {
	EngineStateInfo EngineStateInfo
	Result          searchResultInternal
}

type TakeResult struct {
	EngineStateInfo EngineStateInfo
	Result          takeResultInternal
}

type InventoryResult struct {
	EngineStateInfo EngineStateInfo
	Result          inventoryResultInternal
}

type HealResult struct {
	EngineStateInfo EngineStateInfo
	Result          healResultInternal
}

type TraverseResult struct {
	EngineStateInfo EngineStateInfo
	Result          traverseResultInternal
}

type BattleResult struct {
	EngineStateInfo EngineStateInfo
	Result          battleResultInternal
}

type CombineResult struct {
	EngineStateInfo EngineStateInfo
	Result          combineResultInternal
}

type UseResult struct {
	EngineStateInfo EngineStateInfo
	Result          useResultInternal
}

type MinimapResult struct {
	EngineStateInfo EngineStateInfo
	Result          minimapResultInternal
}

// getEngineStateInfo returns the current engine state info.
func (e *Engine) getEngineStateInfo() *EngineStateInfo {
	engineStateInfo := EngineStateInfo{
		LevelCompletionState: e.LevelCompletionState,
		Mode:                 e.Mode,
		CurrentLevel:         e.Level,
		CurrentFloor:         e.CurrentFloor,
		CurrentRoom:          e.CurrentRoom,
		PlayerHealth:         e.Player.Health,
		FightingEnemy:        e.FightingEnemy,
	}
	if e.LevelCompletionState == LevelCompletionStateComplete {
		engineStateInfo.OutroNarrative = e.Level.OutroNarrative
	}
	return &engineStateInfo
}

func (e *Engine) assertValidEngineState() {
	if e.Mode == Combat && e.FightingEnemy == nil {
		panic("cannot be in combat mode without a fighting enemy")
	}
	if e.Mode == Investigation && e.FightingEnemy != nil {
		panic("cannot be in investigation mode while fighting an enemy")
	}
	if !e.Player.IsAlive() && e.LevelCompletionState != LevelCompletionStateFailed {
		panic("level completion state must be failed when the player is dead")
	}
}

func (e *Engine) checkLevelComplete() error {
	if e.LevelCompletionState == LevelCompletionStateComplete {
		return errors.New("level is already complete")
	}
	if e.Player.Health == world.HealthDead {
		return errors.New("player is dead")
	}
	return nil
}

func (e *Engine) ensureCombatMode() error {
	if e.Mode != Combat {
		return errors.New("cannot perform this action in investigation mode")
	}
	return nil
}

func (e *Engine) ensureInvestigationMode() error {
	if e.Mode != Investigation {
		return errors.New("cannot perform this action in combat mode")
	}
	return nil
}

// --- allowed actions validation ---
//
// These functions validate the engine state for specific actions.
// Certain actions are only allowed in certain modes.
//
// The following actions are allowed in all modes:
// - Observe
// - Inventory
// - Heal
//
// The following actions are allowed in investigation mode:
// - Inspect
// - Uncover
// - Unlock
// - Search
// - Take
// - Traverse
// - Combine
//
// The following actions are allowed in combat mode:
// - Battle

// validateEngineState validates the engine state for all actions.
func (e *Engine) validateEngineState() error {
	e.assertValidEngineState()
	if err := e.checkLevelComplete(); err != nil {
		return err
	}
	return nil
}

// validateEngineStateForInvestigationActions validates the engine state for investigation actions.
func (e *Engine) validateEngineStateForInvestigationActions() error {
	if err := e.validateEngineState(); err != nil {
		return err
	}
	if err := e.ensureInvestigationMode(); err != nil {
		return err
	}
	return nil
}

// validateEngineStateForCombatActions validates the engine state for combat actions.
func (e *Engine) validateEngineStateForCombatActions() error {
	if err := e.validateEngineState(); err != nil {
		return err
	}
	if err := e.ensureCombatMode(); err != nil {
		return err
	}
	return nil
}

// --- public wrapper methods ---
// These wrappers add event handling to the underlying internal methods.
// For the time being, not all internal methods generate events.

// Observe observes the current room.
// Returns an ObserveResult and engine state info.
func (e *Engine) Observe() (*ObserveResult, error) {
	if !e.ValidationDisabled {
		if err := e.validateEngineState(); err != nil {
			return nil, err
		}
	}
	observeResult, err := e.observeInternal()
	if err != nil {
		return nil, err
	}
	return &ObserveResult{
		EngineStateInfo: *e.getEngineStateInfo(),
		Result:          *observeResult,
	}, nil
}

// Inspect inspects an item or door by name.
// Returns an InspectResult and engine state info.
func (e *Engine) Inspect(name string) (*InspectResult, error) {
	if err := e.validateEngineStateForInvestigationActions(); err != nil {
		return nil, err
	}
	inspectResult, err := e.inspectInternal(name)
	if err != nil {
		return nil, err
	}
	return &InspectResult{
		EngineStateInfo: *e.getEngineStateInfo(),
		Result:          *inspectResult,
	}, nil
}

// Uncover uncovers an item by name.
// Returns an UncoverResult and engine state info.
func (e *Engine) Uncover(name string) (*UncoverResult, error) {
	if err := e.validateEngineStateForInvestigationActions(); err != nil {
		return nil, err
	}
	uncoverResult, err := e.uncoverInternal(name)
	if err != nil {
		return nil, err
	}
	return &UncoverResult{
		EngineStateInfo: *e.getEngineStateInfo(),
		Result:          *uncoverResult,
	}, nil
}

// Unlock unlocks a door by name.
// Returns an UnlockResult and engine state info.
func (e *Engine) Unlock(keyNameOrCode string, targetName string) (*UnlockResult, error) {
	if err := e.validateEngineStateForInvestigationActions(); err != nil {
		return nil, err
	}
	unlockResult, err := e.unlockInternal(keyNameOrCode, targetName)
	if err != nil {
		return nil, err
	}
	return &UnlockResult{
		EngineStateInfo: *e.getEngineStateInfo(),
		Result:          *unlockResult,
	}, nil
}

// Search searches a container by name.
// Returns a SearchResult and engine state info.
func (e *Engine) Search(name string) (*SearchResult, error) {
	if err := e.validateEngineStateForInvestigationActions(); err != nil {
		return nil, err
	}
	searchResult, err := e.searchInternal(name)
	if err != nil {
		return nil, err
	}
	return &SearchResult{
		EngineStateInfo: *e.getEngineStateInfo(),
		Result:          *searchResult,
	}, nil
}

// Take takes an item by name.
// Handles the event, possibly triggering a state change.
// Returns a TakeResult and engine state info with state change notification, if applicable.
func (e *Engine) Take(name string) (*TakeResult, error) {
	if err := e.validateEngineStateForInvestigationActions(); err != nil {
		return nil, err
	}
	takeResult, err := e.takeInternal(name)
	if err != nil {
		return nil, err
	}
	stateChange := e.handleEvent(&world.Event{
		Event:    world.EventItemTaken,
		ItemName: takeResult.ItemInfo.Name,
	})
	engineStateInfo := e.getEngineStateInfo()
	engineStateInfo.EngineStateChangeNotification = stateChange
	return &TakeResult{
		EngineStateInfo: *engineStateInfo,
		Result:          *takeResult,
	}, nil
}

// Inventory returns the player's inventory.
// Returns an InventoryResult and engine state info.
func (e *Engine) Inventory() (*InventoryResult, error) {
	if !e.ValidationDisabled {
		if err := e.validateEngineState(); err != nil {
			return nil, err
		}
	}
	inventoryResult, err := e.inventoryInternal()
	if err != nil {
		return nil, err
	}
	return &InventoryResult{
		EngineStateInfo: *e.getEngineStateInfo(),
		Result:          *inventoryResult,
	}, nil
}

// Heal heals the player by name.
// Returns a HealResult and engine state info.
func (e *Engine) Heal(name string) (*HealResult, error) {
	if err := e.validateEngineState(); err != nil {
		return nil, err
	}
	healResult, err := e.healInternal(name)
	if err != nil {
		return nil, err
	}
	return &HealResult{
		EngineStateInfo: *e.getEngineStateInfo(),
		Result:          *healResult,
	}, nil
}

// Traverse traverses to a destination room.
// Handles the event, possibly triggering a state change.
// Returns a TraverseResult and engine state info with state change notification, if applicable.
func (e *Engine) Traverse(destination string) (*TraverseResult, error) {
	if err := e.validateEngineStateForInvestigationActions(); err != nil {
		return nil, err
	}
	traverseResult, err := e.traverseInternal(destination)
	if err != nil {
		return nil, err
	}
	stateChange := e.handleEvent(&world.Event{
		Event:    world.EventRoomEntered,
		RoomName: traverseResult.EnteredRoom.RoomName,
	})
	engineStateInfo := e.getEngineStateInfo()
	engineStateInfo.EngineStateChangeNotification = stateChange
	return &TraverseResult{
		EngineStateInfo: *engineStateInfo,
		Result:          *traverseResult,
	}, nil
}

// Battle battles an enemy.
// Handles the event, possibly triggering a state change.
// Returns a BattleResult and engine state info with state change notification, if applicable.
func (e *Engine) Battle(weaponName string) (*BattleResult, error) {
	if err := e.validateEngineStateForCombatActions(); err != nil {
		return nil, err
	}
	var stateChange *EngineStateChangeNotification
	battleResult, err := e.battleInternal(weaponName)
	if err != nil {
		return nil, err
	}
	if !battleResult.EnemyAlive {
		stateChange = e.handleEvent(&world.Event{
			Event:     world.EventEnemyKilled,
			EnemyName: battleResult.EnemyName,
		})
	}
	if !battleResult.PlayerAlive {
		stateChange = e.handleEvent(&world.Event{
			Event: world.EventPlayerKilled,
		})
	}
	engineStateInfo := e.getEngineStateInfo()
	engineStateInfo.EngineStateChangeNotification = stateChange
	return &BattleResult{
		EngineStateInfo: *engineStateInfo,
		Result:          *battleResult,
	}, nil
}

// Combine crafts a new item by combining two input items.
// Returns a CombineResult and engine state info.
func (e *Engine) Combine(inputItemAName string, inputItemBName string) (*CombineResult, error) {
	if err := e.validateEngineStateForInvestigationActions(); err != nil {
		return nil, err
	}
	combineResult, err := e.combineInternal(inputItemAName, inputItemBName)
	if err != nil {
		return nil, err
	}
	return &CombineResult{
		EngineStateInfo: *e.getEngineStateInfo(),
		Result:          *combineResult,
	}, nil
}

func (e *Engine) Use(itemName string, targetName string) (*UseResult, error) {
	if err := e.validateEngineStateForInvestigationActions(); err != nil {
		return nil, err
	}
	useResult, err := e.useInternal(itemName, targetName)
	if err != nil {
		return nil, err
	}
	return &UseResult{
		EngineStateInfo: *e.getEngineStateInfo(),
		Result:          *useResult,
	}, nil
}

// Minimap returns minimap data for the current floor.
// Returns a MinimapResult with engine state info.
func (e *Engine) Minimap() (*MinimapResult, error) {
	if !e.ValidationDisabled {
		if err := e.validateEngineState(); err != nil {
			return nil, err
		}
	}

	minimapResult, err := e.minimapInternal()
	if err != nil {
		return nil, err
	}

	return &MinimapResult{
		EngineStateInfo: *e.getEngineStateInfo(),
		Result:          *minimapResult,
	}, nil
}

// --- internal helpers ---

func (e *Engine) isItemInInventory(itemName string) bool {
	for _, item := range e.Player.Inventory {
		if item.Name == itemName {
			return true
		}
	}
	return false
}

// Set doors to visible in the minimap for the current room
func (e *Engine) updateMinimapDataForCurrenRoom() {
	for _, conn := range e.CurrentRoom.Connections {
		e.MinimapData[conn.DoorName].Hidden = false
	}
}

// Set all doors to hidden with unknown lock state
func (e *Engine) initializeMinimapData() {
	for _, door := range e.Level.Doors {
		e.MinimapData[door.Name] = &MinimapDoorInfo{
			Name:   door.Name,
			Locked: nil,
			Hidden: true,
		}
	}
	e.updateMinimapDataForCurrenRoom()
}

// updateMinimapForDoor updates the minimap data for a specific door
func (e *Engine) updateMinimapForDoor(doorName string, locked bool) {
	if info, exists := e.MinimapData[doorName]; exists {
		info.Locked = &locked
		info.Hidden = false
	}
}

// ItemInfo contains the basic information about an item, excluding detail.
type ItemInfo struct {
	Name         string
	Description  string
	Location     string
	IsPortable   bool
	IsContainer  bool
	IsConcealer  bool
	IsKey        bool
	IsAmmoBox    bool
	IsWeapon     bool
	IsHealthItem bool
	IsFixture    bool

	// Container-specific fields
	HasKeyLock  bool
	HasCodeLock bool
	IsLocked    bool
	IsSearched  bool
	Contains    string

	// Concealer-specific fields
	IsUncovered bool
}

type DoorInfo struct {
	Name        string
	Description string
	Location    string
	HasKeyLock  bool
	HasCodeLock bool
	IsLocked    bool
	IsStairwell bool
	IsLatched   bool
	LeadsTo     string
}

type FloorInfo struct {
	Name        string
	Description string
}

// ItemInspection contains the details of an inspected item.
type ItemInspection struct {
	ItemInfo
	Detail string
}

// DoorInspection contains the details of an inspected door.
type DoorInspection struct {
	DoorInfo
}

// AmmoCount contains ammo count for a weapon, displayed with inventory.
type AmmoCount struct {
	WeaponName string
	AmmoCount  int
}

// createItemInfo creates an ItemInfo from a world item.
func (e *Engine) createItemInfo(item *world.Item) ItemInfo {
	result := ItemInfo{
		Name:         item.Name,
		Description:  item.Description,
		Location:     item.Location,
		IsPortable:   item.IsPortable(),
		IsContainer:  item.IsContainer(),
		IsConcealer:  item.IsConcealer(),
		IsKey:        item.IsKey(),
		IsAmmoBox:    item.IsAmmoBox(),
		IsWeapon:     item.IsWeapon(),
		IsHealthItem: item.IsHealthItem(),
		IsFixture:    item.IsFixture(),
		IsUncovered:  item.IsConcealer() && item.Concealer.Uncovered,
	}

	if item.IsContainer() {
		result.HasKeyLock = item.Container.HasKeyLock()
		result.HasCodeLock = item.Container.HasCodeLock()
		result.IsLocked = item.Container.IsLocked()
		if item.Container.Searched {
			// Show a container's contents if it has been searched already
			result.IsSearched = true
			if !item.Container.IsEmpty() {
				result.Contains = item.Container.Contains.Name
			}
		}
	}
	return result
}

// createDoorInfo creates an DoorInfo from a world door.
func (e *Engine) createDoorInfo(door *world.Door) DoorInfo {
	result := DoorInfo{
		Name:        door.Name,
		IsStairwell: door.Stairwell,
	}

	// Get the connection from the current room to get room-specific description
	if conn, err := e.CurrentRoom.GetConnection(door.Name); err == nil {
		result.Description = conn.Description
	}

	// Only show lock and latch information if the door has been tried
	if door.Tried {
		result.HasKeyLock = door.HasKeyLock()
		result.HasCodeLock = door.HasCodeLock()
		result.IsLocked = door.IsLocked()
		result.IsLatched = door.IsLatched()
	}

	if door.Traversed {
		if e.CurrentRoom.Name == door.RoomA {
			result.LeadsTo = door.RoomB
		} else {
			result.LeadsTo = door.RoomA
		}
	}

	return result
}

type itemInRoomContainer struct {
	ContainedItem  *world.Item
	ContainingItem *world.Item
}

// findItemInRoomContainer finds an item by name in a searched container in the current room.
func (e *Engine) findItemInRoomContainer(name string) (*itemInRoomContainer, error) {
	for _, item := range e.CurrentRoom.Items {
		if item.IsContainer() {
			if item.Container.Searched {
				if !item.Container.IsEmpty() && item.Container.Contains.Name == name {
					return &itemInRoomContainer{
						ContainedItem:  item.Container.Contains,
						ContainingItem: item,
					}, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("you don't see a %s here", name)
}

// findItem finds an item by name.
func (e *Engine) findItem(name string) (*world.Item, error) {
	// Check the player's inventory first.
	item, err := e.Player.GetItem(name)
	if err == nil {
		return item, nil
	}

	// If not found in inventory, check the current room.
	item, err = e.CurrentRoom.GetItem(name)
	if err == nil {
		return item, nil
	}

	// If not found in room, check open containers in the room that have been searched.
	result, err := e.findItemInRoomContainer(name)
	if err == nil {
		return result.ContainedItem, nil
	}

	return nil, fmt.Errorf("you don't see a %s here", name)
}

// findDoorByName finds a door by name.
func (e *Engine) findDoorByName(name string) (*world.Door, error) {
	// First find the connection in the current room
	conn, err := e.CurrentRoom.GetConnection(name)
	if err != nil {
		return nil, err
	}

	// Then find the actual door in the level
	return e.Level.GetDoor(conn.DoorName), nil
}

func (e *Engine) useHealthItem(healthItem *world.Item) world.HealthState {
	switch healthItem.HealthItem.HealthEffect {
	case world.HealthBoostWeak:
		e.Player.IncreaseHealth()
	case world.HealthBoostStrong:
		e.Player.Health = world.HealthState(world.HealthFine)
	default:
		panic("invalid health effect")
	}
	return e.Player.Health
}

// validateKey validates that the item is in inventory and is a key.
func (e *Engine) validateKey(keyName string) error {
	key, err := e.Player.GetItem(keyName)
	if err != nil {
		return err
	}
	if !key.IsKey() {
		return fmt.Errorf("the %s is not a key", keyName)
	}
	return nil
}

// findDoorByLocation finds a door by location (e.g., "left", "ahead", "back", "right").
func (e *Engine) findDoorByLocation(location string) (*world.Door, error) {
	for _, conn := range e.CurrentRoom.Connections {
		if conn.Location == location {
			// Find the actual door in the level
			return e.Level.GetDoor(conn.DoorName), nil
		}
	}
	return nil, fmt.Errorf("no door to the %s", location)
}

// --- internal results ---

// observeResultInternal is the result of observing the current room.
type observeResultInternal struct {
	RoomName        string
	RoomDescription string
	VisibleItems    []ItemInfo
	Doors           []DoorInfo
}

// inspectResultInternal contains the details of an inspected item or door.
type inspectResultInternal struct {
	*ItemInspection
	*DoorInspection
}

// uncoverResultInternal is the result of uncovering a concealed item.
type uncoverResultInternal struct {
	Name         string
	RevealedItem ItemInfo
}

// unlockResultInternal is the result of unlocking a container or door.
type unlockResultInternal struct {
	Unlocked bool
}

// searchResultInternal is the result of searching a container.
type searchResultInternal struct {
	ContainerName     string
	ContainedItemInfo *ItemInfo
	Unlocked          bool
}

// takeResultInternal is the result of taking an item.
type takeResultInternal struct {
	ItemInfo ItemInfo
}

// inventoryResultInternal is the result of getting the player's inventory.
type inventoryResultInternal struct {
	Items []ItemInfo
	Ammo  []AmmoCount
}

// healResultInternal is the result of healing the player.
type healResultInternal struct {
	Health world.HealthState
}

// traverseResultInternal is the result of traversing between rooms.
type traverseResultInternal struct {
	EnteredRoom  observeResultInternal
	ChangedFloor *FloorInfo
	Unlatched    bool
	Unlocked     bool
}

// battleResultInternal is the result of battling an enemy.
type battleResultInternal struct {
	EnemyName   string
	WonRound    bool
	EnemyAlive  bool
	PlayerAlive bool
}

// combineResultInternal is the result of combining two items.
type combineResultInternal struct {
	CraftedItem ItemInfo
}

type useResultInternal struct {
	FixtureName  string
	UsedItemName string
	ProducedItem *ItemInfo
	IsComplete   bool
}

type minimapResultInternal struct {
	Doors       []MinimapDoorInfo
	Rooms       []MinimapRoomInfo
	CurrentRoom string
}

// --- internal methods ---

// Observe returns the current room's name, description, and visible items and doors.
func (e *Engine) observeInternal() (*observeResultInternal, error) {
	// Choose description based on whether room has been visited
	roomDescription := e.CurrentRoom.Description
	if !e.CurrentRoom.Visited && e.CurrentRoom.InitialDescription != "" {
		roomDescription = e.CurrentRoom.InitialDescription
	}

	result := &observeResultInternal{
		RoomName:        e.CurrentRoom.Name,
		RoomDescription: roomDescription,
	}

	for _, item := range e.CurrentRoom.Items {
		result.VisibleItems = append(result.VisibleItems, e.createItemInfo(item))
	}

	for _, conn := range e.CurrentRoom.Connections {
		// Find the actual door in the level
		door := e.Level.GetDoor(conn.DoorName)
		doorInfo := e.createDoorInfo(door)
		doorInfo.Location = conn.Location
		result.Doors = append(result.Doors, doorInfo)
	}

	// Mark room as visited at the end of observation
	e.CurrentRoom.Visited = true

	return result, nil
}

// Inspect inspects an item or door by name.
// Note: this is mainly intended for use on items, but we handle doors just in case.
func (e *Engine) inspectInternal(name string) (*inspectResultInternal, error) {
	door, err := e.findDoorByName(name)
	if err == nil {
		return &inspectResultInternal{
			DoorInspection: &DoorInspection{
				DoorInfo: e.createDoorInfo(door),
			},
		}, nil
	}
	item, err := e.findItem(name)
	if err == nil {
		return &inspectResultInternal{
			ItemInspection: &ItemInspection{
				ItemInfo: e.createItemInfo(item),
				Detail:   item.Detail,
			},
		}, nil
	}
	return nil, err
}

// Uncover reveals something concealed.
func (e *Engine) uncoverInternal(name string) (*uncoverResultInternal, error) {
	concealer, err := e.CurrentRoom.GetItem(name)
	if err != nil {
		return nil, err
	}

	// Check if the item is a concealer and that it is not already uncovered
	if !concealer.IsConcealer() || concealer.Concealer.Uncovered {
		return nil, fmt.Errorf("the %s cannot conceal anything", name)
	}

	// Reveal the concealed item
	revealedItem, err := concealer.Concealer.Reveal()
	if err != nil {
		return nil, err
	}

	// Move the revealed item to the current room
	e.CurrentRoom.Items = append(e.CurrentRoom.Items, revealedItem)

	return &uncoverResultInternal{
		Name:         name,
		RevealedItem: e.createItemInfo(revealedItem),
	}, nil
}

// Unlocks a container or door with a key or code.
func (e *Engine) unlockInternal(keyNameOrCode string, targetName string) (*unlockResultInternal, error) {
	// Try to unlock a container.
	if item, err := e.CurrentRoom.GetItem(targetName); err == nil {
		if !item.IsContainer() {
			return nil, fmt.Errorf("the %s is not a container", targetName)
		}
		if item.Container.HasCodeLock() {
			err := item.Container.UnlockWithCode(keyNameOrCode)
			if err != nil {
				return nil, err
			}
		} else {
			err := e.validateKey(keyNameOrCode)
			if err != nil {
				return nil, err
			}
			err = item.Container.UnlockWithKey(keyNameOrCode)
			if err != nil {
				return nil, err
			}
			// Remove the key from inventory after successful use
			e.Player.RemoveItem(keyNameOrCode)
		}
		return &unlockResultInternal{Unlocked: true}, nil
	}

	// Try to unlock a door.
	if door, err := e.findDoorByName(targetName); err == nil {
		if door.HasCodeLock() {
			err := door.UnlockWithCode(keyNameOrCode)
			if err != nil {
				return nil, err
			}
		} else {
			err := e.validateKey(keyNameOrCode)
			if err != nil {
				return nil, err
			}
			err = door.UnlockWithKey(keyNameOrCode)
			if err != nil {
				return nil, err
			}
			// Remove the key from inventory after successful use
			e.Player.RemoveItem(keyNameOrCode)
		}
		e.updateMinimapForDoor(door.Name, false)
		return &unlockResultInternal{Unlocked: true}, nil
	}

	return nil, fmt.Errorf("you don't see a %s here", targetName)
}

// Search searches a container in the current room.
func (e *Engine) searchInternal(name string) (*searchResultInternal, error) {
	container, err := e.CurrentRoom.GetItem(name)
	if err != nil {
		return nil, err
	}
	unlocked := false

	if !container.IsContainer() {
		return nil, fmt.Errorf("the %s is not a container", name)
	}

	if container.Container.IsLocked() && container.Container.HasKeyLock() {
		if e.isItemInInventory(container.Container.Locked.KeyName) {
			_, err := e.unlockInternal(container.Container.Locked.KeyName, container.Name)
			if err != nil {
				// Should never happen
				panic("error unlocking container: " + err.Error())
			}
			unlocked = true
		}
	}

	// containedItem may be nil if the container is empty
	containedItem, err := container.Container.Search()
	if err != nil {
		return nil, err
	}

	searchResult := &searchResultInternal{
		ContainerName: name,
		Unlocked:      unlocked,
	}

	if containedItem != nil {
		itemInfo := e.createItemInfo(containedItem)
		searchResult.ContainedItemInfo = &itemInfo
	}

	return searchResult, nil
}

// Take takes an item from the current room.
func (e *Engine) takeInternal(name string) (*takeResultInternal, error) {
	// Try to take from the room
	if item, err := e.CurrentRoom.GetItem(name); err == nil {
		// Special handling for items that conceal another item: "redirect" to uncover.
		if item.IsConcealer() && !item.Concealer.Uncovered {
			uncoverResult, err := e.uncoverInternal(name)
			if err != nil {
				return nil, err
			}
			return &takeResultInternal{ItemInfo: uncoverResult.RevealedItem}, nil
		}
		if !item.IsPortable() {
			return nil, fmt.Errorf("you cannot take the %s", name)
		}
		// Handle ammo and weapon ammo transfer
		if handled := e.handleAmmoTransfer(item); handled {
			// Remove ammo box from room (weapons stay in inventory)
			if item.IsAmmoBox() {
				e.CurrentRoom.RemoveItem(item.Name)
				return &takeResultInternal{ItemInfo: e.createItemInfo(item)}, nil
			}
		}
		// Remove the item from the room when taken (except concealers, handled above)
		e.CurrentRoom.RemoveItem(item.Name)
		e.Player.Inventory = append(e.Player.Inventory, item)
		return &takeResultInternal{ItemInfo: e.createItemInfo(item)}, nil
	}

	// Try to take from a searched container.
	if itemContainer, err := e.findItemInRoomContainer(name); err == nil {
		item := itemContainer.ContainedItem
		if !item.IsPortable() {
			return nil, fmt.Errorf("you cannot take the %s", name)
		}
		removedItem, err := itemContainer.ContainingItem.Container.RemoveItem()
		if err != nil {
			return nil, err
		}
		// Handle ammo and weapon ammo transfer
		if handled := e.handleAmmoTransfer(removedItem); handled {
			// Ammo boxes are consumed, weapons stay in inventory
			if removedItem.IsAmmoBox() {
				return &takeResultInternal{ItemInfo: e.createItemInfo(item)}, nil
			}
		}
		e.Player.Inventory = append(e.Player.Inventory, removedItem)
		return &takeResultInternal{ItemInfo: e.createItemInfo(item)}, nil
	}

	return nil, fmt.Errorf("you don't see a %s here", name)
}

// handleAmmoTransfer handles transferring ammo from items to the player's ammo count.
// Returns true if the item was an ammo box (which should be consumed).
func (e *Engine) handleAmmoTransfer(item *world.Item) bool {
	// Handle ammo boxes: add to ammo count and consume the item
	if item.IsAmmoBox() {
		e.Player.Ammo[item.AmmoBox.WeaponName] += item.AmmoBox.Ammo.Quantity
		return true
	}
	// Handle weapons with ammo: transfer ammo to player and clear weapon ammo
	if item.IsWeapon() && item.Weapon.UsesAmmo() {
		e.Player.Ammo[item.Name] += item.Weapon.Ammo.Quantity
		item.Weapon.Ammo.Quantity = 0 // Clear the weapon's ammo
	}
	return false
}

// Inventory returns the player's inventory.
func (e *Engine) inventoryInternal() (*inventoryResultInternal, error) {
	result := &inventoryResultInternal{}
	for _, item := range e.Player.Inventory {
		result.Items = append(result.Items, e.createItemInfo(item))
	}
	for weaponName, ammoCount := range e.Player.Ammo {
		result.Ammo = append(result.Ammo, AmmoCount{
			WeaponName: weaponName,
			AmmoCount:  ammoCount,
		})
	}
	return result, nil
}

// Heal heals the player with a health item.
func (e *Engine) healInternal(healthItemName string) (*healResultInternal, error) {
	// Find the key in the player's inventory.
	healthItem, err := e.Player.GetItem(healthItemName)
	if err != nil {
		return nil, err
	}

	// Use a health item to heal the player.
	if healthItem.IsHealthItem() {
		if e.Player.Health == world.HealthFine {
			return nil, fmt.Errorf("you are already at full health")
		}
		health := e.useHealthItem(healthItem)
		e.Player.RemoveItem(healthItem.Name)
		return &healResultInternal{
			Health: health,
		}, nil
	}

	return nil, fmt.Errorf("the %s is not a health item", healthItemName)
}

// Traverse moves the player to a destination room if reachable and unlocked.
// Destination can be either a door name or a location (e.g., "left", "ahead", "back", "right").
func (e *Engine) traverseInternal(destination string) (*traverseResultInternal, error) {
	var destinationRoom *world.Room
	var door *world.Door
	var err error

	unlocked := false
	// Try to find the door by name first
	door, err = e.findDoorByName(destination)
	if err != nil {
		// If not found by name, try to find by location
		door, err = e.findDoorByLocation(destination)
		if err != nil {
			return nil, fmt.Errorf("no door named '%s' or no door to the '%s'", destination, destination)
		}
	}

	// Mark the door as tried before checking locks
	door.Tried = true

	// Check if the door is locked.
	if door.IsLocked() {
		e.updateMinimapForDoor(door.Name, true)
		if door.HasKeyLock() {
			if e.isItemInInventory(door.Lock.KeyName) {
				_, err := e.unlockInternal(door.Lock.KeyName, door.Name)
				if err != nil {
					// Should never happen
					panic("error unlocking door: " + err.Error())
				}
				unlocked = true
			} else {
				return nil, fmt.Errorf("the %s is locked", door.Name)
			}
		}
		if door.HasCodeLock() {
			return nil, fmt.Errorf("the %s is locked, it requires a code", door.Name)
		}
	}

	// Check if the door is latched.
	// NOTE: latching is not currently supported in the minimap
	var unlatched bool
	if door.IsLatched() {
		// Check if we can unlatch from this side
		if door.CanUnlatch(e.CurrentRoom.Name) {
			// We can unlatch it, so do so
			door.Unlatch()
			unlatched = true
		} else {
			// We can't unlatch from this side
			return nil, fmt.Errorf("this door is latched from the other side")
		}
	}

	// Determine which room is the destination
	var destinationRoomName string
	if door.RoomA == e.CurrentRoom.Name {
		destinationRoomName = door.RoomB
	} else {
		destinationRoomName = door.RoomA
	}

	var destinationFloor *world.Floor

	if door.Stairwell {
		// Stairwell door: can lead to different floors
		// Find the floor that contains the destination room
		for _, floor := range e.Level.Floors {
			for _, room := range floor.Rooms {
				if room.Name == destinationRoomName {
					destinationFloor = floor
					break
				}
			}
			if destinationFloor != nil {
				break
			}
		}

		if destinationFloor == nil {
			panic(fmt.Sprintf("destination room %s not found on any floor", destinationRoomName))
		}

		destinationRoom = e.Level.GetRoom(destinationFloor.Name, destinationRoomName)
	} else {
		// Regular door: must stay on the same floor
		destinationRoom = e.Level.GetRoom(e.CurrentFloor.Name, destinationRoomName)
		destinationFloor = e.CurrentFloor
	}

	// Move to the destination room and floor
	e.CurrentRoom = destinationRoom
	e.CurrentFloor = destinationFloor

	// Mark the door as traversed and update the minimap
	if !door.Traversed {
		door.Traversed = true
		e.updateMinimapDataForCurrenRoom()
		e.updateMinimapForDoor(door.Name, false)
	}

	// Get the observation result for the entered room (without event handling)
	enteredRoomObs, err := e.observeInternal()
	if err != nil {
		return nil, err
	}

	result := &traverseResultInternal{
		EnteredRoom: *enteredRoomObs,
		Unlatched:   unlatched,
		Unlocked:    unlocked,
	}

	// Populate the changed floor info if we used a stairwell
	if door.Stairwell {
		result.ChangedFloor = &FloorInfo{
			Name:        destinationFloor.Name,
			Description: destinationFloor.Description,
		}
	}

	return result, nil
}

// Battle simulates a round of combat between the player and the enemy.
func (e *Engine) battleInternal(weaponName string) (*battleResultInternal, error) {
	var weaponDamage float64

	if e.FightingEnemy == nil {
		return nil, fmt.Errorf("there is no enemy to fight")
	}

	if weaponName == "" || weaponName == "fists" || weaponName == "hands" {
		weaponDamage = 0.5
	} else {
		weapon, err := e.Player.GetItem(weaponName)
		if err != nil {
			return nil, err
		}
		if !weapon.IsWeapon() {
			return nil, fmt.Errorf("the %s is not a weapon", weaponName)
		}
		if weapon.Weapon.UsesAmmo() {
			err := e.Player.FireWeapon(weaponName)
			if err != nil {
				return nil, err
			}
		}
		weaponDamage = weapon.Weapon.Damage
	}

	wonRound := e.Rng.Float64() < weaponDamage
	if wonRound {
		e.FightingEnemy.InflictDamage()
	} else {
		e.Player.InflictDamage()
	}

	return &battleResultInternal{
		EnemyName:   e.FightingEnemy.Name,
		WonRound:    wonRound,
		EnemyAlive:  e.FightingEnemy.IsAlive(),
		PlayerAlive: e.Player.IsAlive(),
	}, nil
}

// Combine crafts a new item by combining two input items.
func (e *Engine) combineInternal(inputItemAName string, inputItemBName string) (*combineResultInternal, error) {
	// Verify both items are in the player's inventory
	_, err := e.Player.GetItem(inputItemAName)
	if err != nil {
		return nil, err
	}
	_, err = e.Player.GetItem(inputItemBName)
	if err != nil {
		return nil, err
	}

	craftedItem, err := e.Level.CombineItems(inputItemAName, inputItemBName)
	if err != nil {
		return nil, err
	}
	e.Player.RemoveItem(inputItemAName)
	e.Player.RemoveItem(inputItemBName)
	e.Player.Inventory = append(e.Player.Inventory, craftedItem)
	return &combineResultInternal{
		CraftedItem: e.createItemInfo(craftedItem),
	}, nil
}

func (e *Engine) useInternal(itemName string, targetName string) (*useResultInternal, error) {
	// Verify the item is in the player's inventory
	_, err := e.Player.GetItem(itemName)
	if err != nil {
		return nil, err
	}

	// Find the target fixture in the current room
	targetFixture, err := e.CurrentRoom.GetItem(targetName)
	if err != nil {
		return nil, fmt.Errorf("fixture %s not found in current room", targetName)
	}

	if !targetFixture.IsFixture() {
		return nil, fmt.Errorf("%s is not a fixture", targetName)
	}

	// Use the item on the fixture
	result, err := targetFixture.Fixture.UseItem(itemName)
	if err != nil {
		return nil, err
	}

	// Remove the used item from player's inventory
	e.Player.RemoveItem(itemName)

	// If the fixture produced an item, add it to player's inventory
	var producedItemInfo *ItemInfo
	if result.Item != nil {
		e.Player.Inventory = append(e.Player.Inventory, result.Item)
		itemInfo := e.createItemInfo(result.Item)
		itemInfo.Location = "inventory" // Override location since it's now in inventory
		producedItemInfo = &itemInfo
	}

	return &useResultInternal{
		FixtureName:  targetName,
		UsedItemName: itemName,
		ProducedItem: producedItemInfo,
		IsComplete:   targetFixture.Fixture.IsComplete(),
	}, nil
}

// minimapInternal returns minimap data for the current floor.
func (e *Engine) minimapInternal() (*minimapResultInternal, error) {
	result := &minimapResultInternal{
		CurrentRoom: e.CurrentRoom.Name,
	}

	// Add all doors from minimap data
	// Note: this returns doors from all floors
	for doorName, doorInfo := range e.MinimapData {
		result.Doors = append(result.Doors, MinimapDoorInfo{
			Name:   doorName,
			Locked: doorInfo.Locked,
			Hidden: doorInfo.Hidden,
		})
	}

	// Add all rooms from current floor
	for _, room := range e.CurrentFloor.Rooms {
		result.Rooms = append(result.Rooms, MinimapRoomInfo{
			Name:   room.Name,
			Hidden: !room.Visited,
		})
	}

	return result, nil
}

func (e *Engine) DisableValidation() {
	e.ValidationDisabled = true
}

func (e *Engine) EnableValidation() {
	e.ValidationDisabled = false
}

// --- RNG ---

type Rng interface {
	Float64() float64
}

type DefaultRng struct{}

func (g *DefaultRng) Float64() float64 { return rand.Float64() }

type FakeRng struct{ Value float64 }

func (g *FakeRng) Float64() float64       { return g.Value }
func (g *FakeRng) SetValue(value float64) { g.Value = value }

// --- minimap structures ---

// MinimapDoorInfo contains minimap information about a door
type MinimapDoorInfo struct {
	Name   string // door name
	Locked *bool  // nil if unknown, true/false if known
	Hidden bool   // true if the door should be hidden on minimap
}

// MinimapRoomInfo contains minimap information about a room
type MinimapRoomInfo struct {
	Name   string
	Hidden bool
}

// --- debug structures ---

// DebugItemInfo contains complete debug information about an item.
type DebugItemInfo struct {
	Name         string
	Description  string
	Location     string
	Detail       string
	IsPortable   bool
	IsContainer  bool
	IsConcealer  bool
	IsKey        bool
	IsAmmoBox    bool
	IsWeapon     bool
	IsHealthItem bool

	// Container-specific fields
	HasKeyLock  bool
	HasCodeLock bool
	IsLocked    bool
	IsSearched  bool
	Contains    *DebugItemInfo // Nested item info if container has contents

	// Weapon-specific fields
	WeaponDamage float64
	UsesAmmo     bool
	AmmoQuantity int

	// AmmoBox-specific fields
	WeaponName string
	AmmoCount  int

	// HealthItem-specific fields
	HealthEffect string

	// Concealer-specific fields
	IsUncovered bool
	HiddenItem  *DebugItemInfo // Nested item info if concealer has hidden item
}

// DebugDoorInfo contains complete debug information about a door.
type DebugDoorInfo struct {
	Name        string
	Location    string
	RoomA       string
	RoomB       string
	HasKeyLock  bool
	HasCodeLock bool
	IsLocked    bool
	KeyName     string
	Code        string
}

// DebugEnemyInfo contains complete debug information about an enemy.
type DebugEnemyInfo struct {
	Name        string
	Description string
	Room        string
	HP          int
	IsAlive     bool
}

// DebugRoomInfo contains complete debug information about a room.
type DebugRoomInfo struct {
	Name        string
	Description string
	Items       []DebugItemInfo
	Doors       []DebugDoorInfo
	IsCurrent   bool
}

// DebugPlayerInfo contains complete debug information about the player.
type DebugPlayerInfo struct {
	Health    string
	IsAlive   bool
	Inventory []DebugItemInfo
	Ammo      map[string]int
}

// DebugEngineState contains complete debug information about the engine state.
type DebugEngineState struct {
	LevelCompletionState string
	Mode                 string
	FightingEnemy        *DebugEnemyInfo
	CurrentRoom          string
}

// DebugResult contains the complete debug information for the engine.
type DebugResult struct {
	EngineStateInfo EngineStateInfo
	EngineState     DebugEngineState
	Player          DebugPlayerInfo
	Rooms           []DebugRoomInfo
	Enemies         []DebugEnemyInfo
	Triggers        []DebugTriggerInfo
	WinCondition    *DebugEventInfo
}

// PrettyPrint formats the debug result in a readable way.
func (d *DebugResult) PrettyPrint() string {
	var result string

	// Engine State
	result += "=== ENGINE STATE ===\n"
	result += fmt.Sprintf("Mode: %s\n", d.EngineState.Mode)
	result += fmt.Sprintf("Level Completion: %s\n", d.EngineState.LevelCompletionState)
	result += fmt.Sprintf("Current Room: %s\n", d.EngineState.CurrentRoom)
	if d.EngineState.FightingEnemy != nil {
		result += fmt.Sprintf("Fighting Enemy: %s (HP: %d, Alive: %t)\n",
			d.EngineState.FightingEnemy.Name,
			d.EngineState.FightingEnemy.HP,
			d.EngineState.FightingEnemy.IsAlive)
	} else {
		result += "Fighting Enemy: None\n"
	}
	result += "\n"

	// Player State
	result += "=== PLAYER STATE ===\n"
	result += fmt.Sprintf("Health: %s\n", d.Player.Health)
	result += fmt.Sprintf("Alive: %t\n", d.Player.IsAlive)
	result += fmt.Sprintf("Inventory Items: %d\n", len(d.Player.Inventory))
	for i, item := range d.Player.Inventory {
		result += fmt.Sprintf("  %d. %s (%s)\n", i+1, item.Name, item.Description)
	}
	result += fmt.Sprintf("Ammo Types: %d\n", len(d.Player.Ammo))
	for weapon, count := range d.Player.Ammo {
		result += fmt.Sprintf("  %s: %d\n", weapon, count)
	}
	result += "\n"

	// Rooms
	result += "=== ROOMS ===\n"
	for _, room := range d.Rooms {
		current := ""
		if room.IsCurrent {
			current = " (CURRENT)"
		}
		result += fmt.Sprintf("Room: %s%s\n", room.Name, current)
		result += fmt.Sprintf("  Description: %s\n", room.Description)
		result += fmt.Sprintf("  Items: %d\n", len(room.Items))
		for i, item := range room.Items {
			result += fmt.Sprintf("    %d. %s (%s)\n", i+1, item.Name, item.Description)
			if item.IsContainer {
				result += fmt.Sprintf("      Container: Searched=%t, Locked=%t\n", item.IsSearched, item.IsLocked)
				if item.Contains != nil {
					result += fmt.Sprintf("      Contains: %s (%s)\n", item.Contains.Name, item.Contains.Description)
				}
			}
			if item.IsConcealer {
				result += fmt.Sprintf("      Concealer: Uncovered=%t\n", item.IsUncovered)
				if item.HiddenItem != nil {
					result += fmt.Sprintf("      Hidden: %s (%s)\n", item.HiddenItem.Name, item.HiddenItem.Description)
				}
			}
			if item.IsWeapon {
				result += fmt.Sprintf("      Weapon: Damage=%.2f, UsesAmmo=%t\n", item.WeaponDamage, item.UsesAmmo)
			}
			if item.IsAmmoBox {
				result += fmt.Sprintf("      AmmoBox: %s (%d rounds)\n", item.WeaponName, item.AmmoCount)
			}
			if item.IsHealthItem {
				result += fmt.Sprintf("      HealthItem: %s\n", item.HealthEffect)
			}
		}
		result += fmt.Sprintf("  Doors: %d\n", len(room.Doors))
		for i, door := range room.Doors {
			result += fmt.Sprintf("    %d. %s (%s) -> %s\n", i+1, door.Name, door.Location, door.RoomB)
			if door.HasKeyLock || door.HasCodeLock {
				result += fmt.Sprintf("      Locked: %t, KeyLock: %t, CodeLock: %t\n", door.IsLocked, door.HasKeyLock, door.HasCodeLock)
			} else {
				result += fmt.Sprintf("      Lock status: unknown (not tried)\n")
			}
		}
		result += "\n"
	}

	// Enemies
	result += "=== ENEMIES ===\n"
	for i, enemy := range d.Enemies {
		result += fmt.Sprintf("%d. %s (%s)\n", i+1, enemy.Name, enemy.Description)
		result += fmt.Sprintf("   Room: %s, HP: %d, Alive: %t\n", enemy.Room, enemy.HP, enemy.IsAlive)
	}
	result += "\n"

	// Triggers
	result += "=== TRIGGERS ===\n"
	for i, trigger := range d.Triggers {
		result += fmt.Sprintf("%d. Event: %s (%s)\n", i+1, trigger.EventType, trigger.EventName)
		result += fmt.Sprintf("   Effect: %s -> %s\n", trigger.EffectType, trigger.EnemyName)
	}
	result += "\n"

	// Win Condition
	result += "=== WIN CONDITION ===\n"
	if d.WinCondition != nil {
		result += fmt.Sprintf("Event: %s\n", d.WinCondition.EventType)
		if d.WinCondition.RoomName != "" {
			result += fmt.Sprintf("Room: %s\n", d.WinCondition.RoomName)
		}
		if d.WinCondition.ItemName != "" {
			result += fmt.Sprintf("Item: %s\n", d.WinCondition.ItemName)
		}
	} else {
		result += "None\n"
	}

	return result
}

// DebugTriggerInfo contains debug information about a trigger.
type DebugTriggerInfo struct {
	EventType  string
	EventName  string
	EffectType string
	EnemyName  string
}

// DebugEventInfo contains debug information about an event.
type DebugEventInfo struct {
	EventType string
	RoomName  string
	ItemName  string
}

// createDebugItemInfo creates a DebugItemInfo from a world item.
func (e *Engine) createDebugItemInfo(item *world.Item) DebugItemInfo {
	result := DebugItemInfo{
		Name:         item.Name,
		Description:  item.Description,
		Location:     item.Location,
		Detail:       item.Detail,
		IsPortable:   item.IsPortable(),
		IsContainer:  item.IsContainer(),
		IsConcealer:  item.IsConcealer(),
		IsKey:        item.IsKey(),
		IsAmmoBox:    item.IsAmmoBox(),
		IsWeapon:     item.IsWeapon(),
		IsHealthItem: item.IsHealthItem(),
	}

	if item.IsContainer() {
		result.HasKeyLock = item.Container.HasKeyLock()
		result.HasCodeLock = item.Container.HasCodeLock()
		result.IsLocked = item.Container.IsLocked()
		result.IsSearched = item.Container.Searched
		if !item.Container.IsEmpty() {
			result.Contains = e.createDebugItemInfoPtr(item.Container.Contains)
		}
	}

	if item.IsWeapon() {
		result.WeaponDamage = item.Weapon.Damage
		result.UsesAmmo = item.Weapon.UsesAmmo()
		if item.Weapon.Ammo != nil {
			result.AmmoQuantity = item.Weapon.Ammo.Quantity
		}
	}

	if item.IsAmmoBox() {
		result.WeaponName = item.AmmoBox.WeaponName
		result.AmmoCount = item.AmmoBox.Ammo.Quantity
	}

	if item.IsHealthItem() {
		switch item.HealthItem.HealthEffect {
		case world.HealthBoostWeak:
			result.HealthEffect = "weak"
		case world.HealthBoostStrong:
			result.HealthEffect = "strong"
		}
	}

	if item.IsConcealer() {
		result.IsUncovered = item.Concealer.Uncovered
		if item.Concealer.Hidden != nil {
			result.HiddenItem = e.createDebugItemInfoPtr(item.Concealer.Hidden)
		}
	}

	return result
}

// createDebugItemInfoPtr creates a pointer to DebugItemInfo from a world item.
func (e *Engine) createDebugItemInfoPtr(item *world.Item) *DebugItemInfo {
	if item == nil {
		return nil
	}
	info := e.createDebugItemInfo(item)
	return &info
}

// createDebugDoorInfo creates a DebugDoorInfo from a world door.
func (e *Engine) createDebugDoorInfo(door *world.Door, location string) DebugDoorInfo {
	result := DebugDoorInfo{
		Name:     door.Name,
		Location: location,
		RoomA:    door.RoomA,
		RoomB:    door.RoomB,
	}

	// Only show lock information if the door has been tried
	if door.Tried {
		result.HasKeyLock = door.HasKeyLock()
		result.HasCodeLock = door.HasCodeLock()
		result.IsLocked = door.IsLocked()
		if door.Lock != nil {
			result.KeyName = door.Lock.KeyName
			result.Code = door.Lock.Code
		}
	}

	return result
}

// createDebugEnemyInfo creates a DebugEnemyInfo from a world enemy.
func (e *Engine) createDebugEnemyInfo(enemy *world.Enemy) DebugEnemyInfo {
	return DebugEnemyInfo{
		Name:        enemy.Name,
		Description: enemy.Description,
		HP:          enemy.HP,
		IsAlive:     enemy.IsAlive(),
	}
}

// createDebugRoomInfo creates a DebugRoomInfo from a world room.
func (e *Engine) createDebugRoomInfo(room *world.Room, isCurrent bool) DebugRoomInfo {
	result := DebugRoomInfo{
		Name:        room.Name,
		Description: room.Description,
		IsCurrent:   isCurrent,
	}

	// Add items
	for _, item := range room.Items {
		result.Items = append(result.Items, e.createDebugItemInfo(item))
	}

	// Add doors
	for _, conn := range room.Connections {
		// Find the actual door in the level
		door := e.Level.GetDoor(conn.DoorName)
		result.Doors = append(result.Doors, e.createDebugDoorInfo(door, conn.Location))
	}

	return result
}

// createDebugPlayerInfo creates a DebugPlayerInfo from the player.
func (e *Engine) createDebugPlayerInfo() DebugPlayerInfo {
	result := DebugPlayerInfo{
		Health:  string(e.Player.Health),
		IsAlive: e.Player.IsAlive(),
		Ammo:    make(map[string]int),
	}

	// Copy ammo map
	for weapon, count := range e.Player.Ammo {
		result.Ammo[weapon] = count
	}

	// Add inventory items
	for _, item := range e.Player.Inventory {
		result.Inventory = append(result.Inventory, e.createDebugItemInfo(item))
	}

	return result
}

// createDebugTriggerInfo creates a DebugTriggerInfo from a world trigger.
func (e *Engine) createDebugTriggerInfo(trigger *world.Trigger) DebugTriggerInfo {
	return DebugTriggerInfo{
		EventType:  string(trigger.Event.Event),
		EventName:  trigger.Event.RoomName + trigger.Event.ItemName, // Combine for display
		EffectType: string(trigger.Effect.EffectType),
		EnemyName:  trigger.Effect.EnemyName,
	}
}

// createDebugEventInfo creates a DebugEventInfo from a world event.
func (e *Engine) createDebugEventInfo(event *world.Event) *DebugEventInfo {
	if event == nil {
		return nil
	}
	return &DebugEventInfo{
		EventType: string(event.Event),
		RoomName:  event.RoomName,
		ItemName:  event.ItemName,
	}
}

// Debug returns complete debug information about the engine state.
func (e *Engine) Debug() (*DebugResult, error) {
	result := &DebugResult{
		EngineStateInfo: *e.getEngineStateInfo(),
		EngineState: DebugEngineState{
			LevelCompletionState: string(e.LevelCompletionState),
			Mode:                 string(e.Mode),
			CurrentRoom:          e.CurrentRoom.Name,
		},
		Player: e.createDebugPlayerInfo(),
	}

	// Add fighting enemy if in combat
	if e.FightingEnemy != nil {
		result.EngineState.FightingEnemy = e.createDebugEnemyInfoPtr(e.FightingEnemy)
	}

	// Add rooms
	for _, room := range e.CurrentFloor.Rooms {
		result.Rooms = append(result.Rooms, e.createDebugRoomInfo(room, room == e.CurrentRoom))
	}

	// Add enemies
	for _, enemy := range e.Level.Enemies {
		result.Enemies = append(result.Enemies, e.createDebugEnemyInfo(enemy))
	}

	// Add triggers
	for _, trigger := range e.Level.Triggers {
		result.Triggers = append(result.Triggers, e.createDebugTriggerInfo(trigger))
	}

	// Add win condition
	result.WinCondition = e.createDebugEventInfo(e.Level.WinCondition)

	return result, nil
}

// createDebugEnemyInfoPtr creates a pointer to DebugEnemyInfo from a world enemy.
func (e *Engine) createDebugEnemyInfoPtr(enemy *world.Enemy) *DebugEnemyInfo {
	if enemy == nil {
		return nil
	}
	info := e.createDebugEnemyInfo(enemy)
	return &info
}
