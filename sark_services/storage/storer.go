package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"

	"github.com/panjf2000/ants/v2"
)

var ErrFileNotFound = errors.New("file not found")

type Reader interface {
	Read(ctx context.Context, name string) ([]byte, error)
}

type Writer interface {
	Write(ctx context.Context, name string, data []byte) error
	WriteFrom(ctx context.Context, name string, r io.Reader) error
}

type Remover interface {
	Remove(ctx context.Context, name string) error
}

type Lister interface {
	List(ctx context.Context, prefix string) ([]string, error)
}

type FileStorer interface {
	Reader
	Writer
	Remover
	Lister
}

func Replace(ctx context.Context, storer interface {
	Writer
	Remover
}, name string, data []byte) error {
	if err := storer.Remove(ctx, name); err != nil {
		return err
	}
	return storer.Write(ctx, name, data)
}

func WriteBatch(ctx context.Context, writer Writer, files map[string][]byte) error {
	readers := make(map[string]io.Reader, len(files))
	for name, data := range files {
		readers[name] = bytes.NewReader(data)
	}
	return WriteFromBatch(ctx, writer, readers)
}

func WriteFromBatch(ctx context.Context, writer Writer, files map[string]io.Reader) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(files))

	pool, err := ants.NewPool(10)
	if err != nil {
		return err
	}
	defer pool.Release()

	for name, reader := range files {
		fileName := name
		fileReader := reader
		wg.Add(1)
		err := pool.Submit(func() {
			defer wg.Done()
			if err := writer.WriteFrom(ctx, fileName, fileReader); err != nil {
				errCh <- err
			}
		})
		if err != nil {
			wg.Done()
			return err
		}
	}

	wg.Wait()
	close(errCh)
	if len(errCh) > 0 {
		return <-errCh
	}
	return nil
}
