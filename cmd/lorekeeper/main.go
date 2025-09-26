package main

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/riftspire/lorekeeper/pkg/lorekeeper"
	"github.com/spf13/cobra"
)

const (
	// The application name.
	_APP_NAME = "lorekeeper"
)

func main() {
	// Create a new context.
	ctx := context.Background()

	// Initialise the logger.
	initLogging()

	// Execute the cobra.Command.
	if err := newLorekeeperCmd(ctx).Execute(); err != nil {
		log.Fatal(err)
	}
}

func newLorekeeperCmd(ctx context.Context) *cobra.Command {
	var cliArgs Arguments

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s [flags]", _APP_NAME),
		Short: "Keeper of your project's tale, inscribing every release into enduring lore.",
		Long: "Lorekeeper is a release notes generator that transforms commits and tags into a chronicle of your " +
			"project's journey. Instead of scattered changes, you get a cohesive story — a record of growth, fixes, and " +
			"features written like chapters in your code's saga.",
		// Example: "", //TODO: Add this.
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Validate the arguments - populates and arguments from the
			// environment if not provided via flags.
			if err := cliArgs.setAndValidateArgs(); err != nil {
				return err
			}

			// Silence the usage message printing on error from here on.
			cmd.SilenceUsage = true

			// Translate the Mode string to a lorekeeper.mode.
			mode, err := lorekeeper.GetModeByName(cliArgs.Mode)
			if err != nil {
				return err
			}

			// Call Lorekeeper.
			err = lorekeeper.MakeReleaseNotes(
				ctx,
				cliArgs.TagName,
				cliArgs.ReleaseCandidateRegex,
				cliArgs.CurrentBranchName,
				cliArgs.DefaultBranchName,
				mode,
			)
			if err != nil {
				return fmt.Errorf("lorekeeper failed to make release notes: %w", err)
			}

			// Command has completed without error.
			return nil
		},
	}

	// Set the flags for the cobra.Command.
	cliArgs.setFlags(cmd)

	return cmd
}
