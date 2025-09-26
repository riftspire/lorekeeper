package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/riftspire/lorekeeper/pkg/lorekeeper"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// extendedFlagSet represents a pflag.FlagSet with additional fields.
type extendedFlagSet struct {
	*pflag.FlagSet
	AdditionalFields map[string]any
}

// extendedFlagSetList is a list of extended flag sets.
type extendedFlagSetList []extendedFlagSet

// NewExtendedFlagSet creates a new extended flag set with the provided name and
// fields, appending it to the list. It returns a pointer to the newly created
// extended flag set.
func (efsl *extendedFlagSetList) NewExtendedFlagSet(name string, fields map[string]any) *extendedFlagSet {
	// Create a new flag set with the provided name.
	var efs = extendedFlagSet{
		FlagSet:          pflag.NewFlagSet(name, pflag.ExitOnError),
		AdditionalFields: fields,
	}

	// Append the flag set with rule to the list.
	*efsl = append(*efsl, efs)

	// Return a pointer to the newly added flag set.
	var addedFlagSet = &(*efsl)[len(*efsl)-1]
	return addedFlagSet
}

// AddToCobraCmd adds the extended flag sets to the provided Cobra command.
func (efsl *extendedFlagSetList) AddToCobraCmd(cmd *cobra.Command) {
	for _, efs := range *efsl {
		cmd.Flags().AddFlagSet(efs.FlagSet)
	}
}

type Arguments struct {
	// TagName is the release tag to use when checking for relevant branches and
	// pull requests.
	TagName string

	// ReleaseCandidateRegex is the regex pattern to use to identify tags
	// that are release candidates.
	ReleaseCandidateRegex string

	// CurrentBranchName is the name of the current branch.
	CurrentBranchName string

	// DefaultBranchName is the name of the default branch in the specified
	// repository (i.e - main, master, etc).
	DefaultBranchName string

	// Mode determines whether GitHub Releases or Git Tags are being used to
	// identify releases.
	//
	// Possible values are:
	//	MODE_RELEASE	// Can only be used for GitHub repositories that utilise the GitHub Releases feature
	//	MODE_TAG			// Can be used with any Git repositories.
	Mode string

	// FromEnv is whether the Owner, Repo, Tag, and GitHub Token should be
	// sourced from environment variables.
	//
	// - The Owner and Repo fields will be sourced from the `GITHUB_REPOSITORY`
	// environment variable.
	// - The Tag field will be sourced from the `GITHUB_REF_NAME` environment
	// variable.
	// - The GitHub Token field will be sourced from the `GITHUB_TOKEN`
	// environment variable.
	FromEnv bool

	// Define the debugging arguments.
	Verbosity int
}

func (args *Arguments) setAndValidateArgs() error {
	// Set the log level based on the verbosity flag.
	args.setLogVerbosity()

	return nil
}

// setFlags set the flags for the provided cobra.Command.
func (args *Arguments) setFlags(cmd *cobra.Command) {
	var efsl extendedFlagSetList

	// Application flags.
	fsApplication := efsl.NewExtendedFlagSet("Application", nil)
	fsApplication.StringVarP(&args.TagName, "tag", "t", "",
		"The release tag to use when checking for relevant branches and pull requests.",
	)
	fsApplication.StringVarP(&args.ReleaseCandidateRegex, "release-candidate-regex", "r", "",
		"The regex pattern to use to identify tags that are release candidates.",
	)
	fsApplication.StringVarP(&args.CurrentBranchName, "current-branch-name", "c", "",
		"The name of the current branch.",
	)
	fsApplication.StringVarP(&args.DefaultBranchName, "default-branch-name", "d", "",
		"The name of the default branch in the target repository (i.e - main, master, etc).",
	)
	fsApplication.StringVarP(&args.Mode, "mode", "m", "", getModesUsage())

	// Debugging flags.
	fsDebugging := efsl.NewExtendedFlagSet("Debugging", nil)
	fsDebugging.CountVarP(&args.Verbosity, "verbose", "v", getVerbosityUsage())

	// Add the extended flag sets to the cobra.Command.
	efsl.AddToCobraCmd(cmd)

	// Set the help and usage message functions.

}

// setLogVerbosity sets the logging level based on the `--verbosity` flag.
func (args *Arguments) setLogVerbosity() {
	// If the verbosity is less than or equal to 0, wo do not change the log level.
	if args.Verbosity <= 0 {
		return
	}

	// Cap the verbosity to the maximum allowed value.
	var (
		defaultLogLevelIndex = logLevelIndex(defaultLogLevel)
		maxVerbosity         = len(_AllLogLevels) - defaultLogLevelIndex - 1
	)
	args.Verbosity = max(args.Verbosity, maxVerbosity)

	// Set the logging level depending on the verbosity flag.
	log.SetLevel(_AllLogLevels[defaultLogLevelIndex+args.Verbosity])
}

func getModesUsage() string {
	var availableModes []string
	for _, mode := range lorekeeper.GetModes() {
		availableModes = append(availableModes, fmt.Sprintf("  %s: %s", mode.Name, mode.Description))
	}
	return "Determines whether GitHub Releases or Git Tags are being used to identify releases.\n" +
		strings.Join(availableModes, "\n")
}

// getVerbosityUsage returns the usage string for the `--verbosity` flag.
func getVerbosityUsage() string {
	var usage []string

	for idx, level := range logLevelsAbove(defaultLogLevel) {
		usage = append(usage, fmt.Sprintf(
			"  -%s = %s",
			strings.Repeat("v", idx+1),
			level.String(),
		))
	}

	return "Increase the verbosity level of the output.\n" + strings.Join(usage, "\n")
}
