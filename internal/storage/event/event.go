package event

import (
	"context"
	"database/sql"
	"time"

	"digital-wallet/internal/constant/model/db"
	"digital-wallet/internal/constant/model/dto"
	"digital-wallet/internal/constant/model/persistencedb"
	st "digital-wallet/internal/storage"
	"digital-wallet/platform/logger"

	"github.com/jackc/pgtype"
)

type event struct {
	log          logger.Logger
	eventStorage persistencedb.PersistenceDB
}

func Init(
	log logger.Logger,
	eventStorage persistencedb.PersistenceDB,
) st.EventRepository {
	return &event{
		log:          log,
		eventStorage: eventStorage,
	}
}

func (r *event) Save(ctx context.Context, event dto.WebhookEvent) (*dto.WebhookEvent, error) {
	data, err := r.eventStorage.CreateEvent(ctx, db.CreateEventParams{
		EventID:       event.EventID,
		EventType:     string(event.Type),
		Payload:       pgtype.JSONB{Bytes: []byte(event.Payload), Status: pgtype.Present}, // Convert Payload to pgtype.JSONB
		Attempts:      int32(event.Attempts),
		LastAttemptAt: sql.NullTime{Valid: !event.LastAttemptAt.IsZero(), Time: event.LastAttemptAt}, // Correct type for LastAttemptAt
		NextAttemptAt: sql.NullTime{Valid: event.NextAttemptAt != nil, Time: *event.NextAttemptAt},   // Handle NextAttemptAt
		ErrorMessage:  sql.NullString{Valid: event.ErrorMessage != nil, String: *event.ErrorMessage}, // Handle ErrorMessage
	})
	if err != nil {
		return nil, err
	}
	return &dto.WebhookEvent{
		EventID:       data.EventID,
		Type:          dto.EventType(data.EventType),
		Attempts:      int(data.Attempts),
		LastAttemptAt: data.LastAttemptAt.Time,
		NextAttemptAt: func() *time.Time {
			if data.NextAttemptAt.Valid {
				return &data.NextAttemptAt.Time
			}
			return nil
		}(),
		ErrorMessage: func() *string {
			if data.ErrorMessage.Valid {
				return &data.ErrorMessage.String
			}
			return nil
		}(),
	}, nil
}
func (r *event) Exists(ctx context.Context, eventID string) (bool, error) {
	_, err := r.eventStorage.GetEventByEventID(ctx, eventID)
	if err != nil {
		// if errors.i(err, pgx.ErrNoRows) {
		// 	return false, nil
		// }
		return false, err
	}
	return true, nil
}
func (r *event) GetEventByEventID(ctx context.Context,
	eventID string) (*dto.WebhookEvent, error) {
	data, err := r.eventStorage.GetEventByEventID(ctx, eventID)
	if err != nil {
		// if errors.is(err, pgx.ErrNoRows) {
		// 	return false, nil
		// }
	}
	return &dto.WebhookEvent{
		EventID:       data.EventID,
		Type:          dto.EventType(data.EventType),
		Attempts:      int(data.Attempts),
		LastAttemptAt: data.LastAttemptAt.Time,
		NextAttemptAt: func() *time.Time {
			if data.NextAttemptAt.Valid {
				return &data.NextAttemptAt.Time
			}
			return nil
		}(),
		ErrorMessage: func() *string {
			if data.ErrorMessage.Valid {
				return &data.ErrorMessage.String
			}
			return nil
		}(),
	}, nil

}

func (r *event) UpdateStatus(
	ctx context.Context,
	eventID string,
	status dto.EventStatus,
	attempts int,
	lastAttemptAt time.Time,
	nextAttemptAt *time.Time,
	errorMessage *string,
) error {

	return r.eventStorage.UpdateEventStatus(ctx, db.UpdateEventStatusParams{
		ID:            eventID,
		Status:        string(status),
		Attempts:      int32(attempts),
		LastAttemptAt: sql.NullTime{Valid: !lastAttemptAt.IsZero(), Time: lastAttemptAt},
		NextAttemptAt: sql.NullTime{Valid: nextAttemptAt != nil, Time: func() time.Time {
			if nextAttemptAt != nil {
				return *nextAttemptAt
			}
			return time.Time{}
		}()},
		ErrorMessage: sql.NullString{Valid: errorMessage != nil, String: func() string {
			if errorMessage != nil {
				return *errorMessage
			}
			return ""
		}()},
	})
}
