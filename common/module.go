package common

import (
	"context"
)

// Module represents a discoverable data source producing a Report.
// Modules register themselves via init() and are discovered by the console runner.
//
// Lifecycle:
// - Enabled(cfg) tells if the module should run (based on config presence)
// - Run(ctx, cfg) performs fetching/aggregation and returns a Report
//
// All network calls should respect ctx deadline/cancellation.

type Module interface {
	Name() string
	Enabled(cfg *Config) bool
	Run(ctx context.Context, cfg *Config) (*Report, error)
}

var registry []Module

// Register adds a module into the global registry. Should be called in module's init().
func Register(m Module) {
	registry = append(registry, m)
}

// Modules returns a snapshot of registered modules.
func Modules() []Module { return append([]Module(nil), registry...) }
