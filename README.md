[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/dryaf/headless_cms/master/LICENSE)
[![Coverage](https://img.shields.io/badge/coverage-88.0%25-green)](https://img.shields.io/badge/coverage-88.0%25-green)
[![Go Report Card](https://goreportcard.com/badge/github.com/dryaf/headless_cms?style=flat-square)](https://goreportcard.com/report/github.com/dryaf/headless_cms)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/
github.com/dryaf/headless_cms?tab=doc)

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
