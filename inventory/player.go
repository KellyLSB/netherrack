package inventory

import (
	"github.com/NetherrackDev/soulsand"
)

var _ soulsand.PlayerInventory = &PlayerInventory{}

type PlayerInventory struct {
	Type
}

func CreatePlayerInventory() *PlayerInventory {
	return &PlayerInventory{
		Type: Type{
			items:    make([]soulsand.ItemStack, 45),
			Id:       -1,
			watchers: make(map[string]soulsand.Player),
		},
	}
}

func (pi *PlayerInventory) SetSlot(slot int, item soulsand.ItemStack) {
	pi.lock.Lock()
	defer pi.lock.Unlock()
	pi.items[slot] = item
	pi.watcherLock.RLock()
	defer pi.watcherLock.RUnlock()
	for _, p := range pi.watchers {
		p.RunSync(func(se soulsand.SyncEntity) {
			sp := se.(soulsand.SyncPlayer)
			sp.Connection().WriteSetSlot(0, int16(slot), item)
		})
	}
}

func (pi *PlayerInventory) AddWatcher(p soulsand.Player) {
	pi.watcherLock.Lock()
	defer pi.watcherLock.Unlock()
	pi.watchers[p.Name()] = p
	pi.lock.RLock()
	defer pi.lock.RUnlock()
	p.RunSync(func(se soulsand.SyncEntity) {
		sp := se.(soulsand.SyncPlayer)
		sp.Connection().WriteSetWindowItems(0, pi.items)
	})
}

func (pi *PlayerInventory) GetCraftingOutput() soulsand.ItemStack {
	return pi.GetSlot(0)
}

func (pi *PlayerInventory) SetCraftingOutput(item soulsand.ItemStack) {
	pi.SetSlot(0, item)
}

func (pi *PlayerInventory) GetCraftingInput(x, y int) soulsand.ItemStack {
	return pi.GetSlot(1 + x + y*2)
}

func (pi *PlayerInventory) SetCraftingInput(x, y int, item soulsand.ItemStack) {
	pi.SetSlot(1+x+y*2, item)
}

func (pi *PlayerInventory) GetArmourHead() soulsand.ItemStack {
	return pi.GetSlot(5)
}

func (pi *PlayerInventory) SetArmourHead(item soulsand.ItemStack) {
	pi.SetSlot(5, item)
}

func (pi *PlayerInventory) GetArmourChest() soulsand.ItemStack {
	return pi.GetSlot(6)
}

func (pi *PlayerInventory) SetArmourChest(item soulsand.ItemStack) {
	pi.SetSlot(6, item)
}

func (pi *PlayerInventory) GetArmourLegs() soulsand.ItemStack {
	return pi.GetSlot(7)
}

func (pi *PlayerInventory) SetArmourLegs(item soulsand.ItemStack) {
	pi.SetSlot(7, item)
}

func (pi *PlayerInventory) GetArmourFeet() soulsand.ItemStack {
	return pi.GetSlot(8)
}

func (pi *PlayerInventory) SetArmourFeet(item soulsand.ItemStack) {
	pi.SetSlot(8, item)
}

func (pi *PlayerInventory) GetPlayerInventorySlot(slot int) soulsand.ItemStack {
	return pi.GetSlot(9 + slot)
}

func (pi *PlayerInventory) SetPlayerInventorySlot(slot int, item soulsand.ItemStack) {
	pi.SetSlot(9+slot, item)
}

func (pi *PlayerInventory) GetPlayerInventorySize() int {
	return 27
}

func (pi *PlayerInventory) GetHotbarSlot(slot int) soulsand.ItemStack {
	return pi.GetSlot(36 + slot)
}

func (pi *PlayerInventory) SetHotbarSlot(slot int, item soulsand.ItemStack) {
	pi.SetSlot(36+slot, item)
}

func (pi *PlayerInventory) GetHotbarSize() int {
	return 9
}
