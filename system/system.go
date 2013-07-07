package system

import (
	"github.com/NetherrackDev/netherrack/event"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/chat"
)

func init() {
	go watcher()
}

var channel chan func() = make(chan func(), 1000)
var EventSource event.Source

func watcher() {
	for {
		f := <-channel
		f()
	}
}

var (
	entitysById         = make(map[int32]soulsand.Entity)
	playersByName       = make(map[string]soulsand.Player)
	currentEID    int32 = 0
)

func GetEntityCount() int {
	res := make(chan int, 1)
	channel <- func() {
		res <- len(entitysById)
	}
	return <-res
}

func Broadcast(message *chat.Message) {
	channel <- func() {
		for _, p := range playersByName {
			p.SendMessage(message)
		}
	}
}

func AddPlayer(p soulsand.Player) {
	channel <- func() {
		playersByName[p.Name()] = p
		displayName, err := p.DisplayName()
		if err != nil {
			return
		}
		for _, player := range playersByName {
			playerLocale, err := player.Locale()
			if err != nil {
				continue
			}
			player.SendMessage(chat.New().Colour(chat.Yellow).
				TranslateBegin(playerLocale, "message.player.connect").
				Colour(chat.Yellow).Text(displayName).
				TranslateEnd())
		}
	}
}

func RemovePlayer(p soulsand.Player) {
	displayName := (p.(soulsand.SyncPlayer)).DisplayNameSync()
	name := p.Name()
	channel <- func() {
		delete(playersByName, name)
		for _, player := range playersByName {
			playerLocale, err := player.Locale()
			if err != nil {
				continue
			}
			player.SendMessage(chat.New().Colour(chat.Yellow).
				TranslateBegin(playerLocale, "message.player.disconnect").
				Colour(chat.Yellow).Text(displayName).
				TranslateEnd())
		}
	}
}

func GetPlayer(name string) soulsand.Player {
	res := make(chan soulsand.Player)
	channel <- func() {
		res <- playersByName[name]
	}
	return <-res
}

func GetPlayers() []soulsand.Player {
	res := make(chan []soulsand.Player)
	channel <- func() {
		out := make([]soulsand.Player, len(playersByName))
		i := 0
		for _, player := range playersByName {
			out[i] = player
			i++
		}
		res <- out
	}
	return <-res
}

func GetFreeEntityID(e soulsand.Entity) int32 {
	res := make(chan int32, 1)
	channel <- func() {
		for {
			_, ok := entitysById[currentEID]
			if !ok {
				entitysById[currentEID] = e
				res <- currentEID
				currentEID++
				return
			}
			currentEID++
		}
	}
	return <-res
}

func FreeEntityID(e soulsand.Entity) {
	id := e.ID()
	channel <- func() {
		delete(entitysById, id)
	}
}
