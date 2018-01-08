# mohotani
A bot to keep your DNS records up to date

Mohotani is designed to keep your DNS (A) records up to date with your current application needs.

It currently supports a single host, assumes that the whole application is served by a single reverse proxy
(i.e. all DNS have the same A records) and can even act as a dynamic DNS provider.

Mohotani is mainly composed of 3 components:

- DNS lister that provides the list of all domains your application needs
- IP resolver that resolves ip addresses A records shall resolve
- DNS provider to update each A records

## Supported DNS provider

Mohotani supports natively [gandi live DNS](http://doc.livedns.gandi.net/) api as well as logging the changes to stdout.

## Supported IP resolver

Mohotani supports resolving static IP addresses provided on command line as well as polling public IP addresses using
the [IPIFY](https://www.ipify.org/) service.

## Supported domain lister

Mohotani supports resolving required domains provided on command line as well as polling docker setup and extract required domains
from the [traefik Host matcher](https://docs.traefik.io/basics/#matchers). The labels are extracted from the `traefik.frontend.rule` label
from either running containers or created services.

### Docker support

The docker support can be achieved on a single node. In such a case, mohotani should be provided an access to the docker host, either by running
it on the same host, or to provide it the `DOCKER_HOST` (and possibly [other](https://docs.docker.com/engine/reference/commandline/cli/#environment-variables))
environment variable.

It can also support swarm mode deployements. In such a case, mohotani should be provided an access to any swarm manager, as before, either by running on the same
host or by providing the `DOCKER_HOST` environment variable.

The setup can be checked by running:

```
docker container ls # for container support
docker service ls # for swarm mode support
```

# Get mohotani

Mohotani can be installed from any [go environment](https://golang.org/doc/install) by runing the following:

```
go get -u github.com/tjamet/mohotani/cli/mohotani
mohotani --help
```

Or by running a container:

```
docker run --rm tjamet/mohotani --help
```

In order to run mohotani in a container with docker support, you should consider exposing the docker socket to the container:

```
docker run -d --restart always -v /var/run/docker.sock:/var/run/docker.sock tjamet/mohotani <...>
```

When using the swarm mode, you can even create a service and benefit the docker way to manage and share secrets.
For example to run the mohotani using gandi live DNS together with the docker traefik parsing you could run:

```
echo '<api key>' | docker secret create gandi-api-key -
docker service create \
    --name mohotani \
    --mount type=bind,source=/var/run/docker.sock,destination=/var/run/docker.sock \
    --secret gandi-api-key \
    --constraint 'node.role==manager' \
    tjamet/mohotani \
    --gandi --gandi.key-file /run/secrets/gandi-api-key \
    --domains.docker \
    --ips.ipify
```

# Contributing

Mohotani is designed to support much more providers and backends and I will be happy to receive contributions.
If you have questions, suggestions, bugs please open an issue. If you want to submit the code, create the issue and send us a pull request for review.