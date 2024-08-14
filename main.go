package main

import (
	"database/sql"
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

//go:embed static/*
var staticFiles embed.FS

//go:embed views/*
var viewFiles embed.FS

// Define the EmailRequest struct
type EmailRequest struct {
	Name     string `json:"name" form:"name"`
	Instance string `json:"instance" form:"instance"`
	Subject  string `json:"subject" form:"subject"`
	Message  string `json:"message" form:"message"`
	Email 	 string `json:"email" form:"email"`
}

type Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Declare environment variable variables
var (
	FROM      string
	TO        string
	PASSWORD  string
	SMTP_HOST string
	SMTP_PORT string
	APP_PORT  string
	DB_USERNAME string
	DB_PASSWORD string
	DB_HOST    string
	DB_PORT string
	DB_DATABASE string
	connStr = "user=%s password=%s dbname=%s host=%s port=%s sslmode=disable"
	db *sql.DB
)

// Initialize environment variables
func init() {
	FROM 			= os.Getenv("FROM_EMAIL")
	PASSWORD 		= os.Getenv("PASSWORD_EMAIL")
	TO 				= os.Getenv("TO_EMAIL")
	SMTP_HOST 		= os.Getenv("SMTP_HOST")
	SMTP_PORT 		= os.Getenv("SMTP_PORT")
	APP_PORT 		= os.Getenv("APP_PORT")
	DB_USERNAME 	= os.Getenv("DB_USERNAME")
	DB_PASSWORD		= os.Getenv("DB_PASSWORD")
	DB_HOST     	= os.Getenv("DB_HOST")
	DB_PORT 		= os.Getenv("DB_PORT")
	DB_DATABASE 	= os.Getenv("DB_DATABASE")
}


// Main function to start the server
func main() {
	connection := fmt.Sprintf(connStr, DB_USERNAME, DB_PASSWORD, DB_DATABASE, DB_HOST, DB_PORT)
	log.Println(connection)
	dbInstance, err := sql.Open("postgres", connection)
	if err != nil {
		panic(err)
	}
	db = dbInstance
	defer dbInstance.Close()

	t := &Template{
		templates: template.Must(template.ParseFS(viewFiles, "views/*.html")),
	}

	e := echo.New()
	e.Renderer = t
	e.GET("/*", echo.WrapHandler(http.FileServer(http.FS(staticFiles))))
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]string{
			"message": "",
		})
	})
	e.POST("/contact-us", func(c echo.Context) error {
		var emailReq EmailRequest
		if err := c.Bind(&emailReq); err != nil {
			log.Fatalln(err)
		}

		queryRaw := "INSERT INTO contact_us (name, email, organization, subject, message) VALUES ($1, $2, $3, $4, $5) RETURNING id"

		var id int
		err := db.QueryRow(queryRaw, emailReq.Name, emailReq.Email, emailReq.Instance, emailReq.Subject, emailReq.Message).Scan(&id)

		if err != nil {
			log.Println("unable save to DB:", err)
		}

		return c.Render(http.StatusOK, "index.html", map[string]string{
			"message": "Form berhasil di submit, tunggu pihak kami akan menghubungi anda segera",
		})
	})
	e.Logger.Fatal(e.Start(":" + APP_PORT))
}