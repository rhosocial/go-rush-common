package activity

import (
	"errors"
	"sync"
	"time"
)

// Pool defines an activity pool.
// Please do not access any member directly, but use related methods unless you know the consequences of doing so.
type Pool struct {
	// activities stores all the activities.
	// The key of the map is the activity ID.
	activities map[uint64]*Activity
	// activitiesRWMutex describes the access mutex for activities.
	activitiesRWMutex sync.RWMutex
	// activitiesLimit defines the upper limit of the capacity.
	activitiesLimit uint32
	// activitiesIDGenerator defines the function handle that generates the activity ID.
	// After the Pool is initialized, you can also modify this member at any time.
	// If not specified, use NewActivityID()
	activitiesIDGenerator func() uint64
}

// NewActivityPool initializes an activity pool.
func NewActivityPool(limit uint32, IDGenerator func() uint64) *Pool {
	pool := Pool{activities: make(map[uint64]*Activity, limit), activitiesLimit: limit, activitiesIDGenerator: IDGenerator}
	return &pool
}

var ErrActivityPoolReachedLimit = errors.New("the capacity of activity pool has reached the upper limit")

// NewActivity initializes an activity and add it to the pool.
//
// Returns the pointer to the just added activity if added successfully.
//
// If the activity ID already exists and replacement is not allowed, it will return a pointer to the corresponding
// activity and an ErrActivityExisted exception.
//
// If the upper limit is not reached, otherwise an ErrActivityPoolReachedLimit is returned.
func (p *Pool) NewActivity(name string, replace bool) (*Activity, error) {
	p.activitiesRWMutex.Lock()
	defer p.activitiesRWMutex.Unlock()
	id := p.GetActivityIDGenerator()()
	if activity, existed := p.activities[id]; existed && !replace {
		return activity, ErrActivityExisted
	}
	if uint32(len(p.activities)) >= p.activitiesLimit {
		return nil, ErrActivityPoolReachedLimit
	}
	activity := Activity{id, name}
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

// NewActivityID generates an ID with the timestamp of calling this method.
func (p *Pool) NewActivityID() uint64 {
	return uint64(time.Now().Unix())
}

var ErrActivityExisted = errors.New("activity existed")
var ErrActivityNonExists = errors.New("activity not exists")

// RemoveActivity removes the activity
//
// Returns nil if the activity exists and is successfully deleted, otherwise an ErrActivityNonExists is returned.
func (p *Pool) RemoveActivity(id uint64) error {
	p.activitiesRWMutex.Lock()
	defer p.activitiesRWMutex.Unlock()
	if _, exist := p.activities[id]; exist {
		delete(p.activities, id)
		return nil
	}
	return ErrActivityNonExists
}

// GetActivity gets the activity.
//
// If the activity does not exist, an ErrActivityNonExists is returned.
func (p *Pool) GetActivity(id uint64) (*Activity, error) {
	p.activitiesRWMutex.RLock()
	defer p.activitiesRWMutex.RUnlock()
	if activity, exist := p.activities[id]; exist {
		return activity, nil
	}
	return nil, ErrActivityNonExists
}
