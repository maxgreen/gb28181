FROM alpine:latest

ENV TZ=Asia/Shanghai

RUN apk --no-cache add ca-certificates \
	tzdata

WORKDIR /app

ADD ./build/linux_amd64/bin ./
ADD ./configs/config.toml /app/configs/config.toml
ADD ./www /app/www

LABEL Name=gowvp Version=0.0.1

EXPOSE 15123

CMD [ "./bin" ]