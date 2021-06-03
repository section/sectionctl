package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitHTTP "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rs/zerolog/log"
	"github.com/section/sectionctl/api"
)

// GitService interface provides a way to interact with Git
type GitService interface {
	UpdateGitViaGit(ctx *kong.Context, c *DeployCmd, response UploadResponse, logWriters *LogWriters) error
}

// GS ...
type GS struct{}

// This is far less then ideal, however Kong does not seem to provide a way to inject dependencies into its commands so we must use this for testing
var globalGitService GitService = &GS{}

// UpdateGitViaGit clones the application repository to a temporary directory then updates it with the latest payload id and pushes a new commit
func (g *GS) UpdateGitViaGit(ctx *kong.Context, c *DeployCmd, response UploadResponse,logWriters *LogWriters) error {
	app, err := api.Application(c.AccountID, c.AppID)
	if err != nil {
		return err
	}
	appName := strings.ReplaceAll(app.ApplicationName, "/", "")
	cloneDir := fmt.Sprintf("https://aperture.section.io/account/%d/application/%d/%s.git", c.AccountID, c.AppID, appName)
	log.Debug().Msg(fmt.Sprintf(" Begin updating hash in .section-external-source.json:\n\tsection-configmap-tars/%v/%s.tar.gz\n",c.AccountID,response.PayloadID))
	tempDir, err := ioutil.TempDir("", "sectionctl-*")
	if err != nil {
		return err
	}
	log.Debug().Msg(fmt.Sprintln("tempDir: ", tempDir))
	// Git objects storer based on memory
	gitAuth := &gitHTTP.BasicAuth{
		Username: "section-token", // yes, this can be anything except an empty string
		Password: api.Token,
	}
	payload := PayloadValue{ID: response.PayloadID}
	branchRef := fmt.Sprintf("refs/heads/%s",c.Environment)
	var r *git.Repository
	progressOutput := logWriters.CarriageReturnWriter
	log.Info().Msg(fmt.Sprintln("Cloning section config repo for your application to ",tempDir))
	r, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:      cloneDir,
		Auth:     gitAuth,
		Progress: progressOutput,
		ReferenceName: plumbing.ReferenceName(branchRef),
	})
	
	if err != nil {
		log.Error().Err(err).Msg("error cloning")
		return err
	}
	// ... retrieving the branch being pointed by HEAD
	ref, err := r.Head()
	if err != nil {
		log.Error().Err(err).Msg("error retrieving the git HEAD")
		return err
	}
	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		log.Error().Err(err).Msg("error retrieving the commit hash")
		return err
	}
	log.Debug().Msg(fmt.Sprintln("HEAD commit: ", commit))
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
	log.Debug().Str("Old tarball UUID",  content);
	log.Debug().Str("New tarball UUID",  response.PayloadID)
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
	log.Debug().Msg(fmt.Sprintln("git status: ", status))
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
	log.Debug().Msg(fmt.Sprintln("New Commit: ", cmt.String()))
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
	log.Debug().Msg(fmt.Sprintln("contents in new commit: ", ctt))
	
	configFile, err := tree.File("section.config.json")
	if err != nil {
		log.Error().Err(err).Msg("unable to open section.config.json which is used to log the image name and version")
	}
	sectionConfigContents, err := configFile.Contents()
	if err != nil {
		log.Error().Err(err).Msg("unable to open section.config.json which is used to log the image name and version")
	}
	sectionConfig, err := ParseSectionConfig(sectionConfigContents)
	if err != nil{
		log.Error().Err(err).Msg("There was an issue reading the section.config.json")
	}
	// if err := json.Unmarshal(sectionConfigContent.Bytes(), &sectionConfig); err != nil {
	// 	log.Error().Err(err).Msg("unable to decode the json for section.config.json which is used to log the image name and version")
	// }
	moduleVersion := "unknown"
	for _,v := range sectionConfig.Proxychain{
		if(v.Name == c.AppPath){
			moduleVersion = v.Image
		}
	}
	if moduleVersion == "unknown"{
		log.Debug().Msg("failed to pair app path (aka proxy name) with image (version)")
	}
	// for proxy, _ := range sectionConfig["proxychain"]{

	// }
	log.Info().Str("Git Remote",cloneDir).Msg("")
	log.Info().Str("Tarball Source",fmt.Sprintf("%v/%s.tar.gz",c.AccountID,response.PayloadID)).Msg("")
	log.Info().Str("Module Name",c.AppPath).Msg("")
	log.Info().Str("Module Version",moduleVersion).Msg("")
	log.Info().Msg("Validating your app...")
	err = r.Push(&git.PushOptions{Auth: gitAuth, Progress: progressOutput})

	if err != nil {
		return fmt.Errorf("failed to push git changes: %w", err)
	}
	
	return nil
}