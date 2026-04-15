// Package reconcilers — SSH update helper.
//
// runSSHUpdate uses golang.org/x/crypto/ssh to SCP the firmware payload to the
// target node and execute the update command. A single 60-second context
// deadline (set by the caller) governs the entire operation.
package reconcilers

import (
"context"
"fmt"
"io"
"net"
"os"
"path/filepath"
"time"

"golang.org/x/crypto/ssh"

v1 "github.com/bmcdonald3/firmware-management-service/fms/apis/firmware.management.io/v1"
"github.com/bmcdonald3/firmware-management-service/fms/pkg/credentials"
)

const credentialsPath = "/tmp/credentials.json"

// runSSHUpdate SCPs the firmware payload from the local FMLS library to the
// target node and then executes a remote flash command over SSH.
func runSSHUpdate(ctx context.Context, dp v1.DeviceProfile, up v1.UpdateProfile, fmsHostIP string) error {
parser, err := credentials.NewParser(credentialsPath)
if err != nil {
return fmt.Errorf("ssh: load credentials: %w", err)
}

cred, err := parser.Get(dp.Metadata.Name)
if err != nil {
return fmt.Errorf("ssh: get credentials for %q: %w", dp.Metadata.Name, err)
}

sshCfg := &ssh.ClientConfig{
User: cred.Username,
Auth: []ssh.AuthMethod{ssh.Password(cred.Password)},
// InsecureIgnoreHostKey is acceptable in closed management networks; a
// production hardening pass should replace this with a known_hosts check.
HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
Timeout:         10 * time.Second,
}

addr := net.JoinHostPort(dp.Spec.ManagementIP, "22")
client, err := ssh.Dial("tcp", addr, sshCfg)
if err != nil {
return fmt.Errorf("ssh: dial %s: %w", addr, err)
}
defer client.Close()

// Step A: SCP the firmware binary to the remote node.
localPath := filepath.Join("/tmp/firmware", filepath.Base(up.Spec.PayloadPath))
if err := scpFile(ctx, client, localPath, "/tmp/"+filepath.Base(up.Spec.PayloadPath)); err != nil {
return fmt.Errorf("ssh: scp upload: %w", err)
}

// Step B: Execute the remote flash command.
flashCmd := fmt.Sprintf("fwupdate --file /tmp/%s", filepath.Base(up.Spec.PayloadPath))
if err := runRemoteCommand(ctx, client, flashCmd); err != nil {
return fmt.Errorf("ssh: remote exec: %w", err)
}

return nil
}

// scpFile copies a local file to a remote path using the SCP protocol over an
// established SSH client connection.
func scpFile(ctx context.Context, client *ssh.Client, localPath, remotePath string) error {
f, err := os.Open(localPath)
if err != nil {
return fmt.Errorf("open local file %q: %w", localPath, err)
}
defer f.Close()

info, err := f.Stat()
if err != nil {
return fmt.Errorf("stat local file %q: %w", localPath, err)
}

session, err := client.NewSession()
if err != nil {
return fmt.Errorf("new SSH session: %w", err)
}
defer session.Close()

// Feed the file into the SCP sink command via stdin pipe.
stdin, err := session.StdinPipe()
if err != nil {
return fmt.Errorf("stdin pipe: %w", err)
}

errCh := make(chan error, 1)
go func() {
defer stdin.Close()
// SCP header: C<mode> <size> <filename>\n
header := fmt.Sprintf("C0644 %d %s\n", info.Size(), filepath.Base(remotePath))
if _, werr := io.WriteString(stdin, header); werr != nil {
errCh <- werr
return
}
if _, werr := io.Copy(stdin, f); werr != nil {
errCh <- werr
return
}
// SCP null-byte terminator.
_, werr := stdin.Write([]byte{0})
errCh <- werr
}()

done := make(chan error, 1)
go func() {
done <- session.Run(fmt.Sprintf("scp -t %s", filepath.Dir(remotePath)))
}()

select {
case <-ctx.Done():
return ctx.Err()
case scpErr := <-errCh:
if scpErr != nil {
return scpErr
}
case runErr := <-done:
return runErr
}

return <-done
}

// runRemoteCommand executes a single shell command on the remote host,
// returning an error if the command exits non-zero or the context expires.
func runRemoteCommand(ctx context.Context, client *ssh.Client, cmd string) error {
session, err := client.NewSession()
if err != nil {
return fmt.Errorf("new SSH session: %w", err)
}
defer session.Close()

done := make(chan error, 1)
go func() {
done <- session.Run(cmd)
}()

select {
case <-ctx.Done():
return ctx.Err()
case err := <-done:
return err
}
}