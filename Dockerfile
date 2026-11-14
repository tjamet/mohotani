FROM golang:1.26-alpine as build

WORKDIR /src

COPY go.mod go.sum ./
COPY ./vendor ./vendor
COPY ./cli ./cli
COPY ./dns ./dns
COPY ./ip ./ip
COPY ./listener ./listener
COPY ./logger ./logger

RUN CGO_ENABLED=0 go build -o /bin/mohotani ./cli/mohotani

# cannot run the race detector on alpine: https://github.com/golang/go/issues/14481
RUN go test ./cli/... ./dns/... ./ip/... ./listener/... ./logger/...


FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/mohotani /bin/mohotani
ENTRYPOINT ["/bin/mohotani"]
