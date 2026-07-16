package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

//go:embed config.toml
var embeddedFiles embed.FS

const (
	appTitle   = "Codex 配置助手"
	configName = "config.toml"
)

type appUI struct {
	window   *walk.MainWindow
	status   *walk.Label
	details  *walk.TextEdit
	progress *walk.ProgressBar
	close    *walk.PushButton
}

func main() {
	ui := &appUI{}

	mainWindow := MainWindow{
		AssignTo: &ui.window,
		Title:    appTitle,
		MinSize:  Size{Width: 560, Height: 310},
		Size:     Size{Width: 640, Height: 390},
		Layout:   VBox{MarginsZero: false, Spacing: 10},
		Children: []Widget{
			Label{Text: "Codex 配置助手", Font: Font{PointSize: 15, Bold: true}},
			Label{Text: "正在自动查找并配置本机的 Codex。请保持此窗口打开。"},
			ProgressBar{AssignTo: &ui.progress, MinValue: 0, MaxValue: 100, Value: 5},
			Label{AssignTo: &ui.status, Text: "正在准备..."},
			TextEdit{AssignTo: &ui.details, ReadOnly: true, VScroll: true},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{AssignTo: &ui.close, Text: "关闭", Enabled: false, OnClicked: func() { ui.window.Close() }},
				},
			},
		},
	}
	if err := mainWindow.Create(); err != nil {
		panic(err)
	}

	go ui.configure()
	ui.window.Run()
}

func (ui *appUI) configure() {
	ui.update(10, "正在寻找安装的 Codex...", "正在检查系统配置目录。")

	codexDirectory, err := findCodexDirectory(func(location string) {
		ui.append("正在检查: " + location)
	})
	if err != nil {
		ui.finish(false, "未找到 Codex 配置目录", err.Error())
		return
	}

	ui.update(55, "已找到 Codex", "找到配置目录: "+codexDirectory)
	ui.update(70, "正在配置资源...", "正在备份并写入新的 config.toml。")

	backupPath, err := replaceConfig(codexDirectory)
	if err != nil {
		ui.finish(false, "配置失败", err.Error())
		return
	}

	message := "配置文件已更新: " + filepath.Join(codexDirectory, configName)
	if backupPath != "" {
		message += "\r\n已备份原配置: " + backupPath
	}
	ui.finish(true, "已完成", message)
}

func (ui *appUI) update(progress int, status, detail string) {
	ui.window.Synchronize(func() {
		ui.progress.SetValue(progress)
		ui.status.SetText(status)
		ui.details.AppendText(detail + "\r\n")
	})
}

func (ui *appUI) append(detail string) {
	ui.window.Synchronize(func() {
		ui.details.AppendText(detail + "\r\n")
	})
}

func (ui *appUI) finish(success bool, status, detail string) {
	ui.window.Synchronize(func() {
		if success {
			ui.progress.SetValue(100)
		}
		ui.status.SetText(status)
		ui.details.AppendText(detail + "\r\n")
		ui.close.SetEnabled(true)
	})
}

func findCodexDirectory(report func(string)) (string, error) {
	seen := make(map[string]bool)

	for _, candidate := range preferredDirectories() {
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		report(candidate)
		if directoryExists(candidate) {
			return candidate, nil
		}
	}

	for _, root := range userProfileRoots() {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			candidate := filepath.Join(root, entry.Name(), ".codex")
			if seen[candidate] {
				continue
			}
			seen[candidate] = true
			report(candidate)
			if directoryExists(candidate) {
				return candidate, nil
			}
		}
	}

	return "", errors.New("未在本机用户目录中找到 .codex 文件夹。请先安装并启动一次 Windows 版 Codex 后再运行本程序。")
}

func preferredDirectories() []string {
	candidates := []string{
		os.Getenv("CODEX_HOME"),
		joinEnvPath("USERPROFILE", ".codex"),
		joinEnvPath("HOME", ".codex"),
	}

	return candidates
}

func userProfileRoots() []string {
	roots := make([]string, 0)
	for drive := 'A'; drive <= 'Z'; drive++ {
		root := fmt.Sprintf("%c:\\Users", drive)
		if info, err := os.Stat(root); err == nil && info.IsDir() {
			roots = append(roots, root)
		}
	}
	sort.Strings(roots)
	return roots
}

func joinEnvPath(variable, child string) string {
	if parent := os.Getenv(variable); parent != "" {
		return filepath.Join(parent, child)
	}
	return ""
}

func directoryExists(directory string) bool {
	info, err := os.Stat(directory)
	return err == nil && info.IsDir()
}

func replaceConfig(codexDirectory string) (string, error) {
	config, err := embeddedFiles.ReadFile(configName)
	if err != nil {
		return "", fmt.Errorf("读取内置配置失败: %w", err)
	}

	configPath := filepath.Join(codexDirectory, configName)
	backupPath := ""
	if _, err := os.Stat(configPath); err == nil {
		backupPath = nextBackupPath(configPath)
		if err := copyFile(configPath, backupPath); err != nil {
			return "", fmt.Errorf("备份原配置失败: %w", err)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("读取现有配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, config, 0600); err != nil {
		return "", fmt.Errorf("写入新配置失败: %w", err)
	}

	return backupPath, nil
}

func nextBackupPath(configPath string) string {
	stamp := time.Now().Format("20060102-150405")
	base := configPath + ".backup-" + stamp
	for suffix := 0; ; suffix++ {
		candidate := base
		if suffix > 0 {
			candidate = fmt.Sprintf("%s-%d", base, suffix)
		}
		if _, err := os.Stat(candidate); errors.Is(err, fs.ErrNotExist) {
			return candidate
		}
	}
}

func copyFile(source, destination string) error {
	contents, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, contents, 0600)
}
