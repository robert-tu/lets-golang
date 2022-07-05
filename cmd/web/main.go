package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"robert-tu.net/snippetbox/pkg/models"
	"robert-tu.net/snippetbox/pkg/models/mysql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
)

type contextKey string

const contextKeyIsAuthenticated = contextKey("isAuthenticated")

// application struct
// holds application-wide dependencies
type application struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	// inline interface
	snippets interface {
		Insert(string, string, string) (int, error)
		Get(int) (*models.Snippet, error)
		Latest() ([]*models.Snippet, error)
	}
	templateCache map[string]*template.Template
	session       *sessions.Session
	// inline interface
	users interface {
		Insert(string, string, string) error
		Authenticate(string, string) (int, error)
		Get(int) (*models.User, error)
	}
}

func main() {
	// initialize command line flag and parse
	addr := flag.String("addr", ":4000", "HTTP network address")
	// initialize command line flag for MySQL
	ds := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "My SQL data source")
	// define flag for session secret
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret Key")
	flag.Parse()

	// INFO logger
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	// ERROR logger
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// initialize db connection
	db, err := openDB(*ds)
	if err != nil {
		errLog.Fatal(err)
	}
	// defer connection pool close before main
	defer db.Close()

	// initialize template cache
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errLog.Fatal(err)
	}

	// use sessions.New() with secret key to initialize session manager
	session := sessions.New([]byte(*secret))
	// session expires after 12 hours
	session.Lifetime = 12 * time.Hour
	// session secure flag
	session.Secure = true
	// session cookies  attribute
	// session.SameSite = http.SameSiteStrictMode

	// initialize new instance of application
	app := &application{
		infoLog:       infoLog,
		errorLog:      errLog,
		snippets:      &mysql.SnippetModel{DB: db},
		templateCache: templateCache,
		session:       session,
		users:         &mysql.UserModel{DB: db},
	}

	// initialize tls.Config struct
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
		// only accept ECDHE ciphers
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		// allow TLS 1.2 and 1.3
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}

	// initialize http.Server struct
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,     // default Go 3 min
		ReadTimeout:  5 * time.Second, // short read timeout migitigates risk from slow-client attacks
		WriteTimeout: 10 * time.Second,
	}

	// start new web server calling server struct
	// returns error in log
	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errLog.Fatal(err)

}

// openDB() function wraps sql.Open()
// returns sql.DB connection pool for given DS
func openDB(ds string) (*sql.DB, error) {
	db, err := sql.Open("mysql", ds)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
