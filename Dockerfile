# Create the final lightweight image
FROM golang:1.20.1 AS builder

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
