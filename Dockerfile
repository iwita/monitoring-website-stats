FROM golang:1.13-alpine


RUN apk update
RUN apk add git


WORKDIR /go/src/github.com/iwita/monitoring-website-stats
COPY go.mod ./
COPY go.sum ./
RUN go clean -modcache
RUN go mod download
COPY *.go ./
COPY pkg/ ./pkg
COPY files ./files


RUN go build -o ./monitor
CMD [ "./monitor" ]

