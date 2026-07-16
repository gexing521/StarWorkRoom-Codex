# StarWorkRoom Codex 配置助手

这是给 Windows 版 Codex 用户使用的图形工具，不需要知道 Codex 的安装位置，也不需要运行命令。

## 使用方法

1. 关闭 Codex。
2. 双击 `CodexConfigAssistant.exe`。
3. 等待窗口依次显示“正在寻找安装的 Codex”、“正在配置资源”、“已完成”。
4. 显示“已完成”后，关闭窗口并重新打开 Codex。

程序会优先检查系统设置的 Codex 目录和当前用户目录；未找到时会扫描所有磁盘中用户配置目录下的 `.codex` 文件夹。找到后会自动创建或替换 `config.toml`。

原有的 `config.toml` 不会直接丢失。程序会在同一目录生成一个带时间的备份文件，例如 `config.toml.backup-20260716-091800`。

## 完整教程

[Windows 和 macOS 的 Codex 使用教程（飞书）](https://my.feishu.cn/wiki/IO2Kwc1CLi7KieknLJdcwuc3nQd?from=from_copylink)

首次运行若出现 Windows 安全提示，请确认文件来源可信后，点击“更多信息”再点击“仍要运行”。程序不上传电脑中的文件或配置信息。

`config.toml` 是写入的配置模板；`main.go` 和 `build-windows.bat` 仅供需要自行编译的开发者使用。
