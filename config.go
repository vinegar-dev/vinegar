// Copyright vinegar-development 2023

package main

import (
	"errors"
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
	"runtime"
)

// Thank you ayn2op. Thank you so much.

// Primary struct keeping track of vinegar's directories.
type Directories struct {
	Cache  string
	Config string
	Data   string
	Pfx    string
	Log    string
}

type Configuration struct {
	Renderer  string
	ApplyRCO  bool
	AutoRFPSU bool
	Dxvk      bool
	GameMode  bool
	Env       map[string]string
	FFlags    map[string]any
}

var Dirs = defDirs()
var ConfigFilePath = filepath.Join(Dirs.Config, "config.toml")
var Config = loadConfig()

// Define the default values for the Directories
// struct globally for other functions to use it.
func defDirs() Directories {
	homeDir, err := os.UserHomeDir()
	Errc(err)

	xdgDirs := map[string]string{
		"XDG_CACHE_HOME":  filepath.Join(homeDir, ".cache"),
		"XDG_CONFIG_HOME": filepath.Join(homeDir, ".config"),
		"XDG_DATA_HOME":   filepath.Join(homeDir, ".local", "share"),
	}

	// If the variable has already been set, we
	// should use it instead of our own.
	for varName := range xdgDirs {
		value := os.Getenv(varName)

		if value != "" {
			xdgDirs[varName] = value
		}
	}

	dirs := Directories{
		Cache:  filepath.Join(xdgDirs["XDG_CACHE_HOME"], "vinegar"),
		Config: filepath.Join(xdgDirs["XDG_CONFIG_HOME"], "vinegar"),
		Data:   filepath.Join(xdgDirs["XDG_DATA_HOME"], "vinegar"),
	}

	dirs.Pfx = filepath.Join(dirs.Data, "pfx")
	dirs.Log = filepath.Join(dirs.Cache, "logs")

	// Only these Dirs are queued for creation since
	// the other directories are root directories for those.
	CheckDirs(0755, dirs.Log, dirs.Pfx)

	return dirs
}

// Initial default configuration values
func defConfig() Configuration {
	config := Configuration{
		Renderer:  "D3D11",
		Env:       make(map[string]string),
		FFlags:    make(map[string]any),
		ApplyRCO:  true,
		AutoRFPSU: false,
		Dxvk:      true,
		GameMode:  false,
	}

	// Main environment variables initialization
	// Note: these can be overrided by the user.
	config.Env = map[string]string{
		"WINEPREFIX": Dirs.Pfx,
		"WINEARCH":   "win64", // required for rbxfpsunlocker
		// "WINEDEBUG":     "fixme-all,-wininet,-ntlm,-winediag,-kerberos",
		"WINEDEBUG":        "-all",
		"WINEDLLOVERRIDES": "dxdiagn=d;winemenubuilder.exe=d;",

		"DXVK_LOG_LEVEL":        "warn",
		"DXVK_LOG_PATH":         "none",
		"DXVK_STATE_CACHE_PATH": filepath.Join(Dirs.Cache, "dxvk"),

		// these should be overrided by the user.
		"MESA_GL_VERSION_OVERRIDE":    "4.4",
		"__GL_THREADED_OPTIMIZATIONS": "1",
		"DRI_PRIME":                   "1",
		"__NV_PRIME_RENDER_OFFLOAD":   "1",
		"__VK_LAYER_NV_optimus":       "NVIDIA_only",
		"__GLX_VENDOR_LIBRARY_NAME":   "nvidia",
	}

	return config
}

// Write the configuration file with comments linking to
// the Vinegar documentation.
func writeConfigTemplate() {
	// ~/.config/vinegar may not exist yet!
	CheckDirs(0755, Dirs.Config)

	file, err := os.Create(ConfigFilePath)
	Errc(err)
	defer file.Close()

	template := `
# See how to configure Vinegar on the documentation website:
# https://vinegarhq.github.io/Configuration
`
	_, err = file.WriteString(template[1:]) // ignores first newline
	Errc(err)
}

// Load the default configuration, and then append
// the configuration file to the loaded configuration, which is
// global, this can mean that user's variables and fflags can override.
func loadConfig() Configuration {
	config := defConfig()

	if _, err := os.Stat(ConfigFilePath); errors.Is(err, os.ErrNotExist) {
		writeConfigTemplate()
	} else if err != nil {
		panic(err)
	} else {
		_, err := toml.DecodeFile(ConfigFilePath, &config)
		Errc(err, "Could not parse configuration file.")
	}

	if runtime.GOOS == "freebsd" {
		config.Env["WINEARCH"] = "win32"
		config.Env["WINE_NO_WOW64"] = "1"
	}

	for name, value := range config.Env {
		os.Setenv(name, value)
	}

	return config
}
