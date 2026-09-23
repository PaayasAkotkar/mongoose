// Package admin implments the start of the app handling its deps using the uber-fx
package admin

import (
	sdk "app/mongoose/sdk-v1"
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type IApp struct {
	SendAddr    string
	RecieveAddr string
	Addr        string
}

type Handlers struct {
	GET      func(ctx context.Context)
	POST     func(ctx context.Context)
	Download func(ctx context.Context)
}

type Servers struct {
	fx.In
	LC       fx.Lifecycle
	Recieve  *sdk.App
	App      *gin.Engine
	Config   IApp
	Handlers Handlers
}

// Start inits the sdk-2 & starts the app
// todo: implement the proper shutdown
func Start(s Servers) {
	appCtx, appCancel := context.WithCancel(context.Background())
	httpSrv := &http.Server{Addr: s.Config.Addr, Handler: s.App}

	s.LC.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			s.Handlers.GET(appCtx)
			s.Handlers.POST(appCtx)

			go func() {
				log.Println("[download] starting")
				s.Handlers.Download(appCtx)
				log.Println("[download] stopped")
			}()

			go func() {
				log.Printf("[app] starting on %s", s.Config.Addr)
				if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Printf("[app] failed: %v", err)
				}
				log.Println("[app] stopped")
			}()

			go func() {
				log.Println("[udp_client] starting")
				s.Recieve.Connect(appCtx)
				log.Println("[udp_client] stopped")
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("shutting down all servers")
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := httpSrv.Shutdown(shutdownCtx); err != nil {
				log.Printf("[app] graceful shutdown failed: %v", err)
			}
			appCancel()
			return nil
		},
	})
}
