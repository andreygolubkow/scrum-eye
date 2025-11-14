package collector

import (
	"context"
	"scrum-eye/internal/domain"
	"scrum-eye/internal/sources/azureboards"
)

type Collector struct {
	boards *azureboards.Client
	cfg    Config
}

func NewCollector(boards *azureboards.Client) *Collector {
	return &Collector{boards: boards}
}

func (c *Collector) Collect(ctx context.Context) (*domain.Project, error) {
	iteration, err := c.boards.GetCurrentIterationPath(ctx)
	if err != nil {
		return nil, err
	}

	sprint := domain.Sprint{
		ID:   iteration,
		Name: iteration,
	}

	project := &domain.Project{
		CurrentSprint: sprint,
	}

	return project, nil
}
