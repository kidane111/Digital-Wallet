package retry

import (
	"context"
	"fmt"
	"time"

	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/constant/state"
	"digital-wallet/platform/logger"
	"digital-wallet/platform/routine"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
)

func SetParams(params state.RetryParams) state.RetryParams {
	if params.InitialInterval == 0 {
		params.InitialInterval = 10 * time.Second
	}
	if params.RandomizationFactor == 0 {
		params.RandomizationFactor = 0.1
	}
	if params.Multiplier == 0 {
		params.Multiplier = 5
	}
	if params.MaxInterval == 0 {
		params.MaxInterval = 24 * time.Hour
	}
	if params.MaxElapsedTime == 0 {
		params.MaxElapsedTime = 7 * 24 * time.Hour
	}

	return params
}

// PushJob launches a job on a separate routine with an exponential backoff retry.
func PushJob[T any](ctx context.Context, log logger.Logger, retry state.RetryParams, job dto.Job[T]) {
	routine.ExecuteRoutine(ctx, routine.Routine{
		Name: job.Name,
		OnCancelled: func(ctx context.Context, log logger.Logger) {
			if job.OnFailure != nil {
				err := job.OnFailure(ctx, fmt.Errorf("job cancelled"))
				if err != nil {
					log.Error(ctx, "failed to process onFailure of a failed job on context cancellation",
						zap.String("job-name", job.Name),
						zap.Error(err))
				}
			}
		},
		Operation: func(ctx context.Context, log logger.Logger) {
			log = log.Named("retry")
			ctx, cancel := context.WithTimeout(ctx, retry.MaxElapsedTime)
			defer cancel()

			log.Info(ctx, fmt.Sprintf("retry started for job %s", job.Name))
			err := backoff.Retry(func() error {
				if job.Operation == nil {
					return backoff.Permanent(fmt.Errorf("job with no Operation"))
				}

				// return error if context is cancelled
				if err := ctx.Err(); err != nil {
					return backoff.Permanent(fmt.Errorf("job retry cancelled: %v", err))
				}

				// retry operation
				result, err := job.Operation(ctx)
				if err != nil {
					log.Warn(ctx, "retrying job",
						zap.String("job-name", job.Name),
						zap.Error(err))
					return err
				}

				if job.OnSuccess == nil {
					return backoff.Permanent(fmt.Errorf("job with no OnSuccess"))
				}

				// do update
				err = job.OnSuccess(ctx, result)
				if err != nil {
					return backoff.Permanent(err)
				}

				// this job was successful
				return nil
			}, &backoff.ExponentialBackOff{
				InitialInterval:     retry.InitialInterval,
				RandomizationFactor: retry.RandomizationFactor,
				Multiplier:          retry.Multiplier,
				MaxInterval:         retry.MaxInterval,
				MaxElapsedTime:      retry.MaxElapsedTime,
				Stop:                backoff.Stop,
				Clock:               backoff.SystemClock,
			})
			if err != nil {
				log.Error(ctx, "failed to process a retry job",
					zap.String("job-name", job.Name),
					zap.Error(err))

				if job.OnFailure != nil { // not required -> to keep backward compatibility
					if err := job.OnFailure(ctx, err); err != nil {
						log.Error(ctx, "failed to process onFailure of a failed job",
							zap.String("job-name", job.Name),
							zap.Error(err))
					}
				}
			} else {
				log.Info(ctx, "job succeeded!",
					zap.String("job-name", job.Name))
			}
		},
	}, log)
}

// PushJobWithLimitedCount launches a job on a separate routine with an exponential backoff retry
// and a retry count limit.
func PushJobWithLimitedCount[T any](
	ctx context.Context,
	log logger.Logger, retry state.RetryParams,
	job dto.Job[T],
	countLimit int) {
	routine.ExecuteRoutine(ctx, routine.Routine{
		Name: job.Name,
		OnCancelled: func(ctx context.Context, log logger.Logger) {
			if job.OnFailure != nil {
				err := job.OnFailure(ctx, fmt.Errorf("job cancelled"))
				if err != nil {
					log.Error(ctx, "failed to process onFailure of a failed job on context cancellation",
						zap.String("job-name", job.Name),
						zap.Error(err))
				}
			}
		},
		Operation: func(ctx context.Context, log logger.Logger) {
			log = log.Named("retry")
			ctx, cancel := context.WithTimeout(ctx, retry.MaxElapsedTime)
			defer cancel()

			count := 0

			log.Info(ctx, fmt.Sprintf("retry started for job %s", job.Name))
			err := backoff.Retry(func() error {
				if countLimit > 0 && count >= countLimit {
					return backoff.Permanent(fmt.Errorf("reached count limit: %d", countLimit))
				}

				count++

				if job.Operation == nil {
					return backoff.Permanent(fmt.Errorf("job with no Operation"))
				}

				// return error if context is cancelled
				if err := ctx.Err(); err != nil {
					return backoff.Permanent(fmt.Errorf("job retry cancelled: %v", err))
				}

				// retry operation
				result, err := job.Operation(ctx)
				if err != nil {
					log.Warn(ctx, "retrying job",
						zap.String("job-name", job.Name),
						zap.Error(err))
					return err
				}

				if job.OnSuccess == nil {
					return backoff.Permanent(fmt.Errorf("job with no OnSuccess"))
				}

				// do update
				err = job.OnSuccess(ctx, result)
				if err != nil {
					return backoff.Permanent(err)
				}

				// this job was successful
				return nil
			}, &backoff.ExponentialBackOff{
				InitialInterval:     retry.InitialInterval,
				RandomizationFactor: retry.RandomizationFactor,
				Multiplier:          retry.Multiplier,
				MaxInterval:         retry.MaxInterval,
				MaxElapsedTime:      retry.MaxElapsedTime,
				Stop:                backoff.Stop,
				Clock:               backoff.SystemClock,
			})
			if err != nil {
				log.Error(ctx, "failed to process a retry job",
					zap.String("job-name", job.Name),
					zap.Error(err))

				if job.OnFailure != nil { // not required -> to keep backward compatibility
					if err := job.OnFailure(ctx, err); err != nil {
						log.Error(ctx, "failed to process OnFailure of a failed job",
							zap.String("job-name", job.Name),
							zap.Error(err))
					}
				}
			} else {
				log.Info(ctx, "job succeeded!",
					zap.String("job-name", job.Name))
			}
		},
	}, log)
}
