package loafergo

import "context"

// Handler represents the handler function
type Handler func(context.Context, Message) error
