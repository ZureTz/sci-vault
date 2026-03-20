FROM node:22-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm cache clean --force && rm -rf package-lock.json && npm install

COPY . .

RUN npm run build

FROM nginx:alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/build /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]