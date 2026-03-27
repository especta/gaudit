// Copyright IBM Corp. 2019, 2020
// SPDX-License-Identifier: MIT

package analyze

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v25/github"
	"github.com/hashicorp/gaudit/config"
	"github.com/hashicorp/gaudit/state"
	"golang.org/x/oauth2"
)

func Run(options config.Options, audit state.Audit, rules []Rule) error {

	// github client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: options.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// check audit results
	if audit.Results == nil {
		audit.Results = make(map[string]state.Result)
	}

	// loop each repo, for each rule
	for _, k := range audit.Index {
		repo := audit.Repos[k]

		fmt.Println(repo.FullName)
		result := state.Result{}
		repoFiles := []string{}
		repoFilesLoaded := false
		repoFilesLoadErr := error(nil)

		for _, rule := range rules {

			if rule.Type == "" || (rule.Type == "public" && !repo.Private) || (rule.Type == "private" && repo.Private) {

				if hasGlobPattern(rule.Resource) && !repoFilesLoaded && repoFilesLoadErr == nil {
					repoFiles, repoFilesLoadErr = loadRepoFiles(ctx, client, repo)
					repoFilesLoaded = true
				}

				if hasGlobPattern(rule.Resource) && repoFilesLoadErr != nil {
					result.Rules = append(result.Rules, state.Rule{
						Name:    rule.Name,
						Status:  "error",
						Details: []string{fmt.Sprintf("unable to list repository files for glob matching: %s", repoFilesLoadErr.Error())},
					})
					continue
				}

				result.Rules = append(result.Rules, evaluateRule(ctx, client, repo, rule, repoFiles))
			} else {
				result.Rules = append(result.Rules, state.Rule{
					Name:   rule.Name,
					Status: "na",
				})
			}
		}

		// save the result for this repo
		audit.Results[repo.FullName] = result
		err := state.Save(options, audit)
		if err != nil {
			fmt.Println("ERROR: " + err.Error())
		}

	}

	return nil

}

func evaluateRule(ctx context.Context, client *github.Client, repo state.Repo, rule Rule, repoFiles []string) state.Rule {

	result := state.Rule{
		Name: rule.Name,
	}

	resource := normalizeResourcePath(rule.Resource)
	if resource == "" {
		result.Status = "error"
		result.Details = []string{"invalid empty resource"}
		return result
	}

	matchedResources := []string{}
	if hasGlobPattern(rule.Resource) {
		var err error
		matchedResources, err = filterResourcesByPattern(resource, repoFiles)
		if err != nil {
			result.Status = "error"
			result.Details = []string{fmt.Sprintf("invalid glob pattern %q: %s", resource, err.Error())}
			return result
		}
	} else {
		matchedResources = []string{resource}
	}

	action := strings.ToLower(rule.Action)
	if action == "exists" {
		return evaluateExistsRule(ctx, client, repo, result, hasGlobPattern(rule.Resource), resource, matchedResources)
	}

	if action == "not_exists" {
		return evaluateNotExistsRule(ctx, client, repo, result, hasGlobPattern(rule.Resource), resource, matchedResources)
	}

	if action == "contains" {
		return evaluateContainsRule(ctx, client, repo, result, resource, rule.Match, matchedResources)
	}

	result.Status = "error"
	result.Details = []string{fmt.Sprintf("invalid action %q", rule.Action)}
	return result
}

func evaluateExistsRule(ctx context.Context, client *github.Client, repo state.Repo, ruleResult state.Rule, isGlob bool, pattern string, resources []string) state.Rule {
	failedResources := []string{}

	if isGlob {
		if len(resources) == 0 {
			failedResources = append(failedResources, fmt.Sprintf("no resources matched %q", pattern))
		}
	} else {
		for _, resource := range resources {
			_, _, _, err := client.Repositories.GetContents(ctx, repo.Owner, repo.Name, resource, nil)
			if err != nil {
				failedResources = append(failedResources, resource)
			}
		}
	}

	if len(failedResources) > 0 {
		ruleResult.Status = "error"
		ruleResult.Details = failedResources
		return ruleResult
	}

	ruleResult.Status = "success"
	return ruleResult
}

func evaluateNotExistsRule(ctx context.Context, client *github.Client, repo state.Repo, ruleResult state.Rule, isGlob bool, pattern string, resources []string) state.Rule {
	failedResources := []string{}

	if isGlob {
		for _, resource := range resources {
			failedResources = append(failedResources, resource)
		}
	} else {
		for _, resource := range resources {
			_, _, _, err := client.Repositories.GetContents(ctx, repo.Owner, repo.Name, resource, nil)
			if err == nil {
				failedResources = append(failedResources, resource)
			}
		}
	}

	if len(failedResources) > 0 {
		ruleResult.Status = "error"
		ruleResult.Details = failedResources
		return ruleResult
	}

	ruleResult.Status = "success"
	if isGlob && len(resources) == 0 {
		ruleResult.Details = []string{fmt.Sprintf("no resources matched %q", pattern)}
	}
	return ruleResult
}

func evaluateContainsRule(ctx context.Context, client *github.Client, repo state.Repo, ruleResult state.Rule, pattern string, match string, resources []string) state.Rule {
	failedResources := []string{}

	if len(resources) == 0 {
		failedResources = append(failedResources, fmt.Sprintf("no resources matched %q", pattern))
	}

	for _, resource := range resources {
		content, err := resourceContent(ctx, client, repo, resource)
		if err != nil {
			failedResources = append(failedResources, fmt.Sprintf("%s (read failed: %s)", resource, err.Error()))
			continue
		}

		if !strings.Contains(content, match) {
			failedResources = append(failedResources, resource)
		}
	}

	if len(failedResources) > 0 {
		ruleResult.Status = "error"
		ruleResult.Details = failedResources
		return ruleResult
	}

	ruleResult.Status = "success"
	return ruleResult
}

func loadRepoFiles(ctx context.Context, client *github.Client, repo state.Repo) ([]string, error) {
	ref := repo.DefaultBranch
	if ref == "" {
		return nil, fmt.Errorf("repository %q does not define a default branch", repo.FullName)
	}

	branch, _, err := client.Repositories.GetBranch(ctx, repo.Owner, repo.Name, ref)
	if err != nil {
		return nil, err
	}

	branchCommit := branch.GetCommit()
	if branchCommit == nil || branchCommit.GetSHA() == "" {
		return nil, fmt.Errorf("unable to resolve commit SHA for branch %q", ref)
	}

	commit, _, err := client.Git.GetCommit(ctx, repo.Owner, repo.Name, branchCommit.GetSHA())
	if err != nil {
		return nil, err
	}
	if commit.GetTree() == nil || commit.GetTree().GetSHA() == "" {
		return nil, fmt.Errorf("unable to resolve tree SHA for branch %q", ref)
	}

	tree, _, err := client.Git.GetTree(ctx, repo.Owner, repo.Name, commit.GetTree().GetSHA(), true)
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return []string{}, nil
	}

	files := []string{}
	for _, entry := range tree.Entries {
		if entry.GetType() == "blob" {
			files = append(files, normalizeResourcePath(entry.GetPath()))
		}
	}

	return files, nil
}

func resourceContent(ctx context.Context, client *github.Client, repo state.Repo, resource string) (string, error) {
	resp, _, _, err := client.Repositories.GetContents(ctx, repo.Owner, repo.Name, resource, nil)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("resource %q is not a file", resource)
	}

	content, err := resp.GetContent()
	if err == nil {
		return content, nil
	}

	// fallback for unexpected encodings or API responses
	if resp.Content != nil {
		return *resp.Content, nil
	}

	return "", err
}
