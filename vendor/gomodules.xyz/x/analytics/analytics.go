package analytics

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gomodules.xyz/homedir"
)

func ClientID() string {
	dir := filepath.Join(homedir.HomeDir(), ".appscode")
	filename := filepath.Join(dir, "client-id")
	id, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		id := uuid.New().String()
		if e2 := os.MkdirAll(dir, 0o755); e2 == nil {
			os.WriteFile(filename, []byte(id), 0o644)
		}
		return id
	}
	return string(id)
}
