# go-fetcher

This repository contains code for a `go get` routing service which allows
packages from multiple locations to appear to be centralized in one location.
This also allows the packages to be moved to other locations without breaking
imports.

## Development

First, be sure you [install Go](https://golang.org/doc/install). Then clone
the repository:

```
git clone https://github.com/cloudfoundry/go-fetcher.git
cd go-fetcher/
```

We are using [Ginkgo](https://github.com/onsi/ginkgo) to run tests. You may
use the `test` command which uses [counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) to
generate fakes:

```
./bin/test
```

If you are changing interfaces, you can rebuild fakes with the
following:

```
go generate ./...
```


Some environment variables and [configuration](#configuration) are required to run. For local testing, you may use the `run_local` command and point your browser to [localhost:8800](http://localhost:8800):

```
./bin/run_local
```

## Configuration

See the following for an example of what configuration should look like:

```shell
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

## Deploying to Cloud Foundry

Deploying to Cloud Foundry is straight forward, but requires you to do so from the checked out repository so that `cf` can recognize and upload the package. You will need to create a `manifest.yml` to accompany your `config.json`:

```shell
cat > manifest.yml << END
application:
  - name: example
    env:
      CONFIG: config.json
END
```

Then, login to your cf instance and push the application:

```shell
cf login ...
cf push example
```
