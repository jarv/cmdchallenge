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

	tag := os.Getenv("CMD_IMAGE_TAG")
	if tag == "" {
		tag = "latest"
	}

	return &Config{
		RunCmdRegistryImgTag: tag,
		RunCmdRegistryImg:    "registry.gitlab.com/jarv/cmdchallenge",
		RegistryAuth:         "",
		RunCmdTimeout:        6 * time.Second,
		PullImageTimeout:     5 * time.Minute,
		RemoveImageTimeout:   60 * time.Second,
		ROVolumeDir:          getEnv("RO_VOLUME_DIR", roVolumeDir),
		SQLiteDBFile:         getEnv("SQLITE_DB_FILE", sqliteDBFile),
		CMDImgNames:          []string{"cmd", "cmd-no-bin"},
		SolutionsKeyPrefix:   "s/solutions",
	}
}

func (c *Config) RegistryImgURI(imgName string) string {
	return path.Join(c.RunCmdRegistryImg, imgName) + ":" + c.RunCmdRegistryImgTag
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
