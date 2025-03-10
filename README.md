# cfg-manager

A flexible and lightweight configuration manager for Go applications, with support for JSON, YAML, and YML formats.

[![Go Report Card](https://goreportcard.com/badge/github.com/Universal-Cube/cfg-manager)](https://goreportcard.com/report/github.com/Universal-Cube/cfg-manager) [![Go Reference](https://pkg.go.dev/badge/github.com/Universal-Cube/cfg-manager.svg)](https://pkg.go.dev/github.com/Universal-Cube/cfg-manager) [![Build Status](https://github.com/Universal-Cube/cfg-manager/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/Universal-Cube/cfg-manager/actions/workflows/main.yml)

## 📋 Overview

cfg-manager provides a simple yet powerful way to manage configuration in Go applications. It abstracts away the
complexity of handling different configuration formats and provides a unified API for accessing configuration values.

## 🚀 Installation

```bash
go get github.com/Universal-Cube/cfg-manager
```

## ✨ Features

- **Multiple Format Support**: Load configurations from JSON, YAML, and YML files
- **Dot Notation Access**: Retrieve nested values using simple dot notation paths (e.g., `database.host`)
- **Type Conversion**: Built-in methods for converting values to different data types (string, int, bool, float, slices)
- **Mutable Configuration**: Modify and save configuration changes at runtime
- **Developer-Friendly API**: Clean, intuitive interface designed for ease of use
- **Struct Binding**: Automatically bind configuration values to Go structs using tags

## 🔍 Quick Example

```go
package main

import (
	"fmt"
	"github.com/Universal-Cube/cfg-manager/pkg/config"
)

func main() {
	// Load configuration from file
	cfg, err := config.LoadFromFile("config.json")
	if err != nil {
		panic(err)
	}

	// Access values using dot notation
	host := cfg.G("database.host")
	port := cfg.GetInt("database.port")

	fmt.Printf("Database connection: %s:%d\n", host, port)

	// Modify and save configuration
	cfg.Set("database.host", "new-host.example.com")
	cfg.SaveToFile("config.json")
}
```

## 🔄 Struct Binding

Automatically bind configuration to Go structs with tags:

```go
package main

import (
	"fmt"
	"github.com/Universal-Cube/cfg-manager/pkg/config"
)

type Config struct {
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"database"`

	App struct {
		Debug   bool `json:"debug"`
		Timeout int  `json:"timeout"`
	} `json:"app"`
}

func main() {
	cfg := config.New()
	err := cfg.LoadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	var appConfig Config
	err = cfg.Bind(&appConfig)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Database Host: %s\n", appConfig.Database.Host)
}
```

## 📖 Documentation

For full API documentation and examples, visit [pkg.go.dev](https://pkg.go.dev/github.com/Universal-Cube/cfg-manager).

## 📄 License

[MIT License](./LICENSE)
