package workerpool

import (
	"context"
	"sync"
	"time"
)

type JobFunc[J any, R any] func(ctx context.Context, job J) (R, error)

type Result[R any] struct {
	Value R
	Err   error
}

type Config struct {
	WorkerCount int
	QueueSize   int
	MaxRetries  int
	RetryDelay  time.Duration
}

func RunPool[J any, R any](
	ctx context.Context,
	cfg Config,
	jobs []J,
	fn JobFunc[J, R],
) []Result[R] {

	jobCh := make(chan J, cfg.QueueSize)
	resultCh := make(chan Result[R], len(jobs))

	var wg sync.WaitGroup

	// --- worker ---
	worker := func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return

			case job, ok := <-jobCh:
				if !ok {
					return
				}

				var res R
				var err error

				for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
					res, err = fn(ctx, job)
					if err == nil {
						break
					}

					select {
					case <-ctx.Done():
						return
					case <-time.After(cfg.RetryDelay):
					}
				}

				select {
				case resultCh <- Result[R]{Value: res, Err: err}:
				case <-ctx.Done():
					return
				}
			}
		}
	}

	// --- start workers ---
	wg.Add(cfg.WorkerCount)
	for i := 0; i < cfg.WorkerCount; i++ {
		go worker()
	}

	// --- feed jobs ---
	go func() {
		defer close(jobCh)
		for _, j := range jobs {
			select {
			case jobCh <- j:
			case <-ctx.Done():
				return
			}
		}
	}()

	// --- wait workers ---
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// --- collect results ---
	results := make([]Result[R], 0, len(jobs))
	for r := range resultCh {
		results = append(results, r)
	}

	return results
}
