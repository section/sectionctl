package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitHTTP "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/section/sectionctl/api"
)

// GitService interface provides a way to interact with Git
type GitService interface {
	UpdateGitViaGit(ctx context.Context, c *DeployCmd, response UploadResponse) error
}

// GS ...
type GS struct{}

// This is far less then ideal, however Kong does not seem to provide a way to inject dependencies into its commands so we must use this for testing
var globalGitService GitService = &GS{}

// UpdateGitViaGit clones the application repository to a temporary directory then updates it with the latest payload id and pushes a new commit
func (g *GS) UpdateGitViaGit(ctx context.Context, c *DeployCmd, response UploadResponse) error {
	app, err := api.Application(c.AccountID, c.AppID)
	if err != nil {
		return err
	}
	appName := strings.ReplaceAll(app.ApplicationName, "/", "")
	cloneDir := fmt.Sprintf("https://aperture.section.io/account/%d/application/%d/%s.git", c.AccountID, c.AppID, appName)
	log.Printf("[INFO] Begin updating hash in .section-external-source.json:\n\tsection-configmap-tars/%v/%s.tar.gz\n",c.AccountID,response.PayloadID)
	tempDir, err := ioutil.TempDir("", "sectionctl-*")
	if err != nil {
		return err
	}
	log.Println("[DEBUG] tempDir: ", tempDir)
	// Git objects storer based on memory
	gitAuth := &gitHTTP.BasicAuth{
		Username: "section-token", // yes, this can be anything except an empty string
		Password: api.Token,
	}
	payload := PayloadValue{ID: response.PayloadID}
	branchRef := fmt.Sprintf("refs/heads/%s",c.Environment)
	var r *git.Repository
	var progressOutput io.Writer;
	if IsInCtxBool(ctx, "quiet"){
		progressOutput = log.Writer()
	}else{
		progressOutput = log.Writer()
	}
	log.Println("[INFO] cloning application configuration repo to ",tempDir)
	r, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:      cloneDir,
		Auth:     gitAuth,
		Progress: progressOutput,
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
		return fmt.Errorf("couldn't open contents of file: %w", err)
	}
	err = json.Unmarshal([]byte(content), &srcContent)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}
	log.Println("[DEBUG] Old external source contents: ", content)
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
		return fmt.Errorf("could not open contents of new file in git: %w", err)
	}
	log.Println("[DEBUG] contents in new commit: ", ctt)
	
	configFile, err := tree.File("section.config.json")
	if err != nil {
		log.Println("[ERROR] unable to open section.config.json which is used to log the image name and version: %w",err)
	}
	sectionConfigContents, err := configFile.Contents()
	if err != nil {
		log.Println("[ERROR] unable to open section.config.json which is used to log the image name and version: %w" ,err)
	}
	sectionConfig, err := api.ParseSectionConfig(sectionConfigContents)
	if err != nil{
		log.Println("[ERROR] There was an issue reading the section.config.json:",err)
	}
	// if err := json.Unmarshal(sectionConfigContent.Bytes(), &sectionConfig); err != nil {
	// 	log.Println("[ERROR] unable to decode the json for section.config.json which is used to log the image name and version: %w" ,err)
	// }
	moduleVersion := "unknown"
	for _,v := range sectionConfig.Proxychain{
		if(v.Name == c.AppPath){
			moduleVersion = v.Image
		}
	}
	if moduleVersion == "unknown"{
		log.Println("[DEBUG] failed to pair app path (aka proxy name) with image (version)")
	}
	// for proxy, _ := range sectionConfig["proxychain"]{

	// }
	log.Printf("[Info] Pushing %s proxy tarball hash for module image version %s to git remote %s\n", c.AppPath, moduleVersion, cloneDir)
	err = r.Push(&git.PushOptions{Auth: gitAuth, Progress: progressOutput})

	if err != nil {
		return fmt.Errorf("failed to push git changes: %w", err)
	}
	
	return nil
}