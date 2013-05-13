package player

import (
	"github.com/NetherrackDev/netherrack/event"
	"github.com/NetherrackDev/netherrack/protocol"
	"github.com/NetherrackDev/netherrack/system"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/blocks"
	"github.com/NetherrackDev/soulsand/command"
	"github.com/NetherrackDev/soulsand/effect"
	"github.com/NetherrackDev/soulsand/gamemode"
	"log"
	"math"
	"runtime"
)

var packets map[byte]func(c *protocol.Conn, player *Player) = map[byte]func(c *protocol.Conn, player *Player){
	0x00: func(c *protocol.Conn, player *Player) { //Keep Alive
		id := c.ReadKeepAlive()
		if id != player.currentTickID {
			runtime.Goexit()
		}
	},
	0x03: func(c *protocol.Conn, player *Player) { //Chat Message
		msg := c.ReadChatMessage()
		if len(msg) <= 0 {
			return
		}
		if msg[0] == '/' {
			command.Exec(msg[1:], player)
		} else {
			eventType, ev := event.NewMessage(player, msg)
			if !player.Fire(eventType, ev) {
				system.Broadcast(ev.GetMessage())
			}
		}
	},
	0x07: func(c *protocol.Conn, player *Player) { //Use Entity
		c.ReadUseEntity()
	},
	0x0A: func(c *protocol.Conn, player *Player) { //Player
		c.ReadPlayer()
	},
	0x0B: func(c *protocol.Conn, player *Player) { //Player Position
		x, y, _, z, _ := c.ReadPlayerPosition()
		if !player.IgnoreMoveUpdates {
			player.Position.X = x
			player.Position.Y = y
			player.Position.Z = z
		}
	},
	0x0C: func(c *protocol.Conn, player *Player) { //Player Look
		yaw, pitch, _ := c.ReadPlayerLook()
		player.Position.Yaw = yaw
		player.Position.Pitch = pitch
	},
	0x0D: func(c *protocol.Conn, player *Player) { //Player Position and Look
		x, y, _, z, yaw, pitch, _ := c.ReadPlayerPositionLook()
		if !player.IgnoreMoveUpdates {
			player.Position.X = x
			player.Position.Y = y
			player.Position.Z = z
		}
		player.Position.Yaw = yaw
		player.Position.Pitch = pitch
	},
	0x0E: func(c *protocol.Conn, player *Player) { //Player Digging
		status, bx, by, bz, _ := c.ReadPlayerDigging()
		if status != 2 && !(status == 0 && player.gamemode == gamemode.Creative) {
			return
		}
		x := int(bx)
		y := int(by)
		z := int(bz)

		player.World.RunSync(x>>4, z>>4, func(ch soulsand.SyncChunk) {
			chunk := ch.(interface {
				GetPlayerMap() map[string]soulsand.Player
			})
			rx := x - ((x >> 4) << 4)
			rz := z - ((z >> 4) << 4)
			block := ch.GetBlock(rx, y, rz)
			meta := ch.GetMeta(rx, y, rz)
			m := chunk.GetPlayerMap()
			for _, p := range m {
				if p.GetName() != player.GetName() {
					p.PlayEffect(x, y, z, effect.BlockBreak, int(block)|(int(meta)<<12), true)
				}
			}
		})
		player.World.SetBlock(x, y, z, 0, 0)
	},
	0x0F: func(c *protocol.Conn, player *Player) { //Player Block Placement
		bx, by, bz, direction, _, _, _ := c.ReadPlayerBlockPlacement()
		x := int(bx)
		y := int(by)
		z := int(bz)
		switch direction {
		case 0:
			y--
		case 1:
			y++
		case 2:
			z--
		case 3:
			z++
		case 4:
			x--
		case 5:
			x++
		}
		player.World.SetBlock(x, y, z, blocks.Stone.Id(), 0)
	},
	0x10: func(c *protocol.Conn, player *Player) { //Held Item Change
		slotID := c.ReadHeldItemChange()
		player.CurrentSlot = int(slotID)
	},
	0x12: func(c *protocol.Conn, player *Player) { //Animation
		c.ReadAnimation()
	},
	0x13: func(c *protocol.Conn, player *Player) { //Entity Action
		c.ReadEntityAction()
	},
	0x65: func(c *protocol.Conn, player *Player) { //Close Window
		id := c.ReadCloseWindow()
		if id == 5 && player.openInventory != nil {
			player.openInventory.RemoveWatcher(player)
		}
	},
	0x66: func(c *protocol.Conn, player *Player) { //Click Window
		c.ReadClickWindow()
	},
	0x6A: func(c *protocol.Conn, player *Player) { //Confirm Transaction
		c.ReadConfirmTransaction()
	},
	0x6B: func(c *protocol.Conn, player *Player) { //Creative Inventory Action
		c.ReadCreativeInventoryAction()
	},
	0x6C: func(c *protocol.Conn, player *Player) { //Enchant Item
		c.ReadEnchantItem()
	},
	0x82: func(c *protocol.Conn, player *Player) { //Update Sign
		c.ReadUpdateSign()
	},
	0xCA: func(c *protocol.Conn, player *Player) { //Player Abilities
		c.ReadPlayerAbilities()
	},
	0xCB: func(c *protocol.Conn, player *Player) { //Tab-complete
		text := c.ReadTabComplete()
		c.WriteTabComplete(command.Complete(text[1:]))
	},
	0xCC: func(c *protocol.Conn, player *Player) { //Client Settings
		locale, viewDistance, chatFlags, difficulty, showCape := c.ReadClientSettings()
		player.settings.locale = locale
		old := player.settings.viewDistance
		player.settings.viewDistance = int(math.Pow(2, 4-float64(viewDistance)))
		if player.settings.viewDistance > 10 {
			player.settings.viewDistance = 10
		}
		if old != player.settings.viewDistance {
			player.chunkReload(old)
		}
		player.settings.chatFlags = byte(chatFlags)
		player.settings.difficulty = byte(difficulty)
		player.settings.showCape = showCape
	},
	0xCD: func(c *protocol.Conn, player *Player) { //Client Statuses
		c.ReadClientStatuses()
	},
	0xFA: func(c *protocol.Conn, player *Player) { //Plugin Message
		c.ReadPluginMessage()
	},
	0xFF: func(c *protocol.Conn, player *Player) { //Disconnect
		log.Printf("Player %s disconnect %s\n", player.GetName(), c.ReadDisconnect())
		runtime.Goexit()
	},
}
