FROM scratch

WORKDIR /build

COPY public public

COPY server .

CMD ["./server"]

# команда для сборки на винде $env:CGO_ENABLED=0; $env:GOOS="linux"; go build -o server main.go