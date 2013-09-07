package player

import (
	"github.com/NetherrackDev/netherrack/entity"
	"github.com/NetherrackDev/netherrack/log"
	"github.com/NetherrackDev/netherrack/protocol"
)

//A local player is a player connected directly to this server
type LocalPlayer struct {
	entity.LocalEntity

	conn     *protocol.Conn
	username string
	Server   Server

	packetQueue   chan protocol.Packet
	readPackets   chan protocol.Packet
	errorChannel  chan error
	closedChannel chan struct{}
}

func NewLocalPlayer(username string, conn *protocol.Conn, server Server) *LocalPlayer {
	lp := &LocalPlayer{
		username:      username,
		conn:          conn,
		packetQueue:   make(chan protocol.Packet, 200),
		readPackets:   make(chan protocol.Packet, 20),
		errorChannel:  make(chan error, 1),
		closedChannel: make(chan struct{}, 1),
		Server:        server,
	}
	lp.LocalEntity.Server = server
	go lp.packetReader()
	return lp
}

//Queues a packet to be sent to the player
func (lp *LocalPlayer) QueuePacket(packet protocol.Packet) {
	select {
	case lp.packetQueue <- packet:
	case <-lp.closedChannel:
		lp.closedChannel <- struct{}{}
	}
}

//Processes incomming and outgoing packets. Blocks until the player leaves
//or is kicked.
func (lp *LocalPlayer) Start() {
	defer lp.close()
	for {
		select {
		case err := <-lp.errorChannel:
			log.Printf("Player %s error: %s", lp.username, err)
			return
		case packet := <-lp.packetQueue:
			lp.conn.WritePacket(packet)
		case packet := <-lp.readPackets:
			lp.processPacket(packet)
		}
	}
}

//Acts on the passed packet
func (lp *LocalPlayer) processPacket(packet protocol.Packet) {
	switch packet := packet.(type) {
	default:
		log.Printf("Unhandled packet %X from %s", packet.ID(), lp.username)
	}
}

//Close and cleanup the player. The packetReader will close
//once the orginal net.Conn is closed.
func (lp *LocalPlayer) close() {
	lp.closedChannel <- struct{}{}
}

//Reads incomming packets and passes them to the watcher
func (lp *LocalPlayer) packetReader() {
	for {
		packet, err := lp.conn.ReadPacket()
		if err != nil {
			lp.errorChannel <- err
			return
		}
		lp.readPackets <- packet
	}
}