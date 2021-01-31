package gotfs

import (
	"bytes"
	"context"
	"testing"

	"github.com/brendoncarroll/got/pkg/cadata"
	"github.com/stretchr/testify/require"
)

func TestCreateFileFrom(t *testing.T) {
	ctx := context.Background()
	s := cadata.NewMem()
	x, err := New(ctx, s)
	require.NoError(t, err)
	require.NotNil(t, x)
	fileData := []byte("file contents\n")
	x, err = CreateFileFrom(ctx, s, *x, "file.txt", bytes.NewReader(fileData))
	require.NoError(t, err)
	require.NotNil(t, x)
	buf := make([]byte, 128)
	n, err := ReadFileAt(ctx, s, *x, "file.txt", 0, buf)
	require.NoError(t, err)
	require.Equal(t, string(fileData), string(buf[:n]))
}

func TestFileMetadata(t *testing.T) {
	ctx := context.Background()
	s := cadata.NewMem()
	x, err := New(ctx, s)
	require.NoError(t, err)
	require.NotNil(t, x)
	x, err = CreateFileFrom(ctx, s, *x, "file.txt", bytes.NewReader(nil))
	require.NoError(t, err)
	md, err := GetMetadata(ctx, s, *x, "file.txt")
	require.NoError(t, err)
	require.NotNil(t, md)
	require.True(t, md.Mode.IsRegular())
}
