package server

import (
	"html/template"
	"log"
	"net/http"
)

const (
	homeResource = "/"
)

func home(issueHist *IssueIndex, w http.ResponseWriter, r *http.Request) {
	snapshot := issueHist.GetSnapshot()
	page := struct {
		Bugs      map[string]int
		Untriaged map[string]int
	}{
		Bugs:      snapshot.Bugs().CumulativeCount(),
		Untriaged: snapshot.Untriaged().CumulativeCount(),
	}

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Println(err)
		return
	}
	t.Execute(w, page)

	// fmt.Fprintln(w, `<head>`)
	// fmt.Fprintln(w, `  <script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.7.1/Chart.bundle.js"></script>`)
	// fmt.Fprintln(w, `</head>`)
	// fmt.Fprintln(w, `<html>`)
	// fmt.Fprintln(w, `  <body>`)
	// fmt.Fprintln(w, `    <h1>Bugs</h1>`)

	// snapshot := issueHist.Snapshot()

	// for _, issue := range snapshot.Bugs() {
	// 	fmt.Fprintln(w, `    <p>`)
	// 	fmt.Fprintln(w, `      <ul>`)
	// 	fmt.Fprintf(w, "        <li>%s</li>\n", issue.GetTitle())
	// 	fmt.Fprintln(w, `      </ul>`)
	// 	fmt.Fprintln(w, `    </p>`)
	// }
	// fmt.Fprintln(w, `  </body>`)
	// fmt.Fprintln(w, `</html>`)
}

func Serve(port string) {
	// Periodically poll GitHub for issues.
	issueHist, quit := pollGitHub()
	defer close(quit)

	http.HandleFunc(homeResource, func(w http.ResponseWriter, r *http.Request) { home(issueHist, w, r) })

	// Start server.
	log.Printf("Starting server. Listening on port '%s'\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
