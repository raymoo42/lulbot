FROM golang:1.7

LABEL version=1.0
LABEL description="Line Bot"
LABEL maintainer="Stefan Turzer<turzer.stefan@gmail.com"

EXPOSE 3000

WORKDIR /app

ADD lulbot .
COPY web/ /app/web/

CMD ["./lulbot"]