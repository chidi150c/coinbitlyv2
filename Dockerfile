# Create the final lightweight image
FROM golang:1.16 AS builder

# Set the working directory inside the container
WORKDIR /

# Copy the built binary from the host into the container
COPY conbitly .
COPY webClient ./webClient

# Set the environment variable
ENV GOPATH=$HOME/coinbitly.com
ENV PATH=$HOME/coinbitly.com:$PATH
ENV PORT=35259
ENV PORT2=10443
ENV PORT3=35261
ENV PORT4=35260
ENV HOSTEMAIL=mail.privateemail.com
ENV HOSTSITE=https://coinbitly.com
ENV HOSTEXCH1=https://api.hitbtc.com
ENV EMAILPWD=Chid!234

# Expose the port using the environment variable PORT4
EXPOSE $PORT4

# Command to start your Go application
CMD ["./conbitly"]
