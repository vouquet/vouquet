package lock

import (
	"fmt"
	"context"
)

var (
	ERR_CONTEXT_CANCEL error = fmt.Errorf("canceled lock by context.")
)

type TryMutex struct {
	l   chan struct{}
	ctx context.Context
}

func NewTryMutex(ctx context.Context) *TryMutex {
	if ctx == nil {
		ctx = context.Background()
	}
	return &TryMutex{
		l: make(chan struct{}),
		ctx: ctx,
	}
}

func (self *TryMutex) Lock() error {
	select {
		case <- self.ctx.Done():
			return ERR_CONTEXT_CANCEL
		case self.l <- struct{}{}:
			return nil
	}
	return nil
}

func (self *TryMutex) TryLock() (bool, error) {
	select {
		case <- self.ctx.Done():
			return false, ERR_CONTEXT_CANCEL
		case self.l <- struct{}{}:
			return true, nil
		default:
	}
	return false, nil
}

func (self *TryMutex) Unlock() error {
	select {
		case <- self.ctx.Done():
			return ERR_CONTEXT_CANCEL
		case <- self.l:
			return nil
		default:
			panic("lock: unlock of unlocked TryMutex")
	}
	return nil
}
