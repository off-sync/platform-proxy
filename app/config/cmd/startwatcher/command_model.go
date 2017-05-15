package startwatcher

import "context"
import "sync"

type Changes struct {
	Services  []string
	Frontends []string
}

type ChangesCallback func(changes *Changes)

type CommandModel struct {
	WaitGroup *sync.WaitGroup
	Ctx       context.Context
	Callback  ChangesCallback
}
