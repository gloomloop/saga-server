package loader

import (
	"testing"

	world "adventure-engine/world"
)

func TestLoadGame_Demo(t *testing.T) {
	// Load the demo puzzle game
	level, err := LoadGame("../tests/demo.json")
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Test basic game properties
	if level.Name != "demo key puzzle" {
		t.Errorf("Expected game name 'demo key puzzle', got '%s'", level.Name)
	}

	// Test rooms
	if len(level.Rooms) != 4 {
		t.Errorf("Expected 4 rooms, got %d", len(level.Rooms))
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
	waitingRoom := findRoomByName(level.Rooms, "waiting room")
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
	storageRoom := findRoomByName(level.Rooms, "storage room")
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
	office := findRoomByName(level.Rooms, "office")
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
	stairwell := findRoomByName(level.Rooms, "stairwell to roof")
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

	if zombie.Room != "storage room" {
		t.Errorf("Expected zombie to be in storage room, got '%s'", zombie.Room)
	}

	if zombie.TriggerEvent != world.TriggerEventTakeItem {
		t.Errorf("Expected zombie trigger event to be TriggerEventTakeItem, got %s", zombie.TriggerEvent)
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

// Helper functions
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
