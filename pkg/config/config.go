package config

import (
	"path/filepath"
)

var DefaultContent = `{"fastings": []}`
var	file = "fasting.json"
var BaseDB = "FASTING"
var BaseLocal = "DATABASES"
var	BaseBackup = "/media/veikko/VK DATA/"

var LocalFile = filepath.Join(BaseLocal, BaseDB, file)
var BackupFile = filepath.Join(BaseBackup, BaseLocal, BaseDB, file)
