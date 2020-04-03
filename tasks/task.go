package tasks

import (
	"context"
	"fmt"
)

type Task interface {
	Setup(ctx context.Context) error
	Execute(ctx context.Context) error
}

func NewTask(taskName string, tenant string) (Task, error) {
	switch taskName {
	case "catalogue-bulk-index":
		return NewCatalogueBulkIndexTask(tenant)
	case "attributes-bulk-index":
		return NewAttributesBulkIndexTask(tenant)
	default:
		return nil, fmt.Errorf("Task does not exist with name: %s", taskName)
	}
}
