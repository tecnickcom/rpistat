# rpistat

*Web-Service to collext Raspberry pi 4 system statistics*

[![Donate via PayPal](https://img.shields.io/badge/donate-paypal-87ceeb.svg)](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&currency_code=GBP&business=paypal@tecnick.com&item_name=donation%20for%20rpistat%20project)
*Please consider supporting this project by making a donation via [PayPal](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&currency_code=GBP&business=paypal@tecnick.com&item_name=donation%20for%20rpistat%20project)*

* **category**    Application
* **author**      Nicola Asuni <info@tecnick.com>
* **copyright**   2022-2023 Nicola Asuni - Tecnick.com LTD
* **license**     MIT see [LICENSE](LICENSE)
* **link**        https://github.com/tecnickcom/rpistat

-----------------------------------------------------------------

## TOC

* [Description](#description)
* [Quick Start](#quickstart)

-----------------------------------------------------------------

<a name="description"></a>
## Description

Web-Service to collext Raspberry pi 4 system statistics when running Raspbian 64 bit.

-----------------------------------------------------------------

<a name="quickstart"></a>
## Quick Start

This project includes a Makefile that allows you to test and build the project in a Linux-compatible system with simple commands.  
All the artifacts and reports produced using this Makefile are stored in the *target* folder.  

All the packages listed in the *resources/docker/Dockerfile* file are required in order to build and test all the library options in the current environment.
Alternatively, everything can be built inside a [Docker](https://www.docker.com) container using the command "make dbuild".

To see all available options:
```
make help
```

To download all dependencies:
```
make deps
```

To update the mod file:
```
make mod
```

To format the code (please use this command before submitting any pull request):
```
make format
```

To execute all the default test builds and generate reports in the current environment:
```
make qa
```

To build the executable file:
```
make build
```


-----------------------------------------------------------------

