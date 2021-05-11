package exec

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/kubevirt/terraform-provider-kubevirt/ci-tests/terraform/data"
	"github.com/kubevirt/terraform-provider-kubevirt/ci-tests/terraform/exec/lineprinter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// StateFileName is the default name for Terraform state files.
	StateFileName string = "terraform.tfstate"

	// VarFileName is the default name for Terraform var file.
	VarFileName string = "terraform.tfvars"
)

// Apply unpacks the platform-specific Terraform modules into the
// given directory and then runs 'terraform init' and 'terraform
// apply'.  It returns the absolute path of the tfstate file, rooted
// in the specified directory, along with any errors from Terraform.
func Apply(dir string, testName string, terraformVariables []*TfVarFile) (path string, err error) {
	extraArgs, err := unpackAndInit(dir, testName, terraformVariables)
	if err != nil {
		return "", err
	}

	defaultArgs := []string{
		"-auto-approve",
		"-input=false",
		fmt.Sprintf("-state=%s", filepath.Join(dir, StateFileName)),
		fmt.Sprintf("-state-out=%s", filepath.Join(dir, StateFileName)),
	}
	args := append(defaultArgs, extraArgs...)
	args = append(args, dir)
	sf := filepath.Join(dir, StateFileName)

	lpDebug := &lineprinter.LinePrinter{Print: (&lineprinter.Trimmer{WrappedPrint: logrus.Debug}).Print}
	lpError := &lineprinter.LinePrinter{Print: (&lineprinter.Trimmer{WrappedPrint: logrus.Error}).Print}
	defer lpDebug.Close()
	defer lpError.Close()

	errBuf := &bytes.Buffer{}
	if exitCode := execApply(dir, args, lpDebug, io.MultiWriter(errBuf, lpError)); exitCode != 0 {
		return sf, errors.Wrap(Diagnose(errBuf.String()), "failed to apply Terraform")
	}
	return sf, nil
}

// TfVarFile is a file for terraform variables representing.
type TfVarFile struct {
	// Filename is the name of the file.
	Filename string
	// Data is the contents of the file.
	Data []byte
}

// Destroy unpacks the platform-specific Terraform modules into the
// given directory and then runs 'terraform init' and 'terraform
// destroy'.
func Destroy(dir string, testName string) (err error) {
	// extraArgs, err := unpackAndInit(dir, testName, []*TfVarFile{})
	if err != nil {
		return err
	}

	defaultArgs := []string{
		"-auto-approve",
		"-input=false",
		fmt.Sprintf("-state=%s", filepath.Join(dir, StateFileName)),
		fmt.Sprintf("-state-out=%s", filepath.Join(dir, StateFileName)),
	}
	// args := append(defaultArgs, extraArgs...)
	// args = append(args, dir)

	lpDebug := &lineprinter.LinePrinter{Print: (&lineprinter.Trimmer{WrappedPrint: logrus.Debug}).Print}
	lpError := &lineprinter.LinePrinter{Print: (&lineprinter.Trimmer{WrappedPrint: logrus.Error}).Print}
	defer lpDebug.Close()
	defer lpError.Close()

	if exitCode := execDestroy(dir, defaultArgs, lpDebug, lpError); exitCode != 0 {
		return errors.New("failed to destroy using Terraform")
	}
	return nil
}

// unpack unpacks the specific Terraform modules into the
// given directory.
func unpack(dir string, platform string) (err error) {
	err = data.Unpack(dir, platform)
	if err != nil {
		return err
	}

	// err = data.Unpack(filepath.Join(dir, "config.tf"), "config.tf")
	// if err != nil {
	// 	return err
	// }

	return nil
}

// unpackAndInit unpacks the specific Terraform modules into
// the given directory and then runs 'terraform init'.
func unpackAndInit(dir string, dataDir string, terraformVariables []*TfVarFile) (extraArgs []string, err error) {
	extraArgs, err = prepare(dir, dataDir, terraformVariables)
	if err != nil {
		return []string{}, errors.Wrap(err, "failed to prepare")
	}
	err = unpack(dir, dataDir)
	if err != nil {
		return []string{}, errors.Wrap(err, "failed to unpack Terraform modules")
	}

	if err := setupEmbeddedPlugins(dir); err != nil {
		return []string{}, errors.Wrap(err, "failed to setup embedded Terraform plugins")
	}

	lpDebug := &lineprinter.LinePrinter{Print: (&lineprinter.Trimmer{WrappedPrint: logrus.Debug}).Print}
	lpError := &lineprinter.LinePrinter{Print: (&lineprinter.Trimmer{WrappedPrint: logrus.Error}).Print}
	defer lpDebug.Close()
	defer lpError.Close()

	args := []string{
		"-get-plugins=false",
	}
	args = append(args, dir)
	if exitCode := execInit(dir, args, lpDebug, lpError); exitCode != 0 {
		return []string{}, errors.New("failed to initialize Terraform")
	}
	return extraArgs, nil
}

func prepare(dir string, dataDir string, terraformVariables []*TfVarFile) (extraArgs []string, err error) {
	extraArgs = []string{}
	for _, file := range terraformVariables {
		if err := ioutil.WriteFile(filepath.Join(dir, file.Filename), file.Data, 0600); err != nil {
			return []string{}, err
		}
		extraArgs = append(extraArgs, fmt.Sprintf("-var-file=%s", filepath.Join(dir, file.Filename)))
	}
	return extraArgs, nil
}

func setupEmbeddedPlugins(dir string) error {
	execPath, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "failed to find path for the executable")
	}

	pdir := filepath.Join(dir, "plugins")
	if err := os.MkdirAll(pdir, 0777); err != nil {
		return err
	}
	dst := filepath.Join(pdir, "terraform-provider-kubevirt")
	if runtime.GOOS == "windows" {
		dst = fmt.Sprintf("%s.exe", dst)
	}
	if _, err := os.Stat(dst); err == nil {
		// stat succeeded, the plugin already exists.
		return nil
	}
	logrus.Debugf("Symlinking plugin %s src: %q dst: %q", "terraform-provider-kubevirt", execPath, dst)
	if err := os.Symlink(execPath, dst); err != nil {
		return err
	}
	return nil
}
