package activity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var poolDefault *Pool

func setupPool(t *testing.T) {
	poolDefault = NewActivityPool(1, nil)
}

func teardownPool(t *testing.T) {
	poolDefault = nil
}

func TestNewActivityPool(t *testing.T) {
	t.Run("initiate an activity pool with limit 1 and default ID generator", func(t *testing.T) {
		pool := NewActivityPool(1, nil)
		assert.NotNil(t, pool)
		assert.Empty(t, pool.activities)

		activity, err := pool.NewActivity("test-activity-1")
		assert.NotNil(t, activity)
		assert.Equal(t, uint64(time.Now().Unix()), activity.ID)
		assert.Equal(t, "test-activity-1", activity.Name)
		assert.Nil(t, err)
		assert.NotEmpty(t, pool.activities)

		activity, err = pool.NewActivity("test-activity-2")
		assert.ErrorIs(t, err, ErrActivityPoolReachedLimit)
		assert.Nil(t, activity)
		assert.Len(t, pool.activities, 1)
	})

	t.Run("initiate an activity pool with limit 1 and user-defined ID generator", func(t *testing.T) {
		pool := NewActivityPool(1, func() uint64 {
			return uint64(time.Now().Unix()) + 1
		})
		assert.NotNil(t, pool)
		assert.Empty(t, pool.activities)

		activity, err := pool.NewActivity("test-activity-3")
		assert.NotNil(t, activity)
		assert.Equal(t, uint64(time.Now().Unix())+1, activity.ID)
		assert.Equal(t, "test-activity-3", activity.Name)
		assert.Nil(t, err)
		assert.NotEmpty(t, pool.activities)

		activity, err = pool.NewActivity("test-activity-4")
		assert.ErrorIs(t, err, ErrActivityPoolReachedLimit)
		assert.Nil(t, activity)
		assert.Len(t, pool.activities, 1)
	})
}

func TestPool_GetActivity(t *testing.T) {
	setupPool(t)
	defer teardownPool(t)
	activity, err := poolDefault.NewActivity("test-activity-5")
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
	setupPool(t)
	defer teardownPool(t)
	activity, err := poolDefault.NewActivity("test-activity-5")
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
