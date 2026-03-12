package worker

import (
	"context"
	"log"

	"github.com/eduardovfaleiro/gatekeeper/internal/service"
	"github.com/redis/go-redis/v9"
)

type emailWorker struct {
	redis        *redis.Client
	emailService service.EmailService
}

func NewEmailWorker(redis *redis.Client, emailService service.EmailService) *emailWorker {
	return &emailWorker{
		redis:        redis,
		emailService: emailService,
	}
}

func (w *emailWorker) Start(ctx context.Context) {
	for {
		streams, err := w.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "email_processors",
			Consumer: "worker_1",
			Streams:  []string{"email_stream", ">"},
			Count:    1,
			Block:    0,
		}).Result()

		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("Erro ao ler stream: %v", err)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				email, okEmail := msg.Values["email"].(string)
				token, okToken := msg.Values["token"].(string)

				if !okEmail || !okToken {
					log.Printf("Mensagem inválida recebida: %v", msg.Values)
					continue
				}

				err := w.emailService.SendResetLink(
					email,
					token,
				)

				if err == nil {
					w.redis.XAck(ctx, "email_stream", "email_processors", msg.ID)
					log.Printf("Worker: E-mail de reset enviado para %s", email)
				}
			}
		}
	}
}
