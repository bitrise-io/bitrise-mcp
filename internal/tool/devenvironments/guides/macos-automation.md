# macOS session automation from the shell

Most macOS UI actions on a Dev Environments session are scriptable with `bitrise_devenv_execute`. One deterministic command beats a screenshot + coordinate-estimation + click chain: it is faster, cheaper, and cannot miss a coordinate. Reach for the GUI tools (`bitrise_devenv_screenshot`, `_click`, `_type`, `_scroll`, `_mouse_drag`) only when no scriptable path exists, for example inside a third-party app's custom canvas.

## Recipes

- **Open a System Settings pane directly**
  ```
  open "x-apple.systempreferences:com.apple.Network-Settings.extension"
  open "x-apple.systempreferences:com.apple.Displays-Settings.extension"
  open "x-apple.systempreferences:com.apple.Wi-Fi-Settings.extension"
  open "x-apple.systempreferences:com.apple.Accessibility-Settings.extension"
  ```
  Each built-in pane has a bundle id under `/System/Library/ExtensionKit/Extensions` or `/System/Applications/System Settings.app/Contents/Resources`.
- **Launch or focus an app**: `open -a "Safari"` or `osascript -e 'tell application "Safari" to activate'`.
- **Open a URL or file**: `open "https://example.com"`, `open ~/Downloads`.
- **Menu bar, buttons, dialogs** with System Events:
  ```
  osascript -e 'tell application "System Events" to tell process "Safari" to click menu item "New Window" of menu 1 of menu bar item "File" of menu bar 1'
  ```
- **Keystrokes and shortcuts**:
  ```
  timeout 15s osascript -e 'tell application "System Events" to keystroke "hello"'
  timeout 15s osascript -e 'tell application "System Events" to keystroke "t" using {command down}'
  ```
- **Read or change settings**: `defaults read` / `defaults write`, e.g. `defaults write com.apple.dock autohide -bool true && killall Dock`.
- **State checks**: `defaults read`, or System Events queries for the frontmost app and visible windows, are usually faster than reading pixels off a screenshot.
- **Move, select, resize**: `mv` for files, `osascript` + System Events for window positioning, `defaults write` for settings, instead of `bitrise_devenv_mouse_drag`.

## osascript timeout safety net

The common automation scopes (Automation, Accessibility, Screen Recording) are pre-approved on session images, so osascript normally runs without a prompt. An uncommon action can still trigger a TCC permission dialog, and with no human to click "Allow" the command blocks until the execute time cap. Wrap osascript calls in a short timeout so they fail fast and you can fall back to the GUI tools:

```
timeout 15s osascript -e 'tell application "System Events" to ...'
```

## Coordinates when you do click

Call `bitrise_devenv_screenshot` first so the server learns the real screen resolution, then pass `x`/`y` in the coordinate space of the screenshot you look at together with `max_x`/`max_y` (that view's width and height); the server rescales to the real screen.
