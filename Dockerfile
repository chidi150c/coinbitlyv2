# Create the build environment
FROM golang:1.20.1 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o coinbitly .

# Create the final lightweight image
FROM scratch

# Copy the built binary from the build environment
COPY --from=builder /app/coinbitly /coinbitly

# Set the environment variable
ENV PORT=35259

# Expose the port using the environment variable PORT
EXPOSE $PORT

# Command to start your Go application
CMD ["/coinbitly"]
