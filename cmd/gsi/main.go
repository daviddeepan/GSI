package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/briandowns/spinner"
)

func main() {
	printBanner()
	
	var projectName string
	var framework string
	var db string
	var orm string
	var initGit bool

	askOne(&survey.Input{
		Message: "Project name:",
	}, &projectName)

	askOne(&survey.Select{
		Message: "Choose API framework:",
		Options: []string{"Fiber", "Gorilla Mux", "Chi"},
		Default: "Fiber",
	}, &framework)

	askOne(&survey.Select{
		Message: "Choose database:",
		Options: []string{"Postgres", "MySQL", "SQLite", "None"},
		Default: "Postgres",
	}, &db)

	askOne(&survey.Select{
		Message: "Choose ORM:",
		Options: []string{"GORM", "SQLX", "Ent", "None"},
		Default: "GORM",
	}, &orm)

	askOne(&survey.Confirm{
		Message: "Initialize git repo?",
		Default: true,
	}, &initGit)

	spin := spinner.New(spinner.CharSets[14], 120*time.Millisecond)

	runStep(spin, "Creating project structure...", func() {
		createProject(projectName, framework, db, orm)
	})

	runStep(spin, "Initializing go module...", func() {
		run(projectName, "go", "mod", "init", projectName)
	})

	runStep(spin, "Installing framework...", func() {
		switch framework {
		case "Fiber":
			run(projectName, "go", "get", "github.com/gofiber/fiber/v2")
		case "Gorilla Mux":
			run(projectName, "go", "get", "github.com/gorilla/mux")
		case "Chi":
			run(projectName, "go", "get", "github.com/go-chi/chi/v5")
		}
	})

	if db != "None" {
		runStep(spin, "Installing DB driver...", func() {
			switch db {
			case "Postgres":
				run(projectName, "go", "get", "github.com/lib/pq")
			case "MySQL":
				run(projectName, "go", "get", "github.com/go-sql-driver/mysql")
			case "SQLite":
				run(projectName, "go", "get", "github.com/mattn/go-sqlite3")
			}
		})
	}

	if orm != "None" {
		runStep(spin, "Installing ORM...", func() {
			switch orm {
			case "GORM":
				run(projectName, "go", "get", "gorm.io/gorm")
				if db == "Postgres" {
					run(projectName, "go", "get", "gorm.io/driver/postgres")
				} else if db == "MySQL" {
					run(projectName, "go", "get", "gorm.io/driver/mysql")
				} else if db == "SQLite" {
					run(projectName, "go", "get", "gorm.io/driver/sqlite")
				}
			case "SQLX":
				run(projectName, "go", "get", "github.com/jmoiron/sqlx")
			case "Ent":
				run(projectName, "go", "get", "entgo.io/ent/cmd/ent")
			}
		})
	}

	if initGit {
		runStep(spin, "Initializing git...", func() {
			run(projectName, "git", "init")
		})
	}

	fmt.Println("\n🎉 Project ready!")
	fmt.Println("👉 cd", projectName)
	fmt.Println("👉 go run main.go")
}

func printBanner() {
	fmt.Println(`
   ██████╗  ███████╗ ██╗
  ██╔════╝  ██╔════╝ ██║
  ██║  ███╗ ███████╗ ██║
  ██║   ██║ ╚════██║ ██║
  ╚██████╔╝ ███████║ ██║
   ╚═════╝  ╚══════╝ ╚═╝

        GSI (GoStackInit)
        --- Bootstrap Go APIs fast ---
`)
}

func askOne(p survey.Prompt, response interface{}) {
	err := survey.AskOne(p, response)
	if err == terminal.InterruptErr {
		confirmExit()
	} else if err != nil {
		panic(err)
	}
}

func confirmExit() {
	var quit bool
	survey.AskOne(&survey.Confirm{
		Message: "Do you want to quit?",
		Default: true,
	}, &quit)

	if quit {
		fmt.Println("Exiting GSI. See ya!")
		os.Exit(0)
	}
}



func runStep(s *spinner.Spinner, msg string, fn func()) {
	s.Suffix = " " + msg
	s.Start()
	fn()
	s.Stop()
	fmt.Println("✔", msg)
}

func run(dir string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func createProject(name, framework, db, orm string) {
	os.MkdirAll(filepath.Join(name, "cmd"), 0755)

	mainGo := generateMain(name, framework)
	os.WriteFile(filepath.Join(name, "main.go"), []byte(mainGo), 0644)
}

func generateMain(name, framework string) string {
	if framework == "Fiber" {
		return `package main

import (
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from ` + name + ` 🚀")
	})

	app.Listen(":3000")
}
`
	}

	if framework == "Chi" {
		return `package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from ` + name + ` 🚀"))
	})

	http.ListenAndServe(":3000", r)
}
`
	}

	return `package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from ` + name + ` 🚀"))
	})

	http.ListenAndServe(":3000", r)
}
`
}
