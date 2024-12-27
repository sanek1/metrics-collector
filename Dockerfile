FROM golang:1.23

WORKDIR /app

# Copy the entire Go module directory into the container
COPY . .

# Install dependencies using go mod tidy
RUN go mod tidy

# Build the application binary
RUN go build -o zap-app ./cmd/server

# Set the command to run when the container starts
CMD ["./zap-app"]