# go-fetcher

This repository contains code for a `go get` routing service which allows
packages from multiple locations to appear to be centralized in one location.
This also allows the packages to be moved to other locations without breaking
imports.

## Building

In order to build the package, the following dependencies must be met:

* Go should be installed and in the `PATH`
* `GOPATH` should be set as described [here](http://golang.org/doc/code.html)
* Ensure that `$GOPATH/bin` is in your `PATH`
  
To build the Go binary:

```
go get github.com/cloudfoundry/go-fetcher
```

### Running Tests

We are using [Ginkgo](https://github.com/onsi/ginkgo) to run tests. 

Be sure to install [protobuf](github.com/golang/protobuf/proto) with `go get github.com/golang/protobuf/proto`.

Run `ginkgo` from the root of the repository to run all tests.

## Configuring

You will need to create a configuration for `go-fetcher`:

```
cat > config.json << END
{
  "ImportPrefix": "example.com",
  "OrgList": [
    "https://github.com/cloudfoundry/",
    "https://github.com/cloudfoundry-incubator/",
    "https://github.com/cloudfoundry-attic/"
  ],
  "NoRedirectAgents": [
    "Go-http-client",
    "GoDocBot"
  ],
  "Overrides": {
    "stager": "https://github.com/cloudfoundry-incubator/stager"
  }
}
END
```
* The value of "ImportPrefix" is the DNS name of the `go-fetcher` service (ex: example.com).
* The value of "OrgList" is a list of `go get` compatible sites that are searched in order.
* The value of "Overrides" is a dictionary of packages which should not use the normal search path.


## Running locally

The `go-fetcher` program expects environment variables to be set which indicate the port to listen on and the name of the configuration file.

```
./generate_local && PORT=8800 CONFIG="config.json" go run main.go
```

# Deploying to Cloud Foundry

Deploying to Cloud Foundry is straight forward, but requires you to do so from the checked out repository so that `cf` can recognize and upload the package. You will need to create a `manifest.yml` to accompany your `config.json`:

```
cat > manifest.yml << END
application:
  - name: example
env:
  GOPACKAGENAME: github.com/cloudfoundry/go-fetcher/cmd/go-fetcher
  GOVERSION: go1.6
  CONFIG: config.json
END
```


Then, login to your cf instance and push the application:

```
cf login ...
cf push example
```
