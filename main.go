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

	"github.com/bmatcuk/doublestar/v4"
	"github.com/google/shlex"
)

func main() {
	reeveAPI := os.Getenv("REEVE_API")
	if reeveAPI == "" {
		fmt.Println("This docker image is a Reeve CI pipeline step and is not intended to be used on its own.")
		os.Exit(1)
	}

	filePatterns, err := shlex.Split(os.Getenv("FILES"))
	if err != nil {
		panic(fmt.Sprintf("error parsing file pattern list - %s", err))
	}
	files := make([]string, 0, len(filePatterns))
	for _, pattern := range filePatterns {
		matches, err := doublestar.FilepathGlob(pattern, doublestar.WithFilesOnly())
		if err != nil {
			panic(fmt.Sprintf(`error parsing file pattern "%s" - %s`, pattern, err))
		}
		files = append(files, matches...)
	}
	files = distinct(files)

	revisionVar := os.Getenv("REVISION_VAR")
	if revisionVar == "" {
		revisionVar = "FILE_REV"
	}

	revs := make([]RevisionInfo, 0, len(files))
	for _, filename := range files {
		path, err := filepath.Abs(filename)
		if err != nil {
			panic(fmt.Sprintf(`error determining absolute path for "%s" - %s`, filename, err))
		}
		file, err := os.Stat(path)
		if err != nil {
			panic(fmt.Sprintf(`error reading file information for "%s" - %s`, filename, err))
		}
		if !file.Mode().IsRegular() {
			fmt.Printf("skipping non regular file \"%s\"\n", path)
			continue
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf(`error reading file "%s" - %s`, filename, err))
		}
		var uid, gid int
		if stat, ok := file.Sys().(*syscall.Stat_t); ok {
			uid = int(stat.Uid)
			gid = int(stat.Gid)
		} else {
			panic(fmt.Sprintf(`error reading ownership information for "%s"`, filename))
		}
		fmt.Printf("including file \"%s\"\n", filename)
		revs = append(revs, RevisionInfo{
			Uid:     uid,
			Gid:     gid,
			Mode:    uint32(file.Mode()),
			Name:    path,
			Content: contents,
		})
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

func distinct[T comparable](items []T) []T {
	keys := make(map[T]struct{})
	result := make([]T, 0, len(items))
	for _, item := range items {
		if _, exists := keys[item]; !exists {
			keys[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
