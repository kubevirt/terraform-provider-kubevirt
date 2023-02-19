package terraform

import (
	"bytes"
	"context"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kubevirt/terraform-provider-kubevirt/ci-tests/terraform/data"
	"github.com/kubevirt/terraform-provider-kubevirt/ci-tests/terraform/lineprinter"
)

func Init(tfWorkDir string, testName string, tfExecPath string) error {
	if err := unpackTerraformExecDir(tfWorkDir, testName); err != nil {
		return err
	}

	tf, err := newTFExec(tfWorkDir, tfExecPath)
	if err != nil {
		return errors.Wrap(err, "failed to create a new tfexec")
	}

	return errors.Wrap(
		// tf.Init(context.Background(), tfexec.PluginDir(filepath.Join(tfWorkDir, "plugins"))),
		tf.Init(context.Background()),
		"failed doing terraform init",
	)
}

// Apply unpacks the platform-specific Terraform modules into the
// given directory and then runs 'terraform init' and 'terraform
// apply'.
func Apply(tfWorkDir string, tfExecPath string, terraformVariables []*TfVarFile, extraOpts ...tfexec.ApplyOption) error {
	tf, err := newTFExec(tfWorkDir, tfExecPath)
	if err != nil {
		return errors.Wrap(err, "failed to create a new tfexec")
	}

	for _, file := range terraformVariables {
		if err := ioutil.WriteFile(filepath.Join(tfWorkDir, file.Filename), file.Data, 0600); err != nil {
			return err
		}
		extraOpts = append(extraOpts, tfexec.VarFile(filepath.Join(tfWorkDir, file.Filename)))
	}

	return errors.Wrap(
		diagnoseApplyError(tf.Apply(context.Background(), extraOpts...)),
		"failed to apply Terraform",
	)
}

// Destroy unpacks the platform-specific Terraform modules into the
// given directory and then runs 'terraform init' and 'terraform
// destroy'.
func Destroy(tfWorkDir string, tfExecPath string, extraOpts ...tfexec.DestroyOption) error {
	tf, err := newTFExec(tfWorkDir, tfExecPath)
	if err != nil {
		return errors.Wrap(err, "failed to create a new tfexec")
	}

	return errors.Wrap(
		tf.Destroy(context.Background(), extraOpts...),
		"failed doing terraform destroy",
	)
}

// TfVarFile is a file for terraform variables representing.
type TfVarFile struct {
	// Filename is the name of the file.
	Filename string
	// Data is the contents of the file.
	Data []byte
}

// newTFExec creates a tfexec.Terraform for executing Terraform CLI commands.
// The `tfWorkDir` is the location to which the terraform plan (tf files, etc) has been unpacked.
// The `tfExecPath` is the location to which Terraform, provider binaries, & .terraform data dir have been unpacked.
// The stdout and stderr will be sent to the logger at the debug and error levels,
// respectively.
func newTFExec(tfWorkDir string, tfExecPath string) (*tfexec.Terraform, error) {
	tf, err := tfexec.NewTerraform(tfWorkDir, tfExecPath)
	if err != nil {
		return nil, err
	}

	// terraform-exec will not accept debug logs unless a log file path has
	// been specified. And it makes sense since the logging is very verbose.
	if path, ok := os.LookupEnv("TF_LOG_PATH"); ok {
		// These might fail if tf cli does not have a compatible version. Since
		// the exact same check is repeated, we just have to verify error once
		// for all calls
		if err := tf.SetLog(os.Getenv("TF_LOG")); err != nil {
			// We want to skip setting the log path since tf-exec lib will
			// default to TRACE log levels which can risk leaking sensitive
			// data
			logrus.Infof("Skipping setting terraform log levels: %v", err)
		} else {
			tf.SetLogCore(os.Getenv("TF_LOG_CORE"))         //nolint:errcheck
			tf.SetLogProvider(os.Getenv("TF_LOG_PROVIDER")) //nolint:errcheck
			// This never returns any errors despite its signature
			tf.SetLogPath(path) //nolint:errcheck
		}
	}

	// Add terraform info logs to the installer log
	lpDebug := &lineprinter.LinePrinter{Print: (&lineprinter.Trimmer{WrappedPrint: logrus.Debug}).Print}
	lpError := &lineprinter.LinePrinter{Print: (&lineprinter.Trimmer{WrappedPrint: logrus.Error}).Print}
	defer lpDebug.Close()
	defer lpError.Close()

	tf.SetStdout(lpDebug)
	tf.SetStderr(lpError)
	tf.SetLogger(newPrintfer())

	// Set the Terraform data dir to be the same as the dir so that
	// files we unpack are contained and, more importantly, we can ensure the
	// provider binaries unpacked in the Terraform data dir have the same permission
	// levels as the Terraform binary.
	dd := path.Join(tfWorkDir, ".terraform")
	os.Setenv("TF_DATA_DIR", dd)

	return tf, nil
}

// unpack unpacks the platform-specific Terraform modules into the
// given directory.
func unpackTerraformExecDir(tfWorkDir string, testName string) (err error) {
	// 1. Copy terraform files (man.tf & variables.tf) from data dir to exec dir
	err = data.Unpack(tfWorkDir, testName)
	if err != nil {
		return err
	}
	// 2. Add versions file
	if err := addVersionsFiles(tfWorkDir); err != nil {
		return errors.Wrap(err, "failed to write versions.tf files")
	}

	return nil
}

type Provider struct {
	// Name of the provider.
	Name string
	// Source of the provider.
	Source string
}

const versionFileTemplate = `terraform {
  required_version = ">= 1.0.0"
  required_providers {
{{- range .}}
    {{.Name}} = {
      source = "{{.Source}}"
    }
{{- end}}
  }
}
`

func addVersionsFiles(tfWorkDir string) error {
	providers := []Provider{
		{
			Name:   "kubevirt",
			Source: "terraform.local/local/kubevirt",
		},
	}
	tmpl := template.Must(template.New("versions").Parse(versionFileTemplate))
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, providers); err != nil {
		return errors.Wrap(err, "could not create versions.tf from template")
	}
	return addFileToAllDirectories("versions.tf", buf.Bytes(), tfWorkDir)
}

func addFileToAllDirectories(name string, data []byte, tfWorkDir string) error {
	if err := os.WriteFile(filepath.Join(tfWorkDir, name), data, 0666); err != nil {
		return err
	}
	entries, err := os.ReadDir(tfWorkDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if err := addFileToAllDirectories(name, data, filepath.Join(tfWorkDir, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}
