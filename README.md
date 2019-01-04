# Pipe

Powerful continuous stream processing

[![Build Status](https://travis-ci.org/relvacode/pipe.svg?branch=master)](https://travis-ci.org/relvacode/pipe)

Run a series of native or system commands concurrently and manipulate streams using rich native objects

```
brew install relvacode/pipe/pipe
```

### Basics

A pipe takes input objects and produces zero or more output objects. Join pipes together using `::`.

The first pipe in the series is given a `stdin` object and all outputs of the last pipe are echoed to `stdout`.

If the name of a pipe isn't found in the built-in library then that program is executed on the system

#### Context and Tagging


Use `as <name>` to tag each value produced by that pipe to refer to it in a later pipe.

  - `this` is the current object
  - `_index` is the index of the object in the pipe that produced it.

#### Templating

Use Django style templates provided by [Pongo2](https://github.com/flosch/pongo2) in pipe arguments

```
pipe 'print World as name :: print Hello, {{name}}
```

Use `.String` on most objects to get a more human readable representation of an object

```
pipe 'open * :: print {{this.Mode.String}}'
```

#### Filtering

Use the `if` pipe to filter by expression using [Expr](https://github.com/antonmedv/expr)

```bash
# Open all non-empty JSON files and decode their contents
pipe 'open *.json :: if this.Size > 0 :: json'
```

#### Help

All native pipes can be listed with

```
pipe -lib
```

Find out about a specific pipe using

```
pipe -pkg <name>
```

### Advanced

#### Templating

##### Create a temporary file

Use `mktemp`  filter to create a temporary file containing the value's contents.

```bash
pipe 'url.get https://example.org as request :: openssl md5 {{request | mktemp}}'
```


### Examples

Get a new stock quote every 30 seconds

```bash
pipe 'every 30s :: print ilmn as stock :: url.get https://api.iextrading.com/1.0/stock/{{stock}}/quote :: json :: select this.iexRealtimePrice'
```
