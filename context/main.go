package main

import (
	"context"
	"fmt"
)

func enrichContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, "contextId", "123454567")
}

func main() {
	ctx := context.Background()

	newContext := enrichContext(ctx)
	fmt.Println(newContext.Value("contextId"))
}
