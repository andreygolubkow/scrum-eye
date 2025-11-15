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
	sprint, err := c.collectCurrentSprint(ctx)
	if err != nil {
		return nil, err
	}

	project := &domain.Project{
		CurrentSprint: sprint,
	}

	return project, nil
}

func (c *Collector) collectCurrentSprint(ctx context.Context) (*domain.Sprint, error) {
	iteration, err := c.boards.GetCurrentIteration(ctx)
	if err != nil {
		return nil, err
	}

	workItems, err := c.boards.GetIterationWorkItems(iteration.ID, ctx)
	if err != nil {
		return nil, err
	}

	sprint := domain.Sprint{
		ID:        iteration.ID,
		Name:      iteration.Name,
		StartDate: iteration.Attributes.StartDate,
		EndDate:   iteration.Attributes.FinishDate,
		WorkItems: MapODataWorkItems(*workItems),
	}

	return &sprint, nil
}
