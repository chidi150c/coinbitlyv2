# Use an official Golang runtime as a parent image
FROM golang:1.20.0 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application source code to the container
COPY . .

# Build the Go binary
RUN go build -o myapp

# Set the environment variable
ENV PORT4=35260
ENV HOSTSITE=https://resoledge.com

# Expose the port using the environment variable PORT
EXPOSE $PORT4

# Command to start your Go application
CMD ["./myapp"]
