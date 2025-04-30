package dto

import "context"

type Job[T any] struct {
	// Name is A Unique name for this job.
	// Please make sure this name is unique and correctly describes this exact job.
	// This is the only identifier you will have when monitoring ongoing jobs!
	Name string
	// Operation is the function that will do the actual job that is needed to be retried on failure.
	// It will be retried on failure.
	// It should return the result of the operation on success
	Operation func(ctx context.Context) (T, error)
	// OnSuccess is the function that will update the system based on the successful execution of this Job's Request.
	// It won't be retried on failure since this is in-house failure of the system.
	// `result` holds the result from Request
	OnSuccess func(ctx context.Context, result T) error
	// OnFailure is the function that will be called if the retry is stopped because
	// it has reached its retry limit and the job hasn't succeeded.
	OnFailure func(ctx context.Context, err error) error
}

type ContextKey any
