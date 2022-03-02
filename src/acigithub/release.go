package acigithub

import (
	"awesome-ci/src/tools"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v39/github"
)

// CreateRelease
func CreateRelease(version string, releaseBranch string, body string, draft bool) (createdRelease *github.RepositoryRelease, err error) {
	if !isgithubRepository {
		log.Fatalln("make shure the GITHUB_REPOSITORY is available!")
	}
	owner, repo := tools.DevideOwnerAndRepo(githubRepository)

	relName := "Release " + version
	if releaseBranch == "" {
		releaseBranch = tools.GetDefaultBranch()
	}

	// get body for release
	if body != "" {
		bodyFile, err := tools.CheckIsFile(body)
		if err == nil {
			body = bodyFile
		}
	}

	releaseObject := github.RepositoryRelease{
		TargetCommitish: &releaseBranch,
		TagName:         &version,
		Name:            &relName,
		Draft:           &draft,
		Body:            &body,
	}
	createdRelease, _, err = GithubClient.Repositories.CreateRelease(
		ctx,
		owner,
		repo,
		&releaseObject)
	if err != nil {
		err = fmt.Errorf("error at creating github release: %v", err)
		return
	}

	envVars, err := OpenEnvFile()
	if err != nil {
		return nil, err
	}
	envVars.Set("ACI_RELEASE_ID", fmt.Sprintf("%d", *createdRelease.ID))
	err = envVars.SaveEnvFile()
	return
}

// PublishRelease
func PublishRelease(version string, releaseBranch string, body string, releaseId int64, uploadArtifacts *string) (err error) {
	draftFalse := false
	if !isgithubRepository {
		log.Fatalln("make shure the GITHUB_REPOSITORY is available!")
	}
	owner, repo := tools.DevideOwnerAndRepo(githubRepository)

	if releaseId == 0 {
		releaseIdStr, releaseIdBool := os.LookupEnv("ACI_RELEASE_ID")
		if !releaseIdBool {
			log.Println("No release found, creating one...")
			release, err := CreateRelease(version, releaseBranch, body, true)
			if err != nil {
				log.Fatalln(err)
			}
			releaseId = *release.ID
		} else {
			releaseId, err = strconv.ParseInt(releaseIdStr, 10, 64)
			if err != nil {
				fmt.Printf("%s of type %T", releaseIdStr, releaseIdStr)
				os.Exit(2)
			}
		}
	}

	existingRelease, _, err := GithubClient.Repositories.GetRelease(
		ctx,
		owner,
		repo,
		releaseId)
	if err != nil {
		return err
	}

	// upload any given artifacts
	var releaseBodyAssets string = ""
	if *uploadArtifacts != "" {
		filesAndInfos, err := tools.GetAsstes(uploadArtifacts, false)
		if err != nil {
			return err
		}

		releaseBodyAssets = "### Asstes\n"

		for i, fileAndInfo := range filesAndInfos {
			fmt.Printf("uploading %s as asset to release\n", fileAndInfo.Name)
			// Upload assets to GitHub Release
			relAsset, _, err := GithubClient.Repositories.UploadReleaseAsset(
				ctx,
				owner,
				repo,
				releaseId,
				&github.UploadOptions{
					Name: fileAndInfo.Name,
				},
				&fileAndInfo.File)
			if err != nil {
				log.Println("error at uploading asset to release: ", err)
			} else {
				// add asset to release body
				releaseBodyAssets = fmt.Sprintf("%s\n- [%s](%s) `%s`\n  Sha256: `%x`", releaseBodyAssets, fileAndInfo.Name, *relAsset.BrowserDownloadURL, fileAndInfo.Infos.ModTime().Format(time.RFC3339), fileAndInfo.Hash)

				// export Download URL to env. See: #53
				envVars, err := OpenEnvFile()
				if err != nil {
					log.Println("could open envs:", err)
				}
				envVars.Set(fmt.Sprintf("ACI_ARTIFACT_%d_URL", i+1), *relAsset.BrowserDownloadURL)
				err = envVars.SaveEnvFile()
				if err != nil {
					log.Println("could not export atrifact url:", err)
				}
			}
		}
	}

	newReleaseBody := fmt.Sprintf("%s\n\n%s", *existingRelease.Body, releaseBodyAssets)
	existingRelease.Body = &newReleaseBody

	// publishing release
	*existingRelease.Draft = draftFalse
	_, _, err = GithubClient.Repositories.EditRelease(
		ctx,
		owner,
		repo,
		releaseId,
		existingRelease)
	if err != nil {
		return err
	}

	return
}

// GetLatestReleaseVersion
func GetLatestReleaseVersion(owner string, repo string) (latestRelease *github.RepositoryRelease, err error) {

	var releaseMap = make(map[string]*github.RepositoryRelease)

	var listOptions github.ListOptions = github.ListOptions{
		PerPage: 100,
		Page:    0,
	}

	for listOptions.Page >= 0 {
		releases, response, err := GithubClient.Repositories.ListReleases(ctx, owner, repo, &listOptions)

		if err != nil {
			return nil, err
		}

		for _, release := range releases {
			releaseMap[*release.TagName] = release
			fmt.Printf("Release %s %s\n", *release.TargetCommitish, *release.TagName)
		}

		if listOptions.Page == response.NextPage {
			break
		}

		listOptions.Page = response.NextPage
	}

	return findLatestRelease(`/Users/tgr/workspace/daimler/awesome-ci`, releaseMap)
}

func findLatestRelease(directory string, githubReleaseMap map[string]*github.RepositoryRelease) (latestRelease *github.RepositoryRelease, err error) {
	gitRepo, err := git.PlainOpen(directory)

	if err != nil {
		return nil, err
	}

	tagMap, err := getGitTagMap(gitRepo)

	if err != nil {
		return nil, err
	}

	headRef, _ := gitRepo.Head()

	iter, _ := gitRepo.Log(&git.LogOptions{
		From:  headRef.Hash(),
		Order: git.LogOrderCommitterTime,
	})

	var commit *object.Commit
	for commit, err = iter.Next(); commit != nil && err == nil; commit, err = iter.Next() {
		fmt.Printf("Lookup %s ", commit.Hash.String())
		if tagName, found := tagMap[commit.Hash.String()]; found {
			fmt.Println(tagName)
			if latestRelease, found := githubReleaseMap[tagName]; found {
				return latestRelease, nil
			}
		}

	}

	return nil, errors.New("could not find latest release")
}

func getGitTagMap(gitRepo *git.Repository) (tagMap map[string]string, err error) {
	tagMap = make(map[string]string)

	tags, err := gitRepo.Tags()

	if err != nil {
		return nil, err
	}

	tags.ForEach(func(r *plumbing.Reference) error {
		tagMap[r.Hash().String()] = r.Name().Short()
		fmt.Printf("Tag %s %s\n", r.Hash().String(), r.Name().Short())
		return nil
	})

	return tagMap, nil
}
