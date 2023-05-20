package activity

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var poolDefault *Pool

func setupPool(t *testing.T, limit uint32) {
	poolDefault = NewActivityPool(limit, nil)
}

func teardownPool(t *testing.T) {
	poolDefault = nil
}

func TestNewActivityPool(t *testing.T) {
	t.Run("initializes an activity pool with limit 1 and default ID generator", func(t *testing.T) {
		pool := NewActivityPool(1, nil)
		assert.NotNil(t, pool)
		assert.Empty(t, pool.activities)

		activity, err := pool.NewActivity("test-activity-1", true)
		assert.NotNil(t, activity)
		assert.Equal(t, uint64(time.Now().Unix()), activity.ID)
		assert.Equal(t, "test-activity-1", activity.Name)
		assert.Nil(t, err)
		assert.NotEmpty(t, pool.activities)

		activity, err = pool.NewActivity("test-activity-2", true)
		assert.ErrorIs(t, err, ErrActivityPoolReachedLimit)
		assert.Nil(t, activity)
		assert.Len(t, pool.activities, 1)
	})

	t.Run("initializes an activity pool with limit 1 and user-defined ID generator", func(t *testing.T) {
		pool := NewActivityPool(1, func() uint64 {
			return uint64(time.Now().Unix()) + 1
		})
		assert.NotNil(t, pool)
		assert.Empty(t, pool.activities)

		activity, err := pool.NewActivity("test-activity-3", true)
		assert.NotNil(t, activity)
		assert.Equal(t, uint64(time.Now().Unix())+1, activity.ID)
		assert.Equal(t, "test-activity-3", activity.Name)
		assert.Nil(t, err)
		assert.NotEmpty(t, pool.activities)

		activity, err = pool.NewActivity("test-activity-4", true)
		assert.ErrorIs(t, err, ErrActivityPoolReachedLimit)
		assert.Nil(t, activity)
		assert.Len(t, pool.activities, 1)
	})
}

func TestPool_GetActivity(t *testing.T) {
	setupPool(t, 1)
	defer teardownPool(t)
	activity, err := poolDefault.NewActivity("test-activity-5", false)
	assert.Equal(t, uint64(time.Now().Unix()), activity.ID)
	assert.Equal(t, "test-activity-5", activity.Name)
	assert.Nil(t, err)

	t.Run("get existed activity", func(t *testing.T) {
		a, err := poolDefault.GetActivity(activity.ID)
		assert.Equal(t, activity.ID, a.ID)
		assert.Equal(t, activity.Name, a.Name)
		assert.Nil(t, err)
	})
	t.Run("get non-existed activity", func(t *testing.T) {
		a, err := poolDefault.GetActivity(activity.ID + 1)
		assert.Nil(t, a)
		assert.ErrorIs(t, err, ErrActivityNonExists)
	})
}

func TestPool_RemoveActivity(t *testing.T) {
	setupPool(t, 1)
	defer teardownPool(t)
	activity, err := poolDefault.NewActivity("test-activity-5", false)
	assert.Equal(t, uint64(time.Now().Unix()), activity.ID)
	assert.Equal(t, "test-activity-5", activity.Name)
	assert.Nil(t, err)

	t.Run("remove non-existed activity", func(t *testing.T) {
		assert.Len(t, poolDefault.activities, 1)
		err := poolDefault.RemoveActivity(activity.ID + 1)
		assert.ErrorIs(t, err, ErrActivityNonExists)
		assert.Len(t, poolDefault.activities, 1)
	})
	t.Run("remove existed activity", func(t *testing.T) {
		assert.Len(t, poolDefault.activities, 1)
		err := poolDefault.RemoveActivity(activity.ID)
		assert.Nil(t, err)
		assert.Len(t, poolDefault.activities, 0)
	})
}

func TestPool_NewActivity(t *testing.T) {
	const N = 5
	setupPool(t, N)
	defer teardownPool(t)

	t.Run("in parallel", func(t *testing.T) {
		poolDefault.activitiesLimit = N
		index := uint64(10)
		poolDefault.activitiesIDGenerator = func() uint64 {
			return atomic.AddUint64(&index, 1)
		}
		var wgNotification, wgReady sync.WaitGroup
		wgNotification.Add(1)
		wgReady.Add(N)
		for i := 0; i < N; i++ {
			i := i
			go func() {
				wgNotification.Wait()
				activity, err := poolDefault.NewActivity(fmt.Sprintf("%s-%d", "test-activity-6", i), false)
				t.Log(activity)
				assert.NotNil(t, activity)
				assert.Nil(t, err)
				wgReady.Done()
			}()
		}
		wgNotification.Done()
		t.Log("starting...")
		wgReady.Wait()
		t.Log("finished.")
		assert.Len(t, poolDefault.activities, N)
	})
}

func TestPool_NewActivity2(t *testing.T) {
	const N = 5
	setupPool(t, N)
	defer teardownPool(t)

	t.Run("replace existed", func(t *testing.T) {
		poolDefault.activitiesIDGenerator = func() uint64 {
			return 1
		}
		assert.Empty(t, poolDefault.activities)
		activity, err := poolDefault.NewActivity("test-activity-7-1", false)
		assert.NotNil(t, activity)
		assert.Equal(t, activity.ID, uint64(1))
		assert.Equal(t, activity.Name, "test-activity-7-1")
		assert.Nil(t, err)
		assert.Len(t, poolDefault.activities, 1)
		activity, err = poolDefault.NewActivity("test-activity-7-2", false)
		assert.NotNil(t, activity)
		assert.Equal(t, activity.ID, uint64(1))
		assert.Equal(t, activity.Name, "test-activity-7-1")
		assert.ErrorIs(t, err, ErrActivityExisted)
		assert.Len(t, poolDefault.activities, 1)
		activity, err = poolDefault.NewActivity("test-activity-7-3", true)
		assert.NotNil(t, activity)
		assert.Equal(t, activity.ID, uint64(1))
		assert.Equal(t, activity.Name, "test-activity-7-3")
		assert.Nil(t, err)
		assert.Len(t, poolDefault.activities, 1)
	})
}
