package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"dev-digest/console"
	// Side-effect imports to register modules
	_ "dev-digest/modules/azureboards"
	_ "dev-digest/modules/azurerepos"
	_ "dev-digest/modules/teamcity"
)

func main() {
	var cfgPath string
	var noColor bool
	flag.StringVar(&cfgPath, "config", "", "path to YAML config file (default: ./config.yaml or $DEV_DIGEST_CONFIG)")
	flag.BoolVar(&noColor, "no-color", false, "disable colored output")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := console.Run(ctx, cfgPath, noColor); err != nil {
		fmt.Fprintln(os.Stderr, "dev-digest:", err)
		os.Exit(1)
	}
}
