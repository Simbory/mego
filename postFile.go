package mego

import (
	"mime/multipart"
	"os"
	"io"
)

type PostFile struct {
	FileName string
	Size     int64
	Error    error
	File   multipart.File
	Header *multipart.FileHeader
}

func (file *PostFile) Save(path string) error {
	if file.Error != nil {
		return file.Error
	}
	f,err := os.Create(path)
	if err == nil {
		defer f.Close()
		if file.File != nil {
			_,err := io.Copy(f, file.File)
			return err
		}
	}
	return err
}

func (file *PostFile) SaveAndClose(path string) error {
	err := file.Save(path)
	if file.File != nil {
		err1 := file.File.Close()
		if err == nil {
			err = err1
		}
	}
	return err
}