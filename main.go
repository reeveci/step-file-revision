package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/google/shlex"
)

func main() {
	reeveAPI := os.Getenv("REEVE_API")
	if reeveAPI == "" {
		fmt.Println("This docker image is a Reeve CI pipeline step and is not intended to be used on its own.")
		os.Exit(1)
	}

	files, err := shlex.Split(os.Getenv("FILES"))
	if err != nil {
		panic(fmt.Sprintf("error parsing file list - %s", err))
	}
	if len(files) == 0 {
		panic("no files specified")
	}

	revisionVar := os.Getenv("REVISION_VAR")
	if revisionVar == "" {
		revisionVar = "FILE_REV"
	}

	revs := make([]RevisionInfo, len(files))
	for i, name := range files {
		path, err := filepath.Abs(name)
		if err != nil {
			panic(fmt.Sprintf(`error determining absolute path for "%s" - %s`, name, err))
		}
		file, err := os.Stat(path)
		if err != nil {
			panic(fmt.Sprintf(`error reading file information for "%s" - %s`, name, err))
		}
		if !file.Mode().IsRegular() {
			panic(fmt.Sprintf(`error reading file "%s" - not a regular file`, name))
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf(`error reading file "%s" - %s`, name, err))
		}
		var uid, gid int
		if stat, ok := file.Sys().(*syscall.Stat_t); ok {
			uid = int(stat.Uid)
			gid = int(stat.Gid)
		} else {
			panic(fmt.Sprintf(`error reading ownership information for "%s"`, name))
		}
		revs[i] = RevisionInfo{
			Uid:     uid,
			Gid:     gid,
			Mode:    uint32(file.Mode()),
			Name:    path,
			Content: contents,
		}
	}

	sort.Slice(revs, func(i, j int) bool {
		return revs[i].Name < revs[j].Name
	})

	hashData, err := json.Marshal(revs)
	if err != nil {
		panic(fmt.Sprintf("error marshaling revision info - %s", err))
	}
	hasher := sha1.New()
	hasher.Write(hashData)
	revision := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	response, err := http.Post(fmt.Sprintf("%s/api/v1/var?key=%s", reeveAPI, url.QueryEscape(revisionVar)), "text/plain", strings.NewReader(revision))
	if err != nil {
		panic(fmt.Sprintf("error setting revision var - %s", err))
	}
	if response.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("setting revision var returned status %v", response.StatusCode))
	}
	fmt.Printf("Set %s=%s\n", revisionVar, revision)
}

type RevisionInfo struct {
	Uid, Gid int
	Mode     uint32
	Name     string
	Content  []byte
}
