package loader

import (
	"encoding/json"
	"strings"
	"testing"

	world "adventure-engine/internal/world"
)

func TestLoadGame_Demo(t *testing.T) {
	// Load the demo puzzle game
	level, err := LoadGameFromFile("../testdata/demo.json")
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Test basic game properties
	if level.Name != "demo puzzle" {
		t.Errorf("Expected game name 'demo puzzle', got '%s'", level.Name)
	}

	// Test intro and outro narrative (should be empty for demo.json)
	if level.IntroNarrative != "" {
		t.Errorf("Expected intro narrative to be empty, got '%s'", level.IntroNarrative)
	}
	if level.OutroNarrative != "" {
		t.Errorf("Expected outro narrative to be empty, got '%s'", level.OutroNarrative)
	}

	// Test rooms
	allRooms := getAllRooms(level)
	if len(allRooms) != 4 {
		t.Errorf("Expected 4 rooms, got %d", len(allRooms))
	}

	// Test doors
	if len(level.Doors) != 3 {
		t.Errorf("Expected 3 doors, got %d", len(level.Doors))
	}

	// Test enemies
	if len(level.Enemies) != 1 {
		t.Errorf("Expected 1 enemy, got %d", len(level.Enemies))
	}

	// Test triggers
	if len(level.Triggers) != 1 {
		t.Errorf("Expected 1 trigger, got %d", len(level.Triggers))
	}

	// Test win condition
	if level.WinCondition == nil {
		t.Error("Expected win condition to be set")
	} else {
		if level.WinCondition.Event != world.EventRoomEntered {
			t.Errorf("Expected win condition event to be EventRoomEntered, got %s", level.WinCondition.Event)
		}
		if level.WinCondition.RoomName != "stairwell to roof" {
			t.Errorf("Expected win condition room to be 'stairwell to roof', got '%s'", level.WinCondition.RoomName)
		}
	}

	// Test specific room: waiting room
	waitingRoom := findRoomByName(allRooms, "waiting room")
	if waitingRoom == nil {
		t.Fatal("Could not find waiting room")
	}

	if waitingRoom.Description != "a dilapidated waiting room" {
		t.Errorf("Expected waiting room description 'a dilapidated waiting room', got '%s'", waitingRoom.Description)
	}

	// Test waiting room connections
	if len(waitingRoom.Connections) != 3 {
		t.Errorf("Expected waiting room to have 3 connections, got %d", len(waitingRoom.Connections))
	}

	// Test waiting room items
	if len(waitingRoom.Items) != 2 {
		t.Errorf("Expected waiting room to have 2 items, got %d", len(waitingRoom.Items))
	}

	// Test specific item: tattered grey hoodie (concealer)
	hoodie := findItemByName(waitingRoom.Items, "tattered grey hoodie")
	if hoodie == nil {
		t.Fatal("Could not find tattered grey hoodie")
	}

	if !hoodie.IsConcealer() {
		t.Error("Expected tattered grey hoodie to be a concealer")
	}

	if hoodie.Concealer.Hidden == nil {
		t.Error("Expected concealer to have hidden item")
	} else {
		if hoodie.Concealer.Hidden.Name != "ominous note" {
			t.Errorf("Expected hidden item to be 'ominous note', got '%s'", hoodie.Concealer.Hidden.Name)
		}
		if !hoodie.Concealer.Hidden.IsPortable() {
			t.Error("Expected hidden item to be portable")
		}
	}

	// Test specific item: energy drink (health item)
	energyDrink := findItemByName(waitingRoom.Items, "energy drink")
	if energyDrink == nil {
		t.Fatal("Could not find energy drink")
	}

	if !energyDrink.IsHealthItem() {
		t.Error("Expected energy drink to be a health item")
	}

	if energyDrink.HealthItem.HealthEffect != world.HealthBoostWeak {
		t.Errorf("Expected energy drink to have weak health effect, got %s", energyDrink.HealthItem.HealthEffect)
	}

	// Test specific room: storage room
	storageRoom := findRoomByName(allRooms, "storage room")
	if storageRoom == nil {
		t.Fatal("Could not find storage room")
	}

	// Test storage room items
	if len(storageRoom.Items) != 4 {
		t.Errorf("Expected storage room to have 4 items, got %d", len(storageRoom.Items))
	}

	// Test specific item: dark green tarp (concealer with container)
	tarp := findItemByName(storageRoom.Items, "dark green tarp")
	if tarp == nil {
		t.Fatal("Could not find dark green tarp")
	}

	if !tarp.IsConcealer() {
		t.Error("Expected dark green tarp to be a concealer")
	}

	if tarp.Concealer.Hidden == nil {
		t.Error("Expected concealer to have hidden item")
	} else {
		if tarp.Concealer.Hidden.Name != "safe" {
			t.Errorf("Expected hidden item to be 'safe', got '%s'", tarp.Concealer.Hidden.Name)
		}
		if !tarp.Concealer.Hidden.IsContainer() {
			t.Error("Expected hidden item to be a container")
		}
		if !tarp.Concealer.Hidden.Container.HasCodeLock() {
			t.Error("Expected container to have code lock")
		}
		if tarp.Concealer.Hidden.Container.Locked.Code != "2468" {
			t.Errorf("Expected container code to be '2468', got '%s'", tarp.Concealer.Hidden.Container.Locked.Code)
		}
		if tarp.Concealer.Hidden.Container.Contains == nil {
			t.Error("Expected container to have contents")
		} else {
			if tarp.Concealer.Hidden.Container.Contains.Name != "iron key" {
				t.Errorf("Expected container contents to be 'iron key', got '%s'", tarp.Concealer.Hidden.Container.Contains.Name)
			}
			if !tarp.Concealer.Hidden.Container.Contains.IsKey() {
				t.Error("Expected container contents to be a key")
			}
		}
	}

	// Test specific item: filing cabinet (empty container)
	filingCabinet := findItemByName(storageRoom.Items, "filing cabinet")
	if filingCabinet == nil {
		t.Fatal("Could not find filing cabinet")
	}

	if !filingCabinet.IsContainer() {
		t.Error("Expected filing cabinet to be a container")
	}

	if filingCabinet.Container.Contains != nil {
		t.Error("Expected filing cabinet to be empty")
	}

	// Test specific item: metal pipe (weapon)
	metalPipe := findItemByName(storageRoom.Items, "metal pipe")
	if metalPipe == nil {
		t.Fatal("Could not find metal pipe")
	}

	if !metalPipe.IsWeapon() {
		t.Error("Expected metal pipe to be a weapon")
	}

	if metalPipe.Weapon.Damage != 0.7 {
		t.Errorf("Expected metal pipe damage to be 0.7, got %f", metalPipe.Weapon.Damage)
	}

	// Test specific room: office
	office := findRoomByName(allRooms, "office")
	if office == nil {
		t.Fatal("Could not find office")
	}

	// Test office items
	if len(office.Items) != 2 {
		t.Errorf("Expected office to have 2 items, got %d", len(office.Items))
	}

	// Test specific item: desk (container with weapon)
	desk := findItemByName(office.Items, "desk")
	if desk == nil {
		t.Fatal("Could not find desk")
	}

	if !desk.IsContainer() {
		t.Error("Expected desk to be a container")
	}

	if desk.Container.Contains == nil {
		t.Error("Expected desk to have contents")
	} else {
		if desk.Container.Contains.Name != "pistol" {
			t.Errorf("Expected desk contents to be 'pistol', got '%s'", desk.Container.Contains.Name)
		}
		if !desk.Container.Contains.IsWeapon() {
			t.Error("Expected desk contents to be a weapon")
		}
		if desk.Container.Contains.Weapon.Damage != 0.9 {
			t.Errorf("Expected pistol damage to be 0.9, got %f", desk.Container.Contains.Weapon.Damage)
		}
		if !desk.Container.Contains.Weapon.UsesAmmo() {
			t.Error("Expected pistol to use ammo")
		}
		if desk.Container.Contains.Weapon.Ammo.Quantity != 1 {
			t.Errorf("Expected pistol to have 1 ammo, got %d", desk.Container.Contains.Weapon.Ammo.Quantity)
		}
	}

	// Test specific item: cardboard box (container with ammo box)
	cardboardBox := findItemByName(office.Items, "cardboard box")
	if cardboardBox == nil {
		t.Fatal("Could not find cardboard box")
	}

	if !cardboardBox.IsContainer() {
		t.Error("Expected cardboard box to be a container")
	}

	if cardboardBox.Container.Contains == nil {
		t.Error("Expected cardboard box to have contents")
	} else {
		if cardboardBox.Container.Contains.Name != "pistol ammo" {
			t.Errorf("Expected cardboard box contents to be 'pistol ammo', got '%s'", cardboardBox.Container.Contains.Name)
		}
		if !cardboardBox.Container.Contains.IsAmmoBox() {
			t.Error("Expected cardboard box contents to be an ammo box")
		}
		if cardboardBox.Container.Contains.AmmoBox.WeaponName != "pistol" {
			t.Errorf("Expected ammo box weapon name to be 'pistol', got '%s'", cardboardBox.Container.Contains.AmmoBox.WeaponName)
		}
		if cardboardBox.Container.Contains.AmmoBox.Ammo.Quantity != 2 {
			t.Errorf("Expected ammo box to have 2 ammo, got %d", cardboardBox.Container.Contains.AmmoBox.Ammo.Quantity)
		}
	}

	// Test specific room: stairwell to roof
	stairwell := findRoomByName(allRooms, "stairwell to roof")
	if stairwell == nil {
		t.Fatal("Could not find stairwell to roof")
	}

	if stairwell.Description != "a way out" {
		t.Errorf("Expected stairwell description 'a way out', got '%s'", stairwell.Description)
	}

	// Test doors
	storageDoor := findDoorByName(level.Doors, "storage room door")
	if storageDoor == nil {
		t.Fatal("Could not find storage room door")
	}

	if storageDoor.RoomA != "waiting room" || storageDoor.RoomB != "storage room" {
		t.Errorf("Expected storage door to connect waiting room and storage room, got %s and %s", storageDoor.RoomA, storageDoor.RoomB)
	}

	metalDoor := findDoorByName(level.Doors, "metal stairwell door")
	if metalDoor == nil {
		t.Fatal("Could not find metal stairwell door")
	}

	if !metalDoor.IsLocked() {
		t.Error("Expected metal stairwell door to be locked")
	}

	if !metalDoor.HasKeyLock() {
		t.Error("Expected metal stairwell door to have key lock")
	}

	if metalDoor.Lock.KeyName != "iron key" {
		t.Errorf("Expected metal door to require iron key, got '%s'", metalDoor.Lock.KeyName)
	}

	// Test enemy
	zombie := findEnemyByName(level.Enemies, "zombie")
	if zombie == nil {
		t.Fatal("Could not find zombie")
	}

	if zombie.Description != "a wailing zombie" {
		t.Errorf("Expected zombie description 'a wailing zombie', got '%s'", zombie.Description)
	}

	if zombie.HP != 1 {
		t.Errorf("Expected zombie HP to be 1, got %d", zombie.HP)
	}

	// Test trigger
	if len(level.Triggers) != 1 {
		t.Errorf("Expected 1 trigger, got %d", len(level.Triggers))
	} else {
		trigger := level.Triggers[0]
		if trigger.Event.Event != world.EventItemTaken {
			t.Errorf("Expected trigger event to be EventItemTaken, got %s", trigger.Event.Event)
		}
		if trigger.Event.ItemName != "iron key" {
			t.Errorf("Expected trigger item name to be 'iron key', got '%s'", trigger.Event.ItemName)
		}
		if trigger.Effect.EffectType != world.EffectEnterCombat {
			t.Errorf("Expected trigger effect to be EffectEnterCombat, got %s", trigger.Effect.EffectType)
		}
		if trigger.Effect.EnemyName != "zombie" {
			t.Errorf("Expected trigger enemy name to be 'zombie', got '%s'", trigger.Effect.EnemyName)
		}
	}
}

func TestLoadGame_FromRawMessage(t *testing.T) {
	// Test JSON data as raw message
	jsonData := json.RawMessage(`{
		"name": "test game",
		"floors": [
			{
				"name": "test floor",
				"description": "a test floor",
				"rooms": [
					{
						"name": "test room",
						"description": "a test room",
						"items": [
							{
								"name": "test item",
								"description": "a test item",
								"portable": true
							}
						]
					}
				]
			}
		],
		"doors": [],
		"enemies": []
	}`)

	level, err := LoadGame(jsonData)
	if err != nil {
		t.Fatalf("Failed to load game from raw message: %v", err)
	}

	// Test basic properties
	if level.Name != "test game" {
		t.Errorf("Expected game name 'test game', got '%s'", level.Name)
	}

	allRooms := getAllRooms(level)
	if len(allRooms) != 1 {
		t.Errorf("Expected 1 room, got %d", len(allRooms))
	}

	if len(allRooms[0].Items) != 1 {
		t.Errorf("Expected 1 item in room, got %d", len(allRooms[0].Items))
	}

	if allRooms[0].Items[0].Name != "test item" {
		t.Errorf("Expected item name 'test item', got '%s'", allRooms[0].Items[0].Name)
	}
}

func TestLoadGame_ValidationErrors(t *testing.T) {
	// Test invalid item configurations that should fail validation

	// Test 1: Key that is also a container (should fail)
	jsonData1 := json.RawMessage(`{
		"name": "test game",
		"floors": [
			{
				"name": "test floor",
				"description": "a test floor",
				"rooms": [
					{
						"name": "test room",
						"description": "a test room",
						"items": [
							{
								"name": "invalid key",
								"description": "a key that is also a container",
								"key": true,
								"contains": "empty"
							}
						]
					}
				]
			}
		],
		"doors": [],
		"enemies": []
	}`)

	_, err := LoadGame(jsonData1)
	if err == nil {
		t.Error("Expected error for key that is also a container, got nil")
	} else if !strings.Contains(err.Error(), "invalid key") {
		t.Errorf("Expected error about invalid key, got: %v", err)
	}

	// Test 2: Weapon that is also a container (should fail)
	jsonData2 := json.RawMessage(`{
		"name": "test game",
		"floors": [
			{
				"name": "test floor",
				"description": "a test floor",
				"rooms": [
					{
						"name": "test room",
						"description": "a test room",
						"items": [
							{
								"name": "invalid weapon",
								"description": "a weapon that is also a container",
								"weapon_damage": 0.8,
								"contains": "empty"
							}
						]
					}
				]
			}
		],
		"doors": [],
		"enemies": []
	}`)

	_, err = LoadGame(jsonData2)
	if err == nil {
		t.Error("Expected error for weapon that is also a container, got nil")
	} else if !strings.Contains(err.Error(), "invalid weapon") {
		t.Errorf("Expected error about invalid weapon, got: %v", err)
	}

	// Test 3: Container that is also a key (should fail)
	jsonData3 := json.RawMessage(`{
		"name": "test game",
		"floors": [
			{
				"name": "test floor",
				"description": "a test floor",
				"rooms": [
					{
						"name": "test room",
						"description": "a test room",
						"items": [
							{
								"name": "invalid container",
								"description": "a container that is also a key",
								"key": true,
								"contains": "empty"
							}
						]
					}
				]
			}
		],
		"doors": [],
		"enemies": []
	}`)

	_, err = LoadGame(jsonData3)
	if err == nil {
		t.Error("Expected error for container that is also a key, got nil")
	} else if !strings.Contains(err.Error(), "invalid container") {
		t.Errorf("Expected error about invalid container, got: %v", err)
	}

	// Test 4: Nested containers (should fail)
	jsonData4 := json.RawMessage(`{
		"name": "test game",
		"rooms": [
			{
				"name": "test room",
				"description": "a test room",
				"items": [
					{
						"name": "outer container",
						"description": "a container with another container inside",
						"contains": {
							"name": "inner container",
							"description": "a container inside another container",
							"contains": "empty"
						}
					}
				]
			}
		],
		"doors": [],
		"enemies": []
	}`)

	_, err = LoadGame(jsonData4)
	if err == nil {
		t.Error("Expected error for nested containers, got nil")
	} else if !strings.Contains(err.Error(), "container cannot be nested") {
		t.Errorf("Expected error about nested containers, got: %v", err)
	}
}

func TestLoadGame_ReachabilityValidation(t *testing.T) {
	// Test 1: Valid connected level (should pass)
	jsonData1 := json.RawMessage(`{
		"name": "connected level",
		"rooms": [
			{
				"name": "room1",
				"description": "first room",
				"connections": [
					{
						"direction": "east",
						"door_name": "door1"
					}
				]
			},
			{
				"name": "room2",
				"description": "second room",
				"connections": [
					{
						"direction": "west",
						"door_name": "door1"
					}
				]
			}
		],
		"doors": [
			{
				"name": "door1",
				"room_a": "room1",
				"room_b": "room2"
			}
		],
		"enemies": []
	}`)

	_, err := LoadGame(jsonData1)
	if err != nil {
		t.Errorf("Expected connected level to pass validation, got error: %v", err)
	}

	// Test 2: Disconnected level (should fail)
	jsonData2 := json.RawMessage(`{
		"name": "disconnected level",
		"rooms": [
			{
				"name": "room1",
				"description": "first room"
			},
			{
				"name": "room2",
				"description": "second room"
			}
		],
		"doors": [],
		"enemies": []
	}`)

	_, err = LoadGame(jsonData2)
	if err == nil {
		t.Error("Expected disconnected level to fail validation, got nil error")
	} else if !strings.Contains(err.Error(), "unreachable rooms found") {
		t.Errorf("Expected error about unreachable rooms, got: %v", err)
	}

	// Test 3: Level with isolated room (should fail)
	jsonData3 := json.RawMessage(`{
		"name": "level with isolated room",
		"rooms": [
			{
				"name": "room1",
				"description": "first room",
				"connections": [
					{
						"direction": "east",
						"door_name": "door1"
					}
				]
			},
			{
				"name": "room2",
				"description": "second room",
				"connections": [
					{
						"direction": "west",
						"door_name": "door1"
					}
				]
			},
			{
				"name": "isolated_room",
				"description": "room with no connections"
			}
		],
		"doors": [
			{
				"name": "door1",
				"room_a": "room1",
				"room_b": "room2"
			}
		],
		"enemies": []
	}`)

	_, err = LoadGame(jsonData3)
	if err == nil {
		t.Error("Expected level with isolated room to fail validation, got nil error")
	} else if !strings.Contains(err.Error(), "unreachable rooms found") {
		t.Errorf("Expected error about unreachable rooms, got: %v", err)
	} else if !strings.Contains(err.Error(), "isolated_room") {
		t.Errorf("Expected error to mention isolated_room, got: %v", err)
	}

	// Test 4: Complex connected level (should pass)
	jsonData4 := json.RawMessage(`{
		"name": "complex connected level",
		"rooms": [
			{
				"name": "room1",
				"description": "first room",
				"connections": [
					{
						"direction": "east",
						"door_name": "door1"
					},
					{
						"direction": "south",
						"door_name": "door3"
					}
				]
			},
			{
				"name": "room2",
				"description": "second room",
				"connections": [
					{
						"direction": "west",
						"door_name": "door1"
					},
					{
						"direction": "south",
						"door_name": "door2"
					}
				]
			},
			{
				"name": "room3",
				"description": "third room",
				"connections": [
					{
						"direction": "north",
						"door_name": "door2"
					},
					{
						"direction": "west",
						"door_name": "door3"
					}
				]
			}
		],
		"doors": [
			{
				"name": "door1",
				"room_a": "room1",
				"room_b": "room2"
			},
			{
				"name": "door2",
				"room_a": "room2",
				"room_b": "room3"
			},
			{
				"name": "door3",
				"room_a": "room1",
				"room_b": "room3"
			}
		],
		"enemies": []
	}`)

	_, err = LoadGame(jsonData4)
	if err != nil {
		t.Errorf("Expected complex connected level to pass validation, got error: %v", err)
	}
}

func TestLoadGame_JSONStructureValidation(t *testing.T) {
	// Test 1: Missing required field 'name' (should fail)
	jsonData1 := json.RawMessage(`{
		"rooms": [
			{
				"name": "test room",
				"description": "a test room"
			}
		],
		"doors": [],
		"enemies": []
	}`)

	_, err := LoadGame(jsonData1)
	if err == nil {
		t.Error("Expected error for missing 'name' field, got nil")
	} else if !strings.Contains(err.Error(), "missing required field: name") {
		t.Errorf("Expected error about missing name field, got: %v", err)
	}

	// Test 2: Missing required field 'rooms' (should fail)
	jsonData2 := json.RawMessage(`{
		"name": "test game",
		"doors": [],
		"enemies": []
	}`)

	_, err = LoadGame(jsonData2)
	if err == nil {
		t.Error("Expected error for missing 'rooms' field, got nil")
	} else if !strings.Contains(err.Error(), "missing required field: either 'floors' or 'rooms'") {
		t.Errorf("Expected error about missing rooms field, got: %v", err)
	}

	// Test 3: Empty name field (should fail)
	jsonData3 := json.RawMessage(`{
		"name": "",
		"rooms": [
			{
				"name": "test room",
				"description": "a test room"
			}
		],
		"doors": [],
		"enemies": []
	}`)

	_, err = LoadGame(jsonData3)
	if err == nil {
		t.Error("Expected error for empty name field, got nil")
	} else if !strings.Contains(err.Error(), "field 'name' must be a non-empty string") {
		t.Errorf("Expected error about empty name field, got: %v", err)
	}

	// Test 4: Empty rooms array (should fail)
	jsonData4 := json.RawMessage(`{
		"name": "test game",
		"rooms": [],
		"doors": [],
		"enemies": []
	}`)

	_, err = LoadGame(jsonData4)
	if err == nil {
		t.Error("Expected error for empty rooms array, got nil")
	} else if !strings.Contains(err.Error(), "field 'rooms' must be a non-empty array") {
		t.Errorf("Expected error about empty rooms array, got: %v", err)
	}

	// Test 5: Unexpected field (should fail)
	jsonData5 := json.RawMessage(`{
		"name": "test game",
		"rooms": [
			{
				"name": "test room",
				"description": "a test room"
			}
		],
		"doors": [],
		"enemies": [],
		"unexpected_field": "this should not be here"
	}`)

	_, err = LoadGame(jsonData5)
	if err == nil {
		t.Error("Expected error for unexpected field, got nil")
	} else if !strings.Contains(err.Error(), "unexpected field: unexpected_field") {
		t.Errorf("Expected error about unexpected field, got: %v", err)
	}

	// Test 6: Valid structure with all optional fields (should pass)
	jsonData6 := json.RawMessage(`{
		"name": "test game",
		"win_condition": {
			"event": "room_entered",
			"room_name": "test room"
		},
		"rooms": [
			{
				"name": "test room",
				"description": "a test room"
			}
		],
		"doors": [],
		"enemies": [],
		"system_prompt_theme": "adventure"
	}`)

	_, err = LoadGame(jsonData6)
	if err != nil {
		t.Errorf("Expected valid structure to pass validation, got error: %v", err)
	}

	// Test 7: Invalid JSON format (should fail)
	jsonData7 := json.RawMessage(`{
		"name": "test game",
		"rooms": [
			{
				"name": "test room",
				"description": "a test room"
			}
		],
		"doors": [],
		"enemies": [],
		"unclosed": {
	}`)

	_, err = LoadGame(jsonData7)
	if err == nil {
		t.Error("Expected error for invalid JSON format, got nil")
	} else if !strings.Contains(err.Error(), "invalid JSON format") {
		t.Errorf("Expected error about invalid JSON format, got: %v", err)
	}
}

func TestLoadGame_DemoYAML(t *testing.T) {
	// Load the demo puzzle game from YAML
	level, err := LoadGameFromFile("../testdata/demo.yaml")
	if err != nil {
		t.Fatalf("Failed to load game from YAML: %v", err)
	}

	// Test basic game properties
	if level.Name != "demo puzzle" {
		t.Errorf("Expected game name 'demo puzzle', got '%s'", level.Name)
	}

	// Test rooms
	allRooms := getAllRooms(level)
	if len(allRooms) != 4 {
		t.Errorf("Expected 4 rooms, got %d", len(allRooms))
	}

	// Test doors
	if len(level.Doors) != 3 {
		t.Errorf("Expected 3 doors, got %d", len(level.Doors))
	}

	// Test enemies
	if len(level.Enemies) != 1 {
		t.Errorf("Expected 1 enemy, got %d", len(level.Enemies))
	}

	// Test triggers
	if len(level.Triggers) != 1 {
		t.Errorf("Expected 1 trigger, got %d", len(level.Triggers))
	}

	// Test win condition
	if level.WinCondition == nil {
		t.Error("Expected win condition to be set")
	} else {
		if level.WinCondition.Event != world.EventRoomEntered {
			t.Errorf("Expected win condition event to be EventRoomEntered, got %s", level.WinCondition.Event)
		}
		if level.WinCondition.RoomName != "stairwell to roof" {
			t.Errorf("Expected win condition room to be 'stairwell to roof', got '%s'", level.WinCondition.RoomName)
		}
	}

	// Test specific room: waiting room
	waitingRoom := findRoomByName(allRooms, "waiting room")
	if waitingRoom == nil {
		t.Fatal("Could not find waiting room")
	}

	if waitingRoom.Description != "a dilapidated waiting room" {
		t.Errorf("Expected waiting room description 'a dilapidated waiting room', got '%s'", waitingRoom.Description)
	}

	// Test waiting room connections
	if len(waitingRoom.Connections) != 3 {
		t.Errorf("Expected waiting room to have 3 connections, got %d", len(waitingRoom.Connections))
	}

	// Test waiting room items
	if len(waitingRoom.Items) != 2 {
		t.Errorf("Expected waiting room to have 2 items, got %d", len(waitingRoom.Items))
	}
}

// Helper functions
func getAllRooms(level *world.Level) []*world.Room {
	var allRooms []*world.Room
	for _, floor := range level.Floors {
		allRooms = append(allRooms, floor.Rooms...)
	}
	return allRooms
}

func findRoomByName(rooms []*world.Room, name string) *world.Room {
	for _, room := range rooms {
		if room.Name == name {
			return room
		}
	}
	return nil
}

func findItemByName(items []*world.Item, name string) *world.Item {
	for _, item := range items {
		if item.Name == name {
			return item
		}
	}
	return nil
}

func findDoorByName(doors []*world.Door, name string) *world.Door {
	for _, door := range doors {
		if door.Name == name {
			return door
		}
	}
	return nil
}

func findEnemyByName(enemies []*world.Enemy, name string) *world.Enemy {
	for _, enemy := range enemies {
		if enemy.Name == name {
			return enemy
		}
	}
	return nil
}

func TestLoadGame_ComboItems(t *testing.T) {
	// Load the crafting test game
	level, err := LoadGameFromFile("../testdata/crafting.json")
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Test combo items
	if len(level.ComboItems) != 1 {
		t.Errorf("Expected 1 combo item, got %d", len(level.ComboItems))
	}

	combo := level.ComboItems[0]
	if combo.InputItemAName != "fish hook" {
		t.Errorf("Expected input item A to be 'fish hook', got '%s'", combo.InputItemAName)
	}
	if combo.InputItemBName != "dental floss" {
		t.Errorf("Expected input item B to be 'dental floss', got '%s'", combo.InputItemBName)
	}
	if combo.OutputItem.Name != "retrieval tool" {
		t.Errorf("Expected output item to be 'retrieval tool', got '%s'", combo.OutputItem.Name)
	}
}

func TestLoadGame_Fixtures(t *testing.T) {
	// Load the fixture test game
	level, err := LoadGameFromFile("../testdata/fixture.json")
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Test basic game properties
	if level.Name != "fixture test" {
		t.Errorf("Expected game name 'fixture test', got '%s'", level.Name)
	}

	// Test rooms
	allRooms := getAllRooms(level)
	if len(allRooms) != 3 {
		t.Errorf("Expected 3 rooms, got %d", len(allRooms))
	}

	// Test doors
	if len(level.Doors) != 2 {
		t.Errorf("Expected 2 doors, got %d", len(level.Doors))
	}

	// Test combo items
	if len(level.ComboItems) != 1 {
		t.Errorf("Expected 1 combo item, got %d", len(level.ComboItems))
	}

	// Test bathroom fixtures
	bathroom := findRoomByName(allRooms, "bathroom")
	if bathroom == nil {
		t.Fatal("Could not find bathroom")
	}

	// Test bathtub drain fixture
	bathtubDrain := findItemByName(bathroom.Items, "bathtub drain")
	if bathtubDrain == nil {
		t.Fatal("Could not find bathtub drain")
	}

	if !bathtubDrain.IsFixture() {
		t.Error("Expected bathtub drain to be a fixture")
	}

	if bathtubDrain.Fixture == nil {
		t.Fatal("Expected bathtub drain to have fixture data")
	}

	if len(bathtubDrain.Fixture.RequiredItems) != 1 {
		t.Errorf("Expected bathtub drain to require 1 item, got %d", len(bathtubDrain.Fixture.RequiredItems))
	}

	// Check that "retrieval tool" is in the required items map (value should be false initially)
	if _, exists := bathtubDrain.Fixture.RequiredItems["retrieval tool"]; !exists {
		t.Error("Expected bathtub drain to require 'retrieval tool'")
	}

	if bathtubDrain.Fixture.Produces == nil {
		t.Fatal("Expected bathtub drain fixture to have produced item")
	}

	if bathtubDrain.Fixture.Produces.Name != "bedroom key" {
		t.Errorf("Expected bathtub drain to produce 'bedroom key', got '%s'", bathtubDrain.Fixture.Produces.Name)
	}

	if !bathtubDrain.Fixture.Produces.IsKey() {
		t.Error("Expected produced item to be a key")
	}

	// Test bedroom fixtures
	bedroom := findRoomByName(allRooms, "bedroom")
	if bedroom == nil {
		t.Fatal("Could not find bedroom")
	}

	// Test altar fixture
	altar := findItemByName(bedroom.Items, "altar")
	if altar == nil {
		t.Fatal("Could not find altar")
	}

	if !altar.IsFixture() {
		t.Error("Expected altar to be a fixture")
	}

	if altar.Fixture == nil {
		t.Fatal("Expected altar to have fixture data")
	}

	if len(altar.Fixture.RequiredItems) != 2 {
		t.Errorf("Expected altar to require 2 items, got %d", len(altar.Fixture.RequiredItems))
	}

	// Check required items (should be false initially)
	if _, exists := altar.Fixture.RequiredItems["stone"]; !exists {
		t.Error("Expected altar to require 'stone'")
	}

	if _, exists := altar.Fixture.RequiredItems["candle"]; !exists {
		t.Error("Expected altar to require 'candle'")
	}

	if altar.Fixture.Produces == nil {
		t.Fatal("Expected altar fixture to have produced item")
	}

	if altar.Fixture.Produces.Name != "balcony key" {
		t.Errorf("Expected altar to produce 'balcony key', got '%s'", altar.Fixture.Produces.Name)
	}

	if !altar.Fixture.Produces.IsKey() {
		t.Error("Expected produced item to be a key")
	}
}

func TestLoadGame_Latches(t *testing.T) {
	// Load the latch test game
	level, err := LoadGameFromFile("../testdata/latch.json")
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Test basic game properties
	if level.Name != "latch test" {
		t.Errorf("Expected game name 'latch test', got '%s'", level.Name)
	}

	// Test rooms
	allRooms := getAllRooms(level)
	if len(allRooms) != 3 {
		t.Errorf("Expected 3 rooms, got %d", len(allRooms))
	}

	// Test doors
	if len(level.Doors) != 3 {
		t.Errorf("Expected 3 doors, got %d", len(level.Doors))
	}

	// Test door XY - should have a latch
	doorXY := findDoorByName(level.Doors, "door XY")
	if doorXY == nil {
		t.Fatal("Could not find door XY")
	}

	if !doorXY.IsLatched() {
		t.Error("Expected door XY to be latched")
	}

	if doorXY.Latch == nil {
		t.Fatal("Expected door XY to have latch data")
	}

	if doorXY.Latch.LockedFrom != "room Y" {
		t.Errorf("Expected door XY to be latched from 'room Y', got '%s'", doorXY.Latch.LockedFrom)
	}

	if !doorXY.Latch.Locked {
		t.Error("Expected door XY latch to be locked")
	}

	// Test door XZ - should not have a latch
	doorXZ := findDoorByName(level.Doors, "door XZ")
	if doorXZ == nil {
		t.Fatal("Could not find door XZ")
	}

	if doorXZ.IsLatched() {
		t.Error("Expected door XZ to not be latched")
	}

	if doorXZ.Latch != nil {
		t.Error("Expected door XZ to not have latch data")
	}

	// Test door ZY - should not have a latch
	doorZY := findDoorByName(level.Doors, "door ZY")
	if doorZY == nil {
		t.Fatal("Could not find door ZY")
	}

	if doorZY.IsLatched() {
		t.Error("Expected door ZY to not be latched")
	}

	if doorZY.Latch != nil {
		t.Error("Expected door ZY to not have latch data")
	}

	// Test room connections
	roomX := findRoomByName(allRooms, "room X")
	if roomX == nil {
		t.Fatal("Could not find room X")
	}

	if len(roomX.Connections) != 2 {
		t.Errorf("Expected room X to have 2 connections, got %d", len(roomX.Connections))
	}

	roomY := findRoomByName(allRooms, "room Y")
	if roomY == nil {
		t.Fatal("Could not find room Y")
	}

	if len(roomY.Connections) != 2 {
		t.Errorf("Expected room Y to have 2 connections, got %d", len(roomY.Connections))
	}

	roomZ := findRoomByName(allRooms, "room Z")
	if roomZ == nil {
		t.Fatal("Could not find room Z")
	}

	if len(roomZ.Connections) != 2 {
		t.Errorf("Expected room Z to have 2 connections, got %d", len(roomZ.Connections))
	}
}

func TestLoadGame_DoorCodeLocks(t *testing.T) {
	// Test JSON data with a door that has a code lock
	jsonData := `{
		"name": "code lock test",
		"floors": [
			{
				"name": "main floor",
				"description": "test floor",
				"rooms": [
					{
						"name": "room A",
						"description": "first room",
						"connections": [
							{
								"location": "north",
								"door_name": "code_door"
							}
						]
					},
					{
						"name": "room B",
						"description": "second room",
						"connections": [
							{
								"location": "south",
								"door_name": "code_door"
							}
						]
					}
				]
			}
		],
		"doors": [
			{
				"name": "code_door",
				"room_a": "room A",
				"room_b": "room B",
				"locked": true,
				"code": "1234"
			}
		]
	}`

	level, err := LoadGame([]byte(jsonData))
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Test basic game properties
	if level.Name != "code lock test" {
		t.Errorf("Expected game name 'code lock test', got '%s'", level.Name)
	}

	// Test doors
	if len(level.Doors) != 1 {
		t.Errorf("Expected 1 door, got %d", len(level.Doors))
	}

	// Test the code door
	codeDoor := findDoorByName(level.Doors, "code_door")
	if codeDoor == nil {
		t.Fatal("Could not find code_door")
	}

	if !codeDoor.IsLocked() {
		t.Error("Expected code_door to be locked")
	}

	if !codeDoor.HasCodeLock() {
		t.Error("Expected code_door to have a code lock")
	}

	if codeDoor.HasKeyLock() {
		t.Error("Expected code_door to not have a key lock")
	}

	if codeDoor.Lock == nil {
		t.Fatal("Expected code_door to have lock data")
	}

	if codeDoor.Lock.Code != "1234" {
		t.Errorf("Expected code_door code to be '1234', got '%s'", codeDoor.Lock.Code)
	}

	if codeDoor.Lock.KeyName != "" {
		t.Errorf("Expected code_door key name to be empty, got '%s'", codeDoor.Lock.KeyName)
	}

	// Test rooms
	allRooms := getAllRooms(level)
	if len(allRooms) != 2 {
		t.Errorf("Expected 2 rooms, got %d", len(allRooms))
	}

	// Test room connections
	roomA := findRoomByName(allRooms, "room A")
	if roomA == nil {
		t.Fatal("Could not find room A")
	}

	if len(roomA.Connections) != 1 {
		t.Errorf("Expected room A to have 1 connection, got %d", len(roomA.Connections))
	}

	roomB := findRoomByName(allRooms, "room B")
	if roomB == nil {
		t.Fatal("Could not find room B")
	}

	if len(roomB.Connections) != 1 {
		t.Errorf("Expected room B to have 1 connection, got %d", len(roomB.Connections))
	}
}

func TestLoadGame_IntroOutroNarrative(t *testing.T) {
	// Load the kill enemy win test game which has intro and outro narrative
	level, err := LoadGameFromFile("../testdata/kill_enemy_win.json")
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Test intro and outro narrative
	if level.IntroNarrative != "foo" {
		t.Errorf("Expected intro narrative to be 'foo', got '%s'", level.IntroNarrative)
	}
	if level.OutroNarrative != "bar" {
		t.Errorf("Expected outro narrative to be 'bar', got '%s'", level.OutroNarrative)
	}

	// Test that other fields are still loaded correctly
	if level.Name != "kill enemy win" {
		t.Errorf("Expected game name 'kill enemy win', got '%s'", level.Name)
	}
}
