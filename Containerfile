FROM registry.access.redhat.com/ubi8/go-toolset:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN GO111MODULE=on go mod download

COPY . ./

USER root

RUN go build -o app

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/public.crt .

CMD ["/app/app"]
