FROM rust:slim-buster AS builder

WORKDIR /prod
COPY Cargo.lock .
COPY Cargo.toml .
RUN mkdir .cargo src && touch src/main.rs
# This is the trick to speed up the building process.
# RUN cargo vendor > .cargo/config

COPY . .
RUN cargo build --release

# Use any runner as you want
# But beware that some images have old glibc which makes rust unhappy
FROM fedora:34 AS runner
COPY --from=builder /prod/target/release/budget-chat /bin
CMD /bin/budget-chat
