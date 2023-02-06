package loafer_go

import "context"

type Handler func(context.Context, Message) error
