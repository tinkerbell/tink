package vagrant

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
)

type Vagrant struct {
	log                 func(string, ...interface{})
	VagrantBinPath      string
	Name                string
	Workdir             string
	provisioningProcess *os.Process
	async               bool
	sshConn             *ssh.Client
	Stdout              io.ReadWriter
	Stderr              io.ReadWriter
}

type VagrantOpt func(*Vagrant)

func WithStderr(s io.ReadWriter) VagrantOpt {
	return func(v *Vagrant) {
		v.Stderr = s
	}
}

func RunAsync() VagrantOpt {
	return func(v *Vagrant) {
		v.async = true
	}
}

func WithLogger(log func(string, ...interface{})) VagrantOpt {
	return func(v *Vagrant) {
		v.log = log
	}
}

func WithStdout(s io.ReadWriter) VagrantOpt {
	return func(v *Vagrant) {
		v.Stdout = s
	}
}

func WithWorkdir(workdir string) VagrantOpt {
	return func(v *Vagrant) {
		v.Workdir = workdir
	}
}

func WithMachineName(name string) VagrantOpt {
	return func(v *Vagrant) {
		v.Name = name
	}
}

func WithVagrantBinPath(path string) VagrantOpt {
	return func(v *Vagrant) {
		v.VagrantBinPath = path
	}
}

func (v *Vagrant) pipeOutput(ctx context.Context, name string, scanner *bufio.Scanner) {
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			v.log("[pipeOutput %s] %s", name, scanner.Text())
		}
	}
}

func (v *Vagrant) execCmd(ctx context.Context, args ...string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, v.VagrantBinPath, args...)
	cmd.Dir = v.Workdir
	cmd.Stdout = v.Stdout
	cmd.Stderr = v.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("exec error: %v", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("exec error: %v", err)
	}

	go v.pipeOutput(ctx, fmt.Sprintf("%s stderr", cmd.String()), bufio.NewScanner(stderrPipe))
	go v.pipeOutput(ctx, fmt.Sprintf("%s stdout", cmd.String()), bufio.NewScanner(stdoutPipe))

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("exec error: %v", err)
	}

	v.log("executing cmd: %s pid=%d", cmd.String(), cmd.Process.Pid)

	return cmd, err
}

func Up(ctx context.Context, opts ...VagrantOpt) (*Vagrant, error) {
	const (
		defaultVagrantBin = "vagrant"
		defaultName       = "vagrant"
		defaultWorkdir    = "."
	)
	v := &Vagrant{
		VagrantBinPath: defaultVagrantBin,
		Name:           defaultName,
		Workdir:        defaultWorkdir,
		log: func(format string, args ...interface{}) {
			fmt.Println(fmt.Sprintf(format, args))
		},
	}
	for _, opt := range opts {
		opt(v)
	}

	cmd, err := v.execCmd(ctx, "up", "--provision", v.Name)
	if err != nil {
		return nil, err
	}

	v.provisioningProcess = cmd.Process

	if v.async == false {
		err = cmd.Wait()
		if err != nil {
			return nil, fmt.Errorf("exec error \"%s\": %v", err, cmd.String())
		}
	}
	return v, nil
}

func (v *Vagrant) provision(ctx context.Context) error {
	return nil
}

func (v *Vagrant) Destroy(ctx context.Context) error {
	if v.sshConn != nil {
		v.sshConn.Close()
	}

	// A destroy fails if there are any other process locking that particular
	// machine. In our case a worker machine starts async because the
	// provisioning never ends it hangs forever saying "impossible to connect
	// via ssh" when the test succeed and the worker gets destroyed the
	// provisioning process is still up and this prevents the worker to
	// terminate. This procedures kills the provisioning
	if v.provisioningProcess != nil {
		err := v.provisioningProcess.Signal(syscall.Signal(0))
		if err == nil {
			v.log(
				"[Destroy %s]: killing provisioning process because it is still running. pid=%d",
				v.Name,
				v.provisioningProcess.Pid,
			)
			err := syscall.Kill(-v.provisioningProcess.Pid, syscall.SIGKILL)
			if err != nil {
				v.log("error killing provisioning process: %s pid=%d", err, v.provisioningProcess.Pid)
			}
		}
	}

	cmd, err := v.execCmd(ctx, "destroy", "--force", v.Name)
	if err != nil {
		return fmt.Errorf("exec error \"%s\": %v", err, cmd.String())
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("exec error \"%s\": %v", err, cmd.String())
	}
	return nil
}

func (v *Vagrant) Exec(ctx context.Context, args ...string) ([]byte, error) {
	if v.sshConn == nil {
		signer, err := loadPrivateKey()
		if err != nil {
			return nil, err
		}

		config := &ssh.ClientConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			User:            "vagrant",
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer), ssh.Password("vagrant")},
		}

		conn, err := ssh.Dial("tcp", "localhost:2222", config)
		if err != nil {
			return nil, fmt.Errorf("dial failed:%v", err)
		}

		v.sshConn = conn
	}

	session, err := v.sshConn.NewSession()
	if err != nil {
		return nil, err
	}

	defer session.Close()

	v.log("[Exec %s] %s", v.Name, strings.Join(args, " "))

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(strings.Join(args, " "))

	v.log("[Exec %s] %s Output:\n%s", v.Name, strings.Join(args, " "), stdoutBuf.String())

	return stdoutBuf.Bytes(), nil
}

func loadPrivateKey() (ssh.Signer, error) {
	pemBytes, err := ioutil.ReadFile(os.Getenv("HOME") + "/.vagrant.d/insecure_private_key")
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return nil, err
	}

	return signer, nil
}
