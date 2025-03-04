FROM --platform=linux/amd64 ubuntu:20.04

WORKDIR /app

COPY cmd/accrual/accrual_linux_amd64 /app/accrual

RUN chmod +x /app/accrual

EXPOSE 8080

CMD ["./accrual"]