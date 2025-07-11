# gotailwindcss

A small Go wrapper around the tailwind cli.  

## Purpose
I find it offensive to have Node installed on a computer. 
The goal of this project is to allow management of tailwind via the go tool command.

## Installation
Starting with Go 1.24 the [Tools](https://tip.golang.org/doc/modules/managing-dependencies#tools) command was added, 
this adds and tracks the dependencies in the go.mod file
```bash
go get -tool github.com/TaylorOno/gotailwindcss
```

## Execution
After the go tool has been installed in your project, you can run it with the standard tailwind parameters
```bash
go tool gotailwindcss -i ./css/tailwind.css -o ./css/styles.css
```

`gotailwindcss` will default to using the latest tailwindcli [release](https://github.com/tailwindlabs/tailwindcss/releases) there may be scenarios where this is not desirable, 
in the event you want to use a specific version of tailwind you can set the environment variable `TAILWINDCSS_VERSION`

```bash
export TAILWINDCSS_VERSION=v4.1.10 
go tool gotailwindcss -i ./css/tailwind.css -o ./css/styles.css
```