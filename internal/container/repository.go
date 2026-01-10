package container

import (
	"context"

	"github.com/MartinAbdrakhmanov/diploma/internal/repository"
	"github.com/go-faster/errors"
)

func (c *Container) GetRepo(ctx context.Context) (*repository.Repository, error) {
	if c.repo != nil {
		return c.repo, nil
	}

	dbManager, err := c.GetDbManager(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetRepo dbManager init err")
	}

	repo := repository.New(dbManager)
	c.repo = repo

	return repo, nil
}
