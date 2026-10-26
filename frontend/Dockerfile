# Build frontend lalu sajikan sebagai file statis dengan Nginx non-root (bukan vite dev/preview).
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
# API satu domain dengan frontend (Nginx mem-proxy /api ke backend)
ENV VITE_API_URL=/api
RUN npm run build

FROM nginxinc/nginx-unprivileged:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
