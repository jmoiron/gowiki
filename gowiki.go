package main

import (
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/pat"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/mandira"
	"github.com/jmoiron/modl"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/russross/blackfriday"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DEVELOP = iota
	DEPLOY
)

var MODE = DEPLOY

// templates
type M map[string]interface{}

var templates *mandira.Loader
var t = templates.MustGet

// database
var db *sqlx.DB
var dbm *modl.DbMap

// misc
var cookies *sessions.CookieStore
var decoder = schema.NewDecoder()

// Generate a random string of given length, used for cookie secrets
func GenKey(length int) string {
	alphabet := `ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890 ` + "`" +
		`abcdefghijklmnopqrstuvwxyz~!@#$%^&*()-_+={}[]\\|<,>.?/\"';:`
	bytes := make([]byte, length)
	rand.Read(bytes)
	con := byte(len(alphabet))
	for i, b := range bytes {
		bytes[i] = alphabet[b%con]
	}
	return string(bytes)
}

func MustParse(path string) *mandira.Template {
	var t *mandira.Template
	var err error
	if len(path) > 40 {
		t, err = mandira.ParseString(path)
	} else {
		t, err = mandira.ParseFile(path)
	}
	if err != nil {
		log.Fatal(err)
	}
	return t
}

// return an environment key or a fallback
func environ(key, fallback string) string {
	v := os.Getenv(key)
	if len(v) == 0 {
		return fallback
	}
	return v

}

func main() {
	// TODO: user/delete && page/delete
	r := pat.New()
	// user management
	r.Get("/users/create", http.HandlerFunc(createUser))
	r.Post("/users/create", http.HandlerFunc(createUser))
	r.Get("/users/login", http.HandlerFunc(login))
	r.Post("/users/login", http.HandlerFunc(login))
	r.Get("/users/logout", http.HandlerFunc(logout))
	r.Get("/users/{id}", http.HandlerFunc(showUser))
	r.Get("/users", http.HandlerFunc(listUsers))
	// page management
	r.Get("/pages/edit{url:.+}", http.HandlerFunc(editPage))
	r.Post("/pages/edit{url:.+}", http.HandlerFunc(editPage))
	r.Get("/pages", http.HandlerFunc(listPages))
	// wiki site
	r.Get("/{url:.*}", http.HandlerFunc(wikipage))

	http.Handle("/", handlers.LoggingHandler(os.Stdout, r))
	port = environ("GOWIKI_PORT", "2222")
	fmt.Println("Listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func defaults(w http.ResponseWriter, req *http.Request) {
	now := time.Now().UTC()
	ts := now.Format(time.RFC1123)
	ts = strings.Replace(ts, "UTC", "GMT", 1)
	w.Header().Set("Server", "gowiki")
	w.Header().Set("Date", ts)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func abort(w http.ResponseWriter, status int, body []byte) {
	w.WriteHeader(status)
	w.Write(body)
}

// Handles all non-sepcial wiki pages.
func wikipage(w http.ResponseWriter, req *http.Request) {
	defaults(w, req)
	var err error
	page := Page{}
	page.Url = "/" + req.URL.Query().Get(":url")
	err = dbm.Get(&page, page.Url)
	if err != nil {
		w.Write([]byte(t("pagedne.mnd").RenderInLayout(t("base.mnd"), M{"page": page})))
		return
	}
	w.Write([]byte(t("page.mnd").RenderInLayout(t("base.mnd"), M{"page": page})))
}

func listUsers(w http.ResponseWriter, req *http.Request) {
	defaults(w, req)
	users := []*User{}
	db.Select(&users, "SELECT * FROM user")
	c := t("user.mnd").Render(M{"users": users})
	s := t("base.mnd").Render(M{"content": c})
	w.Write([]byte(s))
}

func createUser(w http.ResponseWriter, req *http.Request) {
	defaults(w, req)
	var err error
	user := &User{}

	if req.Method == "POST" {
		req.ParseForm()
		decoder.Decode(user, req.PostForm)
		user.Password = sha1hash(user.Password)
		err = dbm.Insert(user)
		if err == nil {
			http.Redirect(w, req, "/users", 303)
			return
		}
	}

	s := t("usercreate.mnd").RenderInLayout(t("base.mnd"), M{
		"error": err,
		"user":  user,
	})
	w.Write([]byte(s))
}

func showUser(w http.ResponseWriter, req *http.Request) {
	idstr := req.URL.Query().Get(":id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, req, "/users", 301)
		return
	}
	user := &User{}
	err = dbm.Get(user, id)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, req, "/users", 301)
		return
	}
	w.Write([]byte(t("usershow.mnd").RenderInLayout(t("base.mnd"), M{"user": user})))
}

func sha1hash(password string) string {
	sha := sha1.New()
	io.WriteString(sha, password)
	return fmt.Sprintf("%x", sha.Sum(nil))
}

func login(w http.ResponseWriter, req *http.Request) {
	var err error
	user := User{}
	if req.Method == "POST" {
		req.ParseForm()
		decoder.Decode(&user, req.PostForm)
		hash := sha1hash(user.Password)
		err = db.Get(&user, "SELECT * FROM user WHERE email=? AND password=?", user.Email, hash)
		if err == nil {
			session, _ := cookies.Get(req, "gowiki-session")
			session.Values["authenticated"] = true
			session.Values["userid"] = user.Id
			session.Save(req, w)
			http.Redirect(w, req, "/", 303)
			return
		}
	}

	if err == sql.ErrNoRows {
		err = errors.New("Email or Password incorrect.")
	}
	w.Write([]byte(t("login.mnd").RenderInLayout(t("base.mnd"), M{
		"user":  user,
		"error": err,
	})))
}

func logout(w http.ResponseWriter, req *http.Request) {
	session, _ := cookies.Get(req, "gowiki-session")
	session.Values["authenticated"] = false
	delete(session.Values, "userid")
	session.Save(req, w)
	http.Redirect(w, req, "/", 302)
}

func listPages(w http.ResponseWriter, req *http.Request) {
	pages := []*Page{}
	db.Select(&pages, "SELECT * FROM page")
	w.Write([]byte(t("listpages.mnd").RenderInLayout(t("base.mnd"), M{"pages": pages})))
}

func editPage(w http.ResponseWriter, req *http.Request) {
	var err error
	page := &Page{}
	page.Url = req.URL.Query().Get(":url")
	if req.Method == "POST" {
		req.ParseForm()
		err = dbm.Get(page, page.Url)
		decoder.Decode(page, req.PostForm)
		page.Render()
		if err == nil {
			_, err = dbm.Update(page)
		} else {
			err = dbm.Insert(page)
		}
	} else {
		err = dbm.Get(page, page.Url)
	}
	w.Write([]byte(t("editpage.mnd").RenderInLayout(t("base.mnd"), M{
		"page":  page,
		"error": err,
	})))
}

func bundleStatic(w http.ResponseWriter, req *http.Request) {
	f, ok := s_files[strings.TrimLeft(req.URL.Path, "/")]
	if ok {
		w.Write([]byte(f))
	} else {
		http.NotFound(w, req)
	}
}

// db

type User struct {
	Id       int
	Username string
	Password string
	Email    string
}

type Page struct {
	Url      string
	Content  string
	Rendered string
	Title    string
	Locked   bool
	OwnedBy  sql.NullInt64
}

type Config struct {
	Key   string
	Value string
}

// renders a page and sets its Rendered content
func (p *Page) Render() string {
	var flags int
	var extensions int
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS
	extensions |= blackfriday.EXTENSION_HARD_LINE_BREAK
	flags |= blackfriday.HTML_GITHUB_BLOCKCODE
	//flags |= blackfriday.HTML_SAFELINK
	renderer := blackfriday.HtmlRenderer(flags, "", "")
	p.Rendered = string(blackfriday.Markdown([]byte(p.Content), renderer, extensions))
	return p.Rendered
}

func init() {
	var err error
	path := environ("GOWIKI_PATH", "./wiki.sql")
	db, err = sqlx.Connect("sqlite3", path)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	dbm = modl.NewDbMap(&db.DB, modl.SqliteDialect{})
	dbm.AddTable(User{}, "user").SetKeys(true, "id")
	dbm.AddTable(Page{}, "page").SetKeys(false, "url")
	dbm.AddTable(Config{}, "config").SetKeys(false, "key")
	err = dbm.CreateTablesIfNotExists()

	if err != nil {
		log.Fatal("Database not creatable: ", err)
	}
	// if we're developing, use /static/ and /templates/
	if MODE == DEVELOP {
		fmt.Println("Running in development mode without bundled resources.")
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
		templates = mandira.NewLoader("./templates/", false)
		// TODO: load m_default
		t = templates.MustGet
	} else {
		fmt.Println("Running in deployment mode with bundled resources.")
		// decode base64 encoded static files
		for k, v := range s_files {
			v = strings.Trim(v, " \t\n\r")
			b, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				fmt.Println(err)
			} else {
				s_files[k] = string(b)
			}
		}
		http.Handle("/static/", http.HandlerFunc(bundleStatic))
		templates = mandira.NewLoader("/tmp/doesnotexist", true)
		for path, content := range t_files {
			templates.Add(path, MustParse(content))
		}
		t = templates.MustGet
	}

	/* initialize cookie secret if needed and store it in the config table */
	secret := &Config{}
	err = dbm.Get(secret, "secret")
	if err != nil {
		secret = &Config{Key: "secret", Value: GenKey(32)}
		dbm.Insert(secret)
		fmt.Println("Auto-created new cookie secret.")
	}

	index := &Page{}
	err = dbm.Get(index, "/")
	if err != nil {
		index.Content = m_default
		index.Render()
		index.Title = "Welcome to Gowiki"
		index.Url = "/"
		dbm.Insert(index)
		fmt.Println("Auto-created index.")
	}

	cookies = sessions.NewCookieStore([]byte(secret.Value))
}
