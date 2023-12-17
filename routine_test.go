package routine_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dohernandez/errors"
	"github.com/stretchr/testify/require"

	"github.com/dohernandez/routine"
)

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("run successfully", func(t *testing.T) {
		t.Parallel()

		var (
			ctx     = context.Background()
			stopper = make(chan struct{})

			sm  sync.Mutex
			run int
		)

		r := routine.Run(ctx, func(ctx context.Context) error {
			sm.Lock()
			defer sm.Unlock()

			run++

			return nil
		}, stopper)

		close(stopper)

		require.NoError(t, r.Err())
		require.Equal(t, 1, run)
	})

	t.Run("run with error", func(t *testing.T) {
		t.Parallel()

		var (
			ctx     = context.Background()
			stopper = make(chan struct{})
		)

		r := routine.Run(ctx, func(ctx context.Context) error {
			return errors.New("error")
		}, stopper)

		close(stopper)

		require.Error(t, r.Err())
		require.ErrorIs(t, r.Err(), errors.New("error"))
	})

	t.Run("run with stop", func(t *testing.T) {
		t.Parallel()

		var (
			ctx     = context.Background()
			stopper = make(chan struct{})

			sm  sync.Mutex
			run int
		)

		r := routine.Run(ctx, func(ctx context.Context) error {
			t := time.NewTicker(40 * time.Millisecond)

			select {
			case <-stopper:
				return nil
			case <-t.C:
				break
			}

			sm.Lock()
			defer sm.Unlock()

			run++

			return nil
		}, stopper)

		time.Sleep(30 * time.Millisecond)
		close(stopper)

		require.NoError(t, r.Err())
		require.Equal(t, 0, run)
	})
}
