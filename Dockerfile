# Start from the official Go image.
FROM golang:1.24-alpine

# Set the working directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies.
COPY go.mod go.sum ./

# Download dependencies.
RUN go mod download

# Copy the source code.
COPY . .

# Build the Go app.
# The -o flag sets the output file name.
RUN go build -o /noble-api ./

# Expose port 3000 to the outside world.
EXPOSE 3000

# Command to run the executable.
CMD [ "/noble-api" ]
