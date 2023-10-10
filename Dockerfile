# Use the official Golang image as the base image
FROM golang:1.14

# Set the working directory
WORKDIR /go/src/app

COPY go.mod go.sum ./

# Download dependencies using Go modules
RUN go get -d ./...

COPY . .

# Build the Call API tools and place them in the /go/bin directory
RUN GOBIN=/usr/bin make install

# Expose the WebSocket port
EXPOSE 5059

# Default is 'call-api' when the container starts
# to run the others, simply override the CMD
# docker run call-api-client
# docker run call-cmd
CMD ["call-api"]
