package exec

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/kubevirt/terraform-provider-kubevirt/kubevirt"
)

func Serve() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: kubevirt.Provider})
}

func globalPluginDirs(datadir string) ([]string, error) {
	var ret []string
	// Look in ~/.terraform.d/plugins/ , or its equivalent on non-UNIX
	cdir, err := configDir()
	if err != nil {
		return ret, fmt.Errorf("error finding global config directory: %s", err)
	}

	for _, d := range []string{cdir, datadir} {
		machineDir := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
		ret = append(ret, filepath.Join(d, "plugins"))
		ret = append(ret, filepath.Join(d, "plugins", machineDir))
	}

	return ret, nil
}

func configDir() (string, error) {
	dir, err := homeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, ".terraform.d"), nil
}

func homeDir() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try build-in module
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	if user.HomeDir == "" {
		return "", errors.New("blank output")
	}

	return user.HomeDir, nil
}
