package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
	"github.com/go-git/go-git/v5"
	gitHTTP "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/section/sectionctl/api"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitService interface provides a way to interact with Git
type GitService interface {
	UpdateGitViaGit(c *DeployCmd, response UploadResponse) error
}

// GS ...
type GS struct{}

// This is far less then ideal, however Kong does not seem to provide a way to inject dependencies into its commands so we must use this for testing
var globalGitService GitService = &GS{}

// UpdateGitViaGit clones the application repository to a temporary directory then updates it with the latest payload id and pushes a new commit
func (g *GS) UpdateGitViaGit(c *DeployCmd, response UploadResponse) error {
	app, err := api.Application(c.AccountID, c.AppID)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Cloning: https://aperture.section.io/account/%d/application/%d/%s.git ...\n", c.AccountID, c.AppID, app.ApplicationName)
	tempDir, err := ioutil.TempDir("", "sectinoctl-*")
	if err != nil {
		return err
	}
	log.Println("[Debug] tempDir: ", tempDir)
	// Git objects storer based on memory
	gitAuth := &gitHTTP.BasicAuth{
		Username: "section-token", // yes, this can be anything except an empty string
		Password: api.Token,
	}
	payload := PayloadValue{ID: response.PayloadID}
	branchRef := fmt.Sprintf("refs/heads/%s",c.Branch)
	r, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:      fmt.Sprintf("https://aperture.section.io/account/%d/application/%d/%s.git", c.AccountID, c.AppID, app.ApplicationName),
		Auth:     gitAuth,
		Progress: os.Stdout,
		ReferenceName: plumbing.ReferenceName(branchRef),
	})
	if err != nil {
		return err
	}
	// ... retrieving the branch being pointed by HEAD
	ref, err := r.Head()
	if err != nil {
		return err
	}
	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return err
	}
	log.Println("[DEBUG] HEAD commit: ", commit)
	// ... retrieve the tree from the commit
	tree, err := commit.Tree()
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	f, err := tree.File(c.AppPath + "/.section-external-source.json")
	if err != nil {
		return err
	}
	srcContent := PayloadValue{}
	content, err := f.Contents()
	if err != nil {
		return nil
	}
	err = json.Unmarshal([]byte(content), &srcContent)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}
	ct, err := f.Contents()
	if err != nil {
		return fmt.Errorf("couldn't open contents of file: %w", err)
	}
	log.Println("[DEBUG] Old external source contents: ", ct)
	log.Println("[DEBUG] expected new tarball UUID: ", response.PayloadID)
	srcContent.ID = payload.ID
	pl, err := json.MarshalIndent(srcContent, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(tempDir, c.AppPath+"/.section-external-source.json"), pl, 0644)
	if err != nil {
		return err
	}
	_, err = w.Add(c.AppPath + "/.section-external-source.json")
	if err != nil {
		return err
	}

	status, err := w.Status()
	if err != nil {
		return err
	}
	log.Println("[DEBUG] git status: ", status)
	_, err = w.Add(c.AppPath + "/.section-external-source.json")
	if err != nil {
		return err
	}
	commitHash, err := w.Commit("[sectionctl] updated nodejs/.section-external-source.json with new deployment.", &git.CommitOptions{Author: &object.Signature{
		Name:  "sectionctl",
		Email: "noreply@section.io",
		When:  time.Now(),
	}})
	if err != nil {
		return fmt.Errorf("failed to make a commit on the temporary repository: %w", err)
	}
	cmt, err := r.CommitObject(commitHash)
	if err != nil {
		return fmt.Errorf("failed to get commit object: %w", err)
	}
	log.Println("[DEBUG] New Commit: ", cmt.String())
	newTree, err := cmt.Tree()
	if err != nil {
		return err
	}
	newF, err := newTree.File(c.AppPath + "/.section-external-source.json")
	if err != nil {
		return err
	}

	ctt, err := newF.Contents()
	if err != nil {
		return fmt.Errorf("could not open conetnts of new file in git: %w", err)
	}
	log.Println("[DEBUG] contents in new commit: ", ctt)
	err = r.Push(&git.PushOptions{Auth: gitAuth, Progress: os.Stdout})
	if err != nil {
		return fmt.Errorf("failed to push git changes: %w", err)
	}

	return nil
}
