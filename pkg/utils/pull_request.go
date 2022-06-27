package utils

import (
	"errors"
	"log"
	"regexp"
	"strconv"
)

func EvalPrNumber(override *int) (err error) {
	if *override != 0 {
		return nil
	}

	*override, err = GetPrFromMergeMessage()
	if err != nil {
		return err
	}
	return
}

func GetPrFromMergeMessage() (pr int, err error) {
	regex := `.*#([0-9]+).*`
	r := regexp.MustCompile(regex)

	mergeMessage := r.FindStringSubmatch(CMD(`git log -1 --pretty=format:"%s"`, true))
	if len(mergeMessage) > 1 {
		return strconv.Atoi(mergeMessage[1])
	} else {
		log.Printf("no pull-request in merge message found, make shure you match the regex %s \n"+
			"Example: Merge pull request #3 from some-orga/feature/awesome-feature"+
			"Alternativly provide the PR-Number by adding the argument -number <int>", regex)

		return 0, errors.New("no pull-request in merge message found")
	}
}
