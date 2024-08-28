package main

import (
	"deployment/command"
	"github.com/go-git/go-git/v5"
	config2 "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var cmd = command.Command{}
var auth *http.BasicAuth

func main() {
	//var wg sync.WaitGroup
	for true {
		byteArray, err := os.ReadFile(Opts.Config)
		err = yaml.Unmarshal(byteArray, &config)

		auth = &http.BasicAuth{
			Username: config.Username,
			Password: config.Token,
		}

		if err != nil {
			return
		}

		//wg.Add(len(config.Projects))
		for _, project := range config.Projects {
			go func(project Project) {
				//defer wg.Done()
				deploy(project)
			}(project)
		}

		//wg.Wait()
		time.Sleep(time.Duration(config.Settings.CheckTime) * time.Second)
	}
}

func deploy(project Project) {

	_, err := git.PlainClone(project.Path, false, &git.CloneOptions{
		ReferenceName: plumbing.NewBranchReferenceName(project.Branch),
		Auth:          auth,
		URL:           project.Url,
	})

	repo, err := git.PlainOpen(project.Path)
	if err != nil {
		log.Println(project, err)
		return
	}

	spec := "refs/heads/" + project.Branch + ":refs/remotes/origin/" + project.Branch

	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config2.RefSpec{config2.RefSpec(spec)},
		Auth:     auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		log.Println(project, err)
		return
	}

	// Yerel dalı kontrol et
	//localRef := plumbing.NewBranchReferenceName(project.Branch)
	//_, err = repo.Storer.Reference(localRef)
	//if err != nil {
	//	checkout(repo, w, project.Branch)
	//}

	// Yerel ve uzak dalları alın
	localBranch, _ := repo.Branches()
	remoteRefs, _ := repo.Storer.IterReferences()

	// Uzak dalı bulun
	remoteBranchRef, err := remoteRefs.Next()
	for err == nil {
		if remoteBranchRef.Name().String() == "refs/remotes/origin/"+project.Branch {
			break
		}
		remoteBranchRef, err = remoteRefs.Next()
	}
	if err != nil {
		log.Println(project, err)
		return
	}

	// Yerel dalı bulun
	localBranchRef, err := localBranch.Next()
	for err == nil {
		if localBranchRef.Name().String() == "refs/heads/"+project.Branch {
			break
		}
		localBranchRef, err = localBranch.Next()
	}
	if err != nil {
		checkout(project)
		log.Println(project, err)
		return
	}

	// İki dala ait Commit'leri alın
	remoteCommits, _ := repo.Log(&git.LogOptions{From: remoteBranchRef.Hash()})
	localCommits, _ := repo.Log(&git.LogOptions{From: localBranchRef.Hash()})

	remoteUpdated := false
	for {
		if !(localCommits == nil || remoteCommits == nil) {
			remoteCommit, remoteErr := remoteCommits.Next()
			localCommit, localErr := localCommits.Next()

			if remoteErr == io.EOF || localErr == io.EOF {
				break
			}

			if remoteCommit.Hash != localCommit.Hash {
				remoteUpdated = true
				break
			}
		} else {
			remoteUpdated = true
			break
		}

	}

	if remoteUpdated {
		checkout(project)
		run(project)

		log.Println(project, "remote updated")
	} else if project.Check != "" {
		out, _ := cmd.Run(project.Check)

		if strings.Contains(string(out), "start_deployment") {
			checkout(project)
			run(project)
		}
	}
}

func run(project Project) {
	cmd.RunCommand(project.Path, "chmod", "+x", project.Script)
	cmd.RunCommand(project.Path, "bash", "-c", project.Script)
}

func checkout(project Project) {
	cmd := command.Command{}
	cmd.RunCommand(project.Path, "git", "reset", "--hard", "HEAD")
	cmd.RunCommand(project.Path, "git", "clean", "-fd")
	cmd.RunCommand(project.Path, "git", "checkout", "master")
	cmd.RunCommand(project.Path, "git", "branch", "-D", project.Branch)
	cmd.RunCommand(project.Path, "git", "checkout", project.Branch)
	cmd.RunCommand(project.Path, "git", "pull")
}
