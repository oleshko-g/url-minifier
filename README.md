# URL minifier

"URL minifier" is a URL shortener service

## Features

### "minify"

Take a URL and return its shortened version

### "unminify"

Take the shortened URL and return the original URL

## HTTP API

### POST /

Accepts a URL to minify as "text/plain" from the request body and returns the
minified URL, which has the format "http://{minifier address}/{minifiedID}, as
"text/plain" in the response body.

### GET /{minifiedID}

Accepts {minifiedID} as a path parameter and returns the original URL in
"Location" HTTP header. A success response has "307 Temporary Redirect" HTTP
status code.

### POST /api/shorten

Accepts a URL in JSON request body and returns the minified URL in JSON response
body

#### Request body example

```json
{
  "url": "https://example.com/very/long/path"
}
```

#### Response body example

```json
{
  "result": "http://minify.it/io1U12jkk"
}
```
