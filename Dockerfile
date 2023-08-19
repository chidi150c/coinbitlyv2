# Use a base image that matches your environment
FROM debian:bullseye-slim

# Install necessary dependencies
RUN apt-get update && \
    apt-get install -y golang:1.21.0 nano curl git && \
    rm -rf /var/lib/apt/lists/*

# Set the working directory inside the container
WORKDIR /

# Copy the built binary from the host into the container
COPY coinbitly .
COPY webClient ./webClient

# Set the environment variable
ENV PORT=35259

# Expose the port using the environment variable PORT4
EXPOSE $PORT

# Command to start your Go application
CMD ["./coinbitly"]

