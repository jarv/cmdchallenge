package challenge

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

const (
	txtExt = ".txt"
	exeExt = ".exe"
	docExt = ".doc"
	curDir = "."
)

type Check struct {
	log      *slog.Logger
	ch       *Challenge
	oopsDone chan string
}

type CheckFuncType func(*Check) (string, error)

var checkTable = map[string]CheckFuncType{
	"delete_files":                   (*Check).chDeleteFiles,
	"remove_extensions_from_files":   (*Check).chRemoveExtensionsFromFiles,
	"remove_files_with_a_dash":       (*Check).chRemoveFilesWithADash,
	"remove_files_with_extension":    (*Check).chRemoveFilesWithExtension,
	"remove_files_without_extension": (*Check).chRemoveFilesWithoutExtension,
	"replace_text_in_files":          (*Check).chReplaceTextInFiles,
	"create_file":                    (*Check).chCreateFile,
	"create_directory":               (*Check).chCreateDirectory,
	"create_symlink":                 (*Check).chCreateSymlink,
	"copy_file":                      (*Check).chCopyFile,
	"move_file":                      (*Check).chMoveFile,
	"oops_kill_a_process":            (*Check).chOopsKillAProcess,
	"12days_8":                       (*Check).chTwelveDays8,
}

func NewCheck(log *slog.Logger, ch *Challenge, oopsDone chan string) *Check {
	return &Check{log, ch, oopsDone}
}

func (c *Check) RunCheck() (string, error) {
	challengePath := path.Join("/var/challenges", c.ch.Dir())
	if err := os.Chdir(challengePath); err != nil {
		return "Test failed, the challenge directory is missing!", nil
	}

	checkFn, exists := checkTable[c.ch.Slug()]
	if !exists {
		return "", ErrCheckNotExist
	}

	checkResult, err := checkFn(c)
	if err != nil {
		return "", err
	}

	return checkResult, nil
}

func isFile(fname string) (bool, error) {
	if info, err := os.Stat(fname); err == nil {
		return !info.IsDir(), nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("checking isFile for %s failed: %v", fname, err.Error())
	}
}

func isDir(fname string) (bool, error) {
	if info, err := os.Stat(fname); err == nil {
		return info.IsDir(), nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("checking isDir for %s failed: %v", fname, err.Error())
	}
}

func isSymlink(fname string) (bool, error) {
	if info, err := os.Lstat(fname); err == nil {
		return info.Mode()&os.ModeSymlink == os.ModeSymlink, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("checking isSymlink for %s failed: %v", fname, err.Error())
	}
}

func isFileNotSymlink(fname string) (bool, error) {
	isFile, err := isFile(fname)
	if err != nil {
		return false, err
	}

	isSymlink, err := isSymlink(fname)
	if err != nil {
		return false, err
	}

	return (isFile && !isSymlink), nil
}

func fileContents(fname string) (string, error) {
	dat, err := os.ReadFile(fname)
	if err != nil {
		return "", err
	}

	return string(dat), nil
}

type entry struct {
	path  string
	isDir bool
}

func walkDirRec(dirName string) ([]entry, error) {
	entries := []entry{}

	err := filepath.WalkDir(dirName, func(path string, d fs.DirEntry, err error) error {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		} else if err != nil {
			return err
		}

		entries = append(entries, entry{
			path:  path,
			isDir: d.IsDir(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walkDirRec failed for %s: %v", dirName, err.Error())
	}

	// Pop the first item since it includes the root
	return entries[1:], nil
}

// Checks

const fileNotExist = "Test failed, file does not exist"

func (c *Check) chOopsKillAProcess() (string, error) {
	if c.isOopsCmdRunning() {
		return "Test failed, process is still running", nil
	}

	return "", nil
}

func (c *Check) chCreateFile() (string, error) {
	chk, err := isFileNotSymlink("take-the-command-challenge")
	if err != nil {
		return "", err
	}

	if !chk {
		return fileNotExist, nil
	}

	contents, err := fileContents("take-the-command-challenge")
	if err != nil {
		return "", err
	}

	if contents != "" {
		return "Test failed, file is not empty", nil
	}

	return "", nil
}

func (c *Check) chCreateDirectory() (string, error) {
	chk, err := isFile("tmp")
	if err != nil {
		return "", err
	}
	if chk {
		return "Test failed, did you create a file?", nil
	}

	chk, err = isDir("tmp/files")
	if err != nil {
		return "", err
	}

	if !chk {
		return "Test failed, directory does not exist", nil
	}

	return "", nil
}

func (c *Check) chCopyFile() (string, error) {
	chk, err := isFileNotSymlink("tmp/files/take-the-command-challenge")
	if err != nil {
		return "", err
	}

	if !chk {
		return fileNotExist, nil
	}

	chk, err = isFileNotSymlink("take-the-command-challenge")
	if err != nil {
		return "", err
	}

	if !chk {
		return "Test failed, original file was removed", nil
	}

	return "", nil
}

func (c *Check) chMoveFile() (string, error) {
	chk, err := isFileNotSymlink("tmp/files/take-the-command-challenge")
	if err != nil {
		return "", err
	}

	if !chk {
		return fileNotExist, nil
	}

	chk, err = isFile("take-the-command-challenge")
	if err != nil {
		return "", err
	}

	if chk {
		return "Test failed, file was not moved", nil
	}

	contents, err := fileContents("tmp/files/take-the-command-challenge")
	if err != nil {
		return "", err
	}

	if contents != "" {
		return "Test failed, file was modified", nil
	}

	return "", nil
}

func (c *Check) chCreateSymlink() (string, error) {
	const linkName string = "take-the-command-challenge"
	chk, err := isSymlink(linkName)
	if err != nil {
		return "", err
	}

	if !chk {
		return "Test failed, symlink does not exist", nil
	}

	link, err := os.Readlink(linkName)
	if err != nil {
		return "", err
	}

	if link == linkName {
		return "Test failed, link points to itself!", nil
	}

	fpath, err := filepath.EvalSymlinks("/var/challenges/create_symlink/take-the-command-challenge")
	if err != nil {
		return "", err
	}

	if fpath != "/var/challenges/create_symlink/tmp/files/take-the-command-challenge" {
		return "Test failed, symlink does not point to tmp/files/take-the-command-challenge", nil
	}

	return "", nil
}

func (c *Check) chDeleteFiles() (string, error) {
	chk, err := isDir("/var/challenges/delete_files")
	if err != nil {
		return "", err
	}

	if !chk {
		return "Test failed, challenge directory was removed", nil
	}

	entries, err := walkDirRec(curDir)
	if err != nil {
		return "", err
	}

	if len(entries) > 0 {
		return "Test failed, files or directories remain", nil
	}

	return "", nil
}

func (c *Check) chRemoveExtensionsFromFiles() (string, error) {
	entries, err := walkDirRec(curDir)
	if err != nil {
		return "", err
	}

	for _, f := range filesFromEntries(entries) {
		if filepath.Ext(f) != "" {
			return fmt.Sprintf("Test failed, found a file '%s' with an extension", f), nil
		}
	}

	return "", nil
}

func (c *Check) chRemoveFilesWithADash() (string, error) {
	entries, err := walkDirRec(curDir)
	if err != nil {
		return "", err
	}

	files := filesFromEntries(entries)

	if len(files) != 1 {
		return "Test failed, expecting one file", nil
	}

	for _, f := range files {
		if strings.Contains(f, "-") {
			return fmt.Sprintf("Test failed, found a file '%s' with a dash in the name", f), nil
		}
	}

	return "", nil
}

func (c *Check) chRemoveFilesWithExtension() (string, error) {
	entries, err := walkDirRec(curDir)
	if err != nil {
		return "", err
	}

	files := filesFromEntries(entries)

	const expectedFiles = 4
	if len(files) != expectedFiles {
		return fmt.Sprintf("Test failed, got %d files, expected %d", len(files), expectedFiles), nil
	}

	for _, f := range files {
		if filepath.Ext(f) == docExt {
			return fmt.Sprintf("Test failed, found a file '%s' with a .doc extension", f), nil
		}
	}

	return "", nil
}

func (c *Check) chRemoveFilesWithoutExtension() (string, error) {
	entries, err := walkDirRec(curDir)
	if err != nil {
		return "", err
	}
	files := filesFromEntries(entries)

	const expectedFiles = 4
	if len(files) != expectedFiles {
		return fmt.Sprintf("Test failed, got %d files, expected %d", len(files), expectedFiles), nil
	}

	for _, f := range files {
		ext := filepath.Ext(f)
		if !(ext == txtExt || ext == exeExt) {
			return fmt.Sprintf("Test failed, found a file '%s' without a .txt or .exe extension", f), nil
		}
	}

	return "", nil
}

func (c *Check) chReplaceTextInFiles() (string, error) {
	entries, walkErr := walkDirRec(curDir)
	if walkErr != nil {
		return "", walkErr
	}

	files := filesFromEntries(entries)
	for _, f := range files {
		if filepath.Ext(f) == txtExt {
			if contents, err := fileContents(f); err != nil {
				return "", err
			} else if strings.Contains(contents, "challenges are difficult") {
				return "Test failed, found the string 'challenges are difficult'", nil
			}
		}
	}

	const failRemainUnmodified = "Test failed, files without .txt extension must remain unmodified."
	if !slices.Contains(files, "not-a-text-file") {
		return failRemainUnmodified, nil
	}

	contents, err := fileContents("not-a-text-file")
	if err != nil {
		return "", err
	}

	if !strings.Contains(contents, "challenges are difficult") {
		return failRemainUnmodified, nil
	}

	return "", nil
}

func (c *Check) chTwelveDays8() (string, error) {
	const dirName string = "Elves"

	elves, err := walkDirRec(dirName)
	if err != nil {
		return "", err
	}

	l := len(elves)
	if len(elves) != 0 {
		c.log.Info("chTwelveDays8 check failed", "dirName", dirName, "gotLen", l, "expected", 0, "files", elves)
		return fmt.Sprintf("Test failed, elves are still in %s/", dirName), nil
	}

	entries, err := walkDirRec("Workshop")
	if err != nil {
		return "", err
	}
	files := filesFromEntries(entries)

	sort.Strings(files)

	expected := []string{
		"Workshop/Alabaster Snowball", "Workshop/Buddy",
		"Workshop/Bushy Evergreen", "Workshop/Hermey", "Workshop/Pepper Minstix",
		"Workshop/Shinny Upatree", "Workshop/Sugarplum Mary",
		"Workshop/Wunorse Openslae",
	}
	if !testEq(expected, files) {
		c.log.Info("chTwelveDays8 check failed", "expected", expected, "workshop", files)
		return "Test failed, Elves are not in the Workshop!", nil
	}

	return "", nil
}

func testEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func filesFromEntries(entries []entry) []string {
	files := []string{}
	for _, e := range entries {
		if e.isDir {
			continue
		}
		files = append(files, e.path)
	}
	return files
}

func (c *Check) isOopsCmdRunning() bool {
	const oopsRunningTimeout = 100
	select {
	case <-c.oopsDone:
		return false
	case <-time.After(oopsRunningTimeout * time.Millisecond):
		return true
	}
}
