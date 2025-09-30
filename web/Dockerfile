# Базовый nginx образ
FROM nginx:alpine

# Копируем конфиг nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Копируем только фронт: html, css, js
COPY . /usr/share/nginx/html

# Порт 80
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]