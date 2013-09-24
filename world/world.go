/*
   Copyright 2013 Matthew Collins (purggames@gmail.com)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package world

import (
	"bytes"
	"compress/zlib"
	"github.com/NetherrackDev/netherrack/protocol"
)

//Dimensions normally control the lighting and skycolour
type Dimension int8

const (
	Overworld Dimension = 0
	Nether    Dimension = -1
	End       Dimension = 1
)

type World struct {
	name string

	system    System
	generator Generator

	loadedChunks map[uint64]*Chunk

	joinChunk   chan joinChunk
	leaveChunk  chan joinChunk
	placeBlock  chan blockChange
	chunkPacket chan chunkPacket

	//The limiters were added because trying to send/save all the chunks
	//at once caused large amounts of memory usage
	SendLimiter  chan cachedCompressor
	SaveLimiter  chan struct{}
	RequestClose chan *Chunk

	worldData struct {
		Dimension Dimension
	}
}

type cachedCompressor struct {
	buf *bytes.Buffer
	zl  *zlib.Writer
}

func (world *World) init() {
	world.joinChunk = make(chan joinChunk, 500)
	world.leaveChunk = make(chan joinChunk, 500)
	world.placeBlock = make(chan blockChange, 1000)
	world.chunkPacket = make(chan chunkPacket, 1000)
	world.RequestClose = make(chan *Chunk, 20)
}

func (world *World) run() {
	world.generator.Load(world)
	world.loadedChunks = make(map[uint64]*Chunk)
	world.SendLimiter = make(chan cachedCompressor, 20)
	for i := 0; i < cap(world.SendLimiter); i++ {
		buf := &bytes.Buffer{}
		world.SendLimiter <- cachedCompressor{buf, zlib.NewWriter(buf)}
	}
	world.SaveLimiter = make(chan struct{}, 5)
	for i := 0; i < cap(world.SaveLimiter); i++ {
		world.SaveLimiter <- struct{}{}
	}
	for {
		select {
		case jc := <-world.joinChunk:
			world.chunk(jc.x, jc.z).Join(jc.watcher)
		case lc := <-world.leaveChunk:
			world.chunk(lc.x, lc.z).Leave(lc.watcher)
		case bp := <-world.placeBlock:
			cx, cz := bp.X>>4, bp.Z>>4
			world.chunk(cx, cz).blockPlace <- bp
		case cp := <-world.chunkPacket:
			world.chunk(cp.X, cp.Z).chunkPacket <- cp
		case chunk := <-world.RequestClose:
			if chunk.Close() {
				delete(world.loadedChunks, chunkKey(chunk.X, chunk.Z))
			}
		}
	}
}

type blockChange struct {
	X, Y, Z     int
	Block, Data byte
}

type chunkPacket struct {
	Packet protocol.Packet
	UUID   string
	X, Z   int
}

//Sends the packet to all watchers of the chunk apart from the watcher with
//the passed uuid (leave blank to send to all)
func (world *World) QueuePacket(x, z int, uuid string, packet protocol.Packet) {
	world.chunkPacket <- chunkPacket{
		packet,
		uuid,
		x, z,
	}
}

//Sets the block and data at the location
func (world *World) SetBlock(x, y, z int, block, data byte) {
	world.placeBlock <- blockChange{
		x, y, z, block, data,
	}
}

type joinChunk struct {
	x, z    int
	watcher Watcher
}

//Adds the watcher to the chunk at the coordinates. If the chunk isn't loaded
//then it will be loaded.
func (world *World) JoinChunk(x, z int, watcher Watcher) {
	world.joinChunk <- joinChunk{x, z, watcher}
}

//Removes the watcher to the chunk at the coordinates.
func (world *World) LeaveChunk(x, z int, watcher Watcher) {
	world.leaveChunk <- joinChunk{x, z, watcher}
}

//Writes the value into the world's system's storage. This method
//is safe to call from different goroutines when the key is different.
func (world *World) Write(key string, value interface{}) error {
	return world.system.Write(key, value)
}

//Reads the value into the world's system's storage. This method
//is safe to call from different goroutines when the key is different.
func (world *World) Read(key string, value interface{}) error {
	return world.system.Read(key, value)
}

//Gets the loaded chunk or loads it if it isn't loaded
func (world *World) chunk(x, z int) *Chunk {
	if chunk, ok := world.loadedChunks[chunkKey(x, z)]; ok {
		return chunk
	}
	chunk, ok := world.system.Chunk(x, z)
	world.loadedChunks[chunkKey(x, z)] = chunk
	if !ok {
		chunk.Init(world, world.generator, world.system)
	} else {
		chunk.Init(world, nil, world.system)
	}
	return chunk
}

//Returns the worlds dimension
func (world *World) Dimension() Dimension {
	return world.worldData.Dimension
}

func chunkKey(x, z int) uint64 {
	return (uint64(int32(x)) & 0xFFFFFFFF) | ((uint64(int32(z)) & 0xFFFFFFFF) << 32)
}
