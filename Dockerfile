FROM golang:latest AS builder

# Download and install the latest release of dep
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

# Copy the code from the host and compile it
WORKDIR /go/src/github.com/Draska/currencyAlarm/
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY ./dollar-watch/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /dollar-watch .

FROM scratch
COPY --from=builder /dollar-watch ./
ENTRYPOINT ["./dollar-watch"]
