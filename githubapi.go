package controllers

import (
	"fmt"
	"io/ioutil"
	github "github.com/google/go-github/github"
	oauth2 "golang.org/x/oauth2"
)

func GithubCommitFile(filepath, filename string) bool {
	//////////////////////////////////
	// Step 1: create client object //
	//////////////////////////////////
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "token-string"},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	owner := "git-user"
	repo := "repo-name"

	///////////////////////////////////
	// Step 2: get SHA-LATEST-COMMIT //
	///////////////////////////////////
	ref, _, err := client.Git.GetRef(owner, repo, "refs/heads/master")
	if err != nil {
		fmt.Println("Git.GetRef returned error:", err)
		return false
	}
	shaLastestCommit := ref.Object.SHA
	getCommit, _, errGetCommit := client.Git.GetCommit(owner, repo, *shaLastestCommit)
	if errGetCommit != nil {
		fmt.Println("Get Commit error", errGetCommit)
		return false
	}

	///////////////////////////////////////////////
	// Step 3: get SHA-LATEST-TREE of the commit //
	///////////////////////////////////////////////
	shaBaseTree := getCommit.Tree.SHA
	tree, _, errTree := client.Git.GetTree(owner, repo, *shaBaseTree, false)
	if errTree != nil {
		fmt.Println("Git.GetTree returned error:", errTree)
		return false
	}

	//////////////////////////////////////
	// Step 4: create blob for new file //
	//////////////////////////////////////
	var content string
	dat, err := ioutil.ReadFile(filepath)
    if err != nil {
    	fmt.Println("Reading file err", err)
    	return false
    }
  content = string(dat)
	inputBlob := &github.Blob{
		Content:  github.String(content),
		Encoding: github.String("utf-8"),
	}
	blob, _, err := client.Git.CreateBlob(owner, repo, inputBlob)

	//////////////////////////////////////////////////////////
	// Step 5: create new tree by adding the generated blob //
	//////////////////////////////////////////////////////////
	inputTree := []github.TreeEntry{
		{
			Path: github.String(filename),
			Mode: github.String("100644"),
			Type: github.String("blob"),
			SHA:  github.String(*blob.SHA),
		},
	}
	newTree, _, errNewTree := client.Git.CreateTree(owner, repo, *tree.SHA, inputTree)
	if errNewTree != nil {
		fmt.Println("Git.CreateCommit returned error:", errNewTree)
		return false
	} else {
		fmt.Println("New Tree::::", newTree)
	}

	shaNewTree := newTree.SHA

	/////////////////////////////////////////////////////////
	// Step 6: create new commit pointed to generated tree //
	/////////////////////////////////////////////////////////
	inputCommit := &github.Commit{
		Message: github.String("new commit"),
		Tree:    &github.Tree{SHA: github.String(*shaNewTree)},
		Parents: []github.Commit{{SHA: github.String(*shaLastestCommit)}},
	}
	commit, _, errCommit := client.Git.CreateCommit(owner, repo, inputCommit)
	if errCommit != nil {
		fmt.Println("Git.CreateCommit returned error:", errCommit)
		return false
	} else {
		fmt.Println("respone:::", commit)
	}

	///////////////////////////////
	// Step 7: Update branch Ref //
	///////////////////////////////
	updateRef, _, errUpdateRef := client.Git.UpdateRef(owner, repo, &github.Reference{
		Ref:    github.String("refs/heads/master"),
		Object: &github.GitObject{SHA: github.String(*commit.SHA)},
	}, true)
	if errUpdateRef != nil {
		fmt.Println("Git.UpdateRef returned error: ", errUpdateRef)
		return false
	} else {
		fmt.Println("ABC>>>>", updateRef)
	}
	return true
}
