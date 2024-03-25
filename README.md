# encore-tui
Terminal user interface for [encore video encoding engine](https://github.com/svt/encore)

# Requirements
* Go needs to be installed. For instructions on how to install Go,
see [here](https://go.dev/doc/install).
* A text editor is needed for creating new jobs, preferably one that runs in the
terminal such as vim, nano or emacs.

# Installation
```
go install github.com/grusell/encore-tui@latest
```

# Usage

## Environment variables

| env var             | description | default value |
|---------------------| --- | --- |
| ENCORE_URL          | url to encore host | http://localhost:8080 |
| EDITOR, ETUI_EDITOR | editor to use for editing jobs | if not present, encore-tui will check for vim, nano and emacs in that order' |

## Running
The app is started by executing `encore-tui` in a terminal.

## Keyboard navigation
Keyboard shortcuts are described in the app. Escape key is used to
return to previous screen.

## License
encore-tui is licensed with [GNU Public License v3](LICENSE)

