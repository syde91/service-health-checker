FROM golang:1.15.5 as builder
WORKDIR $GOPATH/src/service-health-check
COPY . .
RUN pwd
RUN ls
RUN go build
CMD ["./service-health-check"]
