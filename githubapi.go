package controllers

import (
	//"os"
	"fmt"
	"io/ioutil"
	github "github.com/google/go-github/github"
	oauth2 "golang.org/x/oauth2"
)

func GithubCommitFile(filepath, filename string) bool {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "fc1d0207196310539387a9ee435f3008a61459ce"},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	owner := "chanhlv93"
	repo := "containerization_artifact"
	// blank := "master"

	// get SHA-LATEST-COMMIT
	ref, _, err := client.Git.GetRef(owner, repo, "refs/heads/master")
	if err != nil {
		fmt.Println("Git.GetRef returned error:", err)
		return false
	}
	shaLastestCommit := ref.Object.SHA
	//fmt.Println("shaLastestCommit:", *shaLastestCommit)

	// get latstest commit and then get SHA-BASE-TREE
	getCommit, _, errGetCommit := client.Git.GetCommit(owner, repo, *shaLastestCommit)
	if errGetCommit != nil {
		fmt.Println("Get Commit error", errGetCommit)
		return false
	}
	shaBaseTree := getCommit.Tree.SHA
	//fmt.Println("shaBaseTree:", *shaBaseTree)

	// get SHA tree and set base_tree to the SHA-BASE-TREE
	tree, _, errTree := client.Git.GetTree(owner, repo, *shaBaseTree, false)
	if errTree != nil {
		fmt.Println("Git.GetTree returned error:", errTree)
		return false
	}
	//fmt.Println("Path:", *tree.Entries[2].Path)
	//path := "D:\\docker\\docker-counter-long-running\\README.md"
	var content string
	/*f, _ := os.Open(filepath)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		content += line
	}*/

	dat, err := ioutil.ReadFile(filepath)
    if err != nil {
    	fmt.Println("Reading file err", err)
    	return false
    }
    content = string(dat)
	//fmt.Println(content)
	//create blob
	inputBlob := &github.Blob{
		// SHA:      String("s"),
		Content:  github.String(content),
		Encoding: github.String("utf-8"),
		// Size:     Int(12),
	}
	blob, _, err := client.Git.CreateBlob(owner, repo, inputBlob)

	inputTree := []github.TreeEntry{
		{
			Path: github.String(filename),
			Mode: github.String("100644"),
			Type: github.String("blob"),
			SHA:  github.String(*blob.SHA),
		},
	}

	//create Tree
	newTree, _, errNewTree := client.Git.CreateTree(owner, repo, *tree.SHA, inputTree)
	if errNewTree != nil {
		fmt.Println("Git.CreateCommit returned error:", errNewTree)
		return false
	} else {
		fmt.Println("New Tree::::", newTree)
	}

	shaNewTree := newTree.SHA
	// commit Comment
	inputCommit := &github.Commit{
		Message: github.String("Vinhvdq commited"),
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

	fmt.Println("SHA Refs:", *commit.SHA)
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