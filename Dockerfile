ARG GO_VERSION="latest"
ARG APP_NAME="lorekeeper"

# ---- Build Stage ----
FROM golang:${GO_VERSION} AS builder

# Copy the rest of the source code.
COPY /cmd/${APP_NAME} /cmd/${APP_NAME}

# Build the Go app.
RUN go build -o /cmd/${APP_NAME} /${APP_NAME}

# Define the command to run the app.
ENTRYPOINT ["/${APP_NAME}"]

# Add a default command.
CMD ["--help"]