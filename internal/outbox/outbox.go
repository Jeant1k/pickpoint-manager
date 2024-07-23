package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/kafka"
)

type OutboxEvent struct {
	EventTime  time.Time `json:"event_time"`
	MethodName string    `json:"method_name"`
	RawRequest string    `json:"raw_request"`
}

type Outbox struct {
	DB       *pgxpool.Pool
	Producer *kafka.Producer
}

func NewOutbox(db *pgxpool.Pool, pr *kafka.Producer) *Outbox {
	return &Outbox{
		DB:       db,
		Producer: pr,
	}
}

func (o *Outbox) AddEvent(ctx context.Context, methodName, rawRequest string) error {
	event := OutboxEvent{
		EventTime:  time.Now(),
		MethodName: methodName,
		RawRequest: rawRequest,
	}
	eventJSON, errMarsh := json.Marshal(event)
	if errMarsh != nil {
		return errMarsh
	}
	_, errExec := o.DB.Exec(
		ctx,
		`INSERT INTO outbox (event_time, method_name, raw_request)
		VALUES ($1, $2, $3)`,
		event.EventTime,
		event.MethodName,
		eventJSON,
	)
	return errExec
}

func (o *Outbox) ProcessEvents(ctx context.Context) error {
	rows, errQuery := o.DB.Query(
		ctx,
		`SELECT outbox_id, event_time, method_name, raw_request
		FROM outbox
		WHERE processed = FALSE
		ORDER BY event_time ASC`,
	)
	if errQuery != nil {
		return errQuery
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		var event OutboxEvent
		var rawRequest []byte
		if err := rows.Scan(&id, &event.EventTime, &event.MethodName, &rawRequest); err != nil {
			return err
		}
		if err := json.Unmarshal(rawRequest, &event); err != nil {
			return err
		}
		if err := o.Producer.SendMessage(event.MethodName, event.RawRequest); err != nil {
			return err
		}
		ids = append(ids, id)
	}

	for _, id := range ids {
		_, errExec := o.DB.Exec(
			ctx,
			`UPDATE outbox
			SET processed = TRUE
			WHERE outbox_id = $1`,
			id,
		)
		if errExec != nil {
			return errExec
		}
	}

	return nil
}

func (o *Outbox) StartBackgroundProcessing(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := o.ProcessEvents(ctx)
			if err != nil {
				fmt.Println("Ошибка обработки событий из Outbox:", err)
			}
		}
	}
}
