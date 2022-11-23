FROM golang:1.19

WORKDIR /app/
COPY . /app/
RUN CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o ip.addr main.go

FROM scratch
COPY --from=0 /app/ip.addr .
EXPOSE 8080
ENTRYPOINT ["/ip.addr"]