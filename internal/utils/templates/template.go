package templates

import (
	"bytes"
	"html/template"
	"log"
)

type Options struct {
	Title     string
	Message   string
	Link      string
	LogoCID   string
	BannerCID string
	BtnTitle  string
}

func GenerateTemplate(options Options) string {
	tmplString := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; padding: 20px; background-color: #f4f4f4; color: #333; }
        .container { background-color: #fff; padding: 20px; border-radius: 8px; border: 1px solid #ddd; margin: auto; max-width: 600px; }
        .button { background-color: #007BFF; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block; text-align: center; }
        img { max-width: 100%; height: auto; }
        .footer { margin-top: 20px; font-size: 12px; text-align: center; color: #999; }
    </style>
</head>
<body>
    <div class="container">
        <img src="cid:{{.LogoCID}}" alt="Company Logo">
        <h1>{{.Title}}</h1>
        <img src="cid:{{.BannerCID}}" alt="Banner">
        <p>{{.Message}}</p>
        <a href="{{.Link}}" class="button text-white bg-dark">{{.BtnTitle}}</a>
        <p class="footer">© 2024 Audify. All rights reserved.</p>
    </div>
</body>
</html>
`

	tmpl, err := template.New("emailTemplate").Parse(tmplString)
	if err != nil {
		log.Fatalf("Error parsing template: %s", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, options); err != nil {
		log.Fatalf("Error executing template: %s", err)
	}

	return buf.String()
}
