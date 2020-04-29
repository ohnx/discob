package main

import (
    "log"
    "fmt"

    // go-git libraries
	git "github.com/go-git/go-git/v5"
	gp "github.com/go-git/go-git/v5/plumbing"
    gobj "github.com/go-git/go-git/v5/plumbing/object"
)

type (
    // Git Helper
    GitHelper struct {
        Repo *git.Repository
        Location string
    }
    GitFile struct {
        Filename string
        IsFile bool
    }
)

func (self *GitHelper) InitGitHelper(location string) {
    rhandle, err := git.PlainOpen(location)

    if err != nil {
        log.Fatalf("Failed to open Git repository: %s", err)
    }

    self.Repo = rhandle
    self.Location = location
}

func (self *GitHelper) FetchRevision(identifier string) (gp.Hash, error) {
    // try treating the identifier as a branch first
    branch, err := self.Repo.Branch(identifier)
    if err == nil {
        // branch found, try to get the reference
        ref, err := self.Repo.Reference(branch.Merge, true)

        // get the commit hash
        if err == nil {
            return ref.Hash(), nil
        }
    }

    // try treating the identifier as a tag now
    ref, err := self.Repo.Tag(identifier)
    if err == nil {
        // check if this is an annotated tag
        tagobj, err := self.Repo.TagObject(ref.Hash())
        if err == nil {
            // yup, annotated tag
            commit, err := tagobj.Commit()
            if err != nil {
                // tag does not point to commit
                return gp.ZeroHash, fmt.Errorf("Tag %s is not associated with commit: %s", identifier, err)
            }
            return commit.Hash, nil
        } else {
            // nope, lightweight tag
            return ref.Hash(), nil
        }
    }

    // no such tag found, so treat it as a commit hash
    return gp.NewHash(identifier), nil
}

func (self *GitHelper) FetchTreeAtRevision(hash gp.Hash, path string) ([]GitFile, error) {
    // try to get the commit corresponding to this hash
    commit, err := self.Repo.CommitObject(hash)
    if err != nil {
        // no such commit found
        return nil, fmt.Errorf("Commit %s not found: %s", hash, err)
    }

    // now get the tree associated with this commit
    tree, err := self.Repo.TreeObject(commit.TreeHash)
    if err != nil {
        // no such tree found
        return nil, fmt.Errorf("Could not locate tree associated with commit %s: %s", hash, err)
    }

    // we are trying to get a subdirectory
    if len(path) > 0 {
        tree, err = tree.Tree(path)
        if err != nil {
            // failed to find the path
            return nil, fmt.Errorf("Could not locate path in commit %s: %s", hash, err)
        }
    }

    // iterate through each of the files
    tw := gobj.NewTreeWalker(tree, false, nil)
    files := make([]GitFile, 0)

keep_going:
    name, entry, err := tw.Next()
    if err != nil {
        goto done
    }

    files = append(files, GitFile{
        Filename: name,
        IsFile: entry.Mode.IsFile(),
    })

    goto keep_going
done:
    tw.Close()

    return files, nil
}

func (self *GitHelper) FetchFileAtRevision(hash gp.Hash, path string) (string, error) {
    // try to get the commit corresponding to this hash
    commit, err := self.Repo.CommitObject(hash)
    if err != nil {
        // no such commit found
        return "", fmt.Errorf("Commit %s not found: %s", hash, err)
    }

    // now get the tree associated with this commit
    tree, err := self.Repo.TreeObject(commit.TreeHash)
    if err != nil {
        // no such tree found
        return "", fmt.Errorf("Could not locate tree associated with commit %s: %s", hash, err)
    }

    // now get the file in this tree
    file, err := tree.File(path)
    if err != nil {
        // no such file found
        return "", fmt.Errorf("Could not locate file in commit %s: %s", hash, err)
    }

    // now return the file contents!
    contents, err := file.Contents()
    if err != nil {
        // failed to convert file to string
        return "", fmt.Errorf("Could not output file in commit %s: %s", hash, err)
    }

    return contents, err
}
