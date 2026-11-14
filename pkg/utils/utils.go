package utils

import (
	"fmt"
	"os"
)

func GetBackupFilePath(backupDir, dbName, backupFormat string, noFilestore bool) (string, error) {
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", err
	}
	if noFilestore {
		return fmt.Sprintf("%s/%s_no_fs.%s", backupDir, dbName, backupFormat), nil
	}
	return fmt.Sprintf("%s/%s.%s", backupDir, dbName, backupFormat), nil
}
