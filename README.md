# repgen

Golang implementation of a reporting service

# Setup

## Requirements

- Go 1.18
- PostgreSQL 14.2


## Installation

Install Go modules, located in ```go.mod``` file, by:

```go mod download```

TODO: Init database...

## Running

In order to start the server, following command can be used:

```go run main.go```

## Compile and Run

Run the following code to compile the whole project:

```go build```

After that, run the project by:

```./repgen```

# Features

# TODO

- Report Select API
    - Parse time interval
    - Return report values
- Logging
- User authorization on APIs
- Report history (Create & Edit) (User, time, column name etc.)
    - Report Creation
    - Report Column create & edit
    - Update token
- Report data table index (report_date) performance
