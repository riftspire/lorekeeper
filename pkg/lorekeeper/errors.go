package lorekeeper

import "fmt"

type DefaultBranchReleaseCandidateError struct {
	TagName       string
	DefaultBranch string
}

func (e *DefaultBranchReleaseCandidateError) Error() string {
	return fmt.Sprintf(
		"non-release candidate tags (%s) can only be on the default branch (%s)",
		e.TagName, e.DefaultBranch,
	)
}

type ModeGetByNameError struct {
	Name string
}

func (e *ModeGetByNameError) Error() string {
	return fmt.Sprintf(
		"invalid mode name: expected one of %s, got %s",
		getModeNamesString(), e.Name,
	)
}

type ModeInvalidError struct {
	Mode mode
}

func (e *ModeInvalidError) Error() string {
	return fmt.Sprintf(
		"invalid mode: expected one of %s, got %#v",
		getModeVarNamesString(), e.Mode,
	)
}

type NoPullRequestsFoundError struct {
	Mode      mode
	LatestRef gitReference
}

func (e *NoPullRequestsFoundError) Error() string {
	return fmt.Sprintf(
		"no pull requests merged since latest %s date (%s @ %s) found",
		e.Mode, e.LatestRef.TagName, e.LatestRef.TagName,
	)
}
