# Pipe

Powerful continuous stream processing

[![Build Status](https://travis-ci.org/relvacode/pipe.svg?branch=master)](https://travis-ci.org/relvacode/pipe)

Run a series of commands or pipes concurrently and manipulate streams using rich native objects.


```
go install github.com/relvacode/pipe/cmd/pipe
```


### Basics

#### Pipes

A pipe can either be a built-in native pipe, or fallback to a program on the system.

To join the output of one pipe to another use `::`

The first pipe in a pipeline is given a `stdin` object and all outputs of the last pipe are echoed to `stdout`.

#### Templating

Use Django style template provided by [Pongo2](https://github.com/flosch/pongo2) to access values in pipe arguments.

`this` always refers to the current object

#### Filtering

You can use `if` to filter by expression using [Expr](https://github.com/antonmedv/expr)

```bash
# Open all non-empty JSON files and decode their contents
pipe 'open *.json :: if this.Size > 0 :: json' 
```


### Advanced

#### Templating

##### Create a temporary file

You can use the `mktemp` template filter to generate a temporary file that contains the object's contents.
The value of the filter will be replaced by the path to the temporary file on the system.

```bash
pipe 'url https://example.org as request | openssl md5 {request | mktemp}'
```


### Examples

Get a new stock quote every 30 seconds

```bash
every 30s :: print ilmn as stock :: url https://api.iextrading.com/1.0/stock/{{stock}}/quote :: json :: select this.iexRealtimePrice
```