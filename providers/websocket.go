package providers

import (
	"github.com/samber/do"

	"joints-be/config"
	"joints-be/pkg/event"
	wsPkg "joints-be/pkg/websocket"
)

// ProvideWebSocket registers the WebSocket Hub, EventPublisher, and Handler into the DI injector.
func ProvideWebSocket(i *do.Injector) {
	do.Provide(i, func(i *do.Injector) (wsPkg.Hub, error) {
		cfg := config.AppConfig
		maxConns := 100
		if cfg != nil && cfg.WebSocketMaxConnections > 0 {
			maxConns = cfg.WebSocketMaxConnections
		}
		return wsPkg.NewHub(maxConns), nil
	})

	do.Provide(i, func(i *do.Injector) (event.EventPublisher, error) {
		hub := do.MustInvoke[wsPkg.Hub](i)
		return wsPkg.NewHubEventPublisher(hub), nil
	})

	do.Provide(i, func(i *do.Injector) (*wsPkg.Handler, error) {
		hub := do.MustInvoke[wsPkg.Hub](i)
		return wsPkg.NewHandler(hub), nil
	})
}

