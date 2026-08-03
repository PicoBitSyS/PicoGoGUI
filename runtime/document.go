package runtime

import (
	"strings"

	"github.com/PicoBitSyS/PicoGoGUI/plugin"
	"github.com/PicoBitSyS/PicoGoGUI/web"
)

// Document builds a single HTML document with inlined CSS and JS for SetHtml.
// Prefer DocumentE when plugin activation errors must be reported.
func Document() string {
	document, _ := DocumentE()
	return document
}

// DocumentE builds the runtime document and reports plugin activation errors.
func DocumentE() (string, error) {
	if err := plugin.Activate(); err != nil {
		return "", err
	}

	html := web.MustRead("index.html")
	css := web.MustRead("theme.css") + plugin.CSS()
	js := web.MustRead("app.js") + plugin.JS()

	html = strings.Replace(html, `<link rel="stylesheet" href="theme.css" />`, "<style>\n"+css+"\n</style>", 1)
	html = strings.Replace(html, `<script src="app.js"></script>`, "<script>\n"+js+"\n</script>", 1)
	return html, nil
}
