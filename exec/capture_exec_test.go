package exec

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gitlab.eng.vmware.com/vivienv/flare/script"
)

func TestExecCAPTURE(t *testing.T) {
	tests := []execTest{
		{
			name: "CAPTURE single command",
			source: func() string {
				return "CAPTURE /bin/echo 'HELLO WORLD'"
			},
			exec: func(s *script.Script) error {
				machine := s.Preambles[script.CmdFrom][0].(*script.FromCommand).Machines()[0].Address
				workdir := s.Preambles[script.CmdWorkDir][0].(*script.WorkdirCommand)
				capCmd := s.Actions[0].(*script.CaptureCommand)

				e := New(s)
				if err := e.Execute(); err != nil {
					return err
				}

				fileName := filepath.Join(workdir.Dir(), machine, fmt.Sprintf("%s.txt", flatCmd(capCmd.GetCliString())))
				if _, err := os.Stat(fileName); err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "CAPTURE multiple commands",
			source: func() string {
				return "CAPTURE /bin/echo 'HELLO WORLD'\nCAPTURE ls ."
			},
			exec: func(s *script.Script) error {
				machine := s.Preambles[script.CmdFrom][0].(*script.FromCommand).Machines()[0].Address
				workdir := s.Preambles[script.CmdWorkDir][0].(*script.WorkdirCommand)
				cmd0 := s.Actions[0].(*script.CaptureCommand)
				cmd1 := s.Actions[1].(*script.CaptureCommand)

				e := New(s)
				if err := e.Execute(); err != nil {
					return err
				}

				fname0 := filepath.Join(workdir.Dir(), machine, fmt.Sprintf("%s.txt", flatCmd(cmd0.GetCliString())))
				fname1 := filepath.Join(workdir.Dir(), machine, fmt.Sprintf("%s.txt", flatCmd(cmd1.GetCliString())))
				if _, err := os.Stat(fname0); err != nil {
					return err
				}
				if _, err := os.Stat(fname1); err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "CAPTURE command with user specified",
			source: func() string {
				uid := os.Getuid()
				return fmt.Sprintf("AS %d \nCAPTURE /bin/echo 'HELLO WORLD'", uid)
			},
			exec: func(s *script.Script) error {
				machine := s.Preambles[script.CmdFrom][0].(*script.FromCommand).Machines()[0].Address
				workdir := s.Preambles[script.CmdWorkDir][0].(*script.WorkdirCommand)
				capCmd := s.Actions[0].(*script.CaptureCommand)

				e := New(s)
				if err := e.Execute(); err != nil {
					return err
				}

				fileName := filepath.Join(workdir.Dir(), machine, fmt.Sprintf("%s.txt", flatCmd(capCmd.GetCliString())))
				if _, err := os.Stat(fileName); err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "CAPTURE command as unknown user",
			source: func() string {
				return "AS foo:barr \nCAPTURE /bin/echo 'HELLO WORLD'"
			},
			exec: func(s *script.Script) error {
				e := New(s)
				if err := e.Execute(); err != nil {
					return err
				}
				return nil
			},
			shouldFail: true,
		},
		{
			name: "CAPTURE bad CLI command",
			source: func() string {
				return "CAPTURE ./ffoobarr'"
			},
			exec: func(s *script.Script) error {
				e := New(s)
				if err := e.Execute(); err != nil {
					return err
				}
				return nil
			},
			shouldFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runExecutorTest(t, test)
		})
	}
}
