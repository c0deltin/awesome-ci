package pr

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"fullstack-devops/awesome-ci/pkg/acigithub"
	"fullstack-devops/awesome-ci/pkg/utils"
)

func RunInfo(number int, format string) error {
	_, err := acigithub.NewGitHubClient()
	if err != nil {
		return err
	}

	err = utils.EvalPrNumber(&number)
	if err != nil {
		return err
	}

	prInfos, _, err := acigithub.GetPrInfos(number, "")
	if err != nil {
		return err
	}

	errEnvs := standardPrInfosToEnv(prInfos)
	if format != "" {
		replacer := strings.NewReplacer(
			"pr", fmt.Sprint(prInfos.PrNumber),
			"version", prInfos.NextVersion,
			"latest_version", prInfos.LatestVersion,
			"patchLevel", string(prInfos.PatchLevel))
		output := replacer.Replace(format)
		fmt.Print(output)
	} else {
		fmt.Println("#### Info output:")
		fmt.Printf("Pull Request: %d\n", prInfos.PrNumber)
		fmt.Printf("Latest release version: %s\n", prInfos.LatestVersion)
		fmt.Printf("Patch level: %s\n", prInfos.PatchLevel)
		fmt.Printf("Possible new release version: %s\n", prInfos.NextVersion)
		if errEnvs != nil {
			return errEnvs
		}
	}

	return nil
}

func standardPrInfosToEnv(prInfos *acigithub.StandardPrInfos) (err error) {
	runnerType := "github_runner"
	switch runnerType {
	case "github_runner":
		envVars, err := acigithub.OpenEnvFile()
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
