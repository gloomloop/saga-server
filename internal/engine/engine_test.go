package engine

import (
	"adventure-engine/internal/loader"
	"adventure-engine/internal/world"
	"os"
	"strings"
	"testing"
)

var debugFlag bool

func init() {
	debugFlag = os.Getenv("DEBUG") == "true"
}

func TestInspectInternal(t *testing.T) {
	// Create a simple test level with one room and one item
	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "test_room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{
			"north": {
				BaseEntity: world.BaseEntity{
					Name:        "test_door",
					Description: "A test door",
				},
			},
		},
		Items: []*world.Item{
			{
				BaseEntity: world.BaseEntity{
					Name:        "test_item",
					Description: "A test item",
				},
				Location: "on the floor",
				Detail:   "It looks interesting",
			},
		},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})

	// Test inspecting an item
	result, err := engine.inspectInternal("test_item")
	if err != nil {
		t.Errorf("Inspect failed: %v", err)
	}

	if result.itemInspection == nil {
		t.Fatal("Expected itemInspection, got nil")
	}

	if result.itemInspection.Name != "test_item" {
		t.Errorf("Expected name 'test_item', got '%s'", result.itemInspection.Name)
	}

	if result.itemInspection.Description != "A test item" {
		t.Errorf("Expected description 'A test item', got '%s'", result.itemInspection.Description)
	}

	// Test inspecting a door
	result, err = engine.inspectInternal("test_door")
	if err != nil {
		t.Errorf("Inspect failed: %v", err)
	}

	if result.doorInspection == nil {
		t.Fatal("Expected doorInspection, got nil")
	}

	// --- Inspecting an item in a container ---
	// Add a container with an item inside to the room
	contained := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "hidden_item",
			Description: "A hidden item",
		},
		Location: "box",
		Detail:   "It's hidden in the box.",
		Portable: &world.Portable{},
	}
	box := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "box",
			Description: "A wooden box",
		},
		Location: "test_room",
		Detail:   "It could contain something.",
		Container: &world.Container{
			Contains: contained,
			Searched: false,
			Locked:   nil,
		},
	}
	room.Items = append(room.Items, box)

	// Try to inspect the contained item before searching (should fail)
	_, err = engine.inspectInternal("hidden_item")
	if err == nil {
		t.Error("Expected error inspecting item in unsearched container, got nil")
	}

	// Search the box
	_, err = engine.searchInternal("box")
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	// Now inspect the contained item (should succeed)
	result, err = engine.inspectInternal("hidden_item")
	if err != nil {
		t.Errorf("Inspect failed after searching container: %v", err)
	}
	if result.itemInspection == nil {
		t.Fatal("Expected itemInspection for contained item after search, got nil")
	}
	if result.itemInspection.Name != "hidden_item" {
		t.Errorf("Expected name 'hidden_item', got '%s'", result.itemInspection.Name)
	}
}

func TestObserve_Visibility(t *testing.T) {
	// Regular item
	regularItem := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "regular_item",
			Description: "A regular item",
		},
		Location: "room",
		Detail:   "Just a normal thing.",
	}

	// Concealed item (hidden under a sheet)
	coin := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "coin",
			Description: "A shiny gold coin",
		},
		Location: "room",
		Detail:   "It looks valuable.",
		Portable: &world.Portable{},
	}
	sheet := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "sheet",
			Description: "A white bedsheet",
		},
		Location: "room",
		Detail:   "It's covering something.",
		Concealer: &world.Concealer{
			Hidden:    coin,
			Uncovered: false,
		},
	}

	// Container with an item inside (not searched)
	containedItem := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "gem",
			Description: "A sparkling gem",
		},
		Location: "box",
		Detail:   "It glitters.",
		Portable: &world.Portable{},
	}
	box := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "box",
			Description: "A small box",
		},
		Location: "room",
		Detail:   "It could contain something.",
		Container: &world.Container{
			Contains: containedItem,
			Searched: false,
			Locked:   nil,
		},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{regularItem, sheet, box},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})

	// Initial observe: should see regular item, sheet, and box, but not coin or gem
	obs, err := engine.observeInternal()
	if err != nil {
		t.Errorf("Observe failed: %v", err)
	}
	names := map[string]ItemInfo{}
	for _, info := range obs.VisibleItems {
		names[info.Name] = info
	}
	if _, ok := names["regular_item"]; !ok {
		t.Error("Expected to see regular_item in room")
	}
	if sheetInfo, ok := names["sheet"]; !ok {
		t.Error("Expected to see sheet in room")
	} else if !sheetInfo.IsConcealer {
		t.Error("Expected sheet to be marked as concealer")
	}
	if boxInfo, ok := names["box"]; !ok {
		t.Error("Expected to see box in room")
	} else if !boxInfo.IsContainer {
		t.Error("Expected box to be marked as container")
	}
	if _, ok := names["coin"]; ok {
		t.Error("Should not see coin before revealing")
	}
	if _, ok := names["gem"]; ok {
		t.Error("Should not see gem before box is searched")
	}

	// Reveal the coin (simulate uncovering the sheet)
	revealed, err := sheet.Concealer.Reveal()
	if err != nil {
		t.Errorf("Reveal failed: %v", err)
	}
	room.Items = append(room.Items, revealed)

	obs, err = engine.observeInternal()
	if err != nil {
		t.Errorf("Observe failed: %v", err)
	}
	names = map[string]ItemInfo{}
	for _, info := range obs.VisibleItems {
		names[info.Name] = info
	}
	if _, ok := names["coin"]; !ok {
		t.Error("Expected to see coin after revealing")
	}

	// Search the box (simulate searching the container)
	found, err := box.Container.Search()
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}
	// The gem is still in the box, but box is now searched
	if found != nil {
		// In this engine, Search removes the item from the container, so add it to the room
		room.Items = append(room.Items, found)
	}

	obs, err = engine.observeInternal()
	if err != nil {
		t.Errorf("Observe failed: %v", err)
	}
	names = map[string]ItemInfo{}
	for _, info := range obs.VisibleItems {
		names[info.Name] = info
	}
	if _, ok := names["gem"]; !ok {
		t.Error("Expected to see gem after box is searched")
	}
}

func TestUncoverInternal(t *testing.T) {
	// Create a concealed item
	hidden := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "hidden_gem",
			Description: "A sparkling hidden gem",
		},
		Location: "room",
		Detail:   "It glitters brightly.",
		Portable: &world.Portable{},
	}

	// Create a concealer (e.g., a rug) hiding the item
	rug := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "rug",
			Description: "A dusty old rug",
		},
		Location: "room",
		Detail:   "It looks like it could be hiding something.",
		Concealer: &world.Concealer{
			Hidden:    hidden,
			Uncovered: false,
		},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{rug},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})

	// Uncover the rug
	result, err := engine.uncoverInternal("rug")
	if err != nil {
		t.Errorf("Uncover failed: %v", err)
	}

	if result.name != "rug" {
		t.Errorf("Expected name 'rug', got '%s'", result.name)
	}
	if result.RevealedItem.Name != "hidden_gem" {
		t.Errorf("Expected revealed item 'hidden_gem', got '%s'", result.RevealedItem.Name)
	}

	// The hidden item should now be in the room
	found := false
	for _, item := range room.Items {
		if item.Name == "hidden_gem" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected hidden_gem to be in the room after uncovering")
	}

	// The concealer should now be marked as uncovered
	if !rug.Concealer.Uncovered {
		t.Error("Expected rug to be marked as uncovered after uncovering")
	}

	// Trying to uncover again should fail
	_, err = engine.uncoverInternal("rug")
	if err == nil {
		t.Error("Expected error when uncovering an already uncovered concealer")
	}
}

func TestSearchInternal(t *testing.T) {
	// Create an item to be placed in the container
	key := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "key",
			Description: "A small brass key",
		},
		Location: "box",
		Detail:   "It looks like it might unlock something.",
		Portable: &world.Portable{},
	}

	// Create a container (box) with the key inside
	box := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "box",
			Description: "A wooden box",
		},
		Location: "room",
		Detail:   "It could contain something.",
		Container: &world.Container{
			Contains: key,
			Searched: false,
			Locked:   nil,
		},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{box},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})

	// Search the box
	result, err := engine.searchInternal("box")
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	if result.ContainerName != "box" {
		t.Errorf("Expected ContainerName 'box', got '%s'", result.ContainerName)
	}
	if result.ContainedItemInfo == nil {
		t.Fatal("Expected to find an item in the box on first search")
	}
	if result.ContainedItemInfo.Name != "key" {
		t.Errorf("Expected to find 'key' in the box, got '%s'", result.ContainedItemInfo.Name)
	}

	// The box should now be marked as searched
	if !box.Container.Searched {
		t.Error("Expected box to be marked as searched after searching")
	}

	// Searching again should return the item again (since search no longer removes it)
	result2, err := engine.searchInternal("box")
	if err != nil {
		t.Errorf("Second search failed: %v", err)
	}
	if result2.ContainedItemInfo == nil || result2.ContainedItemInfo.Name != "key" {
		t.Error("Expected to find 'key' again on second search of box")
	}
}

func TestSearch_LockedContainer(t *testing.T) {
	// Create an item to be placed in the locked container
	note := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "note",
			Description: "A secret note",
		},
		Location: "safe",
		Detail:   "It says 'The code is 1234'",
		Portable: &world.Portable{},
	}

	// Create a locked container (safe) with the note inside
	safe := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "safe",
			Description: "A metal safe with a keypad",
		},
		Location: "room",
		Detail:   "It has a keypad with numbers 0-9",
		Container: &world.Container{
			Contains: note,
			Searched: false,
			Locked: &world.Lock{
				Locked:  true,
				KeyName: "",
				Code:    "1234",
			},
		},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{safe},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})

	// Attempt to search the locked safe (should fail)
	_, err := engine.searchInternal("safe")
	if err == nil {
		t.Fatal("Expected error when searching a locked container, got nil")
	}

	// Unlock the safe with the correct code
	err = safe.Container.UnlockWithCode("1234")
	if err != nil {
		t.Errorf("Failed to unlock safe with correct code: %v", err)
	}

	// Now search should succeed
	result, err := engine.searchInternal("safe")
	if err != nil {
		t.Errorf("Search failed after unlocking: %v", err)
	}
	if result.ContainedItemInfo == nil || result.ContainedItemInfo.Name != "note" {
		t.Error("Expected to find 'note' in the safe after unlocking")
	}
}

func TestTakeInternal(t *testing.T) {
	// Portable item in the room
	coin := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "coin",
			Description: "A shiny coin",
		},
		Location: "room",
		Detail:   "It glints in the light.",
		Portable: &world.Portable{},
	}

	// Non-portable item in the room
	rock := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "rock",
			Description: "A heavy rock",
		},
		Location: "room",
		Detail:   "It's too heavy to carry.",
	}

	// Portable item in a container
	key := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "key",
			Description: "A small key",
		},
		Location: "box",
		Detail:   "It might unlock something.",
		Portable: &world.Portable{},
	}
	box := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "box",
			Description: "A wooden box",
		},
		Location: "room",
		Detail:   "It could contain something.",
		Container: &world.Container{
			Contains: key,
			Searched: true,
			Locked:   nil,
		},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{coin, rock, box},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})

	// Take the coin from the room
	result, err := engine.takeInternal("coin")
	if err != nil {
		t.Errorf("Take failed: %v", err)
	}
	if result.ItemInfo.Name != "coin" {
		t.Errorf("Expected to take 'coin', got '%s'", result.ItemInfo.Name)
	}
	if len(engine.Player.Inventory) != 1 || engine.Player.Inventory[0].Name != "coin" {
		t.Error("Coin not found in inventory after take")
	}

	// Try to take the rock (not portable)
	_, err = engine.takeInternal("rock")
	if err == nil {
		t.Error("Expected error when taking non-portable item, got nil")
	}

	// Take the key from the box (container)
	result, err = engine.takeInternal("key")
	if err != nil {
		t.Errorf("Take from container failed: %v", err)
	}
	if result.ItemInfo.Name != "key" {
		t.Errorf("Expected to take 'key', got '%s'", result.ItemInfo.Name)
	}
	if len(engine.Player.Inventory) != 2 {
		t.Error("Expected 2 items in inventory after taking from container")
	}

	// Try to take an item that doesn't exist
	_, err = engine.takeInternal("banana")
	if err == nil {
		t.Error("Expected error when taking non-existent item, got nil")
	}
}

func TestTraverseInternal(t *testing.T) {
	// Create two rooms
	roomA := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "RoomA",
			Description: "The first room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}
	roomB := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "RoomB",
			Description: "The second room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}

	// Create an unlocked door between RoomA and RoomB
	door := &world.Door{
		BaseEntity: world.BaseEntity{
			Name:        "doorAB",
			Description: "A door between RoomA and RoomB",
		},
		RoomA: "RoomA",
		RoomB: "RoomB",
		Lock:  nil, // unlocked
	}
	roomA.Connections["east"] = door
	roomB.Connections["west"] = door

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{roomA, roomB},
		WinCondition: nil,
	})
	engine.CurrentRoom = roomA

	// --- Move from RoomA to RoomB through unlocked door (should succeed) ---
	result, err := engine.traverseInternal("east")
	if err != nil {
		t.Errorf("Traverse failed: %v", err)
	}
	if result == nil || result.ToRoom != "RoomB" {
		t.Error("Expected to traverse to RoomB")
	}
	if engine.CurrentRoom.Name != "RoomB" {
		t.Error("Player should be in RoomB after traverse")
	}

	// --- Attempt to move through a locked door (should fail) ---
	lockedDoor := &world.Door{
		BaseEntity: world.BaseEntity{
			Name:        "lockedDoor",
			Description: "A locked door to nowhere",
		},
		RoomA: "RoomB",
		RoomB: "RoomC",
		Lock: &world.Lock{
			Locked:  true,
			KeyName: "key",
		},
	}
	roomB.Connections["north"] = lockedDoor

	_, err = engine.traverseInternal("north")
	if err == nil {
		t.Error("Expected error when traversing through a locked door, got nil")
	}

	// --- Attempt to move through a non-existent door (should fail) ---
	_, err = engine.traverseInternal("south")
	if err == nil {
		t.Error("Expected error when traversing through a non-existent door, got nil")
	}

	// --- Move back from RoomB to RoomA by door name (should succeed) ---
	result, err = engine.traverseInternal("doorAB")
	if err != nil {
		t.Errorf("Traverse by door name failed: %v", err)
	}
	if result == nil || result.ToRoom != "RoomA" {
		t.Error("Expected to traverse to RoomA by door name")
	}
	if engine.CurrentRoom.Name != "RoomA" {
		t.Error("Player should be in RoomA after traverse by door name")
	}
}

func TestTake_UncoverConcealer(t *testing.T) {
	// Create a hidden item
	hidden := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "hidden_gem",
			Description: "A sparkling hidden gem",
		},
		Location: "room",
		Detail:   "It glitters brightly.",
		Portable: &world.Portable{},
	}

	// Create a concealer (e.g., a rug) hiding the item
	rug := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "rug",
			Description: "A dusty old rug",
		},
		Location: "room",
		Detail:   "It looks like it could be hiding something.",
		Concealer: &world.Concealer{
			Hidden:    hidden,
			Uncovered: false,
		},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{rug},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})

	// Take the rug (should trigger uncover and return the hidden item)
	result, err := engine.takeInternal("rug")
	if err != nil {
		t.Errorf("Take failed: %v", err)
	}
	if result.ItemInfo.Name != "hidden_gem" {
		t.Errorf("Expected to take 'hidden_gem' by uncovering, got '%s'", result.ItemInfo.Name)
	}
	if !rug.Concealer.Uncovered {
		t.Error("Expected rug to be marked as uncovered after take")
	}
	// The hidden item should now be in the room
	found := false
	for _, item := range room.Items {
		if item.Name == "hidden_gem" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected hidden_gem to be in the room after uncovering via take")
	}

	// Try to take the rug again (should fail since it's already uncovered)
	_, err = engine.takeInternal("rug")
	if err == nil {
		t.Error("Expected error when taking a concealer that is already uncovered, got nil")
	} else if err.Error() != "you cannot take the rug" {
		t.Errorf("Expected error message 'you cannot take the rug', got '%s'", err.Error())
	}
}

func TestTake_AmmoBox(t *testing.T) {
	// Create a shotgun (weapon, no ammo in inventory)
	shotgun := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "shotgun",
			Description: "A pump-action shotgun",
		},
		Location: "room",
		Detail:   "It looks powerful.",
		Weapon: &world.Weapon{
			Damage: 0.9,
			Ammo:   &world.Ammo{Quantity: 0}, // Use pointer
		},
		Portable: &world.Portable{},
	}

	// Create an ammo box for the shotgun (2 rounds)
	ammoBox := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "shotgun_ammo_box",
			Description: "A box of shotgun shells",
		},
		Location: "room",
		Detail:   "Contains 2 shells.",
		AmmoBox: &world.AmmoBox{
			WeaponName: "shotgun",
			Ammo:       &world.Ammo{Quantity: 2},
		},
		Portable: &world.Portable{},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{shotgun, ammoBox},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})
	engine.CurrentRoom = room

	// Take the shotgun (should add to inventory, but no ammo)
	_, err := engine.Take("shotgun")
	if err != nil {
		t.Fatalf("Take shotgun failed: %v", err)
	}
	if len(engine.Player.Inventory) != 1 || engine.Player.Inventory[0].Name != "shotgun" {
		t.Fatalf("Expected shotgun in inventory, got %+v", engine.Player.Inventory)
	}
	if engine.Player.Ammo["shotgun"] != 0 {
		t.Errorf("Expected 0 ammo for shotgun, got %d", engine.Player.Ammo["shotgun"])
	}

	// Take the ammo box (should add 2 rounds to shotgun ammo)
	_, err = engine.Take("shotgun_ammo_box")
	if err != nil {
		t.Fatalf("Take ammo box failed: %v", err)
	}
	if engine.Player.Ammo["shotgun"] != 2 {
		t.Errorf("Expected 2 ammo for shotgun after taking ammo box, got %d", engine.Player.Ammo["shotgun"])
	}
}

func TestUnlockInternal(t *testing.T) {
	// Create a key for testing
	key := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "test_key",
			Description: "A test key",
		},
		Location: "inventory",
		Detail:   "It might unlock something.",
		Key:      &world.Key{},
	}

	// Create a container with a key lock
	lockedBox := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "locked_box",
			Description: "A locked wooden box",
		},
		Location: "room",
		Detail:   "It's locked with a key.",
		Container: &world.Container{
			Contains: nil,
			Searched: false,
			Locked: &world.Lock{
				Locked:  true,
				KeyName: "test_key",
				Code:    "",
			},
		},
	}

	// Create a container with a code lock
	lockedSafe := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "locked_safe",
			Description: "A metal safe with a keypad",
		},
		Location: "room",
		Detail:   "It has a keypad with numbers 0-9.",
		Container: &world.Container{
			Contains: nil,
			Searched: false,
			Locked: &world.Lock{
				Locked:  true,
				KeyName: "",
				Code:    "1234",
			},
		},
	}

	// Create a door with a key lock
	lockedDoor := &world.Door{
		BaseEntity: world.BaseEntity{
			Name:        "locked_door",
			Description: "A locked door",
		},
		RoomA: "room",
		RoomB: "other_room",
		Lock: &world.Lock{
			Locked:  true,
			KeyName: "test_key",
			Code:    "",
		},
	}

	// Create a door with a code lock
	codeDoor := &world.Door{
		BaseEntity: world.BaseEntity{
			Name:        "code_door",
			Description: "A door with a keypad",
		},
		RoomA: "room",
		RoomB: "secret_room",
		Lock: &world.Lock{
			Locked:  true,
			KeyName: "",
			Code:    "5678",
		},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{
			"north": lockedDoor,
			"east":  codeDoor,
		},
		Items: []*world.Item{lockedBox, lockedSafe},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})

	// Add the key to player's inventory
	engine.Player.Inventory = append(engine.Player.Inventory, key)

	// --- Test unlocking a container with a key ---
	result, err := engine.unlockInternal("test_key", "locked_box")
	if err != nil {
		t.Errorf("Failed to unlock box with key: %v", err)
	}
	if !result.Unlocked {
		t.Error("Box should be unlocked after using key")
	}
	if lockedBox.Container.IsLocked() {
		t.Error("Box should be unlocked after using key")
	}

	// --- Test unlocking a container with a code ---
	result, err = engine.unlockInternal("1234", "locked_safe")
	if err != nil {
		t.Errorf("Failed to unlock safe with code: %v", err)
	}
	if !result.Unlocked {
		t.Error("Safe should be unlocked after using code")
	}
	if lockedSafe.Container.IsLocked() {
		t.Error("Safe should be unlocked after using code")
	}

	// --- Test unlocking a door with a key ---
	result, err = engine.unlockInternal("test_key", "locked_door")
	if err != nil {
		t.Errorf("Failed to unlock door with key: %v", err)
	}
	if !result.Unlocked {
		t.Error("Door should be unlocked after using key")
	}
	if lockedDoor.IsLocked() {
		t.Error("Door should be unlocked after using key")
	}

	// --- Test unlocking a door with a code ---
	result, err = engine.unlockInternal("5678", "code_door")
	if err != nil {
		t.Errorf("Failed to unlock code door with code: %v", err)
	}
	if !result.Unlocked {
		t.Error("Code door should be unlocked after using code")
	}
	if codeDoor.IsLocked() {
		t.Error("Code door should be unlocked after using code")
	}

	// --- Test error cases ---

	// Try to unlock with wrong key
	_, err = engine.unlockInternal("wrong_key", "locked_box")
	if err == nil {
		t.Error("Expected error when using wrong key")
	}

	// Try to unlock with wrong code
	_, err = engine.unlockInternal("9999", "locked_safe")
	if err == nil {
		t.Error("Expected error when using wrong code")
	}

	// Try to unlock non-existent item
	_, err = engine.unlockInternal("test_key", "non_existent")
	if err == nil {
		t.Error("Expected error when trying to unlock non-existent item")
	}

	// Try to unlock with key not in inventory
	_, err = engine.unlockInternal("missing_key", "locked_box")
	if err == nil {
		t.Error("Expected error when using key not in inventory")
	}

	// Try to unlock non-container item
	rock := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "rock",
			Description: "A heavy rock",
		},
		Location: "room",
		Detail:   "It's too heavy to carry.",
	}
	room.Items = append(room.Items, rock)
	_, err = engine.unlockInternal("test_key", "rock")
	if err == nil {
		t.Error("Expected error when trying to unlock non-container item")
	}

	// Try to unlock already unlocked container
	_, err = engine.unlockInternal("test_key", "locked_box")
	if err == nil {
		t.Error("Expected error when trying to unlock already unlocked container")
	}
}

func TestHealInternal(t *testing.T) {
	// Create a weak health item
	weakPotion := func() *world.Item {
		return &world.Item{
			BaseEntity: world.BaseEntity{
				Name:        "weak_potion",
				Description: "A weak healing potion",
			},
			Location: "inventory",
			Detail:   "It might heal minor wounds.",
			HealthItem: &world.HealthItem{
				HealthEffect: world.HealthBoostWeak,
			},
			Portable: &world.Portable{},
		}
	}

	// Create a strong health item
	strongPotion := func() *world.Item {
		return &world.Item{
			BaseEntity: world.BaseEntity{
				Name:        "strong_potion",
				Description: "A strong healing potion",
			},
			Location: "inventory",
			Detail:   "It can heal serious wounds.",
			HealthItem: &world.HealthItem{
				HealthEffect: world.HealthBoostStrong,
			},
			Portable: &world.Portable{},
		}
	}

	// Create a non-health item
	rock := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "rock",
			Description: "A heavy rock",
		},
		Location: "inventory",
		Detail:   "It's not a health item.",
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})
	engine.CurrentRoom = room

	// Add items to player's inventory
	engine.Player.Inventory = append(engine.Player.Inventory, weakPotion(), strongPotion(), rock)

	// --- Test healing with weak potion when player is injured ---
	// First, injure the player
	engine.Player.InflictDamage()
	if engine.Player.Health == world.HealthFine {
		t.Fatal("Player should be injured after InflictDamage")
	}

	// Heal with weak potion
	result, err := engine.healInternal("weak_potion")
	if err != nil {
		t.Errorf("Failed to heal with weak potion: %v", err)
	}
	if result.Health != world.HealthFine {
		t.Errorf("Expected health to be restored to fine, got %s", result.Health)
	}

	// --- Test healing with strong potion when player is injured ---
	// Injure the player again
	engine.Player.InflictDamage()
	if engine.Player.Health == world.HealthFine {
		t.Fatal("Player should be injured after InflictDamage")
	}
	// Add a new strong potion to inventory
	engine.Player.Inventory = append(engine.Player.Inventory, strongPotion())
	// Heal with strong potion
	result, err = engine.healInternal("strong_potion")
	if err != nil {
		t.Errorf("Failed to heal with strong potion: %v", err)
	}
	if result.Health != world.HealthFine {
		t.Errorf("Expected health to be restored to fine, got %s", result.Health)
	}

	// --- Test error cases ---

	// Try to heal when already at full health
	engine.Player.Inventory = append(engine.Player.Inventory, weakPotion())
	_, err = engine.healInternal("weak_potion")
	if err == nil {
		t.Error("Expected error when trying to heal at full health")
	}

	// Try to heal with non-health item
	engine.Player.InflictDamage() // Injure player again
	_, err = engine.healInternal("rock")
	if err == nil {
		t.Error("Expected error when trying to heal with non-health item")
	}

	// Try to heal with item not in inventory
	_, err = engine.healInternal("missing_potion")
	if err == nil {
		t.Error("Expected error when trying to heal with item not in inventory")
	}

	// --- Test healing when player is at different health states ---
	// Test when player is at critical health (if that's a state)
	// This depends on the health system implementation
	// For now, just test that healing works when player is injured
	engine.Player.InflictDamage()
	engine.Player.Inventory = append(engine.Player.Inventory, strongPotion())
	result, err = engine.healInternal("strong_potion")
	if err != nil {
		t.Errorf("Failed to heal when injured: %v", err)
	}
	if result.Health != world.HealthFine {
		t.Errorf("Expected health to be restored to fine, got %s", result.Health)
	}
}

func TestHeal_RemovesHealthItemFromInventory(t *testing.T) {
	// Create a health item
	potion := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "healing_potion",
			Description: "A small healing potion",
		},
		Location: "inventory",
		Detail:   "Restores your health.",
		HealthItem: &world.HealthItem{
			HealthEffect: world.HealthBoostWeak,
		},
		Portable: &world.Portable{},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})
	engine.CurrentRoom = room
	engine.Player.Inventory = append(engine.Player.Inventory, potion)

	// Injure the player
	engine.Player.InflictDamage()
	if engine.Player.Health == world.HealthFine {
		t.Fatal("Player should be injured after InflictDamage")
	}

	// Heal with the potion
	_, err := engine.healInternal("healing_potion")
	if err != nil {
		t.Fatalf("Heal with potion failed: %v", err)
	}

	// Check that the potion is no longer in inventory
	for _, it := range engine.Player.Inventory {
		if it.Name == "healing_potion" {
			t.Errorf("Expected healing_potion to be removed from inventory after use, but it is still present")
		}
	}
}

func TestBattleInternal(t *testing.T) {
	// Create weapons for testing
	sword := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "sword",
			Description: "A sharp sword",
		},
		Location: "inventory",
		Detail:   "It's a deadly weapon.",
		Weapon: &world.Weapon{
			Damage: 0.8, // High damage weapon
			Ammo:   nil, // No ammo required
		},
	}

	pistol := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "pistol",
			Description: "A loaded pistol",
		},
		Location: "inventory",
		Detail:   "It uses bullets.",
		Weapon: &world.Weapon{
			Damage: 0.9,                      // Very high damage weapon
			Ammo:   &world.Ammo{Quantity: 5}, // Uses ammo
		},
	}

	// Create a non-weapon item
	rock := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "rock",
			Description: "A heavy rock",
		},
		Location: "inventory",
		Detail:   "It's not a weapon.",
	}

	// Create an enemy
	enemy := &world.Enemy{
		BaseEntity: world.BaseEntity{
			Name:        "goblin",
			Description: "A fierce goblin",
		},
		Room:         "room",
		HP:           3,
		TriggerEvent: world.TriggerEventTakeItem,
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})

	// Add items to player's inventory
	engine.Player.Inventory = append(engine.Player.Inventory, sword, pistol, rock)

	// --- Test error cases first ---

	// Try to battle without an enemy
	_, err := engine.battleInternal("sword")
	if err == nil {
		t.Error("Expected error when trying to battle without an enemy")
	}

	// Set up enemy for combat
	engine.FightingEnemy = enemy

	// Try to battle with non-existent weapon
	_, err = engine.battleInternal("non_existent_weapon")
	if err == nil {
		t.Error("Expected error when trying to battle with non-existent weapon")
	}

	// Try to battle with non-weapon item
	_, err = engine.battleInternal("rock")
	if err == nil {
		t.Error("Expected error when trying to battle with non-weapon item")
	}

	// --- Test combat with fists/hands ---
	// Use fake RNG to control the outcome
	fakeRng := &FakeRng{}
	engine.Rng = fakeRng

	// Test winning with fists (RNG < 0.5)
	fakeRng.SetValue(0.3) // 0.3 < 0.5, so player wins
	result, err := engine.battleInternal("")
	if err != nil {
		t.Errorf("Battle with fists failed: %v", err)
	}
	if !result.WonRound {
		t.Error("Expected to win round with fists")
	}
	if !result.EnemyAlive {
		t.Error("Enemy should still be alive after one hit")
	}
	if !result.PlayerAlive {
		t.Error("Player should still be alive")
	}
	if result.EnemyName != "goblin" {
		t.Errorf("Expected enemy name 'goblin', got '%s'", result.EnemyName)
	}

	// Test losing with fists (RNG >= 0.5)
	fakeRng.SetValue(0.7) // 0.7 >= 0.5, so player loses
	result, err = engine.battleInternal("fists")
	if err != nil {
		t.Errorf("Battle with fists failed: %v", err)
	}
	if result.WonRound {
		t.Error("Expected to lose round with fists")
	}
	if !result.EnemyAlive {
		t.Error("Enemy should still be alive")
	}
	if !result.PlayerAlive {
		t.Error("Player should still be alive after one hit")
	}

	// --- Test combat with sword (no ammo required) ---
	fakeRng.SetValue(0.7) // 0.7 < 0.8, so player wins
	result, err = engine.battleInternal("sword")
	if err != nil {
		t.Errorf("Battle with sword failed: %v", err)
	}
	if !result.WonRound {
		t.Error("Expected to win round with sword")
	}
	if !result.EnemyAlive {
		t.Error("Enemy should still be alive after two hits")
	}

	// --- Test combat with pistol (uses ammo) ---
	// Add ammo to player's inventory
	engine.Player.Ammo["pistol"] = 3

	fakeRng.SetValue(0.8) // 0.8 < 0.9, so player wins
	result, err = engine.battleInternal("pistol")
	if err != nil {
		t.Errorf("Battle with pistol failed: %v", err)
	}
	if !result.WonRound {
		t.Error("Expected to win round with pistol")
	}
	if engine.Player.Ammo["pistol"] != 2 {
		t.Errorf("Expected 2 ammo remaining, got %d", engine.Player.Ammo["pistol"])
	}

	// Test running out of ammo
	engine.Player.Ammo["pistol"] = 0
	_, err = engine.battleInternal("pistol")
	if err == nil {
		t.Error("Expected error when trying to fire weapon with no ammo")
	}

	// --- Test enemy death ---
	engine.FightingEnemy = enemy // Restore fighting enemy after defeat
	// Set enemy HP to 1 and ensure player wins
	enemy.HP = 1
	fakeRng.SetValue(0.7) // Player wins
	result, err = engine.battleInternal("sword")
	if err != nil {
		t.Errorf("Battle failed: %v", err)
	}
	if !result.WonRound {
		t.Error("Expected to win round")
	}
	if result.EnemyAlive {
		t.Error("Enemy should be dead after final hit")
	}

	// --- Test player death ---
	// Reset enemy and set player to critical health
	enemy.HP = 3
	engine.FightingEnemy = enemy // Restore fighting enemy after defeat
	engine.Player.Health = world.HealthCrit
	fakeRng.SetValue(0.9) // Player loses
	result, err = engine.battleInternal("sword")
	if err != nil {
		t.Errorf("Battle failed: %v", err)
	}
	if result.WonRound {
		t.Error("Expected to lose round")
	}
	if result.PlayerAlive {
		t.Error("Player should be dead after final hit")
	}
}

func TestBattle_WeaponDamageLogic(t *testing.T) {
	// Create weapons with different damage values
	weakWeapon := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "stick",
			Description: "A weak stick",
		},
		Location: "inventory",
		Detail:   "It's not very effective.",
		Weapon: &world.Weapon{
			Damage: 0.1, // Low damage weapon
			Ammo:   nil,
		},
	}

	strongWeapon := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "sword",
			Description: "A sharp sword",
		},
		Location: "inventory",
		Detail:   "It's very effective.",
		Weapon: &world.Weapon{
			Damage: 0.9, // High damage weapon
			Ammo:   nil,
		},
	}

	enemy := &world.Enemy{
		BaseEntity: world.BaseEntity{
			Name:        "goblin",
			Description: "A fierce goblin",
		},
		Room:         "room",
		HP:           10,
		TriggerEvent: world.TriggerEventTakeItem,
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})
	engine.FightingEnemy = enemy
	engine.Player.Inventory = append(engine.Player.Inventory, weakWeapon, strongWeapon)

	// Use fake RNG to control outcomes
	fakeRng := &FakeRng{}
	engine.Rng = fakeRng

	// Test with RNG value of 0.5
	fakeRng.SetValue(0.5)

	// With weak weapon (damage 0.1): 0.5 < 0.1 = false, so player loses
	result, err := engine.battleInternal("stick")
	if err != nil {
		t.Errorf("Battle with stick failed: %v", err)
	}
	if result.WonRound {
		t.Error("Expected to lose with stick (weak weapon)")
	}

	// Reset enemy HP
	enemy.HP = 10

	// With strong weapon (damage 0.9): 0.5 < 0.9 = true, so player wins
	result, err = engine.battleInternal("sword")
	if err != nil {
		t.Errorf("Battle with sword failed: %v", err)
	}
	if !result.WonRound {
		t.Error("Expected to win with sword (strong weapon)")
	}

	// Now better weapons make it EASIER to win, which is correct!
}

func TestEventHandling_LevelCompletionOnRoomEntry(t *testing.T) {
	// Create demos
	room1 := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "Room1",
			Description: "The first room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}
	room2 := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "Room2",
			Description: "The second room (not win)",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}
	room3 := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "Room3",
			Description: "The third room (win room)",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}

	// Create doors between the rooms
	door1 := &world.Door{
		BaseEntity: world.BaseEntity{
			Name:        "door1",
			Description: "A door between Room1 and Room2",
		},
		RoomA: "Room1",
		RoomB: "Room2",
		Lock:  nil, // unlocked
	}
	door2 := &world.Door{
		BaseEntity: world.BaseEntity{
			Name:        "door2",
			Description: "A door between Room2 and Room3",
		},
		RoomA: "Room2",
		RoomB: "Room3",
		Lock:  nil, // unlocked
	}
	room1.Connections["east"] = door1
	room2.Connections["west"] = door1
	room2.Connections["east"] = door2
	room3.Connections["west"] = door2

	// Set win condition: entering Room3
	winCondition := &world.Event{
		Event:    world.EventRoomEntered,
		RoomName: "Room3",
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room1, room2, room3},
		WinCondition: winCondition,
	})
	engine.CurrentRoom = room1

	// Traverse to Room2 (should NOT trigger win condition)
	result, err := engine.Traverse("east")
	if err != nil {
		t.Fatalf("Traverse to Room2 failed: %v", err)
	}
	if result.EngineStateInfo.EngineStateChangeNotification != nil {
		t.Errorf("Expected no state change notification when entering Room2, got %v", *result.EngineStateInfo.EngineStateChangeNotification)
	}
	if result.EngineStateInfo.LevelCompletionState != LevelCompletionStateInProgress {
		t.Errorf("Expected level completion state %q after entering Room2, got %q", LevelCompletionStateInProgress, result.EngineStateInfo.LevelCompletionState)
	}

	// Traverse to Room3 (should trigger win condition)
	result, err = engine.Traverse("east")
	if err != nil {
		t.Fatalf("Traverse to Room3 failed: %v", err)
	}
	// Check notification
	if result.EngineStateInfo.EngineStateChangeNotification == nil {
		t.Fatal("Expected state change notification, got nil")
	}
	if *result.EngineStateInfo.EngineStateChangeNotification != EngineStateChangeLevelComplete {
		t.Errorf("Expected notification %q, got %q", EngineStateChangeLevelComplete, *result.EngineStateInfo.EngineStateChangeNotification)
	}
	// Check level completion state
	if result.EngineStateInfo.LevelCompletionState != LevelCompletionStateComplete {
		t.Errorf("Expected level completion state %q, got %q", LevelCompletionStateComplete, result.EngineStateInfo.LevelCompletionState)
	}
}

func TestEventHandling_EnterCombatOnTakeGem(t *testing.T) {
	// Create items
	rock := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "rock",
			Description: "Just a rock",
		},
		Location: "room",
		Detail:   "It's a plain rock.",
		Portable: &world.Portable{}, // Make rock portable
	}
	gem := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "gem",
			Description: "A shiny gem",
		},
		Location: "room",
		Detail:   "It sparkles brightly.",
		Portable: &world.Portable{}, // Make gem portable
	}

	// Create an enemy
	enemy := &world.Enemy{
		BaseEntity: world.BaseEntity{
			Name:        "goblin",
			Description: "A sneaky goblin",
		},
		Room:         "room",
		HP:           1,
		TriggerEvent: world.TriggerEventTakeItem,
	}

	// Create a trigger: taking the gem triggers combat with the goblin
	trigger := &world.Trigger{
		Event: world.Event{
			Event:    world.EventItemTaken,
			ItemName: "gem",
		},
		Effect: world.Effect{
			EffectType: world.EffectEnterCombat,
			EnemyName:  "goblin",
		},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{rock, gem},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		Enemies:      []*world.Enemy{enemy},
		Triggers:     []*world.Trigger{trigger},
		WinCondition: nil,
	})
	engine.CurrentRoom = room

	// Take the rock (should NOT trigger combat)
	result, err := engine.Take("rock")
	if err != nil {
		t.Fatalf("Take rock failed: %v", err)
	}
	if result.EngineStateInfo.EngineStateChangeNotification != nil {
		t.Errorf("Expected no state change notification when taking rock, got %v", *result.EngineStateInfo.EngineStateChangeNotification)
	}
	if engine.Mode != Investigation {
		t.Errorf("Expected mode to remain Investigation after taking rock, got %v", engine.Mode)
	}

	// Take the gem (should trigger combat)
	result, err = engine.Take("gem")
	if err != nil {
		t.Fatalf("Take gem failed: %v", err)
	}
	if result.EngineStateInfo.EngineStateChangeNotification == nil {
		t.Fatal("Expected state change notification when taking gem, got nil")
	}
	if *result.EngineStateInfo.EngineStateChangeNotification != EngineStateChangeEnterCombat {
		t.Errorf("Expected notification %q, got %q", EngineStateChangeEnterCombat, *result.EngineStateInfo.EngineStateChangeNotification)
	}
	if engine.Mode != Combat {
		t.Errorf("Expected mode to be Combat after taking gem, got %v", engine.Mode)
	}
	if engine.FightingEnemy == nil || engine.FightingEnemy.Name != "goblin" {
		t.Errorf("Expected FightingEnemy to be goblin, got %+v", engine.FightingEnemy)
	}
}

func TestEventHandling_CombinedEvents(t *testing.T) {
	// Create the key (portable)
	key := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "key",
			Description: "A small brass key",
		},
		Location: "chest",
		Detail:   "It looks like it could unlock something.",
		Portable: &world.Portable{},
		Key:      &world.Key{},
	}

	// Create the chest (container, not locked, not searched)
	chest := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "chest",
			Description: "An old wooden chest",
		},
		Location: "room",
		Detail:   "It might contain something.",
		Container: &world.Container{
			Contains: key,
			Searched: false,
			Locked:   nil,
		},
	}

	// Create the enemy
	enemy := &world.Enemy{
		BaseEntity: world.BaseEntity{
			Name:        "skeleton",
			Description: "A rattling skeleton",
		},
		Room:         "room",
		HP:           1,
		TriggerEvent: world.TriggerEventTakeItem,
	}

	// Create a trigger: taking the key triggers combat with the skeleton
	trigger := &world.Trigger{
		Event: world.Event{
			Event:    world.EventItemTaken,
			ItemName: "key",
		},
		Effect: world.Effect{
			EffectType: world.EffectEnterCombat,
			EnemyName:  "skeleton",
		},
	}

	// Create the win room
	winRoom := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "treasure_room",
			Description: "A room filled with treasure!",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}

	// Create the starting room
	startRoom := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A dusty chamber.",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{chest},
	}

	// Create a locked door between the rooms
	door := &world.Door{
		BaseEntity: world.BaseEntity{
			Name:        "treasure_door",
			Description: "A heavy door to the treasure room",
		},
		RoomA: "room",
		RoomB: "treasure_room",
		Lock: &world.Lock{
			Locked:  true,
			KeyName: "key",
		},
	}
	startRoom.Connections["north"] = door
	winRoom.Connections["south"] = door

	// Set win condition: entering the treasure room
	winCondition := &world.Event{
		Event:    world.EventRoomEntered,
		RoomName: "treasure_room",
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{startRoom, winRoom},
		Enemies:      []*world.Enemy{enemy},
		Triggers:     []*world.Trigger{trigger},
		WinCondition: winCondition,
	})
	engine.CurrentRoom = startRoom

	// 1. Search the chest
	searchResult, err := engine.Search("chest")
	if err != nil {
		t.Fatalf("Search chest failed: %v", err)
	}
	if searchResult.Result.ContainedItemInfo == nil || searchResult.Result.ContainedItemInfo.Name != "key" {
		t.Fatalf("Expected to find key in chest, got %+v", searchResult.Result.ContainedItemInfo)
	}

	// 2. Take the key (should trigger combat)
	takeResult, err := engine.Take("key")
	if err != nil {
		t.Fatalf("Take key failed: %v", err)
	}
	if takeResult.EngineStateInfo.EngineStateChangeNotification == nil {
		t.Fatal("Expected state change notification when taking key, got nil")
	}
	if *takeResult.EngineStateInfo.EngineStateChangeNotification != EngineStateChangeEnterCombat {
		t.Errorf("Expected notification %q, got %q", EngineStateChangeEnterCombat, *takeResult.EngineStateInfo.EngineStateChangeNotification)
	}
	if engine.Mode != Combat {
		t.Errorf("Expected mode to be Combat after taking key, got %v", engine.Mode)
	}
	if engine.FightingEnemy == nil || engine.FightingEnemy.Name != "skeleton" {
		t.Errorf("Expected FightingEnemy to be skeleton, got %+v", engine.FightingEnemy)
	}

	// 3. Defeat the enemy (simulate battle)
	// Use fake RNG to guarantee player wins
	fakeRng := &FakeRng{}
	fakeRng.SetValue(0.0) // Always win
	engine.Rng = fakeRng
	battleResult, err := engine.Battle("") // Use fists
	if err != nil {
		t.Fatalf("Battle failed: %v", err)
	}
	if battleResult.EngineStateInfo.EngineStateChangeNotification == nil {
		t.Fatal("Expected state change notification after defeating enemy, got nil")
	}
	if *battleResult.EngineStateInfo.EngineStateChangeNotification != EngineStateChangeExitCombat {
		t.Errorf("Expected notification %q, got %q", EngineStateChangeExitCombat, *battleResult.EngineStateInfo.EngineStateChangeNotification)
	}
	if engine.Mode != Investigation {
		t.Errorf("Expected mode to be Investigation after defeating enemy, got %v", engine.Mode)
	}
	if engine.FightingEnemy != nil {
		t.Errorf("Expected no FightingEnemy after defeating enemy, got %+v", engine.FightingEnemy)
	}

	// 4. Unlock the door with the key
	unlockResult, err := engine.Unlock("key", "treasure_door")
	if err != nil {
		t.Fatalf("Unlock door failed: %v", err)
	}
	if !unlockResult.Result.Unlocked {
		t.Error("Expected door to be unlocked")
	}

	// 5. Traverse to the treasure room (should trigger win notification)
	traverseResult, err := engine.Traverse("north")
	if err != nil {
		t.Fatalf("Traverse to treasure room failed: %v", err)
	}
	if traverseResult.EngineStateInfo.EngineStateChangeNotification == nil {
		t.Fatal("Expected state change notification when entering treasure room, got nil")
	}
	if *traverseResult.EngineStateInfo.EngineStateChangeNotification != EngineStateChangeLevelComplete {
		t.Errorf("Expected notification %q, got %q", EngineStateChangeLevelComplete, *traverseResult.EngineStateInfo.EngineStateChangeNotification)
	}
	if engine.LevelCompletionState != LevelCompletionStateComplete {
		t.Errorf("Expected level completion state to be complete, got %v", engine.LevelCompletionState)
	}
}

func TestFireWeapon_WithAndWithoutAmmo(t *testing.T) {
	// Create a pistol (weapon, uses ammo, no ammo initially)
	pistol := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "pistol",
			Description: "A standard pistol",
		},
		Location: "room",
		Detail:   "It uses 9mm rounds.",
		Weapon: &world.Weapon{
			Damage: 0.7,
			Ammo:   &world.Ammo{Quantity: 0},
		},
		Portable: &world.Portable{},
	}

	// Create an ammo box for the pistol (3 rounds)
	ammoBox := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "pistol_ammo_box",
			Description: "A box of 9mm rounds",
		},
		Location: "room",
		Detail:   "Contains 3 rounds.",
		AmmoBox: &world.AmmoBox{
			WeaponName: "pistol",
			Ammo:       &world.Ammo{Quantity: 3},
		},
		Portable: &world.Portable{},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{pistol, ammoBox},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})
	engine.CurrentRoom = room

	// Take the pistol
	_, err := engine.Take("pistol")
	if err != nil {
		t.Fatalf("Take pistol failed: %v", err)
	}
	if engine.Player.Ammo["pistol"] != 0 {
		t.Errorf("Expected 0 ammo for pistol, got %d", engine.Player.Ammo["pistol"])
	}

	// Try to fire the pistol with no ammo (should fail)
	err = engine.Player.FireWeapon("pistol")
	if err == nil {
		t.Error("Expected error when firing pistol with no ammo, got nil")
	}

	// Take the ammo box (should add 3 rounds)
	_, err = engine.Take("pistol_ammo_box")
	if err != nil {
		t.Fatalf("Take ammo box failed: %v", err)
	}
	if engine.Player.Ammo["pistol"] != 3 {
		t.Errorf("Expected 3 ammo for pistol after taking ammo box, got %d", engine.Player.Ammo["pistol"])
	}

	// Fire the pistol (should succeed and decrease ammo)
	err = engine.Player.FireWeapon("pistol")
	if err != nil {
		t.Errorf("Expected to fire pistol successfully, got error: %v", err)
	}
	if engine.Player.Ammo["pistol"] != 2 {
		t.Errorf("Expected 2 ammo for pistol after firing, got %d", engine.Player.Ammo["pistol"])
	}
}

func TestTake_RemovesItemFromRoom(t *testing.T) {
	item := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "coin",
			Description: "A shiny coin",
		},
		Location: "room",
		Detail:   "It glints in the light.",
		Portable: &world.Portable{},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{item},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})
	engine.CurrentRoom = room

	// Take the item
	_, err := engine.Take("coin")
	if err != nil {
		t.Fatalf("Take coin failed: %v", err)
	}

	// Check that the item is no longer in the room
	for _, it := range engine.CurrentRoom.Items {
		if it.Name == "coin" {
			t.Errorf("Expected coin to be removed from room after take, but it is still present")
		}
	}
}

func TestValidation_ActionNotAllowedInMode(t *testing.T) {
	// Create a simple room setup
	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})
	engine.CurrentRoom = room

	// Test 1: Cannot battle while in investigation mode
	_, err := engine.Battle("")
	if err == nil {
		t.Error("Expected error when trying to battle in investigation mode, got nil")
	} else if err.Error() != "cannot perform this action in investigation mode" {
		t.Errorf("Expected error message 'cannot perform this action in investigation mode', got '%s'", err.Error())
	}

	// Test 2: Cannot traverse while in combat mode
	// First, enter combat mode
	enemy := &world.Enemy{
		BaseEntity: world.BaseEntity{
			Name:        "goblin",
			Description: "A fierce goblin",
		},
		Room:         "room",
		HP:           1,
		TriggerEvent: world.TriggerEventTakeItem,
	}
	engine.FightingEnemy = enemy
	engine.Mode = Combat

	_, err = engine.Traverse("north")
	if err == nil {
		t.Error("Expected error when trying to traverse in combat mode, got nil")
	} else if err.Error() != "cannot perform this action in combat mode" {
		t.Errorf("Expected error message 'cannot perform this action in combat mode', got '%s'", err.Error())
	}

	// Test 3: Cannot perform actions after level completion
	// Complete the level
	engine.LevelCompletionState = LevelCompletionStateComplete
	engine.Mode = Investigation // Reset to investigation mode
	engine.FightingEnemy = nil  // Clear the fighting enemy

	_, err = engine.Observe()
	if err == nil {
		t.Error("Expected error when trying to observe after level completion, got nil")
	} else if err.Error() != "level is already complete" {
		t.Errorf("Expected error message 'level is already complete', got '%s'", err.Error())
	}
}

func TestEventHandling_CombatEntryExitMultipleEnemies(t *testing.T) {
	// Create demos
	room1 := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "Room1",
			Description: "The first room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}
	room2 := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "Room2",
			Description: "The second room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}
	room3 := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "Room3",
			Description: "The third room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{},
	}

	// Create doors between rooms
	door1 := &world.Door{
		BaseEntity: world.BaseEntity{
			Name:        "door1",
			Description: "A door between Room1 and Room2",
		},
		RoomA: "Room1",
		RoomB: "Room2",
		Lock:  nil,
	}
	door2 := &world.Door{
		BaseEntity: world.BaseEntity{
			Name:        "door2",
			Description: "A door between Room2 and Room3",
		},
		RoomA: "Room2",
		RoomB: "Room3",
		Lock:  nil,
	}
	room1.Connections["east"] = door1
	room2.Connections["west"] = door1
	room2.Connections["east"] = door2
	room3.Connections["west"] = door2

	// Create two enemies
	enemy1 := &world.Enemy{
		BaseEntity: world.BaseEntity{
			Name:        "goblin",
			Description: "A fierce goblin",
		},
		Room: "Room2",
		HP:   1,
	}
	enemy2 := &world.Enemy{
		BaseEntity: world.BaseEntity{
			Name:        "skeleton",
			Description: "A rattling skeleton",
		},
		Room: "Room3",
		HP:   1,
	}

	// Create triggers for entering rooms
	trigger1 := &world.Trigger{
		Event: world.Event{
			Event:    world.EventRoomEntered,
			RoomName: "Room2",
		},
		Effect: world.Effect{
			EffectType: world.EffectEnterCombat,
			EnemyName:  "goblin",
		},
	}
	trigger2 := &world.Trigger{
		Event: world.Event{
			Event:    world.EventRoomEntered,
			RoomName: "Room3",
		},
		Effect: world.Effect{
			EffectType: world.EffectEnterCombat,
			EnemyName:  "skeleton",
		},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room1, room2, room3},
		Enemies:      []*world.Enemy{enemy1, enemy2},
		Triggers:     []*world.Trigger{trigger1, trigger2},
		WinCondition: nil,
	})
	engine.CurrentRoom = room1

	// Set a win condition to prevent nil pointer dereference
	engine.Level.WinCondition = &world.Event{
		Event:    world.EventRoomEntered,
		RoomName: "nonexistent_room", // Room that won't be entered in this test
	}

	// Use fake RNG to control battle outcomes
	fakeRng := &FakeRng{}
	engine.Rng = fakeRng

	// Test 1: Enter Room2 (should trigger combat with goblin)
	result, err := engine.Traverse("east")
	if err != nil {
		t.Fatalf("Traverse to Room2 failed: %v", err)
	}
	if result.EngineStateInfo.EngineStateChangeNotification == nil {
		t.Fatal("Expected state change notification when entering Room2, got nil")
	}
	if *result.EngineStateInfo.EngineStateChangeNotification != EngineStateChangeEnterCombat {
		t.Errorf("Expected notification %q, got %q", EngineStateChangeEnterCombat, *result.EngineStateInfo.EngineStateChangeNotification)
	}
	if engine.Mode != Combat {
		t.Errorf("Expected mode to be Combat after entering Room2, got %v", engine.Mode)
	}
	if engine.FightingEnemy == nil || engine.FightingEnemy.Name != "goblin" {
		t.Errorf("Expected FightingEnemy to be goblin, got %+v", engine.FightingEnemy)
	}

	// Test 2: Defeat the goblin (should exit combat)
	fakeRng.SetValue(0.0) // Always win
	battleResult, err := engine.Battle("")
	if err != nil {
		t.Fatalf("Battle with goblin failed: %v", err)
	}
	if battleResult.EngineStateInfo.EngineStateChangeNotification == nil {
		t.Fatal("Expected state change notification after defeating goblin, got nil")
	}
	if *battleResult.EngineStateInfo.EngineStateChangeNotification != EngineStateChangeExitCombat {
		t.Errorf("Expected notification %q, got %q", EngineStateChangeExitCombat, *battleResult.EngineStateInfo.EngineStateChangeNotification)
	}
	if engine.Mode != Investigation {
		t.Errorf("Expected mode to be Investigation after defeating goblin, got %v", engine.Mode)
	}
	if engine.FightingEnemy != nil {
		t.Errorf("Expected no FightingEnemy after defeating goblin, got %+v", engine.FightingEnemy)
	}

	// Test 3: Enter Room3 (should trigger combat with skeleton)
	result, err = engine.Traverse("east")
	if err != nil {
		t.Fatalf("Traverse to Room3 failed: %v", err)
	}
	if result.EngineStateInfo.EngineStateChangeNotification == nil {
		t.Fatal("Expected state change notification when entering Room3, got nil")
	}
	if *result.EngineStateInfo.EngineStateChangeNotification != EngineStateChangeEnterCombat {
		t.Errorf("Expected notification %q, got %q", EngineStateChangeEnterCombat, *result.EngineStateInfo.EngineStateChangeNotification)
	}
	if engine.Mode != Combat {
		t.Errorf("Expected mode to be Combat after entering Room3, got %v", engine.Mode)
	}
	if engine.FightingEnemy == nil || engine.FightingEnemy.Name != "skeleton" {
		t.Errorf("Expected FightingEnemy to be skeleton, got %+v", engine.FightingEnemy)
	}

	// Test 4: Defeat the skeleton (should exit combat)
	fakeRng.SetValue(0.0) // Always win
	battleResult, err = engine.Battle("")
	if err != nil {
		t.Fatalf("Battle with skeleton failed: %v", err)
	}
	if battleResult.EngineStateInfo.EngineStateChangeNotification == nil {
		t.Fatal("Expected state change notification after defeating skeleton, got nil")
	}
	if *battleResult.EngineStateInfo.EngineStateChangeNotification != EngineStateChangeExitCombat {
		t.Errorf("Expected notification %q, got %q", EngineStateChangeExitCombat, *battleResult.EngineStateInfo.EngineStateChangeNotification)
	}
	if engine.Mode != Investigation {
		t.Errorf("Expected mode to be Investigation after defeating skeleton, got %v", engine.Mode)
	}
	if engine.FightingEnemy != nil {
		t.Errorf("Expected no FightingEnemy after defeating skeleton, got %+v", engine.FightingEnemy)
	}
}

func TestDebug_SimpleRoomSetup(t *testing.T) {
	// Create a gem
	gem := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "gem",
			Description: "A sparkling gem",
		},
		Location: "chest",
		Detail:   "It glitters brightly.",
		Portable: &world.Portable{},
	}

	// Create an unlocked chest containing the gem
	chest := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "chest",
			Description: "A wooden chest",
		},
		Location: "room",
		Detail:   "It could contain something.",
		Container: &world.Container{
			Contains: gem,
			Searched: false,
			Locked:   nil, // unlocked
		},
	}

	// Create a room with the chest
	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A simple test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{chest},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})
	engine.CurrentRoom = room

	// Get debug information
	debug, err := engine.Debug()
	if err != nil {
		t.Fatalf("Debug failed: %v", err)
	}

	// Print debug output if flag is set
	if debugFlag {
		t.Logf("Debug Output:\n%s", debug.PrettyPrint())
	}

	// Verify engine state
	if debug.EngineState.Mode != "investigation" {
		t.Errorf("Expected mode 'investigation', got '%s'", debug.EngineState.Mode)
	}
	if debug.EngineState.LevelCompletionState != "in_progress" {
		t.Errorf("Expected level completion state 'in_progress', got '%s'", debug.EngineState.LevelCompletionState)
	}
	if debug.EngineState.CurrentRoom != "room" {
		t.Errorf("Expected current room 'room', got '%s'", debug.EngineState.CurrentRoom)
	}
	if debug.EngineState.FightingEnemy != nil {
		t.Error("Expected no fighting enemy, got one")
	}

	// Verify player state
	if debug.Player.Health != "fine" {
		t.Errorf("Expected player health 'fine', got '%s'", debug.Player.Health)
	}
	if !debug.Player.IsAlive {
		t.Error("Expected player to be alive")
	}
	if len(debug.Player.Inventory) != 0 {
		t.Errorf("Expected empty inventory, got %d items", len(debug.Player.Inventory))
	}
	if len(debug.Player.Ammo) != 0 {
		t.Errorf("Expected no ammo, got %d ammo types", len(debug.Player.Ammo))
	}

	// Verify rooms
	if len(debug.Rooms) != 1 {
		t.Fatalf("Expected 1 room, got %d", len(debug.Rooms))
	}
	debugRoom := debug.Rooms[0]
	if debugRoom.Name != "room" {
		t.Errorf("Expected room name 'room', got '%s'", debugRoom.Name)
	}
	if debugRoom.Description != "A simple test room" {
		t.Errorf("Expected room description 'A simple test room', got '%s'", debugRoom.Description)
	}
	if !debugRoom.IsCurrent {
		t.Error("Expected room to be marked as current")
	}
	if len(debugRoom.Doors) != 0 {
		t.Errorf("Expected no doors, got %d", len(debugRoom.Doors))
	}

	// Verify items in room
	if len(debugRoom.Items) != 1 {
		t.Fatalf("Expected 1 item in room, got %d", len(debugRoom.Items))
	}
	debugChest := debugRoom.Items[0]
	if debugChest.Name != "chest" {
		t.Errorf("Expected chest name 'chest', got '%s'", debugChest.Name)
	}
	if debugChest.Description != "A wooden chest" {
		t.Errorf("Expected chest description 'A wooden chest', got '%s'", debugChest.Description)
	}
	if debugChest.Location != "room" {
		t.Errorf("Expected chest location 'room', got '%s'", debugChest.Location)
	}
	if debugChest.Detail != "It could contain something." {
		t.Errorf("Expected chest detail 'It could contain something.', got '%s'", debugChest.Detail)
	}
	if debugChest.IsPortable {
		t.Error("Expected chest to not be portable")
	}
	if !debugChest.IsContainer {
		t.Error("Expected chest to be a container")
	}
	if debugChest.IsConcealer {
		t.Error("Expected chest to not be a concealer")
	}
	if debugChest.IsKey {
		t.Error("Expected chest to not be a key")
	}
	if debugChest.IsAmmoBox {
		t.Error("Expected chest to not be an ammo box")
	}
	if debugChest.IsWeapon {
		t.Error("Expected chest to not be a weapon")
	}
	if debugChest.IsHealthItem {
		t.Error("Expected chest to not be a health item")
	}

	// Verify container properties
	if debugChest.HasKeyLock {
		t.Error("Expected chest to not have key lock")
	}
	if debugChest.HasCodeLock {
		t.Error("Expected chest to not have code lock")
	}
	if debugChest.IsLocked {
		t.Error("Expected chest to not be locked")
	}
	if debugChest.IsSearched {
		t.Error("Expected chest to not be searched")
	}

	// Verify contained item
	if debugChest.Contains == nil {
		t.Fatal("Expected chest to contain an item")
	}
	debugGem := debugChest.Contains
	if debugGem.Name != "gem" {
		t.Errorf("Expected gem name 'gem', got '%s'", debugGem.Name)
	}
	if debugGem.Description != "A sparkling gem" {
		t.Errorf("Expected gem description 'A sparkling gem', got '%s'", debugGem.Description)
	}
	if debugGem.Location != "chest" {
		t.Errorf("Expected gem location 'chest', got '%s'", debugGem.Location)
	}
	if debugGem.Detail != "It glitters brightly." {
		t.Errorf("Expected gem detail 'It glitters brightly.', got '%s'", debugGem.Detail)
	}
	if !debugGem.IsPortable {
		t.Error("Expected gem to be portable")
	}
	if debugGem.IsContainer {
		t.Error("Expected gem to not be a container")
	}
	if debugGem.IsConcealer {
		t.Error("Expected gem to not be a concealer")
	}
	if debugGem.IsKey {
		t.Error("Expected gem to not be a key")
	}
	if debugGem.IsAmmoBox {
		t.Error("Expected gem to not be an ammo box")
	}
	if debugGem.IsWeapon {
		t.Error("Expected gem to not be a weapon")
	}
	if debugGem.IsHealthItem {
		t.Error("Expected gem to not be a health item")
	}

	// Verify other collections are empty
	if len(debug.Enemies) != 0 {
		t.Errorf("Expected no enemies, got %d", len(debug.Enemies))
	}
	if len(debug.Triggers) != 0 {
		t.Errorf("Expected no triggers, got %d", len(debug.Triggers))
	}
	if debug.WinCondition != nil {
		t.Error("Expected no win condition, got one")
	}
}

func TestDebug_SearchChestAndTakeGem(t *testing.T) {
	// Create a gem
	gem := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "gem",
			Description: "A sparkling gem",
		},
		Location: "chest",
		Detail:   "It glitters brightly.",
		Portable: &world.Portable{},
	}

	// Create an unlocked chest containing the gem
	chest := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "chest",
			Description: "A wooden chest",
		},
		Location: "room",
		Detail:   "It could contain something.",
		Container: &world.Container{
			Contains: gem,
			Searched: false,
			Locked:   nil, // unlocked
		},
	}

	// Create a room with the chest
	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "room",
			Description: "A simple test room",
		},
		Connections: map[string]*world.Door{},
		Items:       []*world.Item{chest},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})
	engine.CurrentRoom = room

	_, err := engine.Debug()
	if err != nil {
		t.Fatalf("Debug failed: %v", err)
	}

	// Search the chest
	_, err = engine.Search("chest")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Print debug state after searching
	debug, err := engine.Debug()
	if err != nil {
		t.Fatalf("Debug failed: %v", err)
	}
	if debugFlag {
		t.Logf("Debug Output:\n%s", debug.PrettyPrint())
	}

	// Take the gem
	_, err = engine.Take("gem")
	if err != nil {
		t.Fatalf("Take failed: %v", err)
	}

	// Print debug state after taking gem
	debug, err = engine.Debug()
	if err != nil {
		t.Fatalf("Debug failed: %v", err)
	}
	if debugFlag {
		t.Logf("Debug Output:\n%s", debug.PrettyPrint())
	}
}

func TestWeaponAmmo_StartWithOneBullet(t *testing.T) {
	// Create a test level with a weapon that starts with 1 bullet
	weapon := &world.Item{
		BaseEntity: world.BaseEntity{
			Name:        "pistol",
			Description: "a 9mm pistol",
		},
		Location: "on the floor",
		Detail:   "it appears to be in working condition",
		Weapon: &world.Weapon{
			Damage: 0.9,
			Ammo: &world.Ammo{
				Quantity: 1,
			},
		},
		Portable: &world.Portable{},
	}

	room := &world.Room{
		BaseEntity: world.BaseEntity{
			Name:        "test_room",
			Description: "A test room",
		},
		Connections: make(map[string]*world.Door),
		Items:       []*world.Item{weapon},
	}

	engine := NewEngine(&world.Level{
		Rooms:        []*world.Room{room},
		WinCondition: nil,
	})

	// Take the weapon
	takeResult, err := engine.Take("pistol")
	if err != nil {
		t.Fatalf("Take pistol failed: %v", err)
	}

	if takeResult.Result.ItemInfo.Name != "pistol" {
		t.Errorf("Expected to take pistol, got %s", takeResult.Result.ItemInfo.Name)
	}

	// Verify initial ammo count
	if engine.Player.Ammo["pistol"] != 1 {
		t.Errorf("Expected pistol to have 1 ammo, got %d", engine.Player.Ammo["pistol"])
	}

	// Try to fire the weapon (should succeed and consume the ammo)
	err = engine.Player.FireWeapon("pistol")
	if err != nil {
		t.Errorf("Expected to be able to fire pistol with 1 ammo, got error: %v", err)
	}

	// Verify ammo count after firing
	if engine.Player.Ammo["pistol"] != 0 {
		t.Errorf("Expected pistol to have 0 ammo after firing, got %d", engine.Player.Ammo["pistol"])
	}

	// Try to fire again (should fail - out of ammo)
	err = engine.Player.FireWeapon("pistol")
	if err == nil {
		t.Error("Expected error when firing pistol with 0 ammo, got nil")
	}

	// Verify ammo count is still 0
	if engine.Player.Ammo["pistol"] != 0 {
		t.Errorf("Expected pistol to still have 0 ammo after failed fire, got %d", engine.Player.Ammo["pistol"])
	}
}

func TestIntegration_DemoPuzzleComplete(t *testing.T) {
	// Load the demo puzzle game
	level, err := loader.LoadGame("../testdata/demo.json")
	if err != nil {
		t.Fatalf("Failed to load demo game: %v", err)
	}

	// Create engine from the loaded level
	engine := NewEngine(level)

	// 1. Observe the room and pretty print the result
	observeResult, err := engine.Observe()
	if err != nil {
		t.Fatalf("Observe failed: %v", err)
	}

	if debugFlag {
		t.Logf("Starting in room: %s", observeResult.Result.RoomName)
		t.Logf("Initial room observation:\n%s", observeResult.Result.RoomDescription)
		t.Logf("Visible items: %d", len(observeResult.Result.VisibleItems))
		for _, item := range observeResult.Result.VisibleItems {
			if item.IsContainer && item.Contains != "" {
				t.Logf("  - %s: %s (contains %s)", item.Name, item.Description, item.Contains)
			} else {
				t.Logf("  - %s: %s", item.Name, item.Description)
			}
		}
	}

	// Verify we're in the waiting room
	if observeResult.Result.RoomName != "waiting room" {
		t.Errorf("Expected to start in waiting room, got %s", observeResult.Result.RoomName)
	}

	// Verify we can see the hoodie and energy drink
	visibleItems := make(map[string]bool)
	for _, item := range observeResult.Result.VisibleItems {
		visibleItems[item.Name] = true
	}
	if !visibleItems["tattered grey hoodie"] {
		t.Error("Expected to see tattered grey hoodie in waiting room")
	}
	if !visibleItems["energy drink"] {
		t.Error("Expected to see energy drink in waiting room")
	}

	// 2. Uncover the hoodie
	uncoverResult, err := engine.Uncover("tattered grey hoodie")
	if err != nil {
		t.Fatalf("Uncover hoodie failed: %v", err)
	}

	if uncoverResult.Result.RevealedItem.Name != "ominous note" {
		t.Errorf("Expected to reveal ominous note, got %s", uncoverResult.Result.RevealedItem.Name)
	}

	if debugFlag {
		t.Logf("Uncovered: %s", uncoverResult.Result.RevealedItem.Name)
	}

	// 3. Inspect the note
	inspectResult, err := engine.Inspect("ominous note")
	if err != nil {
		t.Fatalf("Inspect note failed: %v", err)
	}

	if inspectResult.Result.itemInspection == nil {
		t.Fatal("Expected item inspection result, got nil")
	}

	if !strings.Contains(inspectResult.Result.itemInspection.Detail, "got to get away from that thing") {
		t.Errorf("Expected to inspect ominous note, got %s", inspectResult.Result.itemInspection.Detail)
	}

	if debugFlag {
		t.Logf("Note detail: %s", inspectResult.Result.itemInspection.Detail)
	}

	// 4. Go north to the office
	traverseResult, err := engine.Traverse("north")
	if err != nil {
		t.Fatalf("Traverse to office failed: %v", err)
	}

	if traverseResult.Result.ToRoom != "office" {
		t.Errorf("Expected to traverse to office, got %s", traverseResult.Result.ToRoom)
	}

	if debugFlag {
		t.Logf("Traversed to: %s", traverseResult.Result.ToRoom)
	}

	// Observe the office
	observeResult, err = engine.Observe()
	if err != nil {
		t.Fatalf("Observe office failed: %v", err)
	}

	if observeResult.Result.RoomName != "office" {
		t.Errorf("Expected to be in office, got %s", observeResult.Result.RoomName)
	}

	if debugFlag {
		t.Logf("Office observation: %s", observeResult.Result.RoomDescription)
		t.Logf("Office visible items: %d", len(observeResult.Result.VisibleItems))
		for _, item := range observeResult.Result.VisibleItems {
			if item.IsContainer && item.Contains != "" {
				t.Logf("  - %s: %s (contains %s)", item.Name, item.Description, item.Contains)
			} else {
				t.Logf("  - %s: %s", item.Name, item.Description)
			}
		}
	}

	// Verify we can see the desk and cardboard box
	visibleItems = make(map[string]bool)
	for _, item := range observeResult.Result.VisibleItems {
		visibleItems[item.Name] = true
	}
	if !visibleItems["desk"] {
		t.Error("Expected to see desk in office")
	}
	if !visibleItems["cardboard box"] {
		t.Error("Expected to see cardboard box in office")
	}

	// 5. Search the desk drawer
	searchResult, err := engine.Search("desk")
	if err != nil {
		t.Fatalf("Search desk failed: %v", err)
	}

	// Observe the office again
	observeResult, err = engine.Observe()
	if err != nil {
		t.Fatalf("Observe office failed: %v", err)
	}

	if debugFlag {
		t.Logf("Office observation after searching desk: %s", observeResult.Result.RoomDescription)
		t.Logf("Office visible items: %d", len(observeResult.Result.VisibleItems))
		for _, item := range observeResult.Result.VisibleItems {
			if item.IsContainer && item.Contains != "" {
				t.Logf("  - %s: %s (contains %s)", item.Name, item.Description, item.Contains)
			} else {
				t.Logf("  - %s: %s", item.Name, item.Description)
			}
		}
	}

	if searchResult.Result.ContainedItemInfo == nil {
		t.Fatal("Expected to find item in desk")
	}

	if searchResult.Result.ContainedItemInfo.Name != "pistol" {
		t.Errorf("Expected to find pistol in desk, got %s", searchResult.Result.ContainedItemInfo.Name)
	}

	if debugFlag {
		t.Logf("Found in desk: %s", searchResult.Result.ContainedItemInfo.Name)
	}

	// 6. Take the pistol
	takeResult, err := engine.Take("pistol")
	if err != nil {
		t.Fatalf("Take pistol failed: %v", err)
	}

	if takeResult.Result.ItemInfo.Name != "pistol" {
		t.Errorf("Expected to take pistol, got %s", takeResult.Result.ItemInfo.Name)
	}

	// Verify pistol is in inventory
	inventoryResult, err := engine.Inventory()
	if err != nil {
		t.Fatalf("Inventory failed: %v", err)
	}

	pistolInInventory := false
	for _, item := range inventoryResult.Result.Items {
		if item.Name == "pistol" {
			pistolInInventory = true
			break
		}
	}
	if !pistolInInventory {
		t.Error("Expected pistol to be in inventory")
	}

	if debugFlag {
		t.Logf("Took pistol, ammo: %d", engine.Player.Ammo["pistol"])
	}

	// 6b. Search the cardboard box
	searchResult, err = engine.Search("cardboard box")
	if err != nil {
		t.Fatalf("Search cardboard box failed: %v", err)
	}

	if searchResult.Result.ContainedItemInfo == nil {
		t.Fatal("Expected to find item in cardboard box")
	}

	if searchResult.Result.ContainedItemInfo.Name != "pistol ammo" {
		t.Errorf("Expected to find pistol ammo in cardboard box, got %s", searchResult.Result.ContainedItemInfo.Name)
	}

	if debugFlag {
		t.Logf("Found in cardboard box: %s", searchResult.Result.ContainedItemInfo.Name)
	}

	// 6c. Take the pistol ammo
	takeResult, err = engine.Take("pistol ammo")
	if err != nil {
		t.Fatalf("Take pistol ammo failed: %v", err)
	}

	if takeResult.Result.ItemInfo.Name != "pistol ammo" {
		t.Errorf("Expected to take pistol ammo, got %s", takeResult.Result.ItemInfo.Name)
	}

	if debugFlag {
		t.Logf("Took pistol ammo, total ammo: %d", engine.Player.Ammo["pistol"])
	}

	// 7. Go back to the first room
	traverseResult, err = engine.Traverse("south")
	if err != nil {
		t.Fatalf("Traverse back to waiting room failed: %v", err)
	}

	if traverseResult.Result.ToRoom != "waiting room" {
		t.Errorf("Expected to traverse back to waiting room, got %s", traverseResult.Result.ToRoom)
	}

	// 8. Go east to storage room
	traverseResult, err = engine.Traverse("east")
	if err != nil {
		t.Fatalf("Traverse to storage room failed: %v", err)
	}

	if traverseResult.Result.ToRoom != "storage room" {
		t.Errorf("Expected to traverse to storage room, got %s", traverseResult.Result.ToRoom)
	}

	// Observe the storage room
	observeResult, err = engine.Observe()
	if err != nil {
		t.Fatalf("Observe storage room failed: %v", err)
	}

	// Verify we can see the tarp, newspapers, filing cabinet, and metal pipe
	visibleItems = make(map[string]bool)
	for _, item := range observeResult.Result.VisibleItems {
		visibleItems[item.Name] = true
	}
	if !visibleItems["dark green tarp"] {
		t.Error("Expected to see dark green tarp in storage room")
	}
	if !visibleItems["stack of newspapers"] {
		t.Error("Expected to see stack of newspapers in storage room")
	}
	if !visibleItems["filing cabinet"] {
		t.Error("Expected to see filing cabinet in storage room")
	}
	if !visibleItems["metal pipe"] {
		t.Error("Expected to see metal pipe in storage room")
	}

	if debugFlag {
		t.Logf("Storage room before uncovering tarp - visible top level items: %d", len(observeResult.Result.VisibleItems))
		for _, item := range observeResult.Result.VisibleItems {
			if item.IsContainer && item.Contains != "" {
				t.Logf("  - %s: %s (contains %s)", item.Name, item.Description, item.Contains)
			} else {
				t.Logf("  - %s: %s", item.Name, item.Description)
			}
		}
	}

	// 9. Uncover the tarp
	uncoverResult, err = engine.Uncover("dark green tarp")
	if err != nil {
		t.Fatalf("Uncover tarp failed: %v", err)
	}

	if uncoverResult.Result.RevealedItem.Name != "safe" {
		t.Errorf("Expected to reveal safe, got %s", uncoverResult.Result.RevealedItem.Name)
	}

	if debugFlag {
		t.Logf("Uncovered: %s", uncoverResult.Result.RevealedItem.Name)
	}

	// Observe the storage room after uncovering the safe
	observeResult, err = engine.Observe()
	if err != nil {
		t.Fatalf("Observe storage room after uncovering failed: %v", err)
	}

	if debugFlag {
		t.Logf("Storage room after uncovering tarp - visible top level items: %d", len(observeResult.Result.VisibleItems))
		for _, item := range observeResult.Result.VisibleItems {
			if item.IsContainer && item.Contains != "" {
				t.Logf("  - %s: %s (contains %s)", item.Name, item.Description, item.Contains)
			} else {
				t.Logf("  - %s: %s", item.Name, item.Description)
			}
		}
	}

	// 10. Enter a wrong code in the safe
	_, err = engine.Unlock("1234", "safe")
	if err == nil {
		t.Error("Expected error when entering wrong code, got nil")
	}

	if debugFlag {
		t.Logf("Wrong code attempt failed as expected")
	}

	// 10b. try searching the safe
	_, err = engine.Search("safe")
	if err == nil {
		t.Fatalf("Expected error when searching locked safe, got nil")
	}

	// 11. Enter the correct code
	unlockResult, err := engine.Unlock("2468", "safe")
	if err != nil {
		t.Fatalf("Unlock safe with correct code failed: %v", err)
	}

	if !unlockResult.Result.Unlocked {
		t.Error("Expected safe to be unlocked")
	}

	if debugFlag {
		t.Logf("Safe unlocked successfully")
	}

	observeResult, err = engine.Observe()
	if err != nil {
		t.Fatalf("Observe safe after unlocking failed: %v", err)
	}

	if debugFlag {
		t.Logf("Storage room after unlocking - visible top level items: %d", len(observeResult.Result.VisibleItems))
		for _, item := range observeResult.Result.VisibleItems {
			if item.IsContainer && item.Contains != "" {
				t.Logf("  - %s: %s (contains %s)", item.Name, item.Description, item.Contains)
			} else {
				t.Logf("  - %s: %s", item.Name, item.Description)
			}
		}
	}

	// 12. Search the safe
	searchResult, _ = engine.Search("safe")

	observeResult, err = engine.Observe()
	if err != nil {
		t.Fatalf("Observe safe after unlocking failed: %v", err)
	}

	if debugFlag {
		t.Logf("Storage room after searching - visible top level items: %d", len(observeResult.Result.VisibleItems))
		for _, item := range observeResult.Result.VisibleItems {
			if item.IsContainer && item.Contains != "" {
				t.Logf("  - %s: %s (contains %s)", item.Name, item.Description, item.Contains)
			} else {
				t.Logf("  - %s: %s", item.Name, item.Description)
			}
		}
	}

	if err != nil {
		t.Fatalf("Search safe failed: %v", err)
	}

	if searchResult.Result.ContainedItemInfo == nil {
		t.Fatal("Expected to find item in safe")
	}

	if searchResult.Result.ContainedItemInfo.Name != "iron key" {
		t.Errorf("Expected to find iron key in safe, got %s", searchResult.Result.ContainedItemInfo.Name)
	}

	if debugFlag {
		t.Logf("Found in safe: %s", searchResult.Result.ContainedItemInfo.Name)
	}

	// 13. Take the iron key (this should trigger combat!)
	takeResult, err = engine.Take("iron key")
	if err != nil {
		t.Fatalf("Take iron key failed: %v", err)
	}

	if takeResult.Result.ItemInfo.Name != "iron key" {
		t.Errorf("Expected to take iron key, got %s", takeResult.Result.ItemInfo.Name)
	}

	// Verify we entered combat mode
	if takeResult.EngineStateInfo.EngineStateChangeNotification == nil {
		t.Fatal("Expected state change notification when taking iron key")
	}
	if *takeResult.EngineStateInfo.EngineStateChangeNotification != EngineStateChangeEnterCombat {
		t.Errorf("Expected enter combat notification, got %s", *takeResult.EngineStateInfo.EngineStateChangeNotification)
	}
	if engine.Mode != Combat {
		t.Errorf("Expected to be in combat mode, got %s", engine.Mode)
	}
	if engine.FightingEnemy == nil || engine.FightingEnemy.Name != "zombie" {
		t.Errorf("Expected to be fighting zombie, got %+v", engine.FightingEnemy)
	}

	if debugFlag {
		t.Logf("Entered combat with: %s", engine.FightingEnemy.Name)
	}

	// 14. Defeat the enemy
	// Use fake RNG to guarantee victory
	fakeRng := &FakeRng{}
	fakeRng.SetValue(0.0) // Always win
	engine.Rng = fakeRng

	if debugFlag {
		t.Logf("Before battle - pistol ammo: %d", engine.Player.Ammo["pistol"])
	}

	battleResult, err := engine.Battle("pistol")
	if err != nil {
		t.Fatalf("Battle with pistol failed: %v", err)
	}

	if debugFlag {
		t.Logf("After battle - pistol ammo: %d", engine.Player.Ammo["pistol"])
	}

	if !battleResult.Result.WonRound {
		t.Error("Expected to win the battle")
	}
	if battleResult.Result.EnemyAlive {
		t.Error("Expected enemy to be dead after battle")
	}

	// Verify we exited combat mode
	if battleResult.EngineStateInfo.EngineStateChangeNotification == nil {
		t.Fatal("Expected state change notification after defeating enemy")
	}
	if *battleResult.EngineStateInfo.EngineStateChangeNotification != EngineStateChangeExitCombat {
		t.Errorf("Expected exit combat notification, got %s", *battleResult.EngineStateInfo.EngineStateChangeNotification)
	}
	if engine.Mode != Investigation {
		t.Errorf("Expected to be back in investigation mode, got %s", engine.Mode)
	}
	if engine.FightingEnemy != nil {
		t.Error("Expected no fighting enemy after battle")
	}

	if debugFlag {
		t.Logf("Defeated zombie, returned to investigation mode")
	}

	// 15. Use the key to complete the level
	// First, go back to waiting room
	_, err = engine.Traverse("west")
	if err != nil {
		t.Fatalf("Traverse back to waiting room failed: %v", err)
	}

	// Unlock the metal stairwell door
	unlockResult, err = engine.Unlock("iron key", "metal stairwell door")
	if err != nil {
		t.Fatalf("Unlock metal stairwell door failed: %v", err)
	}

	if !unlockResult.Result.Unlocked {
		t.Error("Expected door to be unlocked")
	}

	if debugFlag {
		t.Logf("Unlocked metal stairwell door")
	}

	// Traverse to the stairwell (this should trigger win condition!)
	traverseResult, err = engine.Traverse("west")
	if err != nil {
		t.Fatalf("Traverse to stairwell failed: %v", err)
	}

	if traverseResult.Result.ToRoom != "stairwell to roof" {
		t.Errorf("Expected to traverse to stairwell to roof, got %s", traverseResult.Result.ToRoom)
	}

	// Verify we completed the level
	if traverseResult.EngineStateInfo.EngineStateChangeNotification == nil {
		t.Fatal("Expected state change notification when entering stairwell")
	}
	if *traverseResult.EngineStateInfo.EngineStateChangeNotification != EngineStateChangeLevelComplete {
		t.Errorf("Expected level complete notification, got %s", *traverseResult.EngineStateInfo.EngineStateChangeNotification)
	}
	if traverseResult.EngineStateInfo.LevelCompletionState != LevelCompletionStateComplete {
		t.Errorf("Expected level completion state to be complete, got %s", traverseResult.EngineStateInfo.LevelCompletionState)
	}

	if debugFlag {
		t.Logf("Level completed successfully!")
	}

	// Final debug output
	if debugFlag {
		debugResult, err := engine.Debug()
		if err != nil {
			t.Fatalf("Final debug failed: %v", err)
		}
		t.Logf("Final game state:\n%s", debugResult.PrettyPrint())
	}

	if debugFlag {
		t.Logf("Integration test completed successfully!")
	}
}
