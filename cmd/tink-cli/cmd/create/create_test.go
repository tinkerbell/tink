package create_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-cli/cmd/create"
)

func TestCreateCommand(t *testing.T) {
	testcases := []struct {
		Name       string
		Opt        create.Options
		PreExecute func(*cobra.Command)
		Assert     func(*testing.T, io.Reader, error)
	}{
		{
			Name: "Err passing stdin and a file at the same time",
			Opt: create.Options{
				MergeValidateAndCreateFunc: func(t *testing.T) func(io.Reader) (string, error) {
					return func(io.Reader) (string, error) {
						t.Fatal("Didn't expect to end up here")
						return "", nil
					}
				}(t),
			},
			PreExecute: func(cmd *cobra.Command) {
				cmd.SetArgs([]string{"--file", "./input_file.txt"})
				cmd.SetIn(strings.NewReader("ransomstdin"))
			},
			Assert: func(t *testing.T, stdout io.Reader, err error) {
				if !errors.Is(err, create.ErrStdinCollisionWithStdinAndFilePath) {
					t.Errorf("expected err: \"%s\" got \"%s\"", create.ErrStdinCollisionWithStdinAndFilePath, err)
				}
			},
		},
		{
			Name: "Validate that stdin is correctly passed to MergeValidateAndCreateFunc",
			Opt: create.Options{
				MergeValidateAndCreateFunc: func(t *testing.T) func(io.Reader) (string, error) {
					return func(in io.Reader) (string, error) {
						i, err := ioutil.ReadAll(in)
						if err != nil {
							t.Fatal(err)
						}
						if string(i) != "I am an input coming from stdin" {
							t.Fatalf("expecting input: \"I am an input coming from stdin\" got \"%s\"", i)
						}
						return "", nil
					}
				}(t),
			},
			PreExecute: func(cmd *cobra.Command) {
				cmd.SetIn(strings.NewReader("I am an input coming from stdin"))
			},
			Assert: func(t *testing.T, stdout io.Reader, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			Name: "Validate that input file is correctly passed to MergeValidateAndCreateFunc",
			Opt: create.Options{
				MergeValidateAndCreateFunc: func(t *testing.T) func(io.Reader) (string, error) {
					return func(in io.Reader) (string, error) {
						i, err := ioutil.ReadAll(in)
						if err != nil {
							t.Fatal(err)
						}
						if string(i) != "I am a file" {
							t.Fatalf("expecting input: \"I am a file\" got \"%s\"", i)
						}
						return "", nil
					}
				}(t),
			},
			PreExecute: func(cmd *cobra.Command) {
				f, err := ioutil.TempFile(os.TempDir(), "create_test")
				if err != nil {
					t.Fatal(err)
				}
				fmt.Fprint(f, "I am a file")
				t.Cleanup(func() { f.Close() })
				cmd.SetArgs([]string{"--file", f.Name()})
			},
			Assert: func(t *testing.T, stdout io.Reader, err error) {
				if err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, single := range testcases {
		t.Run(single.Name, func(t *testing.T) {
			// Mock the gRPC client becauase it is not needed to
			// test the common part of the create command.
			single.Opt.SetFullClient(&client.FullClient{})

			stdout := bytes.NewBufferString("")
			cmd := create.NewCreateCommand(single.Opt)
			cmd.SetOut(stdout)
			single.PreExecute(cmd)
			err := cmd.Execute()
			single.Assert(t, stdout, err)
		})
	}
}

type req struct {
	Vegetable string
}

func TestCreateCommandWithFlags(t *testing.T) {
	r := req{}
	fs := pflag.NewFlagSet("create", pflag.ContinueOnError)
	fs.StringVar(&r.Vegetable, "vegetable", "", "Specify the vegetable name")

	opt := create.Options{}
	opt.FlagSet = fs
	opt.MergeValidateAndCreateFunc = func(in io.Reader) (string, error) {
		i, err := ioutil.ReadAll(in)
		if err != nil {
			t.Fatal(err)
		}
		if string(i)+r.Vegetable != "I like: tomato" {
			t.Fatalf("expected \"I like: tomato\", got \"%s\"", string(i)+r.Vegetable)
		}
		return "10", nil
	}
	// Mock the gRPC client becauase it is not needed to
	// test the common part of the create command.
	opt.SetFullClient(&client.FullClient{})

	cmd := create.NewCreateCommand(opt)
	stdout := bytes.NewBufferString("")
	cmd.SetOut(stdout)
	cmd.SetIn(bytes.NewBufferString("I like: "))
	cmd.SetArgs([]string{"--vegetable", "tomato"})
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(stdout)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "10\n" {
		t.Fatalf("expected \"10\" got \"%s\"", out)
	}
}
