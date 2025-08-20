# Debugging Golang Applications with mirrord

<div align="center">
  <a href="https://mirrord.dev">
    <img src="images/mirrord.svg" width="150" alt="mirrord Logo"/>
  </a>
  <a href="https://go.dev">
    <img src="images/go.svg" width="150" alt="Go Logo"/>
  </a>
</div>

## Overview

This is a sample Guestbook application built with Go and Redis to demonstrate debugging Kubernetes applications using mirrord. The application allows users to write messages to a guestbook, which are stored in Redis and displayed on the web interface.

## Prerequisites

- Go 1.23.2 or higher
- Docker and Docker Compose
- Kubernetes cluster
- mirrord CLI installed

## Quick Start

1. Clone the repository:

```bash
git clone https://github.com/waveywaves/mirrord-go-debug-example
cd mirrord-go-debug-example
```

2. Deploy to Kubernetes:

```bash
kubectl create -f ./kube
``` 

3. Debug with mirrord:

```bash
mirrord exec -t deployment/guestbook go run main.go
```

The application will be available at http://localhost:3000

## Architecture

The application consists of:
- Go web server
- Redis master for write operations
- Redis replica for read operations

## License

This project is licensed under the MIT License - see the LICENSE file for details.