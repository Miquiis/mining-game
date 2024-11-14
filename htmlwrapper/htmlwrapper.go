package htmlwrapper

import (
	"fmt"
	"html"
	"net/http"

	"github.com/labstack/echo/v4"
)

func PageWrapper(c echo.Context, pageName string, message string) error {
	return c.HTML(http.StatusOK, fmt.Sprintf(`
		<html>
		<head>
			<meta name="color-scheme" content="light dark">
			<title>%s</title>
		</head>
		<body>
			<pre style="word-wrap: break-word; white-space: pre-wrap;" id="dataContainer">%s</pre>
		</body>
		</html>
	`, pageName, html.EscapeString(message)))
}

func PageRedirectWrapper(c echo.Context, pageName string, message string, redirectTo string, interval int) error {
	return c.HTML(http.StatusOK, fmt.Sprintf(`
		<html>
		<head>
			<meta name="color-scheme" content="light dark">
			<title>%s</title>
		</head>
		<script>
			setTimeout(() => {
				location.href = "%s";
			}, %d);
		</script>
		<body>
			<pre style="word-wrap: break-word; white-space: pre-wrap;" id="dataContainer">%s</pre>
		</body>
		</html>
	`, pageName, redirectTo, interval, html.EscapeString(message)))
}
