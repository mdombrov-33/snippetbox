package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql" // prefix it with underscore because we're not using it directly. We need init() function from this so that it can register itself with the database/sql package.
	"github.com/mdombrov-33/snippetbox/internal/models"
)

// Using the config struct to hold the command-line flag values. Same as addr := flag.String("addr", ":4000", "HTTP network address") but more organized and scalable for larger applications.
type config struct {
	addr      string
	staticDir string
	dsn       string
}

// Struct to hold the dependencies for dependency injection into handlers. In this case, we have two loggers for informational messages and errors. This allows us to easily pass these dependencies to our handlers without using global variables.
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	snippets *models.SnippetModel
	config
}

func main() {
	var cfg config

	// Idiomatic way to define a command-line flag in Go that gets read at runtime. The flag is named "addr", has a default value of ":4000", and a description "HTTP network address".
	// addr := flag.String("addr", ":4000", "HTTP network address")
	// In real apps still prefer to use env vars.
	flag.StringVar(&cfg.addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&cfg.staticDir, "static-dir", "./ui/static", "Path to static assets")
	// parseTime=true means we force the MySQL driver to convert TIME and DATE fields to time.Time. Otherwise it returns these as []byte objects. This is specific to this driver.
	flag.StringVar(&cfg.dsn, "dsn", "web:2906@/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	// Create two log.Logger instances for logging informational messages and errors. The infoLog will write to standard output (os.Stdout) and the errorLog will write to standard error (os.Stderr). Both loggers will include the date and time in their output.
	// Use command like go run ./cmd/web >>/tmp/info.log 2>>/tmp/error.log to redirect the logs to separate files.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// DB
	db, err := openDB(cfg.dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// Also defer db db.Close(), so that the connection pool is closed before the main() exits.
	defer db.Close()

	// Initialize a new instance of our application struct, containing our dependencies.
	// cfg is a temporary staging variable — flags are parsed into it first, then
	// embedded into app. After this point everything is accessed through app.
	// This pattern will not work if our handlers are spread across multiple packages.
	// In that case, an alternative approach is to create a config package exporting an Application struct and have your handler functions close over this to form a closure. Very roughly:

	// func main() {
	// 	main() {
	//     app := &config.Application{
	//         ErrorLog: log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	//     }

	//     mux.Handle("/", examplePackage.ExampleHandler(app))
	// }

	// func ExampleHandler(app *config.Application) http.HandlerFunc {
	//     return func(w http.ResponseWriter, r *http.Request) {
	//         ...
	//         ts, err := template.ParseFiles(files...)
	//         if err != nil {
	//             app.ErrorLog.Println(err.Error())
	//             http.Error(w, "Internal Server Error", 500)
	//             return
	//         }
	//         ...
	//     }
	// }

	// https://gist.github.com/alexedwards/5cd712192b4831058b21 - more examples of how to use closure pattern for dependency injection in Go.

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		snippets: &models.SnippetModel{DB: db},
		config:   cfg,
	}

	// Custom HTTP server. We create it mostly because default ListenAndServe uses default error logger and we want to use custom one.
	srv := &http.Server{
		Addr:     app.addr,
		ErrorLog: app.errorLog,
		Handler:  app.routes(),
	}

	app.infoLog.Printf("Starting server on %s", app.addr)
	err = srv.ListenAndServe()
	app.errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Ping is actually initializing the connection pool. sql.Open() does not establish any connections to the database by itself, it just prepares the database connection pool. The first actual connection to the database is established when you call db.Ping() or when you execute a query. By calling db.Ping() immediately after opening the database, we can verify that the connection details are correct and that the database is reachable before we start handling any HTTP requests.
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
