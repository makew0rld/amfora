package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/makeworld-the-better-one/amfora/cache"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var amforaAppData string // Where amfora files are stored on Windows - cached here
var configDir string
var configPath string

var TofuStore = viper.New()
var tofuDBDir string
var tofuDBPath string

func Init() error {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	// Cache AppData path
	if runtime.GOOS == "windows" {
		appdata, ok := os.LookupEnv("APPDATA")
		if ok {
			amforaAppData = filepath.Join(appdata, "amfora")
		} else {
			amforaAppData = filepath.Join(home, filepath.FromSlash("AppData/Roaming/amfora/"))
		}
	}
	// Cache config directory and file paths
	if runtime.GOOS == "windows" {
		configDir = amforaAppData
	} else {
		// Unix / POSIX system
		configDir = filepath.Join(home, ".config", "amfora")
	}
	configPath = filepath.Join(configDir, "config.toml")

	// Cache TOFU db directory and file paths
	// Windows just stores it in APPDATA along with other stuff
	if runtime.GOOS == "windows" {
		tofuDBDir = amforaAppData
	} else {
		// XDG cache dir on POSIX systems
		tofuDBDir = filepath.Join(home, ".cache", "amfora")
	}
	tofuDBPath = filepath.Join(tofuDBDir, "tofu.toml")

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err == nil {
		// Config file doesn't exist yet, write the default one
		_, err = f.Write(defaultConf)
		if err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	err = os.MkdirAll(tofuDBDir, 0755)
	if err != nil {
		return err
	}
	os.OpenFile(tofuDBPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)

	TofuStore.SetConfigFile(tofuDBPath)
	TofuStore.SetConfigType("toml")
	err = TofuStore.ReadInConfig()
	if err != nil {
		return err
	}

	viper.SetDefault("a-general.home", "gemini.circumlunar.space")
	viper.SetDefault("a-general.http", "default")
	viper.SetDefault("a-general.search", "gus.guru/search")
	viper.SetDefault("a-general.color", true)
	viper.SetDefault("a-general.bullets", true)
	viper.SetDefault("a-general.left_margin", 0.15)
	viper.SetDefault("a-general.max_width", 100)
	viper.SetDefault("cache.max_size", 0)
	viper.SetDefault("cache.max_pages", 20)

	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")
	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	// Setup cache from config
	cache.SetMaxSize(viper.GetInt("cache.max_size"))
	cache.SetMaxPages(viper.GetInt("cache.max_pages"))

	return nil
}
