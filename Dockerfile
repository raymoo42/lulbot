FROM golang:1.7

LABEL version=1.0
LABEL description="Line Bot"
LABEL maintainer="Stefan Turzer<turzer.stefan@gmail.com"

EXPOSE 3000

WORKDIR /app
RUN ["mkdir", "-p", "/etc/lulbot/"]

ADD lulbot .

CMD ["lulbot", "-conf", "/etc/lulbot/config.toml"]