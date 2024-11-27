package main

import (
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/PeterHickman/expand_path"
	"github.com/PeterHickman/toolbox"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

const default_database_directory = "~/.config/snippets"
const default_database_file = "snippets.sqlite3"
const real_name = "snip"

type search_result struct {
	nr    int64
	title string
	score float64
}

var database_file string
var db *sql.DB
var option string
var nr int64

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func dropdead(message string) {
	fmt.Println(message)
	os.Exit(1)
}

func isNumber(text string) (int64, error) {
	v, err := strconv.ParseInt(text, 10, 64)
	return v, err
}

func cleanName(filename string) string {
	c := filepath.Base(filename)
	c = strings.ToLower(c)
	c = strings.Replace(c, "_", " ", -1)
	c = strings.Replace(c, ".txt", "", 1)

	return c
}

func newTitle(title string) bool {
	stmt, err := db.Prepare("SELECT nr FROM snippets WHERE title = ?")
	check(err)
	defer stmt.Close()

	var nr int64

	err = stmt.QueryRow(title).Scan(&nr)

	// err is not nil if the record does not exist
	return err != nil
}

func filenameFromTitle(title string) string {
	c := strings.Replace(title, " ", "_", -1)
	c = flag.Arg(0) + "/" + c + ".txt"

	return c
}

func nextNr() int64 {
	rows, err := db.Query("SELECT MAX(nr) AS max_nr FROM snippets")
	check(err)
	defer rows.Close()

	var r int64 = 1
	var max_nr int64

	for rows.Next() {
		err = rows.Scan(&max_nr)
		if err == nil {
			r = max_nr + 1
		}
	}

	err = rows.Err()
	check(err)

	return r
}

func addSnippet(title, filename string) {
	nr := nextNr()

	dat, err := os.ReadFile(filename)
	check(err)

	content := base64.StdEncoding.EncodeToString(dat)

	tx, err := db.Begin()
	check(err)

	stmt, err := tx.Prepare("INSERT INTO snippets(title, nr, content) VALUES (?, ?, ?)")
	check(err)
	defer stmt.Close()

	_, err = stmt.Exec(title, nr, content)
	check(err)

	err = tx.Commit()
	check(err)

	fmt.Printf("#%d : %s\n", nr, title)
}

func exportSnippet(title string, nr int64, content string) {
	filename := filenameFromTitle(title)

	as_text, _ := base64.StdEncoding.DecodeString(content)
	err := os.WriteFile(filename, as_text, 0644)
	check(err)
}

func databaseName() string {
	expanded_default_database_directory, _ := expand_path.ExpandPath(default_database_directory)

	if _, err := os.Stat(expanded_default_database_directory); err != nil {
		err = os.MkdirAll(expanded_default_database_directory, 0755)
		check(err)
	}

	return expanded_default_database_directory + "/" + default_database_file
}

func initialiseDatabase() {
	f, err := os.Create(database_file)
	check(err)
	f.Close()

	db, err = sql.Open("sqlite3", database_file)
	check(err)

	_, err = db.Exec("CREATE TABLE snippets (title TEXT NOT NULL, nr INTEGER NOT NULL, content BLOB NOT NULL)")
	check(err)
	db.Close()
}

func searchTerms() []string {
	l := []string{}

	for _, v := range flag.Args() {
		v = strings.ToLower(v)
		if !slices.Contains(l, v) {
			l = append(l, v)
		}
	}

	return l
}

func searchRecord(search_terms []string, title string) float64 {
	record_terms := []string{}

	for _, v := range strings.Fields(title) {
		v = strings.ToLower(v)
		if !slices.Contains(record_terms, v) {
			record_terms = append(record_terms, v)
		}
	}

	var s float64 = 0.0

	for _, st := range search_terms {
		for _, rt := range record_terms {
			if strings.HasPrefix(rt, st) {
				s += 1.0
				break
			}
		}
	}

	s = s / float64(len(search_terms))

	return s
}

func list() {
	rows, err := db.Query("SELECT title, nr FROM snippets ORDER BY title")
	check(err)

	defer rows.Close()
	for rows.Next() {
		var title string
		var nr int64
		err = rows.Scan(&title, &nr)
		check(err)
		fmt.Printf("#%d : %s\n", nr, title)
	}
	err = rows.Err()
	check(err)
}

func show() {
	stmt, err := db.Prepare("SELECT title, nr, content FROM snippets WHERE nr = ?")
	check(err)
	defer stmt.Close()

	var title string
	var nr int64
	var content string

	err = stmt.QueryRow(flag.Arg(0)).Scan(&title, &nr, &content)
	if err != nil {
		dropdead(fmt.Sprintf("There is no snippet %s", flag.Arg(0)))
	}
	fmt.Printf("#%d : %s\n", nr, title)
	fmt.Println()
	as_text, _ := base64.StdEncoding.DecodeString(content)
	fmt.Println(string(as_text))
}

func delete() {
	stmt, err := db.Prepare("SELECT title, nr FROM snippets WHERE nr = ?")
	check(err)
	defer stmt.Close()

	var title string
	var nr int64

	err = stmt.QueryRow(flag.Arg(0)).Scan(&title, &nr)
	if err != nil {
		dropdead(fmt.Sprintf("There is no snippet %s", flag.Arg(0)))
	}

	_, err = db.Exec(fmt.Sprintf("DELETE FROM snippets WHERE nr = %s", flag.Arg(0)))
	check(err)

	fmt.Printf("Deleting #%d : %s\n", nr, title)
}

func search() {
	search_terms := searchTerms()

	results := []search_result{}

	rows, err := db.Query("SELECT title, nr FROM snippets ORDER BY title")
	check(err)
	defer rows.Close()

	for rows.Next() {
		var title string
		var nr int64
		err = rows.Scan(&title, &nr)
		check(err)
		r := searchRecord(search_terms, title)
		if r > 0.0 {
			results = append(results, search_result{nr: nr, title: title, score: r})
		}
	}
	err = rows.Err()
	check(err)

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	for _, r := range results {
		fmt.Printf("#%d : %s\n", r.nr, r.title)
	}
}

func import_snippets() {
	total := 0
	added := 0

	for _, filename := range flag.Args() {
		total++

		title := cleanName(filename)
		if newTitle(title) {
			addSnippet(title, filename)
			added++
		}
	}

	if added > 0 {
		fmt.Printf("Imported %d files, skipped %d existing\n", added, total-added)
	} else {
		fmt.Println("No files imported")
	}
}

func export_snippets() {
	rows, err := db.Query("SELECT title, nr, content FROM snippets ORDER BY title")
	check(err)

	defer rows.Close()
	for rows.Next() {
		var title string
		var nr int64
		var content string

		err = rows.Scan(&title, &nr, &content)
		check(err)
		exportSnippet(title, nr, content)
	}
	err = rows.Err()
	check(err)
}

func init() {
	var err error

	l := flag.Bool("list", false, "List all the snippets")
	d := flag.Bool("delete", false, "Delete the given snippet")
	i := flag.Bool("import", false, "Import files into the database")
	e := flag.Bool("export", false, "Export the database into files")
	s := flag.Bool("search", false, "Search for snippets")
	f := flag.String("file", "", "Name of an alternate snippets database file")

	flag.Parse()

	if *f != "" {
		database_file = *f
	} else {
		database_file = databaseName()
	}

	fmt.Printf("Using %s\n", database_file)

	if !toolbox.FileExists(database_file) {
		initialiseDatabase()
	}

	db, err = sql.Open("sqlite3", database_file)
	check(err)

	// First check the flags
	if *l {
		option = "list"
	}

	if *d {
		if option == "" {
			if len(flag.Args()) == 1 {
				v, err := isNumber(flag.Arg(0))
				if err == nil {
					nr = v
					option = "delete"
				} else {
					dropdead("Number not supplied")
				}
			} else {
				dropdead("Number not supplied")
			}
		} else {
			dropdead("Only one option at a time")
		}
	}

	if *i {
		if option == "" {
			if len(flag.Args()) >= 1 {
				option = "import"
			} else {
				dropdead("Supply list of files to import")
			}
		} else {
			dropdead("Only one option at a time")
		}
	}

	if *e {
		if option == "" {
			if len(flag.Args()) == 1 {
				if stat, err := os.Stat(flag.Arg(0)); err == nil && stat.IsDir() {
					option = "export"
				} else {
					dropdead("Supply a directory to export into, not a file")
				}
			} else {
				dropdead("Supply a directory to export into")
			}
		} else {
			dropdead("Only one option at a time")
		}
	}

	if *s {
		if option == "" {
			if len(flag.Args()) >= 1 {
				option = "search"
			} else {
				dropdead("Supply terms to search for")
			}
		} else {
			dropdead("Only one option at a time")
		}
	}

	name := filepath.Base(os.Args[0])

	if name != real_name {
		if len(name) == len(real_name)+1 {
			option_char := name[len(real_name):]
			switch option_char {
			case "l":
				option = "list"
			case "d":
				option = "delete"
			case "i":
				option = "import"
			case "e":
				option = "export"
			case "s":
				option = "search"
			default:
				dropdead(fmt.Sprintf("Unknown command name suffix [%s]", option_char))
			}
		}
	}

	// Then see if Arg(0) is a number
	if len(flag.Args()) == 1 && option == "" {
		v, err := isNumber(flag.Arg(0))
		if err == nil {
			nr = v
			option = "show"
		} else {
			dropdead("Number not supplied")
		}
	}

	// Has an option been taken
	if option == "" && nr == 0 {
		dropdead("No option given")
	}
}

func main() {
	switch option {
	case "list":
		list()
	case "show":
		show()
	case "delete":
		delete()
	case "import":
		import_snippets()
	case "export":
		export_snippets()
	case "search":
		search()
	}

	db.Close()
}
