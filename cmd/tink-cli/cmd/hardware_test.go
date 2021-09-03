//nolint:thelper // misuse of test helpers requires a large refactor into subtests
package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func Test_NewHardwareCommand(t *testing.T) {
	subCommand := "hardware"
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
				t.Helper()
				root := c.Root()
				root.SetArgs([]string{subCommand})
				if err := root.Execute(); err != nil {
					t.Logf("%+v", root.Args)
					t.Error("expected an error")
				}
			},
		},
		{
			name: "ID",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{subCommand, "id", "-h"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				if !strings.Contains(out.String(), "get hardware by id") {
					t.Error("expected output should include get hardware by id")
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
				if !strings.Contains(out.String(), "list all known hardware") {
					t.Error("expected output should include list all known hardware")
				}
			},
		},
		{
			name: "IP",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{subCommand, "ip", "-h"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				if !strings.Contains(out.String(), "get hardware by any associated ip") {
					t.Error("expected output should include get hardware by any associated ip")
				}
			},
		},
		{
			name: "MAC",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{subCommand, "mac", "-h"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				if !strings.Contains(out.String(), "get hardware by any associated mac") {
					t.Error("expected output should include get hardware by any associated mac")
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
				want := "Deletes one or more resources"
				if !strings.Contains(out.String(), want) {
					t.Error(fmt.Errorf("unexpected output, looking for %q as a substring in %q", want, out.String()))
				}
			},
		},
		{
			name: "Push",
			args: args{name: testCommand},
			want: &cobra.Command{},
			cmdFunc: func(t *testing.T, c *cobra.Command) {
				root := c.Root()
				out := &bytes.Buffer{}
				root.SetArgs([]string{subCommand, "push", "-h"})
				root.SetOutput(out)
				if err := root.Execute(); err != nil {
					t.Error(err)
				}
				if !strings.Contains(out.String(), "push new hardware to tink") {
					t.Error("expected output should include push new hardware to tink")
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
			rootCmd.AddCommand(NewHardwareCommand())
			tt.cmdFunc(t, rootCmd)
		})
	}
}
