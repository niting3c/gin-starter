# Gin Starter Application

## TL;DR:
Gin + Go Wire + PGX + CRUD + Mockery + Swagger + Prometheus

Run ` go get github.com/google/wire/cmd/wire `
Run: `swag init -g ./cmd/app/main.go`

Run : `brew install mockery`
Make sure the path is up to date for the command to work 
Run : ` go generate ./...`

### Unit Tests

- To run Unit tests please run this:

```
go test  $(go list ./... |  grep -Ev  '/docs|mock|mocks|wire|config|mocks/' )  -coverprofile coverage.out -covermode=atomic -count=1
go tool cover -html=cover.out
```

This should open up your default browser and show the coverage

### SonarQube

- Install sonarqube in local and setup a project
- Remember the authentication and other details
- Once you run the below command , sonarqube should display the project details under its project UI

Run :

```
sonar-scanner -Dsonar.token=token
```