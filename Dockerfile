# Build stage. The migrations are embedded in the binary with go:embed, so the
# runtime stage needs no .sql files and no migration tool.
FROM golang:1.26-alpine AS build

WORKDIR /src

# Dependencies first, so a change to the source does not re-download them.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 for a static binary, which is what lets the runtime stage be
# scratch. -s -w drop the symbol table and DWARF, neither of which is read in
# production, and the service does not print a version.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /bin/api ./cmd

# Runtime stage: the binary, a passwd entry for the non-root user, the CA
# certificates for TLS to PostgreSQL, and nothing else. No shell, no package
# manager, and no Go toolchain.
FROM scratch AS runtime

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /bin/api /bin/api

# nobody, from the base image's /etc/passwd. The service binds ADDRESS, which
# should be a port above 1024, and writes only to standard output.
USER nobody

# Documentation only: the port actually served comes from ADDRESS.
EXPOSE 8080

ENTRYPOINT ["/bin/api"]
