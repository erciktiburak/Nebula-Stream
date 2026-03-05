package worker

import "context"

type Job func(context.Context) error

func RunAsync(ctx context.Context, jobs []Job) error {
	for _, job := range jobs {
		if err := job(ctx); err != nil {
			return err
		}
	}
	return nil
}
