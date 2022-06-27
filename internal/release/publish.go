package release

import (
	"fmt"

	"fullstack-devops/awesome-ci/pkg/acigithub"
	"fullstack-devops/awesome-ci/pkg/semver"
	"fullstack-devops/awesome-ci/pkg/utils"
	"fullstack-devops/awesome-ci/src/tools"
)

func RunPublish(prNumber int, mergeSha, releaseBranch, version, body, patchLevel, assets string, releaseID int64, dryRun, hotfix bool) error {
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

	} else if releaseID == 0 {
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

	if assets != "" {
		_, err = tools.GetAsstes(&assets, true)
		if err != nil {
			return err
		}
	}

	if dryRun {
		fmt.Printf("Would publishing release: %s\n", nextVersion)
	} else {
		fmt.Printf("Publishing release: %s - %d\n", nextVersion, releaseID)
		err = acigithub.PublishRelease(nextVersion, releaseBranch, body, releaseID, &assets)
		if err != nil {
			return err
		}
	}

	return nil
}
