# go-ocelot
A reverse proxy written in golang

## Status
Currently under development (Jan 12, 2017)

## To run locally
    $ go install github.com/ocelotconsulting/go-ocelot
    $ $GOPATH/bin/go-ocelot

## Docker
1. Build the image

        $ docker build --build-arg GO_MAIN=github.com/ocelotconsulting/go-ocelot --build-arg GO_MAIN_EXEC=go-ocelot -t go-ocelot:dev .

2. To run using docker:

        $ docker run --name go-ocelot -p 8082:8080 -d go-ocelot:dev
