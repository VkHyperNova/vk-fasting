package config

import (
	"path/filepath"
)

var DefaultContent = `{"fastings": []}`

var	file = "fasting.json"
var baseLocal = "DATABASES"
var	baseBackup = "/media/veikko/VK DATA/"

var LocalFile = filepath.Join(baseLocal, "FASTING", file)
var BackupFile = filepath.Join(baseBackup, "DATABASES", "FASTING", file)
