# Build stage with static linking
FROM rust:slim AS builder
WORKDIR /usr/src/app
COPY Cargo.toml .
COPY src/ ./src/
RUN rustup target add x86_64-unknown-linux-musl
RUN cargo build --release --target x86_64-unknown-linux-musl
RUN ls -la /usr/src/app/target/x86_64-unknown-linux-musl/release/

# Minimal runtime image
FROM alpine:latest
COPY --from=builder /usr/src/app/target/x86_64-unknown-linux-musl/release/cs3211_assignment_3 /usr/local/bin/app
RUN chmod +x /usr/local/bin/app
CMD ["app"]