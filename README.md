# Pipe

Powerful continuous stream processing

[![Build Status](https://travis-ci.org/relvacode/pipe.svg?branch=master)](https://travis-ci.org/relvacode/pipe)

Run a series of commands or pipes concurrently and manipulate streams using rich native objects.

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

The first pipe in a pipeline is given a `stdin` object and all outputs of the last pipe are echoed to `stdout`.

### Context

With each object passed to a pipe, a context is given which traces the history of that chain.

Use `as <name>` to tag each value produced by that pipe and refer to it in a later pipe.

`this` always refers to the current object.

### Templating

Use Django style template provided by [Pongo2](https://github.com/flosch/pongo2) to access values in pipe arguments.

```
pipe 'print World as name :: print Hello, {{world}}
```

### Filtering

Use the `if` pipe to filter by expression using [Expr](https://github.com/antonmedv/expr)

```bash
# Open all non-empty JSON files and decode their contents
pipe 'open *.json :: if this.Size > 0 :: json.decode'
```


## Advanced

### Templating

#### Create a temporary file

Use `mktemp`  filter to create a temporary file containing the value's contents.

```bash
pipe 'url.get https://example.org as request | openssl md5 {request | mktemp}'
```


## Examples

Get a new stock quote every 30 seconds

```bash
pipe 'every 30s :: print ilmn as stock :: url.get https://api.iextrading.com/1.0/stock/{{stock}}/quote :: json.decode :: select this.iexRealtimePrice'
```
