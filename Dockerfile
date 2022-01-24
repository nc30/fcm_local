FROM golang:1.17.5-alpine3.15 AS builder
WORKDIR /opt
ADD . /opt
RUN go get
RUN go build -o server main.go

FROM alpine:latest AS runner
EXPOSE 8080
COPY --from=builder /opt/server /server

CMD ["/server", "-addr", ":8080"]
