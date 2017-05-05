package filecontent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/oscanner/pkg/configuration"
	"github.com/openshift/oscanner/pkg/scanner/factory"
	"github.com/openshift/oscanner/pkg/types"
)

const (
	scannerName = "filecontent"
)

func init() {
	factory.Register(scannerName, NewFileContentScanner)
}

type FileContentScanner struct {
	config *configuration.Configuration
}

func NewFileContentScanner(config *configuration.Configuration) (types.Scanner, error) {
	return &FileContentScanner{
		config: config,
	}, nil
}

func (s *FileContentScanner) Run(targetDir string) error {
	return filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		fmt.Printf("FileContentScanner: %s\n", path)
		return nil
	})
}
