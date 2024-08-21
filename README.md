# Snippetbox - A Pastebin Clone

## Overview

Snippetbox was developed by coding along the [Let's Go book by Alex Edwards](https://lets-go.alexedwards.net/)

Snippetbox is a full stack web application that attempts to clone the functionality of Pastebin.

## Features

    - Account creation / Authentication
    - Snippet creation, saving and viewing
    - Data persistance through MySql
    - Dynamic HTMl through Go Templates
    - Custom Middleware
    - State Management
    - HTTPS Security
    - Embedded File Systems
    - Unit Tests
    - E2E Tests

## Executing

### Requirements:
    - MySql
    - Go Version 1.20 +

### Configuring:
    - Create a ".Env" file with the following properties:
        - SQL_USER = "Your User"
        - SQL_PASSWORD = "Your Password"
    - Execute the following command to generate the database:
        - go run ./internal/models/database/

### Starting The Application
    - Execute the following command:
        - go run ./cmd/web
