FROM alpine:3.7

# Install cert infrastructure.
RUN apk update \
&& apk add \
     build-base \
     ca-certificates \
     openssl \
&& rm -rf /var/cache/apk/*

# Create working directory.
RUN mkdir app
WORKDIR /app
COPY ./issues /app
COPY ./templates /app/templates

# Run server.
EXPOSE 8080
CMD ./issues
