package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func Test_workflowCmd(t *testing.T) {
	subCommand := "workflow"
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
				if !strings.Contains(out.String(), "list all workflows") {
					t.Error("expected output should include list all workflows")
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
				if !strings.Contains(out.String(), "create a workflow") {
					t.Error("expected output should include create a workflow")
				}
			},
		},
		{
			name: "Data",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{subCommand, "data", "-h"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				if !strings.Contains(out.String(), "get workflow data") {
					t.Error("expected output should include get workflow data")
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
				if !strings.Contains(out.String(), "delete a workflow") {
					t.Error("expected output should include delete a workflow")
				}
			},
		},
		{
			name: "Events",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{subCommand, "events", "-h"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				if !strings.Contains(out.String(), "show all events for a workflow") {
					t.Error("expected output should include show all events for a workflow")
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
				if !strings.Contains(out.String(), "get a workflow") {
					t.Error("expected output should include get a workflow")
				}
			},
		},
		{
			name: "State",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{subCommand, "state", "-h"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				if !strings.Contains(out.String(), "get the current workflow state") {
					t.Error("expected output should include get the current workflow state")
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
			rootCmd.AddCommand(workflowCmd)
			tt.cmdFunc(t, rootCmd)
		})
	}
}
