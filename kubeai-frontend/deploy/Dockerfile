# 第一阶段：构建
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# 第二阶段：运行
FROM nginx:1.27-alpine
# 删除默认配置
RUN rm /etc/nginx/conf.d/default.conf
# 复制自定义 nginx 配置
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf
# 复制构建好的静态文件
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]