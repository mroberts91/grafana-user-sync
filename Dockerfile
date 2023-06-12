# syntax=docker/dockerfile:1

FROM golang:1.20-alpine3.18

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
ADD logging logging
COPY usersync usersync
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /grafana-user-sync

# Run
CMD ["/grafana-user-sync"]