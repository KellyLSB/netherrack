package chunk

import (
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/blocks"
)

func (chunk *Chunk) generate() {
	switch chunk.tryLoad() {
	case 0:
		chunk.World.generator.Generate(int(chunk.X), int(chunk.Z), chunk)
		chunk.needsRelight = true
		chunk.needsSave = true
		if chunk.World.chunkLoaded(chunk.X+1, chunk.Z) {
			chunk.World.RunSync(int(chunk.X+1), int(chunk.Z), relight)
		}

		if chunk.World.chunkLoaded(chunk.X-1, chunk.Z) {
			chunk.World.RunSync(int(chunk.X-1), int(chunk.Z), relight)
		}

		if chunk.World.chunkLoaded(chunk.X, chunk.Z+1) {
			chunk.World.RunSync(int(chunk.X), int(chunk.Z+1), relight)
		}

		if chunk.World.chunkLoaded(chunk.X, chunk.Z-1) {
			chunk.World.RunSync(int(chunk.X), int(chunk.Z-1), relight)
		}
	case 2: //Damaged chunk
		defaultGenerator(0).GenerateBlock(int(chunk.X), int(chunk.Z), chunk, blocks.RedstoneBlock.Id())
		chunk.needsRelight = true
	}
}

func relight(chunk soulsand.SyncChunk) {
	chunk.(*Chunk).needsRelight = true
}

type defaultGenerator int

func (dg defaultGenerator) Generate(x, z int, chunk soulsand.SyncChunk) {
	dg.GenerateBlock(x, z, chunk, blocks.Wool.Id())
}

func (defaultGenerator) GenerateBlock(x, z int, chunk soulsand.SyncChunk, block byte) {
	for y := 0; y < 256; y++ {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				if y <= 64 {
					chunk.SetBlock(x, y, z, block)
					if x == 0 || x == 15 || z == 0 || z == 15 {
						chunk.SetMeta(x, y, z, 1)
					} else {
						chunk.SetMeta(x, y, z, byte((y>>4)+4))
					}
				}
			}
		}
	}
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			chunk.SetBiome(x, z, 1)
		}
	}
}
