package config

import (
	"os"
	"path"
	"runtime"
	"time"
)

type Config struct {
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
	SolutionsKeyPrefix   string
}

func New() *Config {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	roVolumeDir := path.Join(path.Dir(filename), "../../ro_volume")
	sqliteDBFile := path.Join(path.Dir(filename), "../../db.sqlite3")

	cmdImgSuffix := os.Getenv("CMD_IMG_SUFFIX")

	return &Config{
		RunCmdRegistryImg:  "registry.gitlab.com/jarv/cmdchallenge",
		RegistryAuth:       "",
		RunCmdTimeout:      6 * time.Second,
		PullImageTimeout:   5 * time.Minute,
		RemoveImageTimeout: 60 * time.Second,
		ROVolumeDir:        getEnv("RO_VOLUME_DIR", roVolumeDir),
		SQLiteDBFile:       getEnv("SQLITE_DB_FILE", sqliteDBFile),
		CMDImgSuffix:       cmdImgSuffix,
		CMDImgNames:        []string{"cmd" + cmdImgSuffix, "cmd-no-bin" + cmdImgSuffix},
		SolutionsKeyPrefix: "s/solutions",
	}
}

func (c *Config) RegistryImgURI(cmdImgName string) string {
	return path.Join(c.RunCmdRegistryImg, cmdImgName) + c.CMDImgSuffix + ":latest"
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
