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

func TestRunPeriodically(t *testing.T) {
	t.Parallel()

	t.Run("run successfully", func(t *testing.T) {
		t.Parallel()

		var (
			ctx     = context.Background()
			stopper = make(chan struct{})

			sm  sync.Mutex
			run int
		)

		rp := routine.RunPeriodically(ctx, func(ctx context.Context) error {
			sm.Lock()
			defer sm.Unlock()

			run++

			return nil
		}, 5*time.Millisecond, stopper)

		time.Sleep(20 * time.Millisecond)

		close(stopper)

		require.NoError(t, rp.Err())
		require.Equal(t, 4, run)
	})

	t.Run("run with error", func(t *testing.T) {
		t.Parallel()

		var (
			ctx     = context.Background()
			stopper = make(chan struct{})
		)

		defer close(stopper)

		rp := routine.RunPeriodically(
			ctx,
			func(ctx context.Context) error {
				return errors.New("error")
			},
			5*time.Millisecond,
			stopper,
			routine.WithOnError(func(err error) bool {
				return false
			}),
		)

		require.Error(t, rp.Err())
		require.ErrorIs(t, rp.Err(), errors.New("error"))
	})
}
