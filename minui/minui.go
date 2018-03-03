package minui

import (
	"encoding/hex"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gernest/front"
	"github.com/labstack/echo"
	"gopkg.in/russross/blackfriday.v2"

	"github.com/u-speak/core/api"
	"github.com/u-speak/core/config"
	"github.com/u-speak/core/node"
	"github.com/u-speak/core/post"
	"github.com/u-speak/core/tangle"
	"github.com/u-speak/core/tangle/datastore"

	log "github.com/sirupsen/logrus"
	"github.com/u-speak/logrusmiddleware"
)

// Server holds the information about the current node
type Server struct {
	sslkey  string
	sslcert string
	listen  string
	node    *node.Node
	message string
}

type renderer struct {
	templates *template.Template
}

type response struct {
	Theme   string
	Message string
	Data    interface{}
	Path    string
}

var (
	templateMap = template.FuncMap{
		"Post": func(s datastore.Serializable) *post.Post {
			return s.(*post.Post)
		},
		"Title": func(p *post.Post) string {
			m := front.NewMatter()
			m.Handle("---", front.YAMLHandler)
			f, _, _ := m.Parse(strings.NewReader(p.Content))
			if f["title"] == nil {
				return "Untitled Post"
			}
			return f["title"].(string)
		},
		"Body": func(p *post.Post) string {
			m := front.NewMatter()
			m.Handle("---", front.YAMLHandler)
			_, b, err := m.Parse(strings.NewReader(p.Content))
			if err != nil {
				return p.Content
			}
			return b
		},
		"Image": func(p *post.Post) string {
			m := front.NewMatter()
			m.Handle("---", front.YAMLHandler)
			f, _, _ := m.Parse(strings.NewReader(p.Content))
			if f["image"] == nil {
				return "https://picsum.photos/573/190/?random"
			}
			return f["image"].(string)
		},
		"Markdown": func(s string) template.HTML {
			h := blackfriday.Run([]byte(s))
			return template.HTML(string(h))
		},
		"URLEncode": func(s string) string {
			return url.QueryEscape(s)
		},
		"Valid": func(p *post.Post) bool {
			return p.Verify() == nil
		},
		"Fingerprint": func(p *post.Post) string {
			return hex.EncodeToString(p.Pubkey.PrimaryKey.Fingerprint[:])
		},
	}
)

func (r *renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return r.templates.ExecuteTemplate(w, name, data)
}

// New creates a new server
func New(c config.Configuration, n *node.Node) *Server {
	li := c.Web.MinUI.Interface + ":" + strconv.Itoa(c.Web.MinUI.Port)
	return &Server{listen: li, node: n, sslkey: c.Global.SSLKey, sslcert: c.Global.SSLCert, message: c.Global.Message}
}

// Run starts the server
func (s *Server) Run() {
	r := &renderer{
		templates: template.New("").Funcs(templateMap),
	}
	ps, err := WalkDirs("templates", true)
	if err != nil {
		log.Error(err)
		return
	}
	for _, p := range ps {
		bytes, err := ReadFile(p)
		if err != nil {
			log.Error(err)
			return
		}
		r.templates.New(p).Parse(string(bytes))
	}
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Renderer = r

	e.Logger = logrusmiddleware.Logger{log.StandardLogger()}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("theme")
			if err != nil {
				c.Set("theme", "sakura")
				return next(c)
			}
			c.Set("theme", cookie.Value)
			return next(c)
		}
	})
	e.GET("/", s.getIndex)
	e.GET("/themes/:theme", s.switchTheme)
	e.GET("/posts/:hash", s.getPost)
	e.GET("/*", echo.WrapHandler(Handler))

	log.Fatal(e.StartTLS(s.listen, s.sslcert, s.sslkey))
}

func (s *Server) getIndex(c echo.Context) error {
	hs := s.node.Tangle.Hashes()
	for i := range hs {
		j := rand.Intn(i + 1)
		hs[i], hs[j] = hs[j], hs[i]
	}
	sites := []*tangle.Object{}
	l := 10
	if l > len(hs) {
		l = len(hs)
	}
	for _, h := range hs[:l] {
		o := s.node.Tangle.Get(h)
		if o == nil {
			continue
		}
		sites = append(sites, o)
	}
	return c.Render(http.StatusOK, "templates/index.html.tmpl", response{Theme: c.Get("theme").(string), Message: s.message, Data: sites})
}

func (s *Server) getPost(c echo.Context) error {
	h, err := api.DecodeHash(c.Param("hash"))
	if err != nil {
		return s.error404(c)
	}
	o := s.node.Tangle.Get(h)
	if o == nil {
		return s.error404(c)
	}
	if o.Site.Type != "post" {
		return s.error404(c)
	}
	p := o.Data.(*post.Post)
	return c.Render(http.StatusOK, "templates/post.html.tmpl", response{
		Theme:   c.Get("theme").(string),
		Message: s.message,
		Data:    p,
		Path:    "/posts/" + c.Param("hash"),
	})
}

func (s *Server) switchTheme(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "theme"
	switch c.Param("theme") {
	case "default":
		cookie.Value = "sakura"
	case "dark":
		cookie.Value = "sakura-dark"
	case "dark-solarized":
		cookie.Value = "sakura-dark-solarized"
	case "earthly":
		cookie.Value = "sakura-earthly"
	case "vader":
		cookie.Value = "sakura-vader"
	default:
		cookie.Value = "sakura"
	}
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"
	c.SetCookie(cookie)
	rp, err := url.QueryUnescape(c.QueryParam("path"))
	if err != nil || rp == "" {
		rp = "/"
	}
	return c.Redirect(http.StatusTemporaryRedirect, rp)
}

func (s *Server) error404(c echo.Context) error {
	return c.Render(http.StatusNotFound, "templates/404.html.tmpl", response{
		Theme:   c.Get("theme").(string),
		Message: s.message,
		Path:    "/",
	})
}
