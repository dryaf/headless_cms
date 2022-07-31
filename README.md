# headless_cms

## What is it about?

This is a client interface for requests to a headless cms.
An simple POC implementation is done for storyblok

not production ready

## Usage - Storyblok - auto generated website

### 1. Create components and pages in storyblok

```
create pages with nested components as deep as you want
create the blocks/partials like described via github.com/dryaf/templates
```

### 2. Use github/dryaf/templates to render generate the html based on the json tree from storyblok
```gohtml
 {{define "page"}}
<div class="container my-5">
{{range .story.content.body}}
{{ d_block .component . }}
{{end}}
</div>
{{end}}

```
### 3. don't forget to call the emptycache func via a storyblok webhook after publish


## Usage - Storyblok - translations for texts in your local templates

### 1. Create component and pages in storyblok

```
create component with id and value (text) called _translatable_text
create a page in storyblok that only accepts blocks of type _translatable_text
```

### 2. Use the component values via the ids
```gohtml
 {{define "page"}}
<div class="container my-5">
{{ .CMS.text1 }}
{{ .CMS.text2 }}
</div>
{{end}}

```
### 3. don't forget to call the emptycache func via a storyblok webhook after publish

## Todos
[ ] Write Tests
