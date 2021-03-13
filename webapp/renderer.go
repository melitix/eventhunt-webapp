package main

import (
	"html/template"
	"math"
	"net/http"

	"github.com/eventhunt-org/webapp/framework"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func renderPage(a *app, tplHTML string, w http.ResponseWriter, r *http.Request, tplData map[string]interface{}, bases ...string) {

	if len(bases) == 0 {
		bases = append(bases, "default")
	}

	// Load session so that we may get Flashes
	session, _ := store.Get(r, "login")
	flashes := framework.Flashes(session)

	tplData["App"] = map[string]string{
		"Hostname":    viper.GetString("app_host"),
		"Name":        a.Name(),
		"version":     a.Version,
		"environment": viper.GetString("app_environment"),
	}
	tplData["Environment"] = viper.GetString("app_environment")
	tplData["MainNav"] = mainNav.Items
	tplData["Request"] = map[string]string{}
	tplData["Version"] = a.Version
	tplData["Flashes"] = flashes

	mainNav.PreRender(r.URL)

	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"roundInt": func(a float64) int {
			return int(math.Round(a))
		},
	}

	tpl := template.Must(template.New("theme").Funcs(funcMap).ParseGlob(a.ThemePath() + "partials/*.html"))
	tpl, _ = tpl.ParseGlob(a.ThemePath() + "partials/*.js")
	tpl, err := tpl.ParseFiles(
		a.ThemePath()+"sections/"+tplHTML,
		a.ThemePath()+"base/"+bases[len(bases)-1]+".html",
	)

	if err != nil {
		log.Fatal("Theme files are missing.")
	}

	tpl.ExecuteTemplate(w, "base", tplData)
}

/*
 * If text is longer than the limit, cut it to the limit minus 3, then add an
 * elipsis.
 */
func truncateText(text string, limit int) string {

	if len(text) >= limit {
		return text[0:limit-3] + "..."
	}

	return text
}
