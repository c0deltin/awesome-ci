package acigitlab

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"fullstack-devops/awesome-ci/pkg/semver"
	"github.com/xanzy/go-gitlab"
)

var (
	ctx context.Context
	// githubRepository, isgithubRepository = os.LookupEnv("GITHUB_REPOSITORY")
)

type StandardPrInfos struct {
	PrNumber       int
	Owner          string
	Repo           string
	PatchLevel     semver.PatchLevel
	CurrentVersion string
	LatestVersion  string
	NextVersion    string
	Sha            string
	ShaShort       string
	BranchName     string
	MergeCommitSha string
}

func devideOwnerAndRepo(fullRepo string) (owner string, repo string) {
	return strings.Split(fullRepo, "/")[0], strings.Split(fullRepo, "/")[1]
}

// GetPrInfos need the PullRequest-Number
func GetMrInfos(mrNumber int) (standardPrInfos *StandardPrInfos, prInfos *gitlab.MergeRequest, err error) {
	if mrNumber != 0 {
		prInfos, _, err = GitLabClient.MergeRequests.GetMergeRequest(1, mrNumber, nil, nil)
		if err != nil {
			err = errors.New(fmt.Sprintln("could not load any information about the current pull request", err))
			return
		}
	}

	prSHA := prInfos.SHA
	branchName := prInfos.Reference
	patchLevel := semver.ParsePatchLevel(branchName)

	standardPrInfos = &StandardPrInfos{
		PrNumber:   mrNumber,
		BranchName: branchName,
		Sha:        prSHA,
		ShaShort:   prSHA[:8],
		PatchLevel: patchLevel,
	}
	return
}
