# go-config-manager

A flexible and lightweight configuration manager for Go applications, with support for JSON, YAML, and YML formats.

[![Go Report Card](https://goreportcard.com/badge/github.com/Universal-Cube/go-config-manager)](https://goreportcard.com/report/github.com/Universal-Cube/go-config-manager) [![Go Reference](https://pkg.go.dev/badge/github.com/Universal-Cube/go-config-manager.svg)](https://pkg.go.dev/github.com/Universal-Cube/go-config-manager) [![Build Status](https://github.com/Universal-Cube/go-config-manager/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/Universal-Cube/go-config-manager/actions/workflows/main.yml)

## 📋 Overview

go-config-manager provides a simple yet powerful way to manage configuration in Go applications. It abstracts away the
complexity of handling different configuration formats and provides a unified API for accessing configuration values.

## 🚀 Installation

```bash
go get github.com/Universal-Cube/go-config-manager
```

## ✨ Features

- **Multiple Format Support**: Load configurations from JSON, YAML, and YML files
- **Dot Notation Access**: Retrieve nested values using simple dot notation paths (e.g., `database.host`)
- **Type Conversion**: Built-in methods for converting values to different data types (string, int, bool, float, slices)
- **Mutable Configuration**: Modify and save configuration changes at runtime
- **Developer-Friendly API**: Clean, intuitive interface designed for ease of use

## 🔍 Quick Example

```go
package main

import (
    "fmt"
    "github.com/Universal-Cube/go-config-manager/pkg/config"
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

## 📖 Documentation

For full API documentation and examples, visit [pkg.go.dev](https://pkg.go.dev/github.com/Universal-Cube/go-config-manager).

## 📄 License

[MIT License](./LICENSE)
