# nwm-forecasts
This is a containerized application for retrieving National Water Model Short Range stream flow forecasts, Medium Range stream flow forecasts, and Retrospective stream flow products.


## Getting Started

These instructions will cover usage information for the [dewberrycsi/nwm-fetch-forecasts](https://hub.docker.com/repository/docker/dewberrycsi/nwm-fetch-forecasts) docker container. This docker image is a light weight, multistage build containing only the compiled Go executable and any C dependencies. In order to alter the image, forking this repository and altering the Dockerfile may be necessary. Otherwise, if you want to start with a Docker container that only has dependencies and the compiled executable file, begin your Dockerfile with `FROM dewberrycsi/nwm-fetch-forecasts:latest`.

### Prerequisities


In order to run this container you'll need docker installed.

* [Windows](https://docs.docker.com/windows/started)
* [OS X](https://docs.docker.com/mac/started/)
* [Linux](https://docs.docker.com/linux/started/)

### Usage

#### Container Parameters

 - `-date_time`: the date and time you want to pull stream flow for, in the format "2006-01-02-15"
    - If not entered, the most recent forecast is pulled 
    - The earliest available data is from 1993-01-01
    - If the date requested is between 1993-01-01 and 2018-09-17, 5 days before and after the requested date will be pulled from the retrospective product no matter what product is requested
 - `-product`: the product requested. Can be either `short` or `medium`. 
    - If not entered, short range is assumed
    - If `medium` is specified and the date is before 2019-06-02, there will be a single ensemble member. If it is after that day, all ensemble members will be retrieved  
 - `-index`: the location in the netcdf files to pull stream flow
    - all arguments tacked on to the end are assumed to be additional indicies

#### Examples

Grabs short range forecast streamflow for multiple indices
```shell
docker run dewberrycsi/nwm-fetch-forecasts:latest -date_time "2019-01-02-15" -product short -index 900 101 181 1030 ...
```

Grabs the latest short range forecast at index 1234
```shell
docker run dewberrycsi/nwm-fetch-forecasts:latest -index 1234
```

Grabs the medium range forecast at multiple indices
```shell
docker run dewberrycsi/nwm-fetch-forecasts:latest -date_time "2019-01-02-15" -product medium -index 900 101 181 1030 ...
```

Grabs retrospective data at multiple indices
```shell
docker run dewberrycsi/nwm-fetch-forecasts:latest -date_time "1999-01-02-15" -index 900 101 181 1030 ...
```

#### Outputs

Results will be output to `STDOUT` as a json, and can be easily piped to a file as needed. For example:
```shell
docker run dewberrycsi/nwm-fetch-forecasts:latest -index 1234 > results.json
```

#### File locations
The main Go executable is located at `/main`


## Built With

* [Go](https://golang.org/) v1.14
* [Goofys](https://github.com/kahing/goofys) v0.24.0
* [github.com/fhs/go-netcdf](https://github.com/fhs/go-netcdf) v1.1.0

## Find Us

* [GitHub](https://github.com/Dewberry/nwm-forecasts)
* [DockerHub](https://hub.docker.com/repository/docker/dewberrycsi/nwm-fetch-forecasts)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
