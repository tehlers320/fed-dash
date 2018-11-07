FROM alpine

WORKDIR /app

COPY ./main /app

CMD ["/app/main"]
