[![Go Report Card](https://goreportcard.com/badge/github.com/cty3000/superman-detector)](https://goreportcard.com/report/github.com/cty3000/superman-detector) [![GoDoc](http://godoc.org/github.com/cty3000/superman-detector?status.svg)](http://godoc.org/github.com/cty3000/superman-detector)

---

Table of Contents
=================

- [Installation](#installation)
- [Usage](#usage)
- [Example request](#example-request)
- [External Libraries](#external-libraries)
- [Generating test coverage](#generating-test-coverage)
- [Building a docker image](#building-a-docker-image)
- [TODO](#todo)

## Installation
``` bash
$ make
```

## Usage
``` bash
$ ./superman-detector
1970/01/01 00:00:00 Initialized SupermanDetector service at 'http://0.0.0.0:80/'
```

## Example request
``` bash
$ curl -X POST -H "Content-Type: application/json" -d "{\
    \"username\":\"bob\",\
    \"unix_timestamp\":1514761200,\
    \"event_uuid\":\"85ad929a-db03-4bf4-9541-8f728fa12e42\",\
    \"ip_address\":\"91.207.175.104\"\
  }" http://localhost/;
{
  "currentGeo": {
    "lat": 34.0549,
    "lon": -118.2578,
    "radius": 200
  }
}

$ curl -X POST -H "Content-Type: application/json" -d "{\
    \"username\":\"bob\",\
    \"unix_timestamp\":1514851200,\
    \"event_uuid\":\"85ad929a-db03-4bf4-9541-8f728fa12e40\",\
    \"ip_address\":\"24.242.71.20\"\
  }" http://localhost/;
{
  "currentGeo": {
    "lat": 30.3773,
    "lon": -97.71,
    "radius": 5
  },
  "travelToCurrentGeoSuspicious": false,
  "precedingIpAccess": {
    "ip": "91.207.175.104",
    "speed": 49,
    "lat": 34.0549,
    "lon": -118.2578,
    "radius": 200,
    "timestamp": 1514761200
  }
}

$ curl -X POST -H "Content-Type: application/json" -d "{\
    \"username\":\"bob\",\
    \"unix_timestamp\":1514764800,\
    \"event_uuid\":\"85ad929a-db03-4bf4-9541-8f728fa12e41\",\
    \"ip_address\":\"206.81.252.7\"\
  }" http://localhost/;
{
  "currentGeo": {
    "lat": 39.2293,
    "lon": -76.6907,
    "radius": 10
  },
  "travelToCurrentGeoSuspicious": true,
  "travelFromCurrentGeoSuspicious": false,
  "precedingIpAccess": {
    "ip": "91.207.175.104",
    "speed": 2311,
    "lat": 34.0549,
    "lon": -118.2578,
    "radius": 200,
    "timestamp": 1514761200
  },
  "subsequentIpAccess": {
    "ip": "24.242.71.20",
    "speed": 55,
    "lat": 30.3773,
    "lon": -97.71,
    "radius": 5,
    "timestamp": 1514851200
  }
}
```

## External Libraries

External dependencies are listed here:

https://godoc.org/github.com/cty3000/superman-detector?imports

Also [the codes that are documented on the below page](./supermandetector) are generated by [ardielle-tools](https://github.com/ardielle/ardielle-tools)

https://godoc.org/github.com/cty3000/superman-detector/supermandetector

## Generating test coverage
``` bash
$ make coverage
```

Test coverage: [coverage.html](https://cty3000.github.io/superman-detector/coverage.html#file0)

## Building a docker image
``` bash
$ make docker-build
```

The build process and runtime image are both built in [Dockerfile](./Dockerfile) as a multi-stage Docker container.

## TODO
- Improve unit tests coverage
- Support the newest go version 1.12 (currently tested and built on 1.11)
