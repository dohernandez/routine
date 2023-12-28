package routine_test

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dohernandez/errors"
	"github.com/stretchr/testify/require"

	"github.com/dohernandez/routine"
)

func getGoroutineID() string {
	var buf [64]byte

	n := runtime.Stack(buf[:], false)

	// Parse the 4707 out of "goroutine 4707 [running]:"
	return strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
}

func TestRunInPool(t *testing.T) {
	t.Parallel()

	t.Run("run successfully all parallel", func(t *testing.T) {
		t.Parallel()

		var (
			ctx     = context.Background()
			run     int
			stopper = make(chan struct{})

			elms = []any{1, 2, 3, 4, 5}

			sm         sync.Mutex
			routineIDs = make(map[string]bool)
		)

		defer close(stopper)

		rc := routine.RunInPool(ctx, elms, 10, func(ctx context.Context, elm interface{}) (any, error) {
			// Simulate some work to be able to test the concurrency
			time.Sleep(5 * time.Millisecond)

			sm.Lock()
			run++
			sm.Unlock()

			return getGoroutineID(), nil
		}, func(ctx context.Context, elm interface{}) error {
			sm.Lock()
			defer sm.Unlock()

			routineIDs[elm.(string)] = true

			return nil
		}, stopper)

		require.NoError(t, rc.Err())
		require.Equal(t, 5, run)
		require.Len(t, routineIDs, 5)
	})

	t.Run("run successfully 3 in parallel", func(t *testing.T) {
		t.Parallel()

		var (
			ctx     = context.Background()
			run     int
			stopper = make(chan struct{})

			elms = []any{1, 2, 3, 4, 5}

			sm         sync.Mutex
			routineIDs = make(map[string]bool)
		)

		defer close(stopper)

		rc := routine.RunInPool(ctx, elms, 3, func(ctx context.Context, elm interface{}) (any, error) {
			// Simulate some work to be able to test the concurrency
			time.Sleep(5 * time.Millisecond)

			sm.Lock()
			run++
			sm.Unlock()

			return getGoroutineID(), nil
		}, func(ctx context.Context, elm interface{}) error {
			sm.Lock()
			defer sm.Unlock()

			routineIDs[elm.(string)] = true

			return nil
		}, stopper)

		require.NoError(t, rc.Err())
		require.Equal(t, 5, run)
		require.Len(t, routineIDs, 3)
	})

	t.Run("failure one and rest stopped", func(t *testing.T) {
		t.Parallel()

		var (
			ctx     = context.Background()
			run     int
			stopper = make(chan struct{})

			elms = []any{1, 2, 3, 4, 5}

			sm         sync.Mutex
			routineIDs = make(map[string]bool)
		)

		defer close(stopper)

		rc := routine.RunInPool(ctx, elms, 10, func(ctx context.Context, elm interface{}) (any, error) {
			// Simulate some work to be able to test the concurrency
			time.Sleep(5 * time.Millisecond)

			var more bool

			sm.Lock()
			run++

			if run > 1 {
				more = true
			}
			sm.Unlock()

			// Simulate an extra work to test error triggering by the first operation.
			if more {
				time.Sleep(10 * time.Millisecond)
			}

			return getGoroutineID(), nil
		}, func(ctx context.Context, elm interface{}) error {
			sm.Lock()
			defer sm.Unlock()

			routineIDs[elm.(string)] = true

			return errors.New("error")
		}, stopper)

		require.Error(t, rc.Err())
		require.Equal(t, 5, run)
		require.Len(t, routineIDs, 1)
	})
}
