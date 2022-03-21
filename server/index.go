package server

import (
	"bytes"
	"fmt"
	"io"
	textTemplate "text/template"
	// htmlTemplate "html/template"
	"net/http"
	"strings"

	"github.com/miniscruff/dashy/configs"
)

type ContentBuilder func(configs.Content) (string, error)

var defaultStyles = map[string]string{
	":root": `
  /* nord color scheme */
  --layer0: #2e3440;
  --layer1: #3b4252;
  --layer2: #434c5e;
  --layer3: #4c566a;
  --primary2: #d8dee9;
  --primary1: #e5e9f0;
  --primary: #eceff4;
  --accent1: #8fbcbb;
  --accent: #88c0d0;
  --accent2: #81a1c1;
  --accent3: #5e81ac;
  --error: #bf616a;
  --danger: #d08770;
  --warning: #ebcb8b;
  --success: #a3be8c;
  --neutral: #b48ead;`,
	".text-left":   "text-align: left;",
	".text-center": "text-align: center;",
	".text-right":  "text-align: right;",
	".text-large":  "font-size: large;",
	".text-larger": "font-size: larger;",
	"body": `
	background: var(--layer0);
	color: var(--primary);
  padding: 1em;
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  grid-gap: .25rem;`,
	".layer": `
	background: var(--layer1);
	border-radius: 10px;`,
	".two-columns": `
	display: grid;
	grid-template-columns: 1fr 1fr;
	grid-column-gap: 10px;`,
}

var (
	html = `<!DOCTYPE html>
<html lang="en">
	<head>
		<title>{{.title}}</title>
		<meta charset="utf-8" />
		{{.meta}}
		<style>
			{{.styles}}
		</style>
	</head>
	<body>
		{{.body}}
		<script>
			{{.scripts}}
		</script>
	</body>
</html>`
)

const (
	letters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterCount = 52
)

func stringFromIndex(index *int) string {
	res := fmt.Sprintf("%v%v",
		string(letters[*index/letterCount]),
		string(letters[*index%letterCount]),
	)
	*index++
	return res
}

type IndexBuilder struct {
	dashboard    configs.Dashboard
	elements     map[string]string
	contentIndex int
}

func (b *IndexBuilder) buildMeta() (string, error) {
	var builder strings.Builder

	metaFormat := `<meta name="%v" content="%v" />`
	for n, c := range b.dashboard.Meta {
		_, _ = builder.WriteString(fmt.Sprintf(metaFormat, n, c))
	}
	return builder.String(), nil
}

func (b *IndexBuilder) buildLayers() (string, error) {
	var builder strings.Builder
	for _, l := range b.dashboard.Layers {
		layer, err := b.buildLayer(l)
		if err != nil {
			return "", err
		}

		_, _ = builder.WriteString(layer)
	}
	return builder.String(), nil
}

func (b *IndexBuilder) buildLayer(layer configs.Layer) (string, error) {
	var contentBuilder strings.Builder
	for _, c := range layer.Contents {
		content, err := b.buildContent(c)
		if err != nil {
			return "", err
		}

		_, _ = contentBuilder.WriteString(content)
	}

	// TODO:
	// we need to create an extra layer here for the optional title bar
	// additionally, this may need to be recursive to allow layers on layers
	// not sure entirely how that one will work yet.

	return fmt.Sprintf(
		`<div id="%v" class="layer %v" style="grid-column-start: %v;grid-column-end: %v;grid-row-start: %v;grid-row-end: %v;">%v</div>`,
		stringFromIndex(&b.contentIndex),
		layer.Layout,
		layer.X+1,
		layer.X+layer.Width,
		layer.Y+1,
		layer.Y+layer.Height,
		contentBuilder.String(),
	), nil
}

func (b *IndexBuilder) buildContent(content configs.Content) (string, error) {
	var builder ContentBuilder
	switch strings.ToLower(content.Type) {
	case "text":
		builder = b.textContent
	case "constant":
		builder = b.constantContent
	default:
		return "", fmt.Errorf("content type '%v' not found", content.Type)
	}

	return builder(content)
}

func (b *IndexBuilder) textContent(content configs.Content) (string, error) {
	id := stringFromIndex(&b.contentIndex)
	// TODO: this does not allow any option of customizing the text for things
	// like data type conversions, or even formatting really.
	b.elements[id] = fmt.Sprintf("element.innerHTML = `%v`", content.Text)

	return fmt.Sprintf(
		`<div id="%v" class="%v"></div>`,
		id,
		strings.Join(content.Styles, " "),
	), nil
}

func (b *IndexBuilder) constantContent(content configs.Content) (string, error) {
	return fmt.Sprintf(
		`<div class="%v">%v</div>`,
		strings.Join(content.Styles, " "),
		content.Text,
	), nil
}

func (b *IndexBuilder) buildStyles() (string, error) {
	var builder strings.Builder

	allStyles := make(map[string]string, 0)
	for k := range defaultStyles {
		allStyles[k] = defaultStyles[k]
	}
	for k := range b.dashboard.CustomStyles {
		allStyles[k] = defaultStyles[k]
	}

	styleFormat := `%v {%v}`
	for name, content := range allStyles {
		_, _ = builder.WriteString(fmt.Sprintf(styleFormat, name, content))
	}
	return builder.String(), nil
}

func (b *IndexBuilder) buildScripts() (string, error) {
	var builder strings.Builder

	updateFormat := `function update%v(element, data) {
		%v;
	}
	`
	saveElementFormat := `elements["%v"] = document.getElementById("%v");
	`

	for n, m := range b.elements {
		_, _ = builder.WriteString(fmt.Sprintf(updateFormat, n, m))
	}

	builder.WriteString("function updateall(data) {\n")
	for id := range b.elements {
		builder.WriteString(fmt.Sprintf(
			`update%v(elements["%v"], data);
			`,
			id,
			id,
		))
	}
	builder.WriteString("\n}")

	builder.WriteString(`
	const elements = {};
	window.addEventListener('DOMContentLoaded', async () => {
		const response = await fetch('/api/values');
		const values = await response.json();
		`,
	)

	for id := range b.elements {
		_, _ = builder.WriteString(fmt.Sprintf(saveElementFormat, id, id))
	}

	builder.WriteString("updateall(values);")
	builder.WriteString("\n});")

	return builder.String(), nil
}

func (b *IndexBuilder) Write(writer io.Writer) error {
	tmpl, err := textTemplate.New("HTML").Parse(html)
	if err != nil {
		return err
	}

	var (
		meta    string
		body    string
		styles  string
		scripts string
	)

	if meta, err = b.buildMeta(); err != nil {
		return err
	}

	if body, err = b.buildLayers(); err != nil {
		return err
	}

	if styles, err = b.buildStyles(); err != nil {
		return err
	}

	if scripts, err = b.buildScripts(); err != nil {
		return err
	}

	return tmpl.Execute(writer, map[string]interface{}{
		"title":   b.dashboard.Title,
		"meta":    meta,
		"styles":  styles,
		"body":    body,
		"scripts": scripts,
	})
}

func (s *Server) GenerateIndex() error {
	builder := &IndexBuilder{
		dashboard: s.Config.Dashboard,
		elements:  make(map[string]string),
	}

	var bWriter bytes.Buffer
	err := builder.Write(&bWriter)
	if err != nil {
		return err
	}

	fmt.Println("str", bWriter.String())

	s.indexFile = bWriter.Bytes()
	return nil
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" || r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	w.Write(s.indexFile)
}
