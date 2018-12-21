# Pipe

Powerful continuous stream processing

[![Build Status](https://travis-ci.org/relvacode/pipe.svg?branch=master)](https://travis-ci.org/relvacode/pipe)

Run a series of native or system commands concurrently and manipulate streams using rich native objects

```
brew install relvacode/pipe/pipe
```

```
go install github.com/relvacode/pipe/cmd/pipe
```


## Basics

### Pipes

A pipe can either be a built-in native pipe, or fallback to a program on the system.

To join the output of one pipe to the input of the next use `::`.

The first pipe in the series is given a `stdin` object and all outputs of the last pipe are echoed to `stdout`.

### Context and Tagging

With each object passed to a pipe a context is given which traces the history of that object and its parents when tagged.

Use `as <name>` to tag each value produced by that pipe and refer to it in a later pipe.

`this` always refers to the current object.

### Templating

Use Django style templates provided by [Pongo2](https://github.com/flosch/pongo2) in pipe arguments

```
pipe 'print World as name :: print Hello, {{name}}
```

Use `.String` on most objects to get a more human readable representation of an object

```
pipe 'open * :: print {{this.Mode.String}}'
```

### Filtering

Use the `if` pipe to filter by expression using [Expr](https://github.com/antonmedv/expr)

```bash
# Open all non-empty JSON files and decode their contents
pipe 'open *.json :: if this.Size > 0 :: json'
```

## Help

All native pipes can be listed with

```
pipe -lib
```

Find out about a specific pipe using

```
pipe -pkg <name>
```

## Advanced

### Templating

#### Create a temporary file

Use `mktemp`  filter to create a temporary file containing the value's contents.

```bash
pipe 'url.get https://example.org as request :: openssl md5 {{request | mktemp}}'
```


## Examples

Get a new stock quote every 30 seconds

```bash
pipe 'every 30s :: print ilmn as stock :: url.get https://api.iextrading.com/1.0/stock/{{stock}}/quote :: json :: select this.iexRealtimePrice'
```
