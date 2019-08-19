package exec

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"gitlab.eng.vmware.com/vivienv/flare/script"
)

type Executor struct {
	script *script.Script
}

func New(src *script.Script) *Executor {
	return &Executor{script: src}
}

func (e *Executor) Execute() error {
	logrus.Info("Executing flare file")
	// setup FROM
	fromCmds, ok := e.script.Preambles[script.CmdFrom]
	if !ok {
		return fmt.Errorf("Script missing valid %s", script.CmdFrom)
	}
	fromCmd := fromCmds[0].(*script.FromCommand)

	// setup AS instruction
	asCmds, ok := e.script.Preambles[script.CmdAs]
	if !ok {
		return fmt.Errorf("Script missing valid %s", script.CmdAs)
	}
	asCmd := asCmds[0].(*script.AsCommand)
	asUid, asGid, err := asCmd.GetCredentials()
	if err != nil {
		return err
	}

	// setup WORKDIR
	dirs, ok := e.script.Preambles[script.CmdWorkDir]
	if !ok {
		return fmt.Errorf("Script missing valid %s", script.CmdWorkDir)
	}
	workdir := dirs[0].(*script.WorkdirCommand)
	if err := os.MkdirAll(workdir.Dir(), 0744); err != nil && !os.IsExist(err) {
		return err
	}
	logrus.Debugf("Using workdir %s", workdir.Dir())

	// setup ENV
	var envPairs []string
	envCmds := e.script.Preambles[script.CmdEnv]
	for _, envCmd := range envCmds {
		env := envCmd.(*script.EnvCommand)
		if len(env.Envs()) > 0 {
			for _, arg := range env.Envs() {
				envPairs = append(envPairs, arg)
			}
		}
	}

	// process action for each FROM source

	for _, fromSrc := range fromCmd.Sources() {

		for _, action := range e.script.Actions {
			switch cmd := action.(type) {
			case *script.CopyCommand:
				// TODO - COPY uses a go implementation which means uid/guid
				// for the COPY cmd cannot be applied using the flare file.
				// This may need to be changed to a os/cmd external call

				// walk each arg and copy to workdir
				for _, path := range cmd.Args() {
					if relPath, err := filepath.Rel(workdir.Dir(), path); err == nil && !strings.HasPrefix(relPath, "..") {
						logrus.Errorf("%s path %s cannot be relative to workdir %s", cmd.Name(), path, workdir.Dir())
						continue
					}
					logrus.Debugf("Copying content from %s", path)

					err := filepath.Walk(path, func(file string, finfo os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						//TODO subpath calculation flattens the file source, that's wrong.
						// subpath should include full path of file, not just the base.
						subpath := filepath.Join(workdir.Dir(), filepath.Base(file))
						switch {
						case finfo.Mode().IsDir():
							if err := os.MkdirAll(subpath, 0744); err != nil && !os.IsExist(err) {
								return err
							}
							logrus.Debugf("Created subpath %s", subpath)
							return nil
						case finfo.Mode().IsRegular():
							logrus.Debugf("Copying %s -> %s", file, subpath)
							srcFile, err := os.Open(file)
							if err != nil {
								return err
							}
							defer srcFile.Close()

							desFile, err := os.Create(subpath)
							if err != nil {
								return err
							}
							n, err := io.Copy(desFile, srcFile)
							if closeErr := desFile.Close(); closeErr != nil {
								return closeErr
							}
							if err != nil {
								return err
							}

							if n != finfo.Size() {
								return fmt.Errorf("%s did not complet for %s", cmd.Name, file)
							}
						default:
							return fmt.Errorf("%s unknown file type for %s", cmd.Name, file)
						}
						return nil
					})

					if err != nil {
						logrus.Error(err)
					}
				}
			case *script.CaptureCommand:
				// capture command output
				cmdStr := cmd.GetCliString()
				logrus.Debugf("Parsing CLI command %v", cmdStr)
				cliCmd, cliArgs := cmd.GetParsedCli()
				cmdReader, err := CliRun(uint32(asUid), uint32(asGid), envPairs, cliCmd, cliArgs...)
				if err != nil {
					return err
				}
				fileName := fmt.Sprintf("%s.txt", flatCmd(cmdStr))
				filePath := filepath.Join(workdir.Dir(), fileName)
				logrus.Debugf("Capturing command out: [%s] -> %s", cmdStr, filePath)
				if err := writeFile(cmdReader, filePath); err != nil {
					return err
				}
			default:
			}
		}
	}
	return nil
}

func writeFile(source io.Reader, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := io.Copy(file, source); err != nil {
		return err
	}
	return nil
}
