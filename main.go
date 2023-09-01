package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/valyala/fasthttp"
)

func getpwd() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

var process *exec.Cmd

// start starts the godoc server
func start() {
	os.Chdir(getpwd())

	process = exec.Command("godoc", "-http=:8080")
	process.Dir = "."
	process.Stdout = os.Stdout
	process.Stderr = os.Stderr
	process.Start()
}

func stop() {
	process.Process.Kill()
}

func status() bool {
	fmt.Println(process)
	return process != nil && process.Process != nil
}

func restart() {
	if status() {
		stop()
	}
	start()
}

var location = ""

// proxyServer is a proxy server that redirects to the godoc server
func proxyServer() {
	app := fiber.New()

	proxy.WithClient(&fasthttp.Client{
		NoDefaultUserAgentHeader: true,
		DisablePathNormalizing:   true,
	})

	app.Get("/location", func(c *fiber.Ctx) error {
		loc := location
		location = ""
		return c.JSON(fiber.Map{
			"location": loc,
		})
	})

	app.Use(proxy.Balancer(proxy.Config{
		Servers: []string{
			"http://localhost:8080",
		},
		ModifyResponse: func(c *fiber.Ctx) error {
			if strings.Contains(string(c.Response().Header.Peek("Content-Type")), "text/html") {
				res := string(c.Response().Body())
				res = res + `<script>
				setInterval(function() {
					fetch('/location')
					.then(res => res.json())
					.then(data => {
						if (data.location != "" && data.location != window.location.href) {
							window.location.href = data.location
						}
					})
					if (window.document.body.innerText.includes("Scan is not yet complete")) {
						window.location.reload()
					}
				}, 1000)
				</script>
				`
				c.Response().SetBodyString(res)

				if c.Response().StatusCode() != 200 {
					c.Response().SetStatusCode(200)
					c.Response().SetBodyString(res)
				}
			}
			return nil
		},
	}))

	app.Listen(":6060")

}

// getDirsToWatch returns a list of directories to watch
func getDirsToWatch() []string {
	dirs := []string{"."}
	dirs = append(dirs, getDirs(".")...)
	return dirs
}

// hasGoFile checks if the directory contains a .go source file
func hasGoFile(dir string) bool {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".go") {
			return true
		}
	}
	return false
}

// getDirs
// traverse the current directory
// and add all directories to the list
// that are not hidden and not in the .git directory
// except the .git / vendor / node_modules directories
// and the .gitignore file
// and if they do not contain a .go source file
func getDirs(dir string) []string {
	dirs := []string{}
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") && file.Name() != "vendor" && file.Name() != "node_modules" && file.Name() != ".git" && file.Name() != ".idea" && file.Name() != ".vscode" && file.Name() != ".gitignore" && hasGoFile(file.Name()) {
			dirs = append(dirs, file.Name())
			dirs = append(dirs, getDirs(file.Name())...)
		}
	}
	return dirs
}

// readModuleName reads the module name from the go.mod file
func readModuleName() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module") {
			return strings.TrimSpace(strings.Split(line, "module")[1])
		}
	}
	return ""
}

// main is the entry point for the application.
func main() {
	start()
	go proxyServer()

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// model/app.go -> /model
	// pkg/fn.go -> /pkg
	// pkg/a/a/a/a/app.go -> /pkg/a/a/a/a
	// /main.go -> /
	// main.go ->
	fname := regexp.MustCompile(`[^/]+$`)

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Has(fsnotify.Write) {
					log.Printf("modified file: %s", event.Name)
				}
				location = "http://localhost:6060/pkg/" + readModuleName() + "/" + fname.ReplaceAllString(event.Name, "")
				restart()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Add directories to watcher.
	dirs := getDirsToWatch()
	for _, dir := range dirs {
		err = watcher.Add(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	// catch ctrl+c
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		stop()
		os.Exit(0)
	}()

	// Wait forever.
	select {}
}
