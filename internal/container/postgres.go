package container

import (
	"context"

	"github.com/MartinAbdrakhmanov/diploma/pkg/storage"
	"github.com/go-faster/errors"
)

func (c *Container) GetDbManager(ctx context.Context) (*storage.Client, error) {
	if c.dbManager != nil {
		return c.dbManager, nil
	}

	pgxPool, err := storage.InitDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init dbManager")
	}

	dbManager := storage.NewClient(pgxPool, pgxPool)

	c.dbManager = dbManager
	c.closers = append(c.closers, func() {
		pgxPool.Close()
	})

	return c.dbManager, err
}
