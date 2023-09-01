# watchgodoc

## ?

watchgodoc is a tool that runs godoc server and watches for changes in the source code. When a change is detected, it kills the server and restarts it.
Also, it relocates the browser to doc of the package that is changed.

```sh
cd your_go_project
go run github.com/umutbasal/watchgodoc@latest
```

 .vscode/tasks.json to open the browser in vscode

```json
{
    "version": "2.0.0",
    "cwd": "${workspaceFolder}",
    "tasks": [
        {
            "label": "Live Go Doc",
            "command": "${input:browser}",
            "problemMatcher": []
        },
    ],
    "inputs": [
        {
            "id": "browser",
            "type": "command",
            "command": "simpleBrowser.show",
            "args": [
                "http://localhost:6060/pkg/"
            ],
        }
    ]
}
```
