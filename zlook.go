package zlook

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// An Input represents the input to zlook that consists of a slice of files to
// look at and depth to which looking must be done
type Input struct {
	MaxDepth     int
	ArchiveTypes FileExtensions
	PrefixPath   bool
	ExtractEntry string
	Paths        []string
}

// Inspector provides operators inspecting the contents of archives as configured in Input
type Inspector interface {
	List() error
	Extract() error
}

type inspectorImpl struct {
	Input
	ArchiveTypesMap map[string]bool
}

type entryContext struct {
	Reader zip.Reader
	Depth  int
	Parent string
}

type entryCallback func(ctx entryContext, entry *zip.File) (bool, error)

var _ Inspector = inspectorImpl{}

// FileExtensions represents a collection of file extensions (zip/gz, etc) as a slice of strings
type FileExtensions []string

// DefaultArchiveTypes represents the default collection of file extensions treaated as archives
var DefaultArchiveTypes = []string{".zip", ".esa", ".jar", ".ear", ".war"}

// Get the value of FileExtensions
func (t *FileExtensions) String() string {
	return fmt.Sprintf("%v", *t)
}

// Set the value of FileExtensions
func (t *FileExtensions) Set(value string) error {
	*t = strings.Split(value, ",")
	if *t == nil {
		*t = DefaultArchiveTypes
	}
	return nil
}

// NewInspector constructs, initizlies and returns an Inspector from Input
func NewInspector(in Input) Inspector {
	if len(in.ArchiveTypes) == 0 {
		in.ArchiveTypes = DefaultArchiveTypes
	}
	inspector := inspectorImpl{in, slice2Set(in.ArchiveTypes)}
	return inspector
}

func slice2Set(s []string) map[string]bool {
	set := make(map[string]bool)
	for _, item := range s {
		set[item] = true
	}
	return set
}

// List Contents of archives corresponding to in.Paths (currently only zip
// supported) to in.Depth depth
func (ip inspectorImpl) List() error {
	return ip.crawl(ip.listCallback)
}

// Extracts fist entry with suffix matching ip.ExtractEntry in archives
// corresponding to in.Paths if found within ip.MaxDepth or returns error if not found
func (ip inspectorImpl) Extract() error {
	return ip.crawl(ip.extractCallback)
	// for _, p := range ip.Paths {
	// 	r, err := zip.OpenReader(p)
	// 	if err != nil {
	// 		return errors.Wrap(err, fmt.Sprintf("list: Cannot read %s", p))
	// 	}
	// 	defer r.Close()
	// 	ip.doCrawl(entryContext{Reader: r.Reader, Depth: 0}, ip.listCallback)
	// }
	// return nil
}

func (ip inspectorImpl) crawl(cbk entryCallback) error {
	for _, p := range ip.Paths {
		r, err := zip.OpenReader(p)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Cannot read %s", p))
		}
		defer r.Close()
		ip.doCrawl(entryContext{Reader: r.Reader, Depth: 0}, cbk)
	}
	return nil
}

func (ip inspectorImpl) listCallback(ctx entryContext, entry *zip.File) (bool, error) {
	if !strings.HasSuffix(entry.Name, "/") {
		if ip.PrefixPath && ctx.Parent != "" {
			fmt.Printf("%s/%s\n", ctx.Parent, entry.Name)
		} else {
			fmt.Printf("%*s%s\n", ctx.Depth*2, "", entry.Name)
		}
	}
	return true, nil
}

func (ip inspectorImpl) extractCallback(ctx entryContext, entry *zip.File) (bool, error) {
	var entryMatches bool
	if ctx.Parent != "" {
		entryMatches = strings.HasSuffix(ctx.Parent+"/"+entry.Name, ip.ExtractEntry)
	} else {
		entryMatches = strings.HasSuffix(entry.Name, ip.ExtractEntry)
	}
	if entryMatches {
		fr, err := entry.Open()
		if err != nil {
			return false, errors.Wrap(err, fmt.Sprintf("Cannot open matching entry %s", entry.Name))
		}
		_, err = io.Copy(os.Stdout, fr)
		if err != nil {
			return false, errors.Wrap(err, fmt.Sprintf("Cannot not copy matching entry %s", entry.Name))
		}
		return false, nil
	}
	return true, nil
}

func (ip inspectorImpl) doCrawl(ctx entryContext, cbk entryCallback) bool {
	if ctx.Depth > ip.MaxDepth {
		return false
	}
	for _, f := range ctx.Reader.File {
		entryName := f.Name
		proceed, err := cbk(ctx, f)
		if !proceed {
			return false
		} else if err != nil {
			log.Println(err)
		}
		fext := filepath.Ext(entryName)
		if !ip.ArchiveTypesMap[fext] || ctx.Depth+1 > ip.MaxDepth {
			continue
		}
		fr, err := f.Open()
		if err != nil {
			log.Println(fmt.Sprintf("Error opening: %s", entryName))
			continue
		}
		defer fr.Close()
		entryBytes, err := ioutil.ReadAll(fr)
		if err != nil {
			log.Println(fmt.Sprintf("Error reading: %s", entryName))
			continue
		}
		bReader := bytes.NewReader(entryBytes)
		entryZipReader, err := zip.NewReader(bReader, int64(len(entryBytes)))
		if err != nil {
			log.Println(fmt.Sprintf("Error reading : %s", entryName))
			continue
		}
		var newParent string
		if ctx.Parent != "" {
			newParent = ctx.Parent + "/" + entryName
		} else {
			newParent = entryName
		}

		if (ip.doCrawl(entryContext{*entryZipReader, ctx.Depth + 1, newParent}, cbk)) {
			continue
		} else {
			return false
		}
	}
	return true
}
