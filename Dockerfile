FROM golang:1.16

WORKDIR /go/src/mackerel-plugin-scheduler-latency-kvm
COPY . .

RUN go get -d -v ./...
RUN go build
