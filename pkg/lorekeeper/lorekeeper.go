package lorekeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const packageName = "lorekeeper"

type mode struct {
	Name        string
	VarName     string
	Description string
}

var (
	ModeRelease = mode{
		Name:        "release",
		VarName:     "ModeRelease",
		Description: "Can only be used for GitHub repositories that utilise the GitHub Releases feature.",
	}
	ModeTag = mode{
		Name:        "tag",
		VarName:     "ModeTag",
		Description: "Can be used with any Git repositories.",
	}
	nilMode = mode{}
)

func GetModes() []mode {
	return []mode{
		ModeRelease,
		ModeTag,
	}
}

func GetModeByName(name string) (mode, error) {
	for _, mode := range GetModes() {
		if mode.Name == name {
			return mode, nil
		}
	}
	return nilMode, &ModeGetByNameError{Name: name}
}

func getModeNamesString() string {
	var modeStrings []string
	for _, mode := range GetModes() {
		modeStrings = append(modeStrings, mode.Name)
	}
	return strings.Join(modeStrings, ", ")
}

func getModeVarNamesString() string {
	var modeNames []string
	for _, mode := range GetModes() {
		modeNames = append(modeNames, fmt.Sprintf("%s.%s", packageName, mode.VarName))
	}
	return strings.Join(modeNames, ", ")
}

type gitReference struct {
	PublishedAt time.Time `json:"publishedAt"`
	TagName     string    `json:"tagName"`
}

type gitAuthor struct {
	AvatarURL string `json:"avatarUrl"`
	Login     string `json:"login"`
}

type gitCommit struct {
	Authors []gitAuthor `json:"authors"`
}

type gitPullRequest struct {
	Title   string      `json:"title"`
	Body    string      `json:"body"`
	Commits []gitCommit `json:"commits"`
}

// MakeReleaseNotes queries the provided owner/repo with the provided tag to
// build the release notes for a new release, whether it is a release candidate
// or not.
//
// The release notes will be output to stdout.
func MakeReleaseNotes(
	ctx context.Context,

	// tagName is the release tag to use when checking for relevant branches and
	// pull requests.
	tagName string,

	// releaseCandidateRegex is the regex pattern to use to identify tags
	// that are release candidates.
	releaseCandidateRegex string,

	// currentBranchName is the name of the current branch.
	currentBranchName string,

	// defaultBranchName is the name of the default branch in the specified
	// repository (i.e - main, master, etc).
	defaultBranchName string,

	// mode determines whether GitHub Releases or Git Tags are being used to
	// identify releases.
	//
	// Possible values are:
	//	MODE_RELEASE	// Can only be used for GitHub repositories that utilise the GitHub Releases feature
	//	MODE_TAG			// Can be used with any Git repositories.
	mode mode,
) error {
	// The compiled regular expression to identify candidate release tags.
	reReleaseCandidate := regexp.MustCompile(releaseCandidateRegex)

	// Check if the tag is a release candidate.
	tagIsReleaseCandidate := reReleaseCandidate.MatchString(tagName)

	// Check if the tag belongs to the default branch.
	// TODO: May need to remove "refs/heads" from the current branch name, if it
	// comes from github.event.base_ref
	tagIsOnDefaultBranch := currentBranchName == defaultBranchName

	// Initialise the latest reference variables.
	var (
		err           error
		prList        string
		latestRef     gitReference
		latestRefJSON string
	)

	switch {
	case !tagIsOnDefaultBranch && tagIsReleaseCandidate:
		// If the tag IS NOT on the default branch, and IS a release candidate,
		// include the release notes from the associated branch's pull request.

		// Get the SHA of the latest commit for the given tag.
		//
		// This also checks if the tag exists in the repository.
		//
		// `git ref-list` returns commits in reverse chronological order (newest to
		// oldest)
		latestTagCommit, err := runCmd(fmt.Sprintf(
			"git rev-list -n 1 \"%s\"",
			tagName,
		))
		if err != nil {
			// TODO: Handle error from running the command.
		}

		// Get the pull request associated with the latest commit for the given tag.
		//
		// `gh pr list` returns pull requests in reverse chronological order
		// (newewst to oldest) sorted by createdAt, and doesn't let you change it.
		//
		// TODO: This uses the `gh` CLI app, so is locked to GitHub.
		// Find another way to do this without `gh`.
		prList, err = runCmd(fmt.Sprintf(
			"gh pr list "+
				"--search \"sha:%s\" "+
				"--json number | jq '.[].number", latestTagCommit))
		if err != nil {
			// TODO: Handle error from running the command.
		}
	case !tagIsOnDefaultBranch && !tagIsReleaseCandidate:
		// If the tag IS NOT on the default branch, and IS NOT a release candidate,
		// exit with an error as this is not permitted.
		return &DefaultBranchReleaseCandidateError{}
	case tagIsOnDefaultBranch:

		if tagIsReleaseCandidate {
			// If the tag IS on the default branch, and IS a release candidate, include
			// the release notes from ALL pull requests since the the latest (release or
			// tag depending on the mode).
			switch mode {
			case ModeRelease:
				// TODO: This uses the `gh` CLI app, so is locked to GitHub.
				// Find another way to do this without `gh`.
				latestRefJSON, err = runCmd(fmt.Sprintf(
					"gh release view %s",
					tagName,
				))
				if err != nil {
					// TODO: Handle error from running the command.
				}
			case ModeTag:
				latestRefJSON, err = runCmd(
					"git for-each-ref refs/tags " +
						"--sort=-creatordate " +
						"--format '{\"publishedAt\":\"%(creatordate:iso-strict)\",\"tagName\":\"%(refname)\"} | head -n 1",
				)
				if err != nil {
					// TODO: Handle error from running the command.
				}
			default:
				return &ModeInvalidError{Mode: mode}
			}
		} else {
			// If the tag IS on the default branch, and IS NOT a release candidate,
			// include the release notes from ALL pull requests since the the latest
			// non-release candidate ref (release or tag depending on the mode).
			switch mode {
			case ModeRelease:
				// Get the latest ron-RC release date.
				//
				// `gh release list` returns releases in reverse chronological order
				// (newest to oldest) sorted by createdAt.
				//
				// TODO: This uses the `gh` CLI app, so is locked to GitHub.
				// Find another way to do this without `gh`.
				var allReleases string
				allReleases, err = runCmd(
					"gh release list " +
						"--json publishedAt,tagName",
				)
				if err != nil {
					// TODO: Handle error from running the command.
				}

				// Iterate through the releases to find the latest non-RC release.
				for releaseJSON := range strings.SplitSeq(allReleases, "\n") {
					var release gitReference
					err = json.Unmarshal([]byte(releaseJSON), &release)
					if err != nil {
						// TODO: Handle error from running the command.
					}
					if reReleaseCandidate.MatchString(release.TagName) {
						latestRefJSON = releaseJSON
					}
				}
			case ModeTag:
				latestRefJSON, err = runCmd(
					"git for-each-ref refs/tags " +
						"--exclude=\"refs/tags/*-rc*\"" + // TODO: Use the regex here.
						"--sort=-creatordate " +
						"--format '{\"publishedAt\":\"%(creatordate:iso-strict)\",\"tagName\":\"%(refname)\"} | head -n 1",
				)
				if err != nil {
					// TODO: Handle error from running the command.
				}
			default:
				return &ModeInvalidError{Mode: mode}
			}
		}
		// Marshal the latest ref JSON.
		err = json.Unmarshal([]byte(latestRefJSON), &latestRef)
		if err != nil {
			// TODO: Handle error from running the command.
		}

		// Get all pull requests merged after the latestRef.PublishedAt.
		//
		// `gh pr list` returns pull requests in reverse chronological order
		// (newest to oldest) sorted by createdAt, and doesn't let you change it.
		//
		// TODO: This uses the `gh` CLI app, so is locked to GitHub.
		// Find another way to do this without `gh`.
		prList, err = runCmd(fmt.Sprintf(
			"gh pr list "+
				"--state \"merged\" "+
				"--search \"merged:>%s\" "+
				"--json number | jq '.[].number'",
			latestRef.PublishedAt,
		))
		if err != nil {
			// TODO: Handle error from running the command.
		}

		// If there are no pull requests found, exit with an error.
		if prList == "" {
			return &NoPullRequestsFoundError{}
		}
	}

	// Iterate over each pull request.
	for pullRequestNumber := range strings.SplitSeq(prList, "\n") {
		// Get the pull request details.
		//
		// TODO: This uses the `gh` CLI app, so is locked to GitHub.
		// Find another way to do this without `gh`.
		pullRequestJSON, err := runCmd(fmt.Sprintf(
			"gh pr view \"%s\" "+
				"--json title,body,commits",
			pullRequestNumber,
		))
		if err != nil {
			// TODO: Handle error from running the command.
		}

		// Unmarshal the pull request JSON.
		var pullRequest gitPullRequest
		err = json.Unmarshal([]byte(pullRequestJSON), &pullRequest)
		if err != nil {
			// TODO: Handle error from running the command.
		}

		// Output the pull request header.
		fmt.Printf("# %s (#%s)\n\n", pullRequest.Title, pullRequestNumber)

		// Output the pull request authors header.
		fmt.Print("## Authors\n\n")

		// Output the pull request authors.
		var authors []string
		for _, commit := range pullRequest.Commits {
			for _, author := range commit.Authors {
				var reUrl = regexp.MustCompile(`(v=[0-9]+)`)
				avatarUrl := reUrl.ReplaceAllString(author.AvatarURL, "s=64&amp;$1")
				authors = append(authors, fmt.Sprintf("!\"[@%s](%s)", author.Login, avatarUrl))
			}
		}
		fmt.Printf("%s\n\n", strings.Join(authors, " "))

		// Output the pull request body.
		fmt.Printf("%s\n\n", pullRequest.Body)
	}

	return nil
}

// ===
// Helper Functions

func runCmd(command string) (string, error) {
	// Run the command.
	cmd := exec.Command(command, strings.Split(command, " ")...)
	err := cmd.Run()
	if err != nil {
		// TODO: Handle error from running the command.
		fmt.Printf("ERROR: [cmd.Run] %v", err)
		return "", err
	}

	// Get the output.
	output, err := cmd.Output()
	if err != nil {
		// TODO: Handle error from running the command.
		fmt.Printf("ERROR: [cmd.Output] %v", err)
		return "", err
	}

	// TODO: DEBUG: check the output.
	fmt.Printf("DEBUG: %q output: %s", command, output)

	// Return the output from running the command.
	return string(output), nil
}
