# syntax=docker/dockerfile:1

##
## Build
##

FROM golang:1.18-buster AS build
WORKDIR /app
COPY . /app
RUN go build -o /grpc-echo

##
## Deploy
##

FROM gcr.io/distroless/base-debian10
WORKDIR /app
COPY --from=build /grpc-echo /grpc-echo
USER nonroot:nonroot
ENTRYPOINT ["/grpc-echo"]