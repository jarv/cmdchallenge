package config

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"
)

const roVolumeDirName = "ro_volume"
const sqliteDBFileName = "db.sqlite3"
const runCmdRegistryImg = "registry.gitlab.com/jarv/cmdchallenge"
const oopsBin = "oops-this-will-delete-bin-dirs"

var (
	ErrInvalidRegistryImgURI = errors.New("registry image doesn't exist")
)

type Config struct {
	CmdTimeout           time.Duration
	RunCmdRegistryImgTag string
	RunCmdRegistryImg    string
	RegistryAuth         string
	RunCmdTimeout        time.Duration
	RemoveImageTimeout   time.Duration
	PullImageTimeout     time.Duration
	ROVolumeDir          string
	SQLiteDBFile         string
	CMDImgSuffix         string
	CMDImgNames          []string
	OopsBin              string
	SolutionsKeyPrefix   string
	Caller               string
	registryImgURIs      map[string]string
}

func New() *Config {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}

	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	searchDirs := []string{
		path.Join(path.Dir(filename), "../.."),
		"/",
		filepath.Dir(exe),
	}

	var roVolumeDir string
	var sqliteDBFile string
	for _, d := range searchDirs {
		if roVolumeDir == "" && isDir(path.Join(d, roVolumeDirName)) {
			roVolumeDir = path.Join(d, roVolumeDirName)
		}
		if sqliteDBFile == "" && isFile(path.Join(d, sqliteDBFileName)) {
			sqliteDBFile = path.Join(d, sqliteDBFileName)
		}
	}

	if sqliteDBFile == "" {
		sqliteDBFile = path.Join(path.Dir(filename), "../..", sqliteDBFileName)
	}

	cmdImgSuffix := os.Getenv("CMD_IMG_SUFFIX")

	return &Config{
		CmdTimeout:         5 * time.Second,
		RunCmdRegistryImg:  runCmdRegistryImg,
		RegistryAuth:       "",
		RunCmdTimeout:      6 * time.Second,
		PullImageTimeout:   5 * time.Minute,
		RemoveImageTimeout: 60 * time.Second,
		ROVolumeDir:        getEnv("RO_VOLUME_DIR", roVolumeDir),
		SQLiteDBFile:       getEnv("SQLITE_DB_FILE", sqliteDBFile),
		CMDImgSuffix:       cmdImgSuffix,
		OopsBin:            oopsBin,
		SolutionsKeyPrefix: "s/solutions",
		registryImgURIs: map[string]string{
			"cmd":        path.Join(runCmdRegistryImg, "cmd") + cmdImgSuffix + ":latest",
			"cmd-no-bin": path.Join(runCmdRegistryImg, "cmd-no-bin") + cmdImgSuffix + ":latest",
		},
	}
}

func (c *Config) JSONForSlug(slug string) ([]byte, error) {
	f, err := os.Open(c.jsonSlugPath(slug))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	chJSON, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return chJSON, nil
}

func (c *Config) ChallengePath() string {
	return path.Join(c.ROVolumeDir, "ch")
}

func (c *Config) RegistryImgURI(name string) (string, error) {
	if val, ok := c.registryImgURIs[name]; ok {
		return val, nil
	}
	return "", ErrInvalidRegistryImgURI
}

func (c *Config) jsonSlugPath(slug string) string {
	return path.Join(c.ChallengePath(), slug+".json")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func isFile(fname string) bool {
	if info, err := os.Stat(fname); err == nil {
		return !info.IsDir()
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		panic(err)
	}
}

func isDir(fname string) bool {
	if info, err := os.Stat(fname); err == nil {
		return info.IsDir()
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		panic(err)
	}
}
