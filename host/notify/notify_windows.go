//go:build windows

package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

// show uses a PowerShell toast-compatible balloon via Windows Forms NotifyIcon
// as a pragmatic fallback without COM toast XML complexity.
func show(title, body string) error {
	title = escapePS(title)
	body = escapePS(body)
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
$n = New-Object System.Windows.Forms.NotifyIcon
$n.Icon = [System.Drawing.SystemIcons]::Information
$n.Visible = $true
$n.BalloonTipTitle = '%s'
$n.BalloonTipText = '%s'
$n.ShowBalloonTip(4000)
Start-Sleep -Milliseconds 4500
$n.Dispose()
`, title, body)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

func escapePS(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
