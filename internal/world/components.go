package world

import (
	"errors"
	"fmt"
)

// --- base entity ---

// BaseEntity is anything that has a name and description.
type BaseEntity struct {
	Name        string
	Description string
}

// --- components ---

// Portable marks an item that can be taken into inventory.
type Portable struct{}

// Key marks an item that can unlock a Lock.
type Key struct{}

// Ammo is ammunition for a weapon.
type Ammo struct {
	Quantity int
}

// Weapon gives an item the ability to enhance win probability during combat.
// It may or may not use ammo. Weapons that use ammo may come with zero or more rounds.
type Weapon struct {
	Damage float64 // 0.0 to 1.0
	Ammo   *Ammo
}

// Box of ammunition for a weapon.
type AmmoBox struct {
	WeaponName string
	Ammo       *Ammo
}

// Container can hold exactly one item and remembers whether it’s been searched.
type Container struct {
	Contains *Item
	Searched bool
	Locked   *Lock
}

// Conceal hides exactly one item until it is uncovered.
type Concealer struct {
	Hidden    *Item
	Uncovered bool
}

// HealthEffect is the strength of a health item.
type HealthEffect string

const (
	HealthBoostWeak   HealthEffect = "weak"   // increases health to next state
	HealthBoostStrong HealthEffect = "strong" // sets health to maximum
)

// HealthItem restores health.
type HealthItem struct {
	HealthEffect
}

// Fixture is a type of (usually non-portable) item that other items can be "used" on.
// A fixture requires one or more items before it produces a new item.
// Examples include altars, vending machines, a bathtub drain with a key stuck in it, etc.
//
// Note on Fixtures vs Containers:
//
// A fixture is distinct from a Container in that it cannot be searched, does not have a
// standard lock state, and the produced item is automatically added to the player's inventory.
// Additionally, fixtures will also support custom messages upon player interaction.
// Finally, unlike Containers which may be empty or contain non-essential items, completion of
// fixtures is always required to progress through the level.
//
// In the future we will add event handling for fixtures, but for now they only support
// producing items.
type Fixture struct {
	RequiredItems       map[string]bool
	Produces            *Item
	CompletionNarrative string
}

type FixtureUseResult struct {
	Item *Item
	// TODO: custom message when item is accepted
}

// --- fixture component methods ---

// IsComplete checks if all required items have been applied to the fixture.
func (f *Fixture) IsComplete() bool {
	for _, requiredItem := range f.RequiredItems {
		if !requiredItem {
			return false
		}
	}
	return true
}

// UseItem uses an item on a fixture.
func (f *Fixture) UseItem(itemName string) (*FixtureUseResult, error) {
	// Assumes no duplicate items in the level, and that the engine
	// destroys items after successful use on a fixture.
	if _, ok := f.RequiredItems[itemName]; !ok {
		return nil, fmt.Errorf("you can't use a %s on this", itemName)
	}
	f.RequiredItems[itemName] = true
	if f.IsComplete() {
		return &FixtureUseResult{
			Item: f.Produces,
		}, nil
	}
	return &FixtureUseResult{
		Item: nil,
	}, nil
}

// Lock may secure a Portal *or* a Container.
// If KeyName is set, it’s a key lock; if Code is set, it’s a keypad.
type Lock struct {
	Locked  bool
	KeyName string
	Code    string
}

// --- lock component methods ---

func (l *Lock) IsUnlocked() bool {
	return l == nil || !l.Locked
}

// UnlockWithKey unlocks a lock with a key.
func (l *Lock) UnlockWithKey(keyName string) error {
	if l.KeyName == "" {
		return errors.New("lock doesnt not take a key")
	}
	if !l.Locked {
		return errors.New("already unlocked")
	}
	if keyName != l.KeyName {
		return errors.New("wrong key")
	}
	l.Locked = false
	return nil
}

// UnlockWithCode unlocks a lock with a code.
func (l *Lock) UnlockWithCode(code string) error {
	if l.Code == "" {
		return errors.New("lock doesnt not take a code")
	}
	if !l.Locked {
		return errors.New("already unlocked")
	}
	if code != l.Code {
		return errors.New("wrong code")
	}
	l.Locked = false
	return nil
}

// --- container component methods ---

func (c *Container) HasKeyLock() bool  { return c.Locked != nil && c.Locked.KeyName != "" }
func (c *Container) HasCodeLock() bool { return c.Locked != nil && c.Locked.Code != "" }
func (c *Container) HasLock() bool     { return c.HasKeyLock() || c.HasCodeLock() }
func (c *Container) IsLocked() bool    { return c.HasLock() && c.Locked.Locked }
func (c *Container) IsEmpty() bool     { return c.Contains == nil }

// RemoveItem removes the contained item from the container.
func (c *Container) RemoveItem() (*Item, error) {
	if c.IsEmpty() {
		return nil, errors.New("container is empty")
	}
	item := c.Contains
	c.Contains = nil
	return item, nil
}

// Search searches a container.
func (c *Container) Search() (*Item, error) {
	if c.Locked != nil && c.Locked.Locked {
		return nil, errors.New("container is locked")
	}

	c.Searched = true
	return c.Contains, nil
}

// UnlockWithKey unlocks a container with a key.
func (c *Container) UnlockWithKey(keyName string) error {
	if c.Locked == nil {
		return errors.New("container has no lock")
	}
	return c.Locked.UnlockWithKey(keyName)
}

// UnlockWithCode unlocks a container with a code.
func (c *Container) UnlockWithCode(code string) error {
	if c.Locked == nil {
		return errors.New("container has no lock")
	}
	return c.Locked.UnlockWithCode(code)
}

// --- concealer component methods ---

func (c *Concealer) Reveal() (*Item, error) {
	revealed := c.Hidden
	c.Hidden = nil
	c.Uncovered = true
	return revealed, nil
}

// --- weapon component methods ---

func (w *Weapon) UsesAmmo() bool { return w.Ammo != nil }
