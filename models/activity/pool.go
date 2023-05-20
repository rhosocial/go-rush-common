package activity

import (
	"errors"
	"sync"
	"time"
)

type Pool struct {
	activities            map[uint64]*Activity
	activitiesRWMutex     sync.RWMutex
	activitiesLimit       uint32
	activitiesIDGenerator func() uint64
}

func NewActivityPool(limit uint32, IDGenerator func() uint64) *Pool {
	pool := Pool{activities: map[uint64]*Activity{}, activitiesLimit: limit, activitiesIDGenerator: IDGenerator}
	return &pool
}

var ErrActivityPoolReachedLimit = errors.New("the capacity of activity pool has reached the upper limit")

func (p *Pool) NewActivity(name string) (*Activity, error) {
	p.activitiesRWMutex.Lock()
	defer p.activitiesRWMutex.Unlock()
	if uint32(len(p.activities)) >= p.activitiesLimit {
		return nil, ErrActivityPoolReachedLimit
	}
	activity := Activity{p.GetActivityIDGenerator()(), name}
	p.activities[activity.ID] = &activity
	return &activity, nil
}

func (p *Pool) GetActivityIDGenerator() func() uint64 {
	fn := p.activitiesIDGenerator
	if fn == nil {
		fn = p.NewActivityID
	}
	return fn
}

func (p *Pool) NewActivityID() uint64 {
	return uint64(time.Now().Unix())
}

var ErrActivityNonExists = errors.New("activity not exists")

func (p *Pool) RemoveActivity(id uint64) error {
	p.activitiesRWMutex.Lock()
	defer p.activitiesRWMutex.Unlock()
	if _, exist := p.activities[id]; exist {
		delete(p.activities, id)
		return nil
	}
	return ErrActivityNonExists
}

func (p *Pool) GetActivity(id uint64) (*Activity, error) {
	p.activitiesRWMutex.RLock()
	defer p.activitiesRWMutex.RUnlock()
	if activity, exist := p.activities[id]; exist {
		return activity, nil
	}
	return nil, ErrActivityNonExists
}
