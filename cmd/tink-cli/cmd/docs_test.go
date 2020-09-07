package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

const (
	testCommand = "tink-cli"
)

func Test_docsCmd(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *cobra.Command
		cmdFunc func(*testing.T, *cobra.Command)
	}{
		{
			name: "NoArgs",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				root.SetArgs([]string{"docs"})
				if err := root.Execute(); err == nil {
					t.Error("expected an error")
				}
			},
		},
		{
			name: "Help",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{"docs", "--help"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				if !strings.Contains(out.String(), "markdown") {
					t.Error("expected help to include markdown")
				}
			},
		},
		{
			name: "Markdown",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				dir, err := ioutil.TempDir("", "tink-test-*")
				if err != nil {
					t.Fatal(err)
				}
				defer os.RemoveAll(dir)

				root := c.Root()
				root.SetArgs([]string{"docs", "markdown", "--path", dir})

				if err := root.Execute(); err != nil {
					t.Error(err)
				}

				expectFile := testCommand + ".md"
				_, err = os.Stat(path.Join(dir, expectFile))

				if os.IsNotExist(err) {
					t.Errorf("expected to create %s: %s", expectFile, err)
				}

				if err != nil {
					t.Error(err)
				}
			},
		},
		{
			name: "Man",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				dir, err := ioutil.TempDir("", "tink-test-*")
				if err != nil {
					t.Fatal(err)
				}
				defer os.RemoveAll(dir)

				root := c.Root()
				root.SetArgs([]string{"docs", "man", "--path", dir})

				if err := root.Execute(); err != nil {
					t.Error(err)
				}

				expectFile := testCommand + ".1"
				_, err = os.Stat(path.Join(dir, expectFile))

				if os.IsNotExist(err) {
					t.Errorf("expected to create %s: %s", expectFile, err)
				}

				if err != nil {
					t.Error(err)
				}
			},
		},
		{
			name: "BadFormat",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				root.SetArgs([]string{"docs", "invalid"})
				if err := root.Execute(); err == nil {
					t.Error("expected error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := &cobra.Command{
				Use:     testCommand,
				Run:     func(_ *cobra.Command, _ []string) {},
				Version: "test",
			}
			cmd := docsCmd(tt.args.name)
			rootCmd.AddCommand(cmd)
			tt.cmdFunc(t, cmd)
		})
	}
}
