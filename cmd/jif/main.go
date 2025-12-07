// Command jif is a modern, high-performance GIF viewer for the terminal
package main

import (
	"context"
	"fmt"
	"os"

	jif "github.com/Gaurav-Gosain/jif/core"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

// Version information (set by goreleaser)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "jif [gif-file-or-url]",
		Short: "A modern GIF viewer for your terminal",
		Long: `jif - A modern, high-performance GIF viewer for your terminal

Displays GIF animations in your terminal using halfblock rendering for
2x vertical resolution. Supports local files and remote URLs.

Features:
  - Halfblock rendering (2x resolution)
  - High-quality Lanczos3 scaling
  - Pause/resume, frame navigation
  - Progressive loading animation
  - GIF disposal method handling`,
		Example: `  # View a local GIF
  jif animation.gif

  # View a remote GIF
  jif https://example.com/animation.gif

  # Press ? while viewing for keybindings`,
		Version:      version,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return jif.Run(args[0])
		},
	}

	// Execute with fang
	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(fmt.Sprintf("%s\nCommit: %s\nBuilt: %s\nBy: %s", version, commit, date, builtBy)),
	); err != nil {
		os.Exit(1)
	}
}
