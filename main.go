package main

import (
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/zx06/httpdump/ui"
)

const (
	idKey = "id"
)

type Record struct {
	Method     string      `json:"method"`
	URL        url.URL     `json:"url"`
	Proto      string      `json:"proto"`
	Headers    http.Header `json:"headers"`
	Body       []byte      `json:"body"`
	Host       string      `json:"host"`
	RemoteAddr string      `json:"remote_addr"`
}

var (
	store = make(map[string][]Record)
	mu    sync.RWMutex
)

func genID() string {
	// generate random id in a-z
	var (
		id     string
		chars  = []rune("abcdefghijklmnopqrstuvwxyz")
		length = 10
	)
	for i := 0; i < length; i++ {
		id += string(chars[rand.Intn(len(chars))])
	}
	return id
}

func initCookieMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			_, err := c.Cookie(idKey)
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					id := genID()
					cookie := &http.Cookie{
						Name:  idKey,
						Value: id,
						Path:  "/",
					}
					c.SetCookie(cookie)
				} else {
					return err
				}
			}
			return next(c)
		}
	}
}

func main() {
	app := echo.New()
	app.Use(middleware.Logger())
	staticConfig := middleware.StaticConfig{
		Root:       "dist",
		HTML5:      true,
		IgnoreBase: true,
		Filesystem: http.FS(ui.Dist),
	}
	app.Use(middleware.StaticWithConfig(staticConfig))
	app.Any("/x/:id/*", func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.String(http.StatusBadRequest, "id is required")
		}
		bb, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		// write to store
		mu.Lock()
		defer mu.Unlock()
		r := Record{
			Method:     c.Request().Method,
			URL:        *c.Request().URL,
			Proto:      c.Request().Proto,
			Headers:    c.Request().Header,
			Body:       bb,
			Host:       c.Request().Host,
			RemoteAddr: c.Request().RemoteAddr,
		}
		store[id] = append(store[id], r)
		return c.JSON(http.StatusOK, r)
	})
	api := app.Group("/api")
	api.Use(initCookieMiddleware())
	{
		apiRecord := api.Group("/record")
		{
			apiRecord.GET("/", func(c echo.Context) error {
				mu.RLock()
				defer mu.RUnlock()
				return c.JSON(http.StatusOK, store)
			})
			apiRecord.GET("/:id", func(c echo.Context) error {
				id := c.Param("id")
				if id == "" {
					return c.String(http.StatusBadRequest, "id is required")
				}
				mu.RLock()
				defer mu.RUnlock()
				return c.JSON(http.StatusOK, store[id])
			})
		}
	}
	err := app.Start(":1234")
	if err != nil {
		panic(err)
	}
}
