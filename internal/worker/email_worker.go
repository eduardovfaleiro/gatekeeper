package worker

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type EmailWorker struct {
	redis *redis.Client
	emailService
}

func (w *EmailWorker) Start(ctx context.Context) {
	for {
		streams, err := w.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "email_processors",
			Consumer: "worker_1",
			Streams:  []string{"email_stream", ">"},
			Count:    1,
			Block:    0,
		}).Result()

		if err != nil {
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				err := w.emailService.SendResetLink(
					msg.Values["email"].(string),
					msg.Values["token"].(string),
				)

				if err == nil {
					w.redis.XAck(ctx, "email_stream", "email_processors", msg.ID)
				}
			}
		}
	}
}
