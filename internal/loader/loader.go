package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"adventure-engine/internal/world"

	"gopkg.in/yaml.v3"
)

// ComboItemData represents a combination item in the JSON
type ComboItemData struct {
	InputItemAName string   `json:"input_item_a_name"`
	InputItemBName string   `json:"input_item_b_name"`
	OutputItem     ItemData `json:"output_item"`
}

// GameData represents the top-level JSON structure
type FloorData struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Rooms       []RoomData `json:"rooms"`
}

type GameData struct {
	Name         string          `json:"name"`
	WinCondition *EventData      `json:"win_condition"`
	Floors       []FloorData     `json:"floors,omitempty"`
	Rooms        []RoomData      `json:"rooms,omitempty"` // For backward compatibility
	DoorData     []DoorData      `json:"doors"`
	Enemies      []EnemyData     `json:"enemies"`
	ComboItems   []ComboItemData `json:"combo_items,omitempty"`
}

// EventData represents an event in the JSON
type EventData struct {
	Event    string `json:"event"`
	RoomName string `json:"room_name"`
	ItemName string `json:"item_name"`
}

// RoomData represents a room in the JSON
type RoomData struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Connections []ConnectionData `json:"connections,omitempty"`
	Items       []ItemData       `json:"items,omitempty"`
}

// ConnectionData represents a room connection in the JSON
type ConnectionData struct {
	Location string `json:"location"`
	DoorName string `json:"door_name"`
}

// ContainerContents can be either an ItemData or the string "empty"
type ContainerContents struct {
	Item  *ItemData
	Empty bool
}

// UnmarshalJSON implements custom unmarshaling for ContainerContents
func (cc *ContainerContents) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str == "empty" {
			cc.Empty = true
			return nil
		}
	}

	// Try to unmarshal as ItemData
	var item ItemData
	if err := json.Unmarshal(data, &item); err == nil {
		cc.Item = &item
		return nil
	}

	return fmt.Errorf("invalid container contents")
}

// FixtureData represents a fixture in the JSON
type FixtureData struct {
	RequiredItems []string  `json:"required_items"`
	Produces      *ItemData `json:"produces"`
}

// ItemData represents an item in the JSON
type ItemData struct {
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Location     string             `json:"location,omitempty"`
	Detail       string             `json:"detail,omitempty"`
	Portable     bool               `json:"portable,omitempty"`
	Key          bool               `json:"key,omitempty"`
	WeaponDamage float64            `json:"weapon_damage,omitempty"`
	Ammo         int                `json:"ammo,omitempty"`
	WeaponName   string             `json:"weapon_name,omitempty"`
	HealthEffect string             `json:"health_effect,omitempty"`
	Code         string             `json:"code,omitempty"`
	Conceals     *ItemData          `json:"conceals,omitempty"`
	Contains     *ContainerContents `json:"contains,omitempty"`
	Fixture      *FixtureData       `json:"fixture,omitempty"`
}

// DoorData represents a door in the JSON
type DoorData struct {
	Name            string `json:"name"`
	RoomA           string `json:"room_a"`
	RoomB           string `json:"room_b"`
	Locked          bool   `json:"locked,omitempty"`
	RequiredKeyName string `json:"required_key_name,omitempty"`
	Code            string `json:"code,omitempty"`
	Stairwell       bool   `json:"stairwell,omitempty"`
	LatchedFrom     string `json:"latched_from,omitempty"`
}

// EnemyData represents an enemy in the JSON
type EnemyData struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	HP          int          `json:"hp"`
	Room        string       `json:"room"`
	Trigger     *TriggerData `json:"trigger,omitempty"`
}

// TriggerData represents a trigger in the JSON
type TriggerData struct {
	Event    string `json:"event"`
	ItemName string `json:"item_name"`
}

// LoadGame loads a game from JSON data
func LoadGame(data json.RawMessage) (*world.Level, error) {
	// Sanity check the JSON structure first
	if err := validateJSONStructure(data); err != nil {
		return nil, fmt.Errorf("JSON structure validation failed: %w", err)
	}

	var gameData GameData
	if err := json.Unmarshal(data, &gameData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Create rooms map for easy lookup across all floors
	roomsMap := make(map[string]*world.Room)

	// Create floors
	var floors []*world.Floor

	// Handle new floors format
	if len(gameData.Floors) > 0 {
		for _, floorData := range gameData.Floors {
			floor := &world.Floor{
				Name:        floorData.Name,
				Description: floorData.Description,
				Rooms:       []*world.Room{},
			}

			// First pass: create all rooms for this floor
			for _, roomData := range floorData.Rooms {
				room := &world.Room{
					BaseEntity: world.BaseEntity{
						Name:        roomData.Name,
						Description: roomData.Description,
					},
					Connections: []*world.Connection{},
					Items:       []*world.Item{},
				}
				roomsMap[roomData.Name] = room
				floor.Rooms = append(floor.Rooms, room)
			}

			floors = append(floors, floor)
		}
	} else {
		// Handle old rooms format for backward compatibility
		floor := &world.Floor{
			Name:        "main floor",
			Description: "the main floor",
			Rooms:       []*world.Room{},
		}

		// First pass: create all rooms
		for _, roomData := range gameData.Rooms {
			room := &world.Room{
				BaseEntity: world.BaseEntity{
					Name:        roomData.Name,
					Description: roomData.Description,
				},
				Connections: []*world.Connection{},
				Items:       []*world.Item{},
			}
			roomsMap[roomData.Name] = room
			floor.Rooms = append(floor.Rooms, room)
		}

		floors = append(floors, floor)
	}

	// Create doors map for easy lookup
	doorsMap := make(map[string]*world.Door)

	// Second pass: create all doors
	for _, doorData := range gameData.DoorData {
		var lock *world.Lock
		if doorData.Locked {
			lock = &world.Lock{
				Locked:  true,
				KeyName: doorData.RequiredKeyName,
				Code:    doorData.Code,
			}
		}

		var latch *world.Latch
		if doorData.LatchedFrom != "" {
			latch = &world.Latch{
				Locked:     true,
				LockedFrom: doorData.LatchedFrom,
			}
		}

		door := &world.Door{
			Name:      doorData.Name,
			RoomA:     doorData.RoomA,
			RoomB:     doorData.RoomB,
			Lock:      lock,
			Stairwell: doorData.Stairwell,
			Latch:     latch,
		}
		doorsMap[doorData.Name] = door
	}

	// Third pass: populate room connections and items
	if len(gameData.Floors) > 0 {
		// Handle new floors format
		for _, floorData := range gameData.Floors {
			for _, roomData := range floorData.Rooms {
				room := roomsMap[roomData.Name]

				// Add connections
				for _, conn := range roomData.Connections {
					if _, exists := doorsMap[conn.DoorName]; exists {
						connection := &world.Connection{
							DoorName: conn.DoorName,
							Location: conn.Location,
						}
						room.Connections = append(room.Connections, connection)
					}
				}

				// Add items
				for _, itemData := range roomData.Items {
					item, err := createItem(itemData)
					if err != nil {
						return nil, fmt.Errorf("failed to create item %s: %w", itemData.Name, err)
					}
					room.Items = append(room.Items, item)
				}
			}
		}
	} else {
		// Handle old rooms format for backward compatibility
		for _, roomData := range gameData.Rooms {
			room := roomsMap[roomData.Name]

			// Add connections
			for _, conn := range roomData.Connections {
				if _, exists := doorsMap[conn.DoorName]; exists {
					connection := &world.Connection{
						DoorName: conn.DoorName,
						Location: conn.Location,
					}
					room.Connections = append(room.Connections, connection)
				}
			}

			// Add items
			for _, itemData := range roomData.Items {
				item, err := createItem(itemData)
				if err != nil {
					return nil, fmt.Errorf("failed to create item %s: %w", itemData.Name, err)
				}
				room.Items = append(room.Items, item)
			}
		}
	}

	// Create enemies
	var enemies []*world.Enemy
	for _, enemyData := range gameData.Enemies {
		var triggerEvent world.TriggerEvent
		if enemyData.Trigger != nil {
			switch enemyData.Trigger.Event {
			case "take_item":
				triggerEvent = world.TriggerEventTakeItem
			case "enter_room":
				triggerEvent = world.TriggerEventEnterRoom
			}
		}

		enemy := &world.Enemy{
			BaseEntity: world.BaseEntity{
				Name:        enemyData.Name,
				Description: enemyData.Description,
			},
			HP:           enemyData.HP,
			TriggerEvent: triggerEvent,
		}
		enemies = append(enemies, enemy)
	}

	// Create win condition event
	var winCondition *world.Event
	if gameData.WinCondition != nil {
		var eventType world.EventType
		switch gameData.WinCondition.Event {
		case "enter_room":
			eventType = world.EventRoomEntered
		}

		winCondition = &world.Event{
			Event:    eventType,
			RoomName: gameData.WinCondition.RoomName,
		}
	}

	// Create triggers
	var triggers []*world.Trigger
	for _, enemyData := range gameData.Enemies {
		if enemyData.Trigger != nil {
			var eventType world.EventType
			switch enemyData.Trigger.Event {
			case "take_item":
				eventType = world.EventItemTaken
			case "enter_room":
				eventType = world.EventRoomEntered
			}

			trigger := world.Trigger{
				Event: world.Event{
					Event:    eventType,
					ItemName: enemyData.Trigger.ItemName,
				},
				Effect: world.Effect{
					EffectType: world.EffectEnterCombat,
					EnemyName:  enemyData.Name,
				},
			}
			triggers = append(triggers, &trigger)
		}
	}

	// Convert rooms map to slice, preserving original order
	var rooms []*world.Room
	if len(gameData.Floors) > 0 {
		// Handle new floors format
		for _, floorData := range gameData.Floors {
			for _, roomData := range floorData.Rooms {
				room := roomsMap[roomData.Name]
				rooms = append(rooms, room)
			}
		}
	} else {
		// Handle old rooms format for backward compatibility
		for _, roomData := range gameData.Rooms {
			room := roomsMap[roomData.Name]
			rooms = append(rooms, room)
		}
	}

	// Convert doors map to slice
	var doors []*world.Door
	for _, door := range doorsMap {
		doors = append(doors, door)
	}

	// Create combo items
	var comboItems []*world.ComboItem
	for _, comboItemData := range gameData.ComboItems {
		outputItem, err := createItem(comboItemData.OutputItem)
		if err != nil {
			return nil, fmt.Errorf("failed to create combo item output %s: %w", comboItemData.OutputItem.Name, err)
		}

		comboItem := &world.ComboItem{
			InputItemAName: comboItemData.InputItemAName,
			InputItemBName: comboItemData.InputItemBName,
			OutputItem:     outputItem,
		}
		comboItems = append(comboItems, comboItem)
	}

	// Create level
	level := &world.Level{
		Name:         gameData.Name,
		Floors:       floors,
		Doors:        doors,
		Enemies:      enemies,
		Triggers:     triggers,
		WinCondition: winCondition,
		ComboItems:   comboItems,
	}

	// Validate reachability
	if err := validateReachability(level); err != nil {
		return nil, fmt.Errorf("reachability validation failed: %w", err)
	}

	return level, nil
}

// LoadGameFromFile loads a game from a JSON or YAML file
func LoadGameFromFile(filename string) (*world.Level, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Check if file is YAML based on extension
	if strings.HasSuffix(strings.ToLower(filename), ".yaml") || strings.HasSuffix(strings.ToLower(filename), ".yml") {
		// Convert YAML to JSON
		var v any
		if err := yaml.Unmarshal(data, &v); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}

		// Convert to JSON
		jsonData, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
		}

		return LoadGame(jsonData)
	}

	// Treat as JSON
	return LoadGame(data)
}

// validateJSONStructure performs a sanity check on the JSON input structure
func validateJSONStructure(data json.RawMessage) error {
	// Parse the JSON into a generic map to check structure
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// Check for required fields
	requiredFields := []string{"name"}
	for _, field := range requiredFields {
		if _, exists := jsonMap[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// Check for either floors or rooms (for backward compatibility)
	if _, hasFloors := jsonMap["floors"]; !hasFloors {
		if _, hasRooms := jsonMap["rooms"]; !hasRooms {
			return fmt.Errorf("missing required field: either 'floors' or 'rooms'")
		}
	}

	// Check for optional fields (these are allowed but not required)
	optionalFields := []string{"win_condition", "doors", "enemies", "system_prompt_theme", "combo_items"}

	// Check for any unexpected fields
	allowedFields := make(map[string]bool)
	for _, field := range requiredFields {
		allowedFields[field] = true
	}
	allowedFields["floors"] = true
	allowedFields["rooms"] = true // For backward compatibility
	for _, field := range optionalFields {
		allowedFields[field] = true
	}

	for field := range jsonMap {
		if !allowedFields[field] {
			return fmt.Errorf("unexpected field: %s (allowed fields: %v)", field, getSortedKeys(allowedFields))
		}
	}

	// Validate that required fields have the correct types
	if name, ok := jsonMap["name"].(string); !ok || name == "" {
		return fmt.Errorf("field 'name' must be a non-empty string")
	}

	// Validate floors field if present
	if floors, ok := jsonMap["floors"].([]interface{}); ok {
		if len(floors) == 0 {
			return fmt.Errorf("field 'floors' must be a non-empty array")
		}
	}

	// Validate rooms field if present (for backward compatibility)
	if rooms, ok := jsonMap["rooms"].([]interface{}); ok {
		if len(rooms) == 0 {
			return fmt.Errorf("field 'rooms' must be a non-empty array")
		}
	}

	return nil
}

// getSortedKeys returns the keys of a map as a sorted slice
func getSortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// validateReachability ensures that all rooms in the level are reachable from each other
// by performing a breadth-first search starting from the first room
func validateReachability(level *world.Level) error {
	if len(level.Floors) == 0 {
		return fmt.Errorf("level has no floors")
	}

	// Collect all rooms from all floors
	var allRooms []*world.Room
	for _, floor := range level.Floors {
		if len(floor.Rooms) == 0 {
			return fmt.Errorf("floor %s has no rooms", floor.Name)
		}
		allRooms = append(allRooms, floor.Rooms...)
	}

	// Use BFS to find all reachable rooms
	visited := make(map[string]bool)
	queue := []string{allRooms[0].Name} // Start from the first room
	visited[allRooms[0].Name] = true

	for len(queue) > 0 {
		currentRoomName := queue[0]
		queue = queue[1:]

		// Find the current room
		var currentRoom *world.Room
		for _, room := range allRooms {
			if room.Name == currentRoomName {
				currentRoom = room
				break
			}
		}

		if currentRoom == nil {
			return fmt.Errorf("room %s not found in level", currentRoomName)
		}

		// Check all connections from this room
		for _, conn := range currentRoom.Connections {
			// Find the actual door in the level
			var door *world.Door
			for _, d := range level.Doors {
				if d.Name == conn.DoorName {
					door = d
					break
				}
			}
			if door == nil {
				continue
			}

			// Determine the other room connected by this door
			var otherRoomName string
			if door.RoomA == currentRoomName {
				otherRoomName = door.RoomB
			} else {
				otherRoomName = door.RoomA
			}

			// If we haven't visited this room yet, add it to the queue
			if !visited[otherRoomName] {
				visited[otherRoomName] = true
				queue = append(queue, otherRoomName)
			}
		}
	}

	// Check if all rooms were visited
	var unreachableRooms []string
	for _, room := range allRooms {
		if !visited[room.Name] {
			unreachableRooms = append(unreachableRooms, room.Name)
		}
	}

	if len(unreachableRooms) > 0 {
		return fmt.Errorf("unreachable rooms found: %v", unreachableRooms)
	}

	return nil
}

// createItem recursively creates an item and its nested items
func createItem(itemData ItemData) (*world.Item, error) {
	item := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        itemData.Name,
			Description: itemData.Description,
		},
		Location: itemData.Location,
		Detail:   itemData.Detail,
	}

	// Handle portable items
	if itemData.Portable {
		item.Portable = &world.Portable{}
	}

	// Handle keys
	if itemData.Key {
		item.Key = &world.Key{}
		// Keys are always portable
		if item.Portable == nil {
			item.Portable = &world.Portable{}
		}
	}

	// Handle weapons
	if itemData.WeaponDamage > 0 {
		var ammo *world.Ammo
		if itemData.Ammo > 0 {
			ammo = &world.Ammo{
				Quantity: itemData.Ammo,
			}
		}

		item.Weapon = &world.Weapon{
			Damage: itemData.WeaponDamage,
			Ammo:   ammo,
		}
		// Weapons are always portable
		if item.Portable == nil {
			item.Portable = &world.Portable{}
		}
	}

	// Handle health items
	if itemData.HealthEffect != "" {
		var healthEffect world.HealthEffect
		switch itemData.HealthEffect {
		case "weak":
			healthEffect = world.HealthBoostWeak
		case "strong":
			healthEffect = world.HealthBoostStrong
		}

		item.HealthItem = &world.HealthItem{
			HealthEffect: healthEffect,
		}
		// Health items are always portable
		if item.Portable == nil {
			item.Portable = &world.Portable{}
		}
	}

	// Handle ammo boxes
	if itemData.WeaponName != "" && itemData.Ammo > 0 {
		item.AmmoBox = &world.AmmoBox{
			WeaponName: itemData.WeaponName,
			Ammo: &world.Ammo{
				Quantity: itemData.Ammo,
			},
		}
		// Ammo boxes are always portable
		if item.Portable == nil {
			item.Portable = &world.Portable{}
		}
	}

	// Handle containers
	if itemData.Contains != nil {
		var contains *world.Item
		if itemData.Contains.Empty {
			// Empty container
			contains = nil
		} else if itemData.Contains.Item != nil {
			// Container with item
			var err error
			contains, err = createItem(*itemData.Contains.Item)
			if err != nil {
				return nil, fmt.Errorf("failed to create contained item: %w", err)
			}
		}

		var lock *world.Lock
		if itemData.Code != "" {
			lock = &world.Lock{
				Locked: true,
				Code:   itemData.Code,
			}
		}

		item.Container = &world.Container{
			Contains: contains,
			Searched: false,
			Locked:   lock,
		}
	}

	// Handle concealers
	if itemData.Conceals != nil {
		hidden, err := createItem(*itemData.Conceals)
		if err != nil {
			return nil, fmt.Errorf("failed to create concealed item: %w", err)
		}
		item.Concealer = &world.Concealer{
			Hidden:    hidden,
			Uncovered: false,
		}
	}

	// Handle fixtures
	if itemData.Fixture != nil {
		// Convert the list of required items to a map with all items initially false
		requiredItems := make(map[string]bool)
		for _, itemName := range itemData.Fixture.RequiredItems {
			requiredItems[itemName] = false
		}

		// Create the produced item if specified
		var producedItem *world.Item
		if itemData.Fixture.Produces != nil {
			var err error
			producedItem, err = createItem(*itemData.Fixture.Produces)
			if err != nil {
				return nil, fmt.Errorf("failed to create produced item for fixture %s: %w", itemData.Name, err)
			}
		}

		item.Fixture = &world.Fixture{
			RequiredItems: requiredItems,
			Produces:      producedItem,
		}
	}

	// Validate the item's initial state
	if err := item.ValidateInitialState(); err != nil {
		return nil, fmt.Errorf("invalid item %s: %w", item.Name, err)
	}

	return item, nil
}
