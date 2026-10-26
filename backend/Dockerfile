# Build multi-stage: hasil akhir hanya berisi binary, berjalan sebagai user non-root.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# timetzdata: data zona waktu ikut di binary, supaya TZ=Asia/Jakarta berlaku di image distroless
RUN CGO_ENABLED=0 go build -tags timetzdata -trimpath -ldflags="-s -w" -o /out/api .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
USER nonroot:nonroot
ENV APP_ENV=production
EXPOSE 8080
ENTRYPOINT ["/api"]
