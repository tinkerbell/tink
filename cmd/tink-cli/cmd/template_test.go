package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func Test_templateCmd(t *testing.T) {
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
				if !strings.Contains(out.String(), "list all saved templates") {
					t.Error("expected output should include list all saved templates")
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
				if !strings.Contains(out.String(), "create a workflow template") {
					t.Error("expected output should include create a workflow template")
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
				if !strings.Contains(out.String(), "delete a template") {
					t.Error("expected output should include delete a template")
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
				if !strings.Contains(out.String(), "get a template") {
					t.Error("expected output should include get a template")
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
				if !strings.Contains(out.String(), "update a template") {
					t.Error("expected output should include update a template")
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
			rootCmd.AddCommand(templateCmd)
			tt.cmdFunc(t, rootCmd)
		})
	}
}
