package interceptor

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// RateLimitInterceptor cria um limitador de requisições baseado no IP ou ID do usuário usando Redis
func RateLimitInterceptor(rdb *redis.Client, limit int, window time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 1. Identificar o cliente (Prioridade: user_id no context -> IP do Peer)
		identifier := "anonymous"
		if userID := ctx.Value("user_id"); userID != nil {
			identifier = fmt.Sprintf("user:%v", userID)
		} else if p, ok := peer.FromContext(ctx); ok {
			host, _, err := net.SplitHostPort(p.Addr.String())
			if err == nil {
				identifier = fmt.Sprintf("ip:%s", host)
			} else {
				identifier = fmt.Sprintf("ip:%s", p.Addr.String())
			}
		}

		// 2. Chave do Redis (janela de tempo fixa)
		// Ex: ratelimit:ip:127.0.0.1:/auth.AuthService/Login:1710292800
		now := time.Now().UnixNano() / int64(window)
		key := fmt.Sprintf("ratelimit:%s:%s:%d", identifier, info.FullMethod, now)

		// 3. Incrementar contador no Redis
		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window*2) // Expira um pouco depois da janela
		_, err := pipe.Exec(ctx)
		if err != nil {
			// Em caso de erro no Redis, permitimos a passagem (fail-open) ou bloqueamos?
			// Geralmente deixamos passar para não derrubar a API por causa do Redis
			return handler(ctx, req)
		}

		// 4. Verificar limite
		if incr.Val() > int64(limit) {
			return nil, status.Error(codes.ResourceExhausted, "too many requests, please try again later")
		}

		return handler(ctx, req)
	}
}
