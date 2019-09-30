package main

import (
	"io/ioutil"
	"os"
)

type Filetype int
const (
	FILE Filetype = iota
	FOLDER
)
// When showing the contents of a folder, this is used to get info about the type without getting the contents
type FileInfo struct {
	Path, Name, Owner string
	Filetype Filetype
}
type File struct {
	Path, FileContents, Owner string
	Filetype Filetype
	FolderContents []FileInfo
}

type Files interface {
	Init(string) error // Currently location, may change to adapt to new methods
	Get(string, string) (File, error) //Owner, Path
	Edit(string, string, string) (error) //Owner, Path, Contents
	Rename(string, string, string) (error) //Owner, Path, newPath
	Remove(string, string) error //Owner, Path
	New(string, string) (File, error) //Owner, Path
	NewFolder(string, string) (File, error) //Owner, Path
	NewUser(string) error //Username
}

type FSFiles struct {
	Location string
}
func (f *FSFiles) Init(location string) error {
	f.Location = location
	return nil
}
func (f FSFiles) Get(owner, path string) (File, error) {
	fi, err := os.Lstat(f.Location + "/" + owner + "/" + path)
	if err != nil {
		return File{}, err
	}
	file := File{
		Owner: owner,
		Path: path,
	}
	if fi.IsDir() {
		file.Filetype = FOLDER
		folder, err := ioutil.ReadDir(f.Location + "/" + owner + "/" + path)
		if err != nil {
			return File{}, err
		}
		file.FolderContents = make([]FileInfo, 0, len(folder))
		for _, info := range folder {
			var filetype Filetype
			if info.IsDir() {
				filetype = FOLDER
			} else {
				filetype = FILE
			}
			file.FolderContents = append(file.FolderContents, FileInfo{
				Path: path + "/" + info.Name(),
				Name: info.Name(),
				Owner: owner,
				Filetype: filetype,
			})
		}
	} else {
		file.Filetype = FILE
		contents, err := ioutil.ReadFile(f.Location + "/" + owner + "/" + path)
		if err != nil {
			return File{}, err
		}
		file.FileContents = string(contents)
	}
	return file, nil
}

func (f FSFiles) Edit(owner, path, contents string) error {
	//should I put in some more checks here? I think this is alright
	return ioutil.WriteFile(f.Location + "/" + owner + "/" + path, []byte(contents), 0644)
}

func (f FSFiles) Rename(owner, oldpath, newpath string) error {
	return os.Rename(f.Location + "/" + owner + "/" + oldpath, f.Location + "/" + owner + "/" + newpath)
}

func (f FSFiles) Remove(owner, path string) error {
	return os.RemoveAll(f.Location + "/" + owner + "/" + path)
}

func (f FSFiles) New(owner, path string) (File, error) {
	err := f.Edit(owner, path, "")
	if err != nil {
		return File{}, err
	}
	return File{
		Owner: owner,
		Path: path,
		Filetype: FILE,
		FileContents: "",
	}, nil
}

func (f FSFiles) NewFolder(owner, path string) (File, error) {
	err := os.MkdirAll(f.Location + "/" + owner + "/" + path, os.ModeDir + 0755)
	if err != nil {
		return File{}, err
	}
	return File{
		Owner: owner,
		Path: path,
		Filetype: FOLDER,
		FolderContents: []FileInfo{},
	}, nil
}
func (f FSFiles) NewUser(username string) error {
	err := os.MkdirAll(f.Location + "/" + username, os.ModeDir + 0755)
	if err != nil {return err}
	defaultStyle, err := ioutil.ReadFile("views/default.css")
	if err != nil {return err}
	err = ioutil.WriteFile(f.Location + "/" + username + "/.style.css", defaultStyle, 0644)
	if err != nil {return err}
	defaultScript, err := ioutil.ReadFile("views/default.js")
	if err != nil {return err}
	return ioutil.WriteFile(f.Location + "/" + username + "/.userScript.js", defaultScript, 0644)
}
