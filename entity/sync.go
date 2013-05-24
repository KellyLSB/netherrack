package entity

import (
	"errors"
	"github.com/NetherrackDev/soulsand"
)

func (e *Entity) PositionSync() (float64, float64, float64) {
	return e.position.X, e.position.Y, e.position.Z
}

func (e *Entity) SetPositionSync(x, y, z float64) {
	e.position.X, e.position.Y, e.position.Z = x, y, z
}

func (e *Entity) LookSync() (float32, float32) {
	return e.position.Yaw, e.position.Pitch
}

func (e *Entity) SetLookSync(yaw, pitch float32) {
	e.position.Yaw, e.position.Pitch = yaw, pitch
}

func (e *Entity) RunSync(f func(soulsand.SyncEntity)) error {
	select {
	case e.EventChannel <- f:
	case <-e.EntityDead:
		e.EntityDead <- struct{}{}
		return errors.New("Entity removed")
	}
	return nil
}

func (e *Entity) CallSync(f func(soulsand.SyncEntity, chan interface{})) (interface{}, error) {
	ret := make(chan interface{}, 1)
	err := e.RunSync(func(soulsand.SyncEntity) {
		f(e, ret)
	})
	if err == nil {
		select {
		case val := <-ret:
			return val, err
		case <-e.EntityDead:
			e.EntityDead <- struct{}{}
			return nil, errors.New("Entity removed")
		}
	}
	return nil, err
}

func (e *Entity) EntityMetadata() soulsand.EntityMetadata {
	return e.metadata
}
