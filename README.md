# Simiload

Simulation of an Overload Protection System

`make` compiles `server.go` and runs a docker-compose cluster with:

- Simulation server: `localhost:8080`
- Metrics server: `localhost:3000`
- Load generator: `ruby generate.rb`
