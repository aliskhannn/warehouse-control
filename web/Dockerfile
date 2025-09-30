FROM node:20-alpine AS build
WORKDIR /app
COPY . .
# No build step needed for pure HTML/JS/CSS
FROM nginx:alpine
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=build /app /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
