# Start from golang v1.12.6 base image
FROM golang:1.12.6

# Add Maintainer Info
LABEL maintainer="Anders Kvist <anderskvist@gmail.com>"

# Build dependencies
RUN apt-get update && apt-get install libmagic-dev libmagickwand-dev -y

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/anderskvist/GoPDF2PNG

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Download all the dependencies
# https://stackoverflow.com/questions/28031603/what-do-three-dots-mean-in-go-command-line-invocations
RUN go get -d -v ./...


# Install the package set date and git revision as version
RUN go install -ldflags "-X github.com/anderskvist/GoHelpers/version.Version=`date -u '+%Y%m%d-%H%M%S'`-`git rev-parse --short HEAD`" -v ./...

# Run the executable
CMD ["/go/bin/GoPDF2PNG","config.ini"]
