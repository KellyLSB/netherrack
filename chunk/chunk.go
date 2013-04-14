package chunk

import (
	"bitbucket.org/Thinkofdeath/soulsand"
)

var (
	_ soulsand.SyncChunk = &Chunk{}
)

type Chunk struct {
	World          *World
	X, Z           int32
	SubChunks      []*SubChunk
	Biome          []byte
	Players        map[int32]soulsand.Player
	Entitys        map[int32]soulsand.Entity
	requests       chan *ChunkRequest
	watcherJoin    chan *chunkWatcherRequest
	watcherLeave   chan *chunkWatcherRequest
	entityJoin     chan *chunkEntityRequest
	entityLeave    chan *chunkEntityRequest
	messageChannel chan *chunkMessage
	eventChannel   chan func(soulsand.SyncChunk)
	blockQueue     []blockChange
}
type SubChunk struct {
	Type       []byte
	MetaData   []byte
	BlockLight []byte
	SkyLight   []byte
	Blocks     uint
}
type ChunkPosition struct {
	X, Z int32
}
type blockChange struct {
	X, Y, Z     int
	Block, Meta byte
}
type ChunkRequest struct {
	X, Z int32
	Stop chan struct{}
	Ret  chan [][]byte
}
type chunkEntityRequest struct {
	Pos ChunkPosition
	E   soulsand.Entity
}
type chunkWatcherRequest struct {
	Pos ChunkPosition
	P   soulsand.Player
}
type chunkMessage struct {
	Pos ChunkPosition
	Msg func(soulsand.SyncEntity)
	ID  int32
}
type chunkBlocksRequest struct {
	Pos     ChunkPosition
	X, Y, Z int
	Ret     chan []byte
}
type chunkEvent struct {
	Pos ChunkPosition
	F   func(soulsand.SyncChunk)
}

func (c *Chunk) GetPlayerMap() map[int32]soulsand.Player {
	return c.Players
}

func (c *Chunk) AddChange(x, y, z int, block, meta byte) {
	c.blockQueue = append(c.blockQueue, blockChange{x, y, z, block, meta})
}

func (c *Chunk) SetBlock(x, y, z int, blType byte) {
	sec := y >> 4
	if c.SubChunks[sec] == nil {
		c.SubChunks[sec] = CreateSubChunk()
	}
	ind := ((y & 15) << 8) | (z << 4) | x
	if c.SubChunks[sec].Type[ind] == 0 && blType != 0 {
		c.SubChunks[sec].Blocks++
	} else if c.SubChunks[sec].Type[ind] != 0 && blType == 0 {
		c.SubChunks[sec].Blocks--
	}
	c.SubChunks[sec].Type[ind] = blType
	if c.SubChunks[sec].Blocks == 0 {
		c.SubChunks[sec] = nil
	}
}

func (c *Chunk) GetBlock(x, y, z int) byte {
	sec := y >> 4
	if sec < 0 || sec >= 16 || c.SubChunks[sec] == nil {
		return 0
	}
	ind := ((y & 15) << 8) | (z << 4) | x
	return c.SubChunks[sec].Type[ind]
}
func (c *Chunk) SetMeta(x, y, z int, data byte) {
	sec := y >> 4
	if c.SubChunks[sec] == nil {
		return
	}
	i := ((y & 15) << 8) | (z << 4) | x
	if i&1 == 0 {
		c.SubChunks[sec].MetaData[i>>1] &= 0xF0
		c.SubChunks[sec].MetaData[i>>1] |= data & 0xF
	} else {
		c.SubChunks[sec].MetaData[i>>1] &= 0xF
		c.SubChunks[sec].MetaData[i>>1] |= (data & 0xF) << 4
	}
}

func (c *Chunk) GetMeta(x, y, z int) byte {
	sec := y >> 4
	if sec < 0 || sec >= 16 || c.SubChunks[sec] == nil {
		return 0
	}
	i := ((y & 15) << 8) | (z << 4) | x
	if i&1 == 0 {
		return c.SubChunks[sec].MetaData[i>>1] & 0xF
	} else {
		return c.SubChunks[sec].MetaData[i>>1] >> 4
	}
	return 0
}

func (c *Chunk) SetBlockLight(x, y, z int, data byte) {
	sec := y >> 4
	if c.SubChunks[sec] == nil {
		return
	}
	i := ((y & 15) << 8) | (z << 4) | x
	if i&1 == 0 {
		c.SubChunks[sec].BlockLight[i>>1] &= 0xF0
		c.SubChunks[sec].BlockLight[i>>1] |= data & 0xF
	} else {
		c.SubChunks[sec].BlockLight[i>>1] &= 0xF
		c.SubChunks[sec].BlockLight[i>>1] |= (data & 0xF) << 4
	}
}
func (c *Chunk) SetSkyLight(x, y, z int, data byte) {
	sec := y >> 4
	if c.SubChunks[sec] == nil {
		return
	}
	i := ((y & 15) << 8) | (z << 4) | x
	if i&1 == 0 {
		c.SubChunks[sec].SkyLight[i>>1] &= 0xF0
		c.SubChunks[sec].SkyLight[i>>1] |= data & 0xF
	} else {
		c.SubChunks[sec].SkyLight[i>>1] &= 0xF
		c.SubChunks[sec].SkyLight[i>>1] |= (data & 0xF) << 4
	}
}

func (c *Chunk) SetBiome(x, z int, biome byte) {
	c.Biome[x|(z<<4)] = biome
}

func CreateChunk(x, z int32) *Chunk {
	chunk := &Chunk{
		X:              x,
		Z:              z,
		SubChunks:      make([]*SubChunk, 16),
		Biome:          make([]byte, 256),
		Players:        make(map[int32]soulsand.Player),
		Entitys:        make(map[int32]soulsand.Entity),
		requests:       make(chan *ChunkRequest, 500),
		watcherJoin:    make(chan *chunkWatcherRequest, 200),
		watcherLeave:   make(chan *chunkWatcherRequest, 200),
		entityJoin:     make(chan *chunkEntityRequest, 200),
		entityLeave:    make(chan *chunkEntityRequest, 200),
		messageChannel: make(chan *chunkMessage, 1000),
		eventChannel:   make(chan func(soulsand.SyncChunk), 500),
		blockQueue:     make([]blockChange, 0, 3),
	}
	return chunk
}

func CreateSubChunk() *SubChunk {
	subChunk := &SubChunk{
		Type:       make([]byte, 16*16*16),
		MetaData:   make([]byte, (16*16*16)/2),
		BlockLight: make([]byte, (16*16*16)/2),
		SkyLight:   make([]byte, (16*16*16)/2),
	}
	return subChunk
}

type chunkMessageEvent interface {
	Run(interface{})
	GetEID() int32
}
