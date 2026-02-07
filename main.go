package main

import (
	"context"
	"fmt"
	"time"

	"github.com/faisal-porag/production_grade_worker_pool_example/workerpool"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	cfg := workerpool.Config{
		WorkerCount: 5,
		QueueSize:   20,
		MaxRetries:  2,
		RetryDelay:  time.Millisecond * 200,
	}

	results := workerpool.RunPool(ctx, cfg, jobs,
		func(ctx context.Context, job int) (int, error) {
			time.Sleep(time.Millisecond * 300)
			return job * job, nil
		},
	)

	for _, r := range results {
		if r.Err != nil {
			fmt.Println("error:", r.Err)
			continue
		}
		fmt.Println("result:", r.Value)
	}
}
