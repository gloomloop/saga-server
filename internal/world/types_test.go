package world

import "testing"

func TestRoomWithNonPortableItem(t *testing.T) {
	// Create a non-portable item
	item := &Item{
		BaseEntity: BaseEntity{
			Name:        "rock",
			Description: "A heavy rock that cannot be moved",
		},
		Location: "test_room",
		Detail:   "It's just sitting there",
	}

	if err := item.ValidateInitialState(); err != nil {
		t.Errorf("Failed to validate item: %v", err)
	}

	// Create a room with no doors and the item
	room := &Room{
		BaseEntity: BaseEntity{
			Name:        "test_room",
			Description: "A simple test room",
		},
		Connections: []*Connection{}, // no doors
		Items:       []*Item{item},
	}

	// Test that we can get the item
	if len(room.Items) != 1 {
		t.Errorf("Expected 1 item in room, got %d", len(room.Items))
	}

	if room.Items[0].Name != "rock" {
		t.Errorf("Expected item name 'rock', got '%s'", room.Items[0].Name)
	}

	if room.Items[0].IsPortable() {
		t.Error("Expected item to not be portable")
	}
}

func TestPortableItemTransfer(t *testing.T) {
	// Create a portable item
	item := &Item{
		BaseEntity: BaseEntity{
			Name:        "key",
			Description: "A small brass key",
		},
		Location: "test_room",
		Detail:   "It looks like it might unlock something",
		Portable: &Portable{}, // Make it portable
	}

	if err := item.ValidateInitialState(); err != nil {
		t.Errorf("Failed to validate item: %v", err)
	}

	// Create a room with the portable item
	room := &Room{
		BaseEntity: BaseEntity{
			Name:        "test_room",
			Description: "A simple test room",
		},
		Connections: []*Connection{},
		Items:       []*Item{item},
	}

	// Create a player
	player := &Player{
		Inventory: []*Item{},
		Health:    HealthState(HealthFine),
	}

	// Test that the item is portable
	if !item.IsPortable() {
		t.Error("Expected item to be portable")
	}

	// Test that we can remove the item from the room
	removedItem, err := room.RemoveItem("key")
	if err != nil {
		t.Errorf("Failed to remove item: %v", err)
	}

	if removedItem.Name != "key" {
		t.Errorf("Expected removed item name 'key', got '%s'", removedItem.Name)
	}

	// Test that the room no longer has the item
	if len(room.Items) != 0 {
		t.Errorf("Expected 0 items in room after removal, got %d", len(room.Items))
	}

	// Test that we can add the item to player inventory
	player.Inventory = append(player.Inventory, removedItem)
	if len(player.Inventory) != 1 {
		t.Errorf("Expected 1 item in player inventory, got %d", len(player.Inventory))
	}

	if player.Inventory[0].Name != "key" {
		t.Errorf("Expected inventory item name 'key', got '%s'", player.Inventory[0].Name)
	}
}

func TestContainerSearch(t *testing.T) {
	// Create a container with no lock and no contents
	container := &Item{
		BaseEntity: BaseEntity{
			Name:        "chest",
			Description: "An empty wooden chest",
		},
		Location: "test_room",
		Detail:   "It looks like it might contain something",
		Container: &Container{
			Contains: nil, // no contents
			Searched: false,
			Locked:   nil, // no lock
		},
	}

	if err := container.ValidateInitialState(); err != nil {
		t.Errorf("Failed to validate container: %v", err)
	}

	// Note: container is created but not added to a room for this test

	// Test that the item is a container
	if !container.IsContainer() {
		t.Error("Expected item to be a container")
	}

	// Test that the container has no lock
	if container.Container.HasLock() {
		t.Error("Expected container to have no lock")
	}

	// Test that the container is not locked
	if container.Container.IsLocked() {
		t.Error("Expected container to not be locked")
	}

	// Test searching the container
	foundItem, err := container.Container.Search()
	if err != nil {
		t.Errorf("Failed to search container: %v", err)
	}

	// Should return nil since container is empty
	if foundItem != nil {
		t.Errorf("Expected no item found, got %v", foundItem)
	}

	// Test that the container is now marked as searched
	if !container.Container.Searched {
		t.Error("Expected container to be marked as searched")
	}

	// Test searching again should return nil (already searched)
	foundItem2, err := container.Container.Search()
	if err != nil {
		t.Errorf("Failed to search container again: %v", err)
	}

	if foundItem2 != nil {
		t.Errorf("Expected no item found on second search, got %v", foundItem2)
	}
}

func TestLockedSafeWithKeypad(t *testing.T) {
	// Create a note to put inside the safe
	note := &Item{
		BaseEntity: BaseEntity{
			Name:        "note",
			Description: "A handwritten note",
		},
		Location: "safe",
		Detail:   "It says 'The code is 1234'",
		Portable: &Portable{}, // Make it portable so it can be taken out
	}

	// Create a locked safe with keypad containing the note
	safe := &Item{
		BaseEntity: BaseEntity{
			Name:        "safe",
			Description: "A metal safe with a digital keypad",
		},
		Location: "test_room",
		Detail:   "It has a keypad with numbers 0-9",
		Container: &Container{
			Contains: note,
			Searched: false,
			Locked: &Lock{
				Locked:  true,
				KeyName: "",     // no key required
				Code:    "1234", // keypad code
			},
		},
	}

	if err := safe.ValidateInitialState(); err != nil {
		t.Errorf("Failed to validate safe: %v", err)
	}

	// Test that the safe is a container
	if !safe.IsContainer() {
		t.Error("Expected safe to be a container")
	}

	// Test that the safe has a lock
	if !safe.Container.HasLock() {
		t.Error("Expected safe to have a lock")
	}

	// Test that the safe is locked
	if !safe.Container.IsLocked() {
		t.Error("Expected safe to be locked")
	}

	// Test trying to search the safe while locked (should fail)
	_, err := safe.Container.Search()
	if err == nil {
		t.Error("Expected search to fail when safe is locked")
	}

	// Test unlocking with wrong code (should fail)
	err = safe.Container.UnlockWithCode("0000")
	if err == nil {
		t.Error("Expected unlock to fail with wrong code")
	}

	// Test that the safe is still locked after wrong code
	if !safe.Container.IsLocked() {
		t.Error("Expected safe to still be locked after wrong code")
	}

	// Test unlocking with correct code
	err = safe.Container.UnlockWithCode("1234")
	if err != nil {
		t.Errorf("Failed to unlock safe with correct code: %v", err)
	}

	// Test that the safe is now unlocked
	if safe.Container.IsLocked() {
		t.Error("Expected safe to be unlocked after correct code")
	}

	// Test searching the safe now that it's unlocked
	foundItem, err := safe.Container.Search()
	if err != nil {
		t.Errorf("Failed to search unlocked safe: %v", err)
	}

	// Should find the note
	if foundItem == nil {
		t.Error("Expected to find note in safe")
	}

	// Test that the safe is now marked as searched
	if !safe.Container.Searched {
		t.Error("Expected safe to be marked as searched")
	}

	// Test that the safe still contains the note
	if safe.Container.Contains == nil {
		t.Error("Expected safe to still contain the note after search")
	}
}

func TestConcealedPortableItem(t *testing.T) {
	// Create a portable item to be concealed
	coin := &Item{
		BaseEntity: BaseEntity{
			Name:        "coin",
			Description: "A shiny gold coin",
		},
		Location: "table",
		Detail:   "It looks valuable",
		Portable: &Portable{},
	}

	// Create a sheet that conceals the coin
	sheet := &Item{
		BaseEntity: BaseEntity{
			Name:        "sheet",
			Description: "A white bedsheet",
		},
		Location: "table",
		Detail:   "It's covering something",
		Concealer: &Concealer{
			Hidden:    coin,
			Uncovered: false,
		},
	}

	// Create a room
	room := &Room{
		BaseEntity: BaseEntity{
			Name:        "test_room",
			Description: "A simple test room",
		},
		Connections: []*Connection{},
		Items:       []*Item{sheet}, // Start with just the sheet
	}

	// Validate both items
	if err := coin.ValidateInitialState(); err != nil {
		t.Errorf("Failed to validate coin: %v", err)
	}

	if err := sheet.ValidateInitialState(); err != nil {
		t.Errorf("Failed to validate sheet: %v", err)
	}

	// Test that the coin is portable
	if !coin.IsPortable() {
		t.Error("Expected coin to be portable")
	}

	// Test that the sheet is a concealer
	if !sheet.IsConcealer() {
		t.Error("Expected sheet to be a concealer")
	}

	// Test revealing the concealed item
	revealedItem, err := sheet.Concealer.Reveal()
	if err != nil {
		t.Errorf("Failed to reveal concealed item: %v", err)
	}

	// Should reveal the coin
	if revealedItem.Name != "coin" {
		t.Errorf("Expected to reveal 'coin', got '%s'", revealedItem.Name)
	}

	// Test that the sheet is now marked as uncovered
	if !sheet.Concealer.Uncovered {
		t.Error("Expected sheet to be marked as uncovered")
	}

	// Test that the sheet no longer conceals anything
	if sheet.Concealer.Hidden != nil {
		t.Error("Expected sheet to no longer conceal anything after reveal")
	}

	// Simulate engine orchestration: add the revealed item to the room
	room.Items = append(room.Items, revealedItem)

	// Test that the room now contains both the sheet and the revealed coin
	if len(room.Items) != 2 {
		t.Errorf("Expected 2 items in room after reveal, got %d", len(room.Items))
	}

	// Verify the coin is now in the room
	coinFound := false
	for _, item := range room.Items {
		if item.Name == "coin" {
			coinFound = true
			break
		}
	}
	if !coinFound {
		t.Error("Expected to find coin in room after reveal")
	}
}

func TestChestWithKeyLock(t *testing.T) {
	// Create two different keys
	correctKey := &Item{
		BaseEntity: BaseEntity{
			Name:        "silver_key",
			Description: "A silver key with intricate patterns",
		},
		Location: "test_room",
		Detail:   "It looks like it might fit a specific lock",
		Portable: &Portable{},
		Key:      &Key{},
	}

	wrongKey := &Item{
		BaseEntity: BaseEntity{
			Name:        "brass_key",
			Description: "A brass key with simple design",
		},
		Location: "test_room",
		Detail:   "It looks like it might fit a different lock",
		Portable: &Portable{},
		Key:      &Key{},
	}

	// Create a locked chest that requires the silver key
	chest := &Item{
		BaseEntity: BaseEntity{
			Name:        "treasure_chest",
			Description: "A wooden chest with a silver lock",
		},
		Location: "test_room",
		Detail:   "It's locked with a silver lock",
		Container: &Container{
			Contains: nil, // empty for this test
			Searched: false,
			Locked: &Lock{
				Locked:  true,
				KeyName: "silver_key", // requires the silver key
				Code:    "",           // no code required
			},
		},
	}

	// Validate all items
	if err := correctKey.ValidateInitialState(); err != nil {
		t.Errorf("Failed to validate correct key: %v", err)
	}

	if err := wrongKey.ValidateInitialState(); err != nil {
		t.Errorf("Failed to validate wrong key: %v", err)
	}

	if err := chest.ValidateInitialState(); err != nil {
		t.Errorf("Failed to validate chest: %v", err)
	}

	// Test that both keys are portable and are keys
	if !correctKey.IsPortable() || !correctKey.IsKey() {
		t.Error("Expected correct key to be portable and a key")
	}

	if !wrongKey.IsPortable() || !wrongKey.IsKey() {
		t.Error("Expected wrong key to be portable and a key")
	}

	// Test that the chest is a container with a lock
	if !chest.IsContainer() {
		t.Error("Expected chest to be a container")
	}

	if !chest.Container.HasLock() {
		t.Error("Expected chest to have a lock")
	}

	if !chest.Container.IsLocked() {
		t.Error("Expected chest to be locked")
	}

	// Test trying to unlock with wrong key (should fail)
	err := chest.Container.UnlockWithKey("brass_key")
	if err == nil {
		t.Error("Expected unlock to fail with wrong key")
	}

	// Test that the chest is still locked after wrong key
	if !chest.Container.IsLocked() {
		t.Error("Expected chest to still be locked after wrong key")
	}

	// Test unlocking with correct key
	err = chest.Container.UnlockWithKey("silver_key")
	if err != nil {
		t.Errorf("Failed to unlock chest with correct key: %v", err)
	}

	// Test that the chest is now unlocked
	if chest.Container.IsLocked() {
		t.Error("Expected chest to be unlocked after correct key")
	}

	// Test that we can now search the chest
	foundItem, err := chest.Container.Search()
	if err != nil {
		t.Errorf("Failed to search unlocked chest: %v", err)
	}

	// Should return nil since chest is empty
	if foundItem != nil {
		t.Errorf("Expected no item found in empty chest, got %v", foundItem)
	}

	// Test that the chest is now marked as searched
	if !chest.Container.Searched {
		t.Error("Expected chest to be marked as searched")
	}
}

func TestConcealedCabinet(t *testing.T) {
	// Create a cabinet (container) to be concealed
	cabinet := &Item{
		BaseEntity: BaseEntity{
			Name:        "cabinet",
			Description: "A wooden cabinet with drawers",
		},
		Location: "corner",
		Detail:   "It looks like it might contain something",
		Container: &Container{
			Contains: nil, // empty for this test
			Searched: false,
			Locked:   nil, // no lock
		},
	}

	// Create a sheet that conceals the cabinet
	sheet := &Item{
		BaseEntity: BaseEntity{
			Name:        "dusty_sheet",
			Description: "A dusty white sheet",
		},
		Location: "corner",
		Detail:   "It's covering something large",
		Concealer: &Concealer{
			Hidden:    cabinet,
			Uncovered: false,
		},
	}

	// Create a room
	room := &Room{
		BaseEntity: BaseEntity{
			Name:        "test_room",
			Description: "A simple test room",
		},
		Connections: []*Connection{},
		Items:       []*Item{sheet}, // Start with just the sheet
	}

	// Validate both items
	if err := cabinet.ValidateInitialState(); err != nil {
		t.Errorf("Failed to validate cabinet: %v", err)
	}

	if err := sheet.ValidateInitialState(); err != nil {
		t.Errorf("Failed to validate sheet: %v", err)
	}

	// Test that the cabinet is a container
	if !cabinet.IsContainer() {
		t.Error("Expected cabinet to be a container")
	}

	// Test that the sheet is a concealer
	if !sheet.IsConcealer() {
		t.Error("Expected sheet to be a concealer")
	}

	// Test revealing the concealed cabinet
	revealedItem, err := sheet.Concealer.Reveal()
	if err != nil {
		t.Errorf("Failed to reveal concealed cabinet: %v", err)
	}

	// Should reveal the cabinet
	if revealedItem.Name != "cabinet" {
		t.Errorf("Expected to reveal 'cabinet', got '%s'", revealedItem.Name)
	}

	// Test that the sheet is now marked as uncovered
	if !sheet.Concealer.Uncovered {
		t.Error("Expected sheet to be marked as uncovered")
	}

	// Test that the sheet no longer conceals anything
	if sheet.Concealer.Hidden != nil {
		t.Error("Expected sheet to no longer conceal anything after reveal")
	}

	// Simulate engine orchestration: add the revealed cabinet to the room
	room.Items = append(room.Items, revealedItem)

	// Test that the room now contains both the sheet and the revealed cabinet
	if len(room.Items) != 2 {
		t.Errorf("Expected 2 items in room after reveal, got %d", len(room.Items))
	}

	// Verify the cabinet is now in the room
	cabinetFound := false
	for _, item := range room.Items {
		if item.Name == "cabinet" {
			cabinetFound = true
			break
		}
	}
	if !cabinetFound {
		t.Error("Expected to find cabinet in room after reveal")
	}

	// Test that we can search the revealed cabinet
	foundItem, err := revealedItem.Container.Search()
	if err != nil {
		t.Errorf("Failed to search revealed cabinet: %v", err)
	}

	// Should return nil since cabinet is empty
	if foundItem != nil {
		t.Errorf("Expected no item found in empty cabinet, got %v", foundItem)
	}

	// Test that the cabinet is now marked as searched
	if !revealedItem.Container.Searched {
		t.Error("Expected cabinet to be marked as searched")
	}
}

func TestLockedDoor(t *testing.T) {
	// Create a locked door
	door := &Door{
		BaseEntity: BaseEntity{
			Name:        "iron_door",
			Description: "A heavy iron door",
		},
		RoomA: "room_a",
		RoomB: "room_b",
		Lock: &Lock{
			Locked:  true,
			KeyName: "iron_key", // requires iron key
			Code:    "",         // no code required
		},
	}

	// Test that the door has a lock
	if !door.HasLock() {
		t.Error("Expected door to have a lock")
	}

	// Test that the door is locked
	if !door.IsLocked() {
		t.Error("Expected door to be locked")
	}

	// Test unlocking the door with the correct key
	err := door.UnlockWithKey("iron_key")
	if err != nil {
		t.Errorf("Failed to unlock door with correct key: %v", err)
	}

	// Test that the door is now unlocked
	if door.IsLocked() {
		t.Error("Expected door to be unlocked after correct key")
	}

	// Test that the lock is unlocked
	if door.Lock.Locked {
		t.Error("Expected lock to be unlocked")
	}
}
