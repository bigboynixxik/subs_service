package main

import (
	"context"
	"subs_service/internal/app"
)

// @title Subscription Service API
// @version 1.0
// @description REST API для агрегации данных об онлайн подписках пользователей.
// @host localhost:8080
// @BasePath /
func main() {
	ctx := context.Background()
	a, err := app.New(ctx)
	if err != nil {
		panic(err)
	}
	a.Run()
}
