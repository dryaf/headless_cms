[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/dryaf/headless_cms/master/LICENSE)
[![Coverage](https://raw.githubusercontent.com/dryaf/headless_cms/master/coverage.svg)](https://raw.githubusercontent.com/dryaf/headless_cms/master/coverage.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/dryaf/headless_cms?style=flat-square)](https://goreportcard.com/report/github.com/dryaf/headless_cms)
[![GoDoc](https://godoc.org/github.com/dryaf/headless_cms?status.svg)](https://godoc.org/github.com/dryaf/headless_cms)


# Headless CMS Go Client 

A Go client for the Storyblok Headless CMS API.

Implemented SaaS providers:
[x] storyblok.com
[ ] contentful.com


## Features

- Fetches data from Storyblok API
- Caching support for fetched data (in memory or redis)
- Request storyblok data in JSON or map[string]any format for complete website generation 
(see d_block in github.com/dryaf/templates)
- Requests storyblok data as map[string]map[string]any format where storyblok blocks need to contain an id so then can be accessed in go templates via .Texts.id.value (for simple i18n support in non dynamicly rendered pages)
- Empty cache with a Token triggered via a webhook by the headless cms provider

## License
MIT
