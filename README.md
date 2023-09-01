# watchgodoc

## ?

watchgodoc is a tool that runs godoc server and watches for changes in the source code. When a change is detected, it kills the server and restarts it.
Also, it relocates the browser to doc of the package that is changed.

```sh
cd your_go_project
go run github.com/umutbasal/watchgodoc@latest
```

 .vscode/tasks.json to open the browser in vscode (CMD + SHIFT + P > RUN TASK > LIVE GO DOC)

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


<img width="1422" alt="image" src="https://github.com/umutbasal/watchgodoc/assets/21194079/38ce8d70-aa57-4104-94be-022835b6239a">
