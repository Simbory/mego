package mego

import (
	"mime/multipart"
	"os"
	"io"
)

type UploadFile struct {
	FileName string
	Size     int64
	Error    error
	File   multipart.File
	Header *multipart.FileHeader
}

func (file *UploadFile) Save(path string) error {
	if file.Error != nil {
		return file.Error
	}
	f,err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0666)
	if err == nil {
		defer f.Close()
		if file.File != nil {
			_,err := io.Copy(f, file.File)
			return err
		}
	}
	return err
}

func (file *UploadFile) SaveAndClose(path string) error {
	err := file.Save(path)
	if file.File != nil {
		err1 := file.File.Close()
		if err == nil {
			err = err1
		}
	}
	return err
}