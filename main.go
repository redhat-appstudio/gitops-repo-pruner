//
// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

var appDataOrg = "redhat-appstudio-appdata"

func main() {
	if os.Getenv("GITHUB_TOKEN") == "" {
		fmt.Println("GITHUB_TOKEN must be set as an environment variable")
	}
	githubToken := os.Getenv("GITHUB_TOKEN")

	// Initialize an authenticated github client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	authClient := github.NewClient(tc)

	// Parse command line flags to determine which operation to perform
	// Options are:
	// Delete repository matching a given name
	// Delete all invalid repositories (ones starting with a dash)
	// Listing all repositories
	var operation, keyword, repo string
	flag.StringVar(&operation, "operation", "", "The operation to perform. One of: delete-repo or delete-invalid")
	flag.StringVar(&keyword, "keyword", "", "The keyword(s) to match gitops repositories on")
	flag.StringVar(&repo, "repo", "", "The name of a repository")
	flag.Parse()

	// Check the values of the flags before proceding
	if operation == "" {
		log.Fatal("usage: --operation must be set as a command-line flag")
	}

	if operation != "delete-invalid" && operation != "list-all" && operation != "delete-repo" {
		log.Fatal("usage: The only valid options for '--operation' are delete-repo, delete-invalid, or list-all")
	}

	if operation == "delete-invalid" {
		invalidRepos, err := listInvalidRepos(ctx, authClient)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = deleteRepos(ctx, authClient, invalidRepos)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else if operation == "list-all" {
		allRepos, err := listAllRepos(ctx, authClient)
		if err != nil {
			log.Fatal(err.Error())
		}
		for _, repo := range allRepos {
			fmt.Println(*repo.Name)
		}

	} else {
		if repo == "" {
			log.Fatal("usage: --repo <repo-name> must be passed in as a flag when using the 'delete' operation")
		}
		err := deleteRepo(ctx, authClient, repo)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

}

// listAllRepos returns the list of all github repositories in the org
func listAllRepos(ctx context.Context, client *github.Client) ([]*github.Repository, error) {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, appDataOrg, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos, nil

}

// listInvalidRepos return any gitops repositories in the org starting with '-'
// Created due to a bug in HAS. These should no longer get created
func listInvalidRepos(ctx context.Context, client *github.Client) ([]*github.Repository, error) {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allRepos []*github.Repository
	count := 0
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, appDataOrg, opt)
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			repoName := *repo.Name

			if repoName[0:1] == "-" {
				allRepos = append(allRepos, repo)
			}
		}

		// ToDo: Cleanup
		// There's a lot (> 10k) of invalid repos right now. Limit to 1000 returned to avoid rate limiting the GitHub token
		count++
		if count == 40 {
			break
		}
		// ToDo: cleanup
		// By default it seems go-github returns the oldest repositories first, rather than newest. So after we get the first set of results,
		// navigate to the "last page" (the newest repositories) and move backwards
		// This is because most of the invalid repositories are front loaded (i.e. newer) so we don't want to waste API calls going through
		// pages of old, valid repositories
		if count == 1 {
			//fmt.Printf("Last page: %d\n", resp.LastPage)
			opt.Page = resp.LastPage
		} else {
			//fmt.Printf("Prev page: %d\n", resp.PrevPage)
			opt.Page = resp.PrevPage
		}

		if opt.Page == 0 {
			break
		}

	}
	return allRepos, nil
}

// deleteRepos takes in a list of Git repository objects and deletes each one using go-github
func deleteRepos(ctx context.Context, client *github.Client, repos []*github.Repository) error {
	for _, repo := range repos {
		err := deleteRepo(ctx, client, *repo.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

// deleteRepo deletes the given git repository
func deleteRepo(ctx context.Context, client *github.Client, repo string) error {
	fmt.Println("Deleting repo: " + repo)
	_, err := client.Repositories.Delete(ctx, appDataOrg, repo)
	if err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}
