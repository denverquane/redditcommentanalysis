package filesystem

import (
	"log"
	"os"
)

func DoesFolderExist(path string) bool {
	if v, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else if v.IsDir() {
		return true
	} else {
		log.Println("Folder " + path + " does not exist, but it is a file!")
		return false
	}
}

func CreateFolder(path string) error {
	err := os.Mkdir(path, os.FileMode(0522))
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Created " + path)
	}
	return err
}

func DoesFileExist(path string) bool {
	if v, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else if v.IsDir() {
		return false
	} else {
		return true
	}
}

func CreateSubdirectoryStructure(basedir, month, year string) {
	if !DoesFolderExist(basedir + "/Extracted") {
		CreateFolder(basedir + "/Extracted")
	}
	if !DoesFolderExist(basedir + "/Extracted/" + year) {
		CreateFolder(basedir + "/Extracted/" + year)
	}
	if !DoesFolderExist(basedir + "/Extracted/" + year + "/" + month) {
		CreateFolder(basedir + "/Extracted/" + year + "/" + month)
	}
}
