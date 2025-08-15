package world

import (
	"errors"
	"fmt"
)

// --- entities ---

// Item is anything that can exist in a room.
type Item struct {
	BaseEntity
	Location string
	Detail   string

	// Optional capabilities (nil if absent)
	Portable   *Portable
	Key        *Key
	Weapon     *Weapon
	Container  *Container
	Concealer  *Concealer
	AmmoBox    *AmmoBox
	HealthItem *HealthItem
	Fixture    *Fixture
}

// Latch locks a door from one side only.
type Latch struct {
	Locked     bool
	LockedFrom string // "room_a" or "room_b"
}

// Door connects two rooms; it may be locked.
type Door struct {
	Name      string
	RoomA     string
	RoomB     string
	Lock      *Lock
	Stairwell bool // true if the door is a stairwell (connects floors)
	Latch     *Latch
	Traversed bool
	Tried     bool
}

// Connection represents a door as seen from a specific room.
// The location is relative to the room from which it is observed.
type Connection struct {
	DoorName    string
	Location    string
	Description string
}

// Room is a location the player can occupy.
type Room struct {
	BaseEntity
	InitialDescription string
	Connections        []*Connection
	Items              []*Item
	Visited            bool // true if the player has entered this room
}

// ComboItem contains a combination item and the names of the required input items.
type ComboItem struct {
	InputItemAName string
	InputItemBName string
	OutputItem     *Item
}

// Enemy is an NPC that must be defeated to return to investigation mode.
type Enemy struct {
	BaseEntity
	HP int
}

// --- enemy methods ---

// InflictDamage decrements the enemy's HP.
func (e *Enemy) InflictDamage() {
	e.HP--
}

func (e *Enemy) IsAlive() bool {
	return e.HP > 0
}

// --- room methods ---

// GetConnection returns a connection from the room by door name.
func (r *Room) GetConnection(doorName string) (*Connection, error) {
	for _, conn := range r.Connections {
		if conn.DoorName == doorName {
			return conn, nil
		}
	}
	return nil, fmt.Errorf("no door named %s in this room", doorName)
}

// GetItem returns an item from the room.
func (r *Room) GetItem(name string) (*Item, error) {
	for _, item := range r.Items {
		if item.Name == name {
			return item, nil
		}
	}
	return nil, fmt.Errorf("you don't see a %s here", name)
}

// RemoveItem removes an item from the room.
// Used when the player picks up an item.
func (r *Room) RemoveItem(name string) (*Item, error) {
	items := r.Items
	for i, it := range items {
		if it.Name == name {
			copy(items[i:], items[i+1:])
			r.Items = items[:len(items)-1]
			return it, nil
		}
	}
	return nil, fmt.Errorf("you don't see a %s here", name)
}

// --- item methods ---

func (it *Item) IsPortable() bool   { return it.Portable != nil }
func (it *Item) IsKey() bool        { return it.Key != nil }
func (it *Item) IsWeapon() bool     { return it.Weapon != nil }
func (it *Item) IsContainer() bool  { return it.Container != nil }
func (it *Item) IsConcealer() bool  { return it.Concealer != nil }
func (it *Item) IsAmmoBox() bool    { return it.AmmoBox != nil }
func (it *Item) IsHealthItem() bool { return it.HealthItem != nil }
func (it *Item) IsFixture() bool    { return it.Fixture != nil }

// Validate a newly created item.
func (it *Item) ValidateInitialState() error {
	if it.IsKey() {
		if !it.IsPortable() || it.IsContainer() || it.IsConcealer() || it.IsWeapon() {
			return errors.New("invalid key")
		}
	}
	if it.IsWeapon() {
		if !it.IsPortable() || it.IsContainer() || it.IsConcealer() || it.IsKey() {
			return errors.New("invalid weapon")
		}
	}
	if it.IsContainer() {
		if it.IsPortable() || it.IsConcealer() || it.IsKey() || it.IsWeapon() {
			return errors.New("invalid container")
		}
		if it.Container.Contains != nil && it.Container.Contains.IsContainer() {
			return errors.New("container cannot be nested")
		}
		if it.Container.HasLock() && !it.Container.IsLocked() {
			return errors.New("container with lock must start in a locked state")
		}
	}
	if it.IsConcealer() {
		if it.IsPortable() || it.IsContainer() || it.IsKey() || it.IsWeapon() {
			return errors.New("invalid concealer")
		}
		if it.Concealer.Hidden != nil && it.Concealer.Hidden.IsConcealer() {
			return errors.New("concealers cannot be nested")
		}
	}
	if it.IsFixture() {
		if it.IsPortable() || it.IsContainer() || it.IsConcealer() || it.IsKey() || it.IsWeapon() {
			return errors.New("invalid fixture")
		}
	}
	return nil
}

// --- door methods ---

func (d *Door) HasKeyLock() bool                { return d.Lock != nil && d.Lock.KeyName != "" }
func (d *Door) HasCodeLock() bool               { return d.Lock != nil && d.Lock.Code != "" }
func (d *Door) HasLock() bool                   { return d.HasKeyLock() || d.HasCodeLock() }
func (d *Door) IsLocked() bool                  { return d.HasLock() && d.Lock.Locked }
func (d *Door) IsLatched() bool                 { return d.Latch != nil && d.Latch.Locked }
func (d *Door) CanUnlatch(roomName string) bool { return d.Latch.LockedFrom == roomName }
func (d *Door) Unlatch()                        { d.Latch.Locked = false }

// UnlockWithKey unlocks a door with a key.
func (d *Door) UnlockWithKey(keyName string) error {
	if d.Lock == nil {
		return fmt.Errorf("the %s has no lock", d.Name)
	}
	return d.Lock.UnlockWithKey(keyName)
}

// UnlockWithCode unlocks a door with a code.
func (d *Door) UnlockWithCode(code string) error {
	if d.Lock == nil {
		return fmt.Errorf("the %s has no lock", d.Name)
	}
	return d.Lock.UnlockWithCode(code)
}

// --- player ---

type HealthState string

const (
	HealthFine HealthState = "fine"
	HealthHurt HealthState = "hurt"
	HealthCrit HealthState = "critical"
	HealthDead HealthState = "dead"
)

func (p *Player) IsAlive() bool {
	return p.Health != HealthDead
}

type Player struct {
	Inventory []*Item
	Health    HealthState
	Ammo      map[string]int // weapon name -> ammo quantity
}

// GetItem returns an item from the player's inventory.
func (p *Player) GetItem(name string) (*Item, error) {
	for _, item := range p.Inventory {
		if item.Name == name {
			return item, nil
		}
	}
	return nil, fmt.Errorf("you don't have a %s in your inventory", name)
}

// RemoveItem removes an item from the player's inventory.
func (p *Player) RemoveItem(name string) (*Item, error) {
	items := p.Inventory
	for i, it := range items {
		if it.Name == name {
			copy(items[i:], items[i+1:])
			p.Inventory = items[:len(items)-1]
			return it, nil
		}
	}
	return nil, fmt.Errorf("you don't have a %s in your inventory", name)
}

func (p *Player) IncreaseHealth() {
	switch p.Health {
	case HealthHurt:
		p.Health = HealthFine
	case HealthCrit:
		p.Health = HealthHurt
	default:
		panic("invalid health state")
	}
}

func (p *Player) InflictDamage() {
	switch p.Health {
	case HealthFine:
		p.Health = HealthHurt
	case HealthHurt:
		p.Health = HealthCrit
	case HealthCrit:
		p.Health = HealthDead
	default:
		panic("invalid health state")
	}
}

func (p *Player) FireWeapon(weaponName string) error {
	if p.Ammo[weaponName] == 0 {
		return fmt.Errorf("the %s is out of ammo", weaponName)
	}
	p.Ammo[weaponName]--
	return nil
}

// --- events and triggers ---

type EventType string

const (
	EventEnemyKilled  EventType = "enemy_killed"
	EventPlayerKilled EventType = "player_killed"
	EventItemTaken    EventType = "item_taken"
	EventRoomEntered  EventType = "room_entered"
)

type Event struct {
	Event     EventType
	EnemyName string
	RoomName  string
	ItemName  string
}

type EffectType string

const (
	EffectEnterCombat EffectType = "enter_combat"
)

type Effect struct {
	EffectType
	EnemyName string
}

type Trigger struct {
	Event
	Effect
}

// --- level ---

type Floor struct {
	Name        string
	Description string
	Rooms       []*Room
}

type Level struct {
	Name         string
	Floors       []*Floor
	Doors        []*Door
	Enemies      []*Enemy
	Triggers     []*Trigger
	WinCondition *Event
	ComboItems   []*ComboItem
}

// CombineItems crafts a new item by combining two input items.
// Returns the new item or an error if the combination is not possible.
// The engine is responsible for removing input items from the player's inventory.
func (l *Level) CombineItems(inputItemAName string, inputItemBName string) (*Item, error) {
	for _, comboItem := range l.ComboItems {
		if comboItem.InputItemAName == inputItemAName && comboItem.InputItemBName == inputItemBName ||
			comboItem.InputItemAName == inputItemBName && comboItem.InputItemBName == inputItemAName {
			return comboItem.OutputItem, nil
		}
	}
	return nil, fmt.Errorf("you can't combine the %s and %s", inputItemAName, inputItemBName)
}

// GetEnemy returns an enemy by name.
func (e *Level) GetEnemy(name string) *Enemy {
	for _, enemy := range e.Enemies {
		if enemy.Name == name {
			return enemy
		}
	}
	panic(fmt.Sprintf("no enemy named %s", name))
}

// GetFloor returns a floor by name.
func (e *Level) GetFloor(name string) *Floor {
	for _, floor := range e.Floors {
		if floor.Name == name {
			return floor
		}
	}
	panic(fmt.Sprintf("no floor named %s", name))
}

// GetRoom returns a room by name.
func (e *Level) GetRoom(floorName string, roomName string) *Room {
	floor := e.GetFloor(floorName)
	for _, room := range floor.Rooms {
		if room.Name == roomName {
			return room
		}
	}
	panic(fmt.Sprintf("no room named %s on floor %s", roomName, floorName))
}

// GetDoor returns a door by name.
func (e *Level) GetDoor(name string) *Door {
	for _, door := range e.Doors {
		if door.Name == name {
			return door
		}
	}
	panic(fmt.Sprintf("no door named %s", name))
}
