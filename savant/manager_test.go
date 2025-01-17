package savant

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yetio/savant.mqtt/config"
)

func TestLightsManager_Poll(t *testing.T) {
	opts := &config.Config{
		RegistryFilePath: testRegistryPath,
		PollSeconds:      1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(1200 * time.Millisecond)
		cancel()
	}()

	manager, err := New(opts)
	assert.NoError(t, err)

	mock := mockSetup()
	mock.output = make([]string, len(manager.Lights))
	for i := range mock.output {
		mock.output[i] = "0"
	}
	mock.output[0] = "100"
	mock.error = nil

	// Ensure we are getting expected state changes
	count := struct{ A, B int }{}
	manager.Poll(ctx, func(stateChange StateChange) {
		if stateChange.ID == manager.ids[0] {
			count.A += 1
			assert.Equal(t, 100, stateChange.Level)
			assert.Equal(t, "ON", stateChange.State())
		}
		if stateChange.ID == manager.ids[1] {
			count.B += 1
			assert.Equal(t, 0, stateChange.Level)
			assert.Equal(t, "OFF", stateChange.State())
		}
	})
	assert.Equal(t, 1, count.A)
	assert.Equal(t, 1, count.B)

	// Ensure we're calling the command line as expected
	assert.Len(t, mock.runs, 1, "Savant command line was called once")
	fmt.Println(mock.runs[0])
	assert.Equal(t, "readstate", mock.runs[0][0], "readstate was called")
	assert.Len(t, mock.runs[0], len(manager.Lights)+1, "readstate was called with a state per light")
}

func TestLightsManager_TurnOn(t *testing.T) {
	opts := &config.Config{
		RegistryFilePath: testRegistryPath,
		PollSeconds:      1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(35 * time.Millisecond)
		cancel()
	}()

	manager, err := New(opts)
	assert.NoError(t, err)

	mock := mockSetup()
	mock.runs = make([][]string, 0)
	mock.output = make([]string, len(manager.Lights))
	for i := range mock.output {
		mock.output[i] = "0"
	}
	mock.error = nil

	// Ensure we are getting expected state changes
	switchId := manager.ids[2]
	dimmerId := manager.ids[0]
	randomId := manager.ids[1]

	count := map[string]int{
		switchId: 0,
		dimmerId: 0,
		randomId: 0,
	}

	manager.TurnOn(switchId)
	manager.TurnOn(dimmerId)

	manager.Poll(ctx, func(stateChange StateChange) {
		if stateChange.ID == switchId {
			assert.Equal(t, 100, stateChange.Level)
			assert.Equal(t, "ON", stateChange.State())
			count[switchId] += 1
		}
		if stateChange.ID == dimmerId {
			assert.Equal(t, 100, stateChange.Level)
			assert.Equal(t, "ON", stateChange.State())
			count[dimmerId] += 1
		}
		if stateChange.ID == manager.ids[1] {
			count[randomId] += 1
		}
	})
	assert.Equal(t, 1, count[switchId])
	assert.Equal(t, 1, count[dimmerId])
	assert.Equal(t, 0, count[randomId])

	// Ensure we're calling the command line as expected
	assert.Len(t, mock.runs, 1, "Savant command line was called once")

	assert.Equal(t, "writestate", mock.runs[0][0], "writestate was called")
	assert.Len(t, mock.runs[0][1:], 4, "writestate with 4 arguments (two pairs)")
	assert.Equal(t, "100", mock.runs[0][2], "Level is set to 100 for on")
	assert.Equal(t, "100", mock.runs[0][4], "Level is set to 100 for on")
}

func TestLightsManager_TurnOff(t *testing.T) {
	opts := &config.Config{
		RegistryFilePath: testRegistryPath,
		PollSeconds:      1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(35 * time.Millisecond)
		cancel()
	}()

	manager, err := New(opts)
	assert.NoError(t, err)

	mock := mockSetup()
	mock.runs = make([][]string, 0)
	mock.output = make([]string, len(manager.Lights))
	for i := range mock.output {
		mock.output[i] = "100"
	}
	mock.error = nil

	// Ensure we are getting expected state changes
	switchId := manager.ids[2]
	dimmerId := manager.ids[0]
	randomId := manager.ids[1]

	count := map[string]int{
		switchId: 0,
		dimmerId: 0,
		randomId: 0,
	}

	manager.TurnOff(switchId)
	manager.TurnOff(dimmerId)

	manager.Poll(ctx, func(stateChange StateChange) {
		if stateChange.ID == switchId {
			assert.Equal(t, 0, stateChange.Level)
			assert.Equal(t, "OFF", stateChange.State())
			count[switchId] += 1
		}
		if stateChange.ID == dimmerId {
			assert.Equal(t, 0, stateChange.Level)
			assert.Equal(t, "OFF", stateChange.State())
			count[dimmerId] += 1
		}
		if stateChange.ID == randomId {
			count[randomId] += 1
		}
	})
	assert.Equal(t, 1, count[switchId])
	assert.Equal(t, 1, count[dimmerId])
	assert.Equal(t, 0, count[randomId])

	// Ensure we're calling the command line as expected
	assert.Len(t, mock.runs, 1, "Savant command line was called once")

	assert.Equal(t, "writestate", mock.runs[0][0], "writestate was called")
	assert.Len(t, mock.runs[0][1:], 4, "writestate with 4 arguments (two pairs)")
	assert.Equal(t, "0", mock.runs[0][2], "Level is set to 0 for off")
	assert.Equal(t, "0", mock.runs[0][4], "Level is set to 0 for off")
}

func TestLightsManager_Set(t *testing.T) {
	opts := &config.Config{
		RegistryFilePath: testRegistryPath,
		PollSeconds:      1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(35 * time.Millisecond)
		cancel()
	}()

	manager, err := New(opts)
	assert.NoError(t, err)

	mock := mockSetup()
	mock.runs = make([][]string, 0)
	mock.output = make([]string, len(manager.Lights))
	for i := range mock.output {
		mock.output[i] = "100"
	}
	mock.error = nil

	// Ensure we are getting expected state changes
	dimmerId := manager.ids[0]
	randomId := manager.ids[1]

	count := map[string]int{
		dimmerId: 0,
		randomId: 0,
	}

	manager.Set(dimmerId, 50)

	manager.Poll(ctx, func(stateChange StateChange) {
		if stateChange.ID == dimmerId {
			assert.Equal(t, 50, stateChange.Level)
			assert.Equal(t, "ON", stateChange.State())
			count[dimmerId] += 1
		}
		if stateChange.ID == randomId {
			count[randomId] += 1
		}
	})
	assert.Equal(t, 1, count[dimmerId])
	assert.Equal(t, 0, count[randomId])

	// Ensure we're calling the command line as expected
	assert.Len(t, mock.runs, 1, "Savant command line was called once")

	assert.Equal(t, "writestate", mock.runs[0][0], "writestate was called")
	assert.Len(t, mock.runs[0][1:], 2, "writestate with 2 arguments (one pair)")
	assert.Equal(t, "50", mock.runs[0][2], "Level is set to 50")
}
