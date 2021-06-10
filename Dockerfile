FROM golang:1.15.5 as builder
WORKDIR $GOPATH/src/service-health-checker
COPY . .
RUN go build
CMD ["./service-health-checker"]
