package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	// Create a new ServeMux object to handle HTTP requests. This is the main router for our application.
	mux := http.NewServeMux()

	// Set up a file server to serve static files from the "./ui/static/" directory. The http.FileServer function returns a handler that serves HTTP requests with the contents of the specified file system.
	// Can also serve individual files with http.ServeFile(w, r, "./ui/static/css/main.css")
	fileServer := http.FileServer(http.Dir(app.staticDir))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Route handlers for different URL paths. Each handler function will be called when a request is made to the corresponding path.
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet/view", app.snippetView)
	mux.HandleFunc("/snippet/create", app.snippetCreate)

	return mux
}
