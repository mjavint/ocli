package utils

import (
	"fmt"
	"os"
)

func GetBackupFilePath(backupDir, dbName, backupFormat string, noFilestore bool) string {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return ""
	}
	if noFilestore {
		return fmt.Sprintf("%s/%s_no_fs.%s", backupDir, dbName, backupFormat)
	}
	return fmt.Sprintf("%s/%s.%s", backupDir, dbName, backupFormat)
}
