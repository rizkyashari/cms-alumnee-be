FROM node:20-alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY package.json yarn.lock ./

RUN yarn install

COPY . .

EXPOSE 1337

CMD ["yarn", "develop"]
