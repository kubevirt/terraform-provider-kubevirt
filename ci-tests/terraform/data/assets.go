package data

import (
	"net/http"
	"os"
)

// Assets contains project assets.
var Assets http.FileSystem

func init() {
	dir := os.Getenv("OPENSHIFT_INSTALL_DATA")
	if dir == "" {
		dir = "../terraform/data"
	}
	Assets = http.Dir(dir)
}
