package release

import (
	"fmt"

	"fullstack-devops/awesome-ci/pkg/acigithub"
	"fullstack-devops/awesome-ci/pkg/semver"
	"fullstack-devops/awesome-ci/pkg/utils"
)

func RunCreate(prNumber int, mergeSha, releaseBranch, version, body, patchLevel string, dryRun, hotfix bool) error {
	var nextVersion string
	_, err := acigithub.NewGitHubClient()
	if err != nil {
		return err
	}

	if version != "" && patchLevel != "" {
		parsedPatchLevel := semver.ParsePatchLevel(patchLevel)
		nextVersion, err = semver.IncreaseVersion(parsedPatchLevel, version)
		if err != nil {
			return err
		}
	} else if version != "" && patchLevel == "" {
		nextVersion = version
	} else if hotfix {
		release, err := acigithub.GetLatestReleaseVersion()
		if err != nil {
			return err
		}

		nextVersion, err = semver.IncreaseVersion(semver.Bugfix, *release.TagName)
		if err != nil {
			return err
		}

	} else {
		// if no merge commit sha is provided, the pull request number should either be specified or evaluated from the merge message (fallback)
		if mergeSha == "" {
			err := utils.EvalPrNumber(&prNumber)
			if err != nil {
				return err
			}
		}

		prInfos, _, err := acigithub.GetPrInfos(prNumber, mergeSha)
		if err != nil {
			return err
		}

		nextVersion = prInfos.NextVersion
		if errEnvs := acigithub.StandardPrInfosToEnv(prInfos); errEnvs != nil {
			return errEnvs
		}
	}

	if dryRun {
		fmt.Printf("Would create new release with version: %s\n", nextVersion)
	} else {
		fmt.Printf("Writing new release: %s\n", nextVersion)
		createdRelease, err := acigithub.CreateRelease(nextVersion, releaseBranch, body, true)
		if err != nil {
			return err
		}
		fmt.Println("Create release successful. ID:", *createdRelease.ID)

		return nil
	}

	return nil
}
