# go-tailwindcss

A small go wrapper around the tailwind cli.  

## Purpose
I find it offensive to have Node installed on a computer. 
The goal of this project is to allow management of tailwind via the go tool command.

## Installation
Starting with go 1.24 [Tools](https://tip.golang.org/doc/modules/managing-dependencies#tools) this adds and tracks the dependencies in the go.mod file
```bash
go get -tool github.com/TaylorOno/gotailwindcss
```

## Execution
After the go tool has been installed in your project, you can run it with the standard tailwind parameters
```bash
go tool gotailwindcss -i ./css/tailwind.css -o ./css/styles.css
```