package console

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"

	"dev-digest/common"
)

// Run loads configuration, discovers modules, collects data, and renders to console.
func Run(ctx context.Context, configPath string, forceNoColor bool) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	cfg.ResolveEnv()
	if err := cfg.Validate(); err != nil {
		// Non-fatal: allow running with partial config; just warn.
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	if forceNoColor || (cfg.Console.Color != nil && !*cfg.Console.Color) {
		color.NoColor = true
	}

	facts := &common.Facts{}
	ctx = common.WithFacts(ctx, facts)

	// Collect reports
	var reports []*common.Report
	for _, m := range common.Modules() {
		if !m.Enabled(cfg) {
			continue
		}
		timeout := cfg.DefaultPerModuleTimeout()
		if cfg.TeamCity != nil && m.Name() == "TeamCity" && cfg.TeamCity.Timeout > 0 {
			timeout = cfg.TeamCity.Timeout
		}
		if cfg.Azure != nil && (m.Name() == "Azure Boards" || m.Name() == "Azure Repos") && cfg.Azure.Timeout > 0 {
			timeout = cfg.Azure.Timeout
		}

		mctx, cancel := context.WithTimeout(ctx, timeout)
		start := time.Now()
		report, err := m.Run(mctx, cfg)
		cancel()
		if err != nil {
			// Create an error report to show instead of failing the whole app
			reports = append(reports, &common.Report{Title: m.Name(), Summary: fmt.Sprintf("error: %v", err), Errors: []string{err.Error()}})
		} else if report != nil {
			// Attach duration
			duration := time.Since(start)
			if report.Meta == nil {
				report.Meta = map[string]any{}
			}
			report.Meta["duration"] = duration.String()
			reports = append(reports, report)
		}
	}

	if len(reports) == 0 {
		return errors.New("no enabled modules / nothing to report; check your config")
	}

	RenderReports(reports)
	return nil
}

func loadConfig(path string) (*common.Config, error) {
	if path == "" {
		if v := os.Getenv("DEV_DIGEST_CONFIG"); v != "" {
			path = v
		} else {
			path = "config.yaml"
		}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg common.Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	return &cfg, nil
}
