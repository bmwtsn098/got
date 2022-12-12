package cells

import (
	"context"

	"github.com/brendoncarroll/go-state/cells"
	"github.com/brendoncarroll/go-state/cells/cryptocell"
)

type (
	ErrTooLarge = cells.ErrTooLarge
)

type Cell = cells.Cell

func Apply(ctx context.Context, cell Cell, fn func([]byte) ([]byte, error)) error {
	return cells.Apply(ctx, cell, fn)
}

func GetBytes(ctx context.Context, cell Cell) ([]byte, error) {
	return cells.GetBytes(ctx, cell)
}

func NewMem() cells.Cell {
	return cells.NewMem(1 << 16)
}

func NewEncrypted(inner cells.Cell, secret *[32]byte) Cell {
	return cryptocell.NewChaCha20Poly1305(inner, secret)
}
