// get-build-number project main.go
package main

// This hands out build numbers roughly following the scheme described at
// https://tidbits.com/2020/07/08/how-to-decode-apple-version-and-build-numbers/

// To test, run with
// CIRRUS_CHANGE_IN_REPO=23456789 CIRRUS_BRANCH=foo CIRRUS_DEFAULT_BRANCH=bar RELEASE_ID_FOR_STORAGE=33980128 CIRRUS_REPO_NAME=ISO CIRRUS_REPO_OWNER=helloSystem BRANCH=bar GITHUB_TOKEN=...

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"
	"context"
	"os"
	"fmt"
	"sort"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func mapToJsonString(m map[string]int) string {
	j, _ := json.Marshal(m)
	jsonString := string(j)
	return (jsonString)
}

func jsonStringToMap(jsonString string) map[string]int {
	var result map[string]int
	json.Unmarshal([]byte(jsonString), &result)
	return (result)
}

func letterToNumber(letter string) string {
	char := []rune(letter)[0] // rune, not string
	ascii := int(char) - 64
	return (fmt.Sprint(ascii))
}

func numberToLetter(number int) string {
	asciiNum := number // Uppercase A
	character := string(asciiNum + 64)
	return (string(fmt.Sprint(character)))
}

func main() {

	// Ensure that we have all required environment variables
	vars := []string{"BRANCH", "GITHUB_TOKEN", "CIRRUS_REPO_OWNER", "CIRRUS_REPO_NAME", "RELEASE_ID_FOR_STORAGE", "CIRRUS_DEFAULT_BRANCH", "CIRRUS_CHANGE_IN_REPO"}
	for _, v := range vars {
		vv := os.Getenv(v)
		if len(vv) == 0 {
			log.Println("Missing " + v + " environment variable; not running on Cirrus CI?")
			os.Exit(1)
		}
	}

	branch := os.Getenv("BRANCH")
	token := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("CIRRUS_REPO_OWNER")
	reponame := os.Getenv("CIRRUS_REPO_NAME")
	var releaseId int64
	n, err := strconv.ParseInt(os.Getenv("RELEASE_ID_FOR_STORAGE"), 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	releaseId = n // 33980128 - the release body ("description") of this GitHub Release ID will be used as Storage

	// If this build was triggered by a PR, then return the Git SHA
	if os.Getenv("CIRRUS_PR") != "" {
		fmt.Println(os.Getenv("CIRRUS_CHANGE_IN_REPO"))
		os.Exit(0)
	}

	// Check if we are at the default branch. For the default branch it is not needed that the name starts
	// with something like 0A... (called "buildtrainAndMinor"),
	// so that we can give it a different name, such as "experimental".
	// This is so that everyone immediately sees what is being offered
	weAreOnDefaultBranch := false
	if os.Getenv("BRANCH") == os.Getenv("CIRRUS_DEFAULT_BRANCH") {
		weAreOnDefaultBranch = true
	}

	// Check if the BRANCH environment variable starts with 0A... (in our example), or something like it.
	// If no and if we are not on the default branch then return the Git SHA
	r := regexp.MustCompile(`^([0-9]*)([A-Z]*)`)
	matches := r.FindStringSubmatch(branch)
	if (len(matches) != 3 || matches[1] == "" || matches[2] == "") && weAreOnDefaultBranch == false {
		fmt.Println(os.Getenv("CIRRUS_CHANGE_IN_REPO"))
		os.Exit(0)
	}

	// If yes, then get the last build number from the Storage, increment it by one, and return it as part of OAxxx
	// If the Storage does not have a build number for 0A..., then start with 1
	// If we are on the default branch, then use the highest 0A... from Storage we can find.

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// Load stored build numbers from Storage
	var buildNumbersMap map[string]int
	release, _, err := client.Repositories.GetRelease(ctx, owner, reponame, releaseId)
	if err != nil {
		log.Println(err)
	}
	buildNumbersMap = jsonStringToMap(release.GetBody())

	var buildtrain string
	var minor string

	if weAreOnDefaultBranch == false {
		buildtrain = matches[1]
		minor = matches[2]
	} else {
		var keys []string
		for k := range buildNumbersMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buildtrainAndMinor := keys[len(keys)-1]
		r := regexp.MustCompile(`^([0-9]*)([A-Z]*)`)
		matches := r.FindStringSubmatch(buildtrainAndMinor)
		if (len(matches) == 3 && ( matches[1] != "" || matches[2] != "")) {
			buildtrain = matches[1]
			minor = matches[2]
		} else {
			log.Fatal("Could not get the highest existing buildtrainAndMinor")
		}
	}

	var buildNumberForThisBuildtrainAndMinor int
	if val, ok := buildNumbersMap[buildtrain+minor]; ok {
		buildNumberForThisBuildtrainAndMinor = val + 1
		buildNumbersMap[buildtrain+minor] = buildNumberForThisBuildtrainAndMinor
	} else {
		// Handle the case when there is no entry in the JSON for this key (buildtrain + letter),
		// in this case start with 1
		buildNumberForThisBuildtrainAndMinor = 1
		buildNumbersMap[buildtrain+minor] = buildNumberForThisBuildtrainAndMinor
	}

	fmt.Printf("%#s%#s%#v\n", buildtrain, minor, buildNumberForThisBuildtrainAndMinor)

	// Save the new JSON string in Storage
	buildNumbersJsonString := mapToJsonString(buildNumbersMap)
	// fmt.Println(buildNumbersJsonString)
	release.Body = github.String(buildNumbersJsonString)
	client.Repositories.EditRelease(ctx, owner,reponame,releaseId,release)

}
