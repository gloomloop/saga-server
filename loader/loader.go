package loader

import (
	"encoding/json"
	"fmt"
	"os"

	"adventure-engine/world"
)

// GameData represents the top-level JSON structure
type GameData struct {
	Name              string      `json:"name"`
	SystemPromptTheme string      `json:"system_prompt_theme"`
	WinCondition      *EventData  `json:"win_condition"`
	Rooms             []RoomData  `json:"rooms"`
	DoorData          []DoorData  `json:"doors"`
	Enemies           []EnemyData `json:"enemies"`
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
	Direction string `json:"direction"`
	DoorName  string `json:"door_name"`
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
}

// DoorData represents a door in the JSON
type DoorData struct {
	Name            string `json:"name"`
	RoomA           string `json:"room_a"`
	RoomB           string `json:"room_b"`
	Locked          bool   `json:"locked,omitempty"`
	RequiredKeyName string `json:"required_key_name,omitempty"`
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

// LoadGame loads a game from a JSON file
func LoadGame(filename string) (*world.Level, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var gameData GameData
	if err := json.Unmarshal(data, &gameData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Create rooms map for easy lookup
	roomsMap := make(map[string]*world.Room)

	// First pass: create all rooms
	for _, roomData := range gameData.Rooms {
		room := &world.Room{
			BaseEntity: world.BaseEntity{
				Name:        roomData.Name,
				Description: roomData.Description,
			},
			Connections: make(map[string]*world.Door),
			Items:       []*world.Item{},
		}
		roomsMap[roomData.Name] = room
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
			}
		}

		door := &world.Door{
			BaseEntity: world.BaseEntity{
				Name:        doorData.Name,
				Description: fmt.Sprintf("a door leading from %s to %s", doorData.RoomA, doorData.RoomB),
			},
			RoomA: doorData.RoomA,
			RoomB: doorData.RoomB,
			Lock:  lock,
		}
		doorsMap[doorData.Name] = door
	}

	// Third pass: populate room connections and items
	for _, roomData := range gameData.Rooms {
		room := roomsMap[roomData.Name]

		// Add connections
		for _, conn := range roomData.Connections {
			if door, exists := doorsMap[conn.DoorName]; exists {
				room.Connections[conn.Direction] = door
			}
		}

		// Add items
		for _, itemData := range roomData.Items {
			item := createItem(itemData)
			room.Items = append(room.Items, item)
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
			Room:         enemyData.Room,
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
	for _, roomData := range gameData.Rooms {
		room := roomsMap[roomData.Name]
		rooms = append(rooms, room)
	}

	// Convert doors map to slice
	var doors []*world.Door
	for _, door := range doorsMap {
		doors = append(doors, door)
	}

	// Create level
	level := &world.Level{
		Name:         gameData.Name,
		Rooms:        rooms,
		Doors:        doors,
		Enemies:      enemies,
		Triggers:     triggers,
		WinCondition: winCondition,
	}

	return level, nil
}

// createItem recursively creates an item and its nested items
func createItem(itemData ItemData) *world.Item {
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
			contains = createItem(*itemData.Contains.Item)
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
		hidden := createItem(*itemData.Conceals)
		item.Concealer = &world.Concealer{
			Hidden:    hidden,
			Uncovered: false,
		}
	}

	return item
}
