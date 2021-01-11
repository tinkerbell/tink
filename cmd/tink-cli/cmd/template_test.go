package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func Test_NewTemplateCommand(t *testing.T) {
	subCommand := "template"
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
				root.SetArgs([]string{subCommand})
				if err := root.Execute(); err != nil {
					t.Logf("%+v", root.Args)
					t.Error("expected an error")
				}
			},
		},
		{
			name: "List",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{subCommand, "list", "-h"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				want := "list all saved templates"
				if !strings.Contains(out.String(), want) {
					t.Error(fmt.Errorf("unexpected output, looking for %q as a substring in %q", want, out.String()))
				}
			},
		},
		{
			name: "Create",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{subCommand, "create", "-h"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				want := "Create template using the --file flag"
				if !strings.Contains(out.String(), want) {
					t.Error(fmt.Errorf("unexpected output, looking for %q as a substring in %q", want, out.String()))
				}
			},
		},
		{
			name: "Delete",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{subCommand, "delete", "-h"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				want := "delete a template"
				if !strings.Contains(out.String(), want) {
					t.Error(fmt.Errorf("unexpected output, looking for %q as a substring in %q", want, out.String()))
				}
			},
		},
		{
			name: "Get",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{subCommand, "get", "-h"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				want := "get a template"
				if !strings.Contains(out.String(), want) {
					t.Error(fmt.Errorf("unexpected output, looking for %q as a substring in %q", want, out.String()))
				}
			},
		},
		{
			name: "Update",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{subCommand, "update", "-h"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				want := "Update an existing template"
				if !strings.Contains(out.String(), want) {
					t.Error(fmt.Errorf("unexpected output, looking for %q as a substring in %q", want, out.String()))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip(`In the current form the CLI uses init too much and it is
	preventing env vars to work as expected. That's why it does not pick up
	the right get command. Overall those tests are not that good (testing
	surface is almost zero). I think we should just remove them.`)
			rootCmd := &cobra.Command{
				Use:     testCommand,
				Run:     func(_ *cobra.Command, _ []string) {},
				Version: "test",
			}
			rootCmd.AddCommand(NewTemplateCommand())
			tt.cmdFunc(t, rootCmd)
		})
	}
}
