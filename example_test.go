package canonlog_test

import (
	"context"
	"errors"

	"github.com/nhalm/canonlog"
)

func ExampleSetupGlobalLogger() {
	canonlog.SetupGlobalLogger("info", "text")
}

func ExampleLogger() {
	ctx := context.Background()
	l := canonlog.New()
	defer l.Flush(ctx)

	l.InfoAdd("user_id", "123")
	l.InfoAdd("action", "create_order")
	l.InfoAddMany(map[string]any{
		"amount":   99.99,
		"currency": "USD",
	})
}

func ExampleInfoAdd() {
	ctx := canonlog.NewContext(context.Background())
	defer canonlog.Flush(ctx)

	canonlog.InfoAdd(ctx, "user_id", "123")
	canonlog.InfoAdd(ctx, "action", "fetch_profile")
}

func ExampleInfoAddMany() {
	ctx := canonlog.NewContext(context.Background())
	defer canonlog.Flush(ctx)

	canonlog.InfoAddMany(ctx, map[string]any{
		"user_id": "123",
		"org_id":  "456",
		"role":    "admin",
	})
}

func ExampleErrorAdd() {
	ctx := canonlog.NewContext(context.Background())
	defer canonlog.Flush(ctx)

	canonlog.ErrorAdd(ctx, errors.New("payment failed"))
}

func ExampleLogger_chainable() {
	ctx := context.Background()
	l := canonlog.New()
	defer l.Flush(ctx)

	l.DebugAdd("cache", "hit").
		InfoAdd("user_id", "123").
		InfoAdd("action", "login").
		InfoAddMany(map[string]any{
			"ip":         "192.168.1.1",
			"user_agent": "Mozilla/5.0",
		})
}
