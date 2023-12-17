package routine

import (
	"context"

	"github.com/bool64/ctxd"
	"golang.org/x/sync/errgroup"
)

// RunConcurrent runs the operation concurrently.
//
// The operation is executed in a goroutine per element of the slice but limited by the given limit.
// The operation will be stopped when an error, when the given context is canceled or when the given stopper channel is closed.
func RunConcurrent(
	ctx context.Context,
	elms []any,
	limit int,
	op func(ctx context.Context, elm any) (any, error),
	opDone func(ctx context.Context, elm any) error,
	stopper <-chan struct{},
) *Routine {
	rc := &Routine{
		done: make(chan struct{}),
	}

	go func() {
		defer func() {
			rc.Stop()
		}()

		if len(elms) == 0 {
			return
		}

		limit = min(limit, len(elms))

		pool := make(chan any, len(elms))

		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(limit)

		for i := 0; i < limit; i++ {
			j := i

			g.Go(func() error {
				rctx := ctxd.AddFields(gctx, "c_routine", j+1)

				for {
					select {
					case <-stopper:
						return nil
					case <-rctx.Done():
						return nil
					case e, ok := <-pool:
						if !ok {
							return nil
						}

						r, err := op(rctx, e)
						if err != nil {
							return err
						}

						if opDone == nil {
							continue
						}

						// Check if the routine has been canceled during the op.
						select {
						case <-stopper:
							return nil
						case <-rctx.Done():
							return nil
						default:
							err = opDone(rctx, r)
							if err != nil {
								return err
							}

							break
						}
					}
				}
			})
		}

		for _, e := range elms {
			pool <- e
		}

		close(pool)

		if err := g.Wait(); err != nil {
			rc.err = err
		}
	}()

	return rc
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
