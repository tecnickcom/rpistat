<!-- Space: <no value> -->
<!-- Parent: Projects -->
<!-- Title: rpistat -->
<!-- Attachment: doc/images/logo.png -->

# rpistat

*Web-Service to collect system usage statistics.*

[![Donate via PayPal](https://img.shields.io/badge/donate-paypal-87ceeb.svg)](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&currency_code=GBP&business=paypal@tecnick.com&item_name=donation%20for%20rpistat%20project)
*Please consider supporting this project by making a donation via [PayPal](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&currency_code=GBP&business=paypal@tecnick.com&item_name=donation%20for%20rpistat%20project)*

![rpistat logo](doc/images/logo.png)

* **category:**    Application
* **copyright:**   2022-2024 Nicola Asuni - Tecnick.com LTD
* **license:**     [LICENSE](https://github.com/tecnickcom/rpistat/blob/main/LICENSE)
* **cvs:**         https://github.com/tecnickcom/rpistat

[![check](https://github.com/tecnickcom/rpistat/actions/workflows/check.yaml/badge.svg)](https://github.com/tecnickcom/rpistat/actions/workflows/check.yaml)

----------

## TOC
* [Description](#description)
* [Documentation](#documentation)
	* [public](#documentation_public)
		* [General](#documentation_public_general)
* [Development](#development)
* [Deployment](#deployment)

----------

<a name="description"></a>
## Description
Web-Service to collect system usage statistics.

This project was started to collect system usage statistics from a remote [Raspberry PI 4](https://www.raspberrypi.com/products/raspberry-pi-4-model-b/) but it can be compiled and used with other systems too.

The statistics are exposed both via the ```/stats``` Web API endpoint and [Prometheus](https://prometheus.io/) metrics at ```/metrics```.

An [HomeAssistant](https://www.home-assistant.io/) sensor and template to collect those metrics can be found in the ```resources/HomeAssistant/``` directory.


### REST API Example:

```
curl 192.168.1.2:65501/stats

{

      "datetime": "2023-06-27T07:03:54Z",
      "timestamp": 1687849434569669934,
      "hostname": "rpi4",
      "uptime": 680189000000000,
      "memory_total": 1005973504,
      "memory_free": 587198464,
      "memory_used": 418775040,
      "memory_usage": 0.41628833993623754,
      "load_1m": 0,
      "load_5m": 0,
      "load_15m": 0,
      "temperature_cpu": 44.303,
      "disk_total": 31109140480,
      "disk_free": 27586793472,
      "disk_used": 3522347008,
      "disk_usage": 0.11322546858099501,
      "network": [
            {
                  "nic": "eth0",
                  "rx": 163621268,
                  "tx": 113155073
            },
            {
                  "nic": "wlan0",
                  "rx": 8225055,
                  "tx": 1327982
            },
            {
                  "nic": "eth1",
                  "rx": 24906794,
                  "tx": 18615044
            }
      ]

}
```

### HomeAssistant Screenshot

![HomeAssistant screenshot](doc/images/homeassistant.png)


----------



<a name="documentation"></a>
## Documentation
<a name="documentation_public"></a>
* public
	<a name="documentation_public_general"></a>
	* General  
	_General project documentation_
		* [GitHup project page](https://github.com/tecnickcom/rpistat)


----------



<a name="development"></a>
## Development
### TOC

* [Style and Conventions](#style)
* [Requirements](#requirements)
* [Quick Start](#quickstart)
* [Running all tests](#runtest)
* [Documentation](#gendoc)
* [Usage](#usage)
* [Configuration](CONFIG.md)
* [Examples](#examples)
* [Logs](#logs)
* [Metrics](#metrics)
* [Profiling](#profiling)
* [OpenAPI](#openapi)
* [Docker](#docker)


<a name="style"></a>
## Style and Conventions

For the general style and conventions, please refer to external documents:
https://github.com/uber-go/guide/blob/master/style.md


<a name="requirements"></a>
## Requirements

* [check-jsonschema](https://pypi.org/project/check-jsonschema/) to check the validity of the JSON configuration files against the JSON schema.

```bash
sudo pip install --upgrade check-jsonschema
```


<a name="quickstart"></a>
## Developers' Quick Start

To quickly get started with this project, follow these steps:

1. Ensure you ahev installed the latest Go version and Python3 for some extra tests.
1. Clone the repository: `git clone https://github.com/tecnickcom/rpistat.git`.
2. Change into the project directory: `cd rpistat`.
3. Install the required dependencies and test everything: `DEVMODE=LOCAL make x`.

Now you are ready to start developing with rpistat!


This project includes a *Makefile* that allows you to test and build the project in a Linux-compatible system with simple commands.  
All the artifacts and reports produced using this *Makefile* are stored in the *target* folder.  

Alternatively, everything can be built inside a [Docker](https://www.docker.com) container using the command `make dbuild` that uses the environment defined at `resources/docker/Dockerfile.dev`.

To see all available options:
```bash
make help
```


<a name="runtest"></a>
## Running all tests

Before committing the code, please check if it passes all tests using
```bash
make x
```

that is an alias for:
```bash
DEVMODE=LOCAL make format clean mod deps generate qa build docker dockertest
```


<a name="gendoc"></a>
## Documentation

The `README.md` documentation file is generated using the source templates in `doc/src` via `make gendoc` command.

To update links and common information edit the file `doc/src/config.yaml` in YAML format.
The schema of the configuration file is defined by the JSON schema: `doc/src/config.schema.json`.
The document templates are defined by the `*.tmpl` files in [gomplate](https://docs.gomplate.ca)-compatible format.

To regenerate the static documentation file:
```bash
make gendoc
```


<a name="usage"></a>
## Usage

```bash
rpistat [flags]

Flags:

-c, --configDir  string  Configuration directory to be added on top of the search list
-f, --logFormat  string  Logging format: CONSOLE, JSON
-o, --loglevel   string  Log level: EMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFO, DEBUG
```

<a name="examples"></a>
## Examples

Once the application has being compiled with `make build`, it can be quickly tested:

```bash
target/usr/bin/rpistat -c resources/test/etc/rpistat
```


<a name="logs"></a>
## Logs

This program logs the log messages in JSON format:

```
{
	"level": "info",
	"timestamp": 1595942715776382171,
	"msg": "Request",
	"program": "rpistat",
	"version": "0.0.0",
	"release": "0",
    "hostname":"myserver",
	"request_id": "c4iah65ldoyw3hqec1rluoj93",
	"request_method": "GET",
	"request_path": "/uid",
	"request_query": "",
	"request_uri": "/uid",
	"request_useragent": "curl/7.69.1",
	"remote_ip": "[::1]:36790",
	"response_code": 200,
	"response_message": "OK",
	"response_status": "success",
	"response_data": "avxkjeyk43av"
}
```

Logs are sent to stderr by default.

The log level can be set either in the configuration or as command argument (`logLevel`).


<a name="metrics"></a>
## Metrics

This service provides [Prometheus](https://prometheus.io/) metrics at the `/metrics` endpoint.


<a name="profiling"></a>
## Profiling

This service provides [PPROF](https://github.com/google/pprof) profiling data at the `/pprof` endpoint.

The pprof data can be analyzed and displayed using the pprof tool:

```
go get github.com/google/pprof
```

Example:

```
pprof -seconds 10 -http=localhost:8182 http://INSTANCE_URL:PORT/pprof/profile
```


<a name="openapi"></a>
## OpenAPI

The rpistat API is specified via the [OpenAPI 3](https://www.openapis.org/) file: `openapi.yaml`.

The openapi file can be edited using the Swagger Editor:

```
docker pull swaggerapi/swagger-editor
docker run -p 8056:8080 swaggerapi/swagger-editor
```

and pointing the Web browser to http://localhost:8056


<a name="docker"></a>
## Docker

To build a Docker scratch container for the rpistat executable binary execute the following command:
```
make docker
```

### Useful Docker commands

To manually create the container you can execute:
```
docker build --tag="tecnickcom/rpistatdev" .
```

To log into the newly created container:
```
docker run -t -i tecnickcom/rpistatdev /bin/bash
```

To get the container ID:
```
CONTAINER_ID=`docker ps -a | grep tecnickcom/rpistatdev | cut -c1-12`
```

To delete the newly created docker container:
```
docker rm -f $CONTAINER_ID
```

To delete the docker image:
```
docker rmi -f tecnickcom/rpistatdev
```

To delete all containers
```
docker rm $(docker ps -a -q)
```

To delete all images
```
docker rmi $(docker images -q)
```


----------

<a name="deployment"></a>
## Deployment
After building the executable binary with `make build`:

* Copy the `target/usr/bin/rpistat` file into `/usr/bin/rpistat` in the target system.
* Set the file is executable with: `sudo chmod +x /usr/bin/rpistat`.
* Copy and edit the configuration file `resources/etc/rpistat/config.json` into `/etc/rpistat/config.json` in the target system.
* Copy the service file `resources/etc/systemd/system/rpistat.service` into `/etc/systemd/system/rpistat.service` in the target system and enable it:
```
sudo systemctl daemon-reload
sudo systemctl enable rpistat.service
sudo systemctl start rpistat.service
```


----------


