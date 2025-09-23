# ステージ 1: React アプリのビルド
FROM node:20-alpine AS build

WORKDIR /app
COPY ./package.json ./package-lock.json ./
RUN npm install
COPY ./ /app
RUN npm run build 

# ステージ 2: Nginx に React アプリを提供
FROM nginx:alpine

# React アプリのビルド結果を Nginx のコンテンツディレクトリにコピー
COPY --from=build /app/dist /usr/share/nginx/html 

# コンテナ起動時に Nginx を起動
CMD ["nginx", "-g", "daemon off;"]
