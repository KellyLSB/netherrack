package items

import (
	"github.com/thinkofdeath/netherrack/nbt"
	"github.com/thinkofdeath/soulsand"
	"sync"
)

//Compile time checks
var _ soulsand.ItemStack = &ItemStack{}

func CreateItemStack(id, data int16, count byte) *ItemStack {
	return &ItemStack{
		ID:       id,
		Damage:   data,
		Count:    count,
		metadata: make(map[string]interface{}),
	}
}

type ItemStack struct {
	Lock     sync.RWMutex
	ID       int16
	Count    byte
	Damage   int16
	Tag      nbt.Type
	metaLock sync.RWMutex
	metadata map[string]interface{}
}

func (i *ItemStack) GetID() int16 {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	return i.ID
}

func (i *ItemStack) SetID(id int16) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	i.ID = id
}

func (i *ItemStack) GetData() int16 {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	return i.Damage
}

func (i *ItemStack) SetData(data int16) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	i.Damage = data
}

func (i *ItemStack) GetCount() byte {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	return i.Count
}

func (i *ItemStack) SetCount(count byte) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	i.Count = count
}

func (i *ItemStack) SetDisplayName(name string) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	if i.Tag == nil {
		i.Tag = nbt.NewNBT()
	}

	display, _ := i.Tag.GetCompound("display", true)
	display.Set("Name", name)
}

func (i *ItemStack) ClearLore() {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	if i.Tag == nil {
		return
	}
	if display, ok := i.Tag.GetCompound("display", false); ok {
		display.Remove("Lore")
	}
}

func (i *ItemStack) AddLore(line string) {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	if i.Tag == nil {
		i.Tag = nbt.NewNBT()
	}
	display, _ := i.Tag.GetCompound("display", true)
	lore, _ := display.GetList("Lore", true)
	lore = append(lore, line)
	display.Set("Lore", lore)
}

func (i *ItemStack) SetMetadata(key string, value interface{}) {
	i.metaLock.Lock()
	defer i.metaLock.Unlock()
	i.metadata[key] = value
}

func (i *ItemStack) GetMetadata(key string) interface{} {
	i.metaLock.RLock()
	defer i.metaLock.RUnlock()
	return i.metadata[key]
}

func (i *ItemStack) RemoveMetadata(key string) {
	i.metaLock.Lock()
	defer i.metaLock.Unlock()
	delete(i.metadata, key)
}
