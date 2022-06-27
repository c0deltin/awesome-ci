package acigithub

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"fullstack-devops/awesome-ci/pkg/semver"
	"fullstack-devops/awesome-ci/src/tools"

	"github.com/google/go-github/v44/github"
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

func StandardPrInfosToEnv(prInfos *StandardPrInfos) (err error) {
	runnerType := "github_runner"
	switch runnerType {
	case "github_runner":
		envVars, err := OpenEnvFile()
		if err != nil {
			return err
		}
		envVars.Set("ACI_PR", strconv.Itoa(prInfos.PrNumber))
		envVars.Set("ACI_PR_SHA", prInfos.Sha)
		envVars.Set("ACI_PR_SHA_SHORT", prInfos.ShaShort)
		envVars.Set("ACI_PR_BRANCH", prInfos.BranchName)
		envVars.Set("ACI_MERGE_COMMIT_SHA", prInfos.MergeCommitSha)

		envVars.Set("ACI_OWNER", prInfos.Owner)
		envVars.Set("ACI_REPO", prInfos.Repo)
		envVars.Set("ACI_PATCH_LEVEL", string(prInfos.PatchLevel))
		envVars.Set("ACI_VERSION", prInfos.NextVersion)
		envVars.Set("ACI_LATEST_VERSION", prInfos.LatestVersion)

		err = envVars.SaveEnvFile()
		if err != nil {
			return err
		}
	default:
		log.Println("Runner Type not implemented!")
	}
	return
}

// GetPrInfos need the PullRequest-Number
func GetPrInfos(prNumber int, mergeCommitSha string) (standardPrInfos *StandardPrInfos, prInfos *github.PullRequest, err error) {
	if !isgithubRepository {
		log.Fatalln("make shure the GITHUB_REPOSITORY is available!")
	}
	owner, repo := tools.DevideOwnerAndRepo(githubRepository)
	if prNumber != 0 {
		prInfos, _, err = GithubClient.PullRequests.Get(ctx, owner, repo, prNumber)
		if err != nil {
			return nil, nil, fmt.Errorf("could not load any information about the given pull request  %d: %v", prNumber, err)
		}
	}
	if mergeCommitSha != "" && prNumber == 0 {
		prOpts := github.PullRequestListOptions{
			State:     "all",
			Sort:      "updated",
			Direction: "desc",
			ListOptions: github.ListOptions{
				PerPage: 10,
			},
		}
		pullRequests, _, err := GithubClient.PullRequests.List(ctx, owner, repo, &prOpts)
		if err != nil {
			return nil, nil, fmt.Errorf("could not load any information about the given pull request  %d: %v", prNumber, err)
		}
		var found int = 0
		for _, pr := range pullRequests {
			if *pr.MergeCommitSHA == mergeCommitSha {
				prInfos = pr
				found = found + 1
			}
		}
		if found > 1 {
			return nil, nil, fmt.Errorf("found more than one pull request, this should not be possible. please open an issue with all log files")
		}
	}

	if prInfos == nil {
		return nil, nil, fmt.Errorf("no pull request found, please check if all resources are specified")
	}

	isCI, isCIBool := os.LookupEnv("CI")
	_, isSilentBool := os.LookupEnv("ACI_SILENT")
	if isCIBool && !isSilentBool {
		if *prInfos.State == "open" && isCI == "true" {
			err = CommentHelpToPullRequest(*prInfos.Number)
			if err != nil {
				log.Println(err)
			}
		}
	}

	prSHA := *prInfos.Head.SHA
	branchName := *prInfos.Head.Ref
	patchLevel := semver.ParsePatchLevel(branchName)

	var version = ""
	var latestVersion = ""
	// if an comment exists with aci_patch_level=major, make a major version!
	issueComments, err := GetIssueComments(prNumber)
	if err != nil {
		return nil, nil, err
	}

	for _, comment := range issueComments {
		// FIXME: access must be restricted but GITHUB_TOKEN doesn't get informations.
		// Refs: https://docs.github.com/en/rest/collaborators/collaborators#list-repository-collaborators
		// Refs: https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token

		// Must be a collaborator to have permission to create an override
		// isCollaborator, resp, err := GithubClient.Repositories.IsCollaborator(ctx, owner, repo, *comment.User.Login)
		// if err != nil {
		// 	return nil, nil, err
		// }
		// fmt.Println(resp.StatusCode)

		// if isCollaborator {
		if true {
			aciVersionOverride := regexp.MustCompile(`aci_version_override: ([0-9]+\.[0-9]+\.[0-9]+)`)
			aciPatchLevel := regexp.MustCompile(`aci_patch_level: ([a-zA-Z]+)`)

			if aciVersionOverride.MatchString(*comment.Body) {
				version = aciVersionOverride.FindStringSubmatch(*comment.Body)[1]
				break
			}

			if aciPatchLevel.MatchString(*comment.Body) {
				patchLevel = semver.ParsePatchLevel(aciPatchLevel.FindStringSubmatch(*comment.Body)[1])
				break
			}

		}
	}

	if version == "" {
		repositoryRelease, err := GetLatestReleaseVersion()
		if err == nil {
			latestVersion = *repositoryRelease.TagName
			version, err = semver.IncreaseVersion(patchLevel, latestVersion)
		} else {
			version, err = semver.IncreaseVersion(patchLevel, "0.0.0")
		}

		if err != nil {
			return nil, nil, err
		}
	}

	standardPrInfos = &StandardPrInfos{
		PrNumber:       prNumber,
		Owner:          owner,
		Repo:           repo,
		BranchName:     branchName,
		Sha:            prSHA,
		ShaShort:       prSHA[:8],
		PatchLevel:     patchLevel,
		LatestVersion:  latestVersion,
		CurrentVersion: "",
		NextVersion:    version,
		MergeCommitSha: *prInfos.MergeCommitSHA,
	}
	return
}
