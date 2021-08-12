FROM golang:1.16-alpine


WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
COPY input.yaml ./

RUN go build -o /monitor
CMD [ "/monitor" ]

