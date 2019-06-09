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
	Path, Owner string
	Filetype Filetype
}
type File struct {
	Path, FileContents, Owner string
	Filetype Filetype
	FolderContents []FileInfo
}

type Files interface {
	Get(string, string) (File, error) //Owner, Path
	Edit(string, string, string) (error) //Owner, Path, Contents
	Rename(string, string, string) (error) //Owner, Path, newPath
	New(string, string) (File, error) //Owner, Path
	NewFolder(string, string) (File, error) //Owner, Path
}

type FSFiles struct {
	Location string
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
		folder, err := ioutil.ReadDir(f.Location + owner + "/" + path)
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
	return ioutil.WriteFile(f.Location + "/" + owner + "/" + path, []byte(contents), 0600)
}

func (f FSFiles) Rename(owner, oldpath, newpath string) error {
	return os.Rename(f.Location + "/" + owner + "/" + oldpath, f.Location + "/" + owner + "/" + newpath)
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
	err := os.MkdirAll(path, 0600)
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
