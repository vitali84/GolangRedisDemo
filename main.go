package main

import (
	"github.com/go-redis/redis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"io"
	"net/http"
	"os"
)

const REDIS_HASHNAME = "DEMO_HASH"

var RedisCleint *redis.Client

func NewClient() {
	redisHost := "localhost:6379"
	if os.Getenv("REDIS_HOST") != "" {
		redisHost = os.Getenv("REDIS_HOST")
	}

	RedisCleint = redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	NewClient()
	e := echo.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("*.html")),
	}
	e.Renderer = renderer
	e.GET("/", index).Name = "index"
	e.POST("/clear", clear).Name = "clear"
	e.POST("/", create).Name = "create"
	e.GET("/hash", hash).Name = "hash"
	e.Logger.Fatal(e.Start(":3000"))
}

func index(c echo.Context) error {
	redisContent, err := RedisCleint.HGetAll(REDIS_HASHNAME).Result()
	if err != nil {
		panic(err)
	}
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"redisContent": redisContent,
	})
}

func create(c echo.Context) error {
	err := RedisCleint.HSet(REDIS_HASHNAME, c.FormValue("key"), c.FormValue("val")).Err()
	if err != nil {
		panic(err)
	}
	return c.Redirect(http.StatusMovedPermanently, "/")
}

func clear(c echo.Context) error {
	err := RedisCleint.Del(REDIS_HASHNAME).Err()
	if err != nil {
		panic(err)
	}
	return c.Redirect(http.StatusMovedPermanently, "/")
}

func hash(c echo.Context) error {
	str := c.Param("string")
	bytes, _ := bcrypt.GenerateFromPassword([]byte(str), 14)
	return c.String(http.StatusOK, string(bytes))
}
