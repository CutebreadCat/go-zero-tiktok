# go-zero-tiktok 前端（Vue 3）

配套后端：仓库根目录的 go-zero 网关（默认 `:8888`）。配套教学文档：[docs/frontend/](../docs/frontend/README.md)。

## 技术栈

Vue 3（组合式 API + `<script setup>`）· Vue Router · Pinia · Axios · Vite —— 无 UI 组件库，设计令牌手写（深色短视频主题）。

## 运行

```bash
cd frontend
npm install
npm run dev        # http://localhost:5173
```

后端启动（仓库根目录）：`make infra-up && make migrate-up && make build-local`，然后按 Makefile 逐个 `run-*-local`。

## 切换后端地址（配置化核心）

三种方式，优先级从高到低：

1. **运行时**：页面右上角「连接设置」→ 输入如 `http://192.168.1.5:8888` → 可先「测试连接」再保存。存 localStorage，立即生效，无需刷新。
2. **环境变量**：`VITE_API_BASE`（构建期默认值，`.env.development` / `.env.production` 默认 `/api`）。
3. **同源模式**：留空或根相对路径（`/api`）= 请求发到前端所在域名，开发由 `vite.config.js` 的 `/api` 代理、生产由 Nginx 反代转发到网关（见 `deploy/nginx/tiktok-web.conf`）。

> Cookie 鉴权说明：登录态是 HttpOnly Cookie（SameSite=Lax）。前端与后端**同站**（同 IP 或经代理）时 Cookie 自动携带；跨站部署时需后端调整 CORS + SameSite 策略，详见教学文档 03。

## 目录结构

```
src/
├── config/appConfig.js   # ★ 运行时 API 地址：三层优先级，切换即生效
├── api/http.js           # ★ axios 单例：动态 Base + Cookie + 信封解析 + 401 刷新重放
├── api/{user,video,interaction,communication,tracking}.js  # 五大业务域各一文件
├── stores/{auth,messages}.js   # Pinia：会话探测 / 未读轮询
├── router/index.js       # 路由 + 登录守卫
├── utils/tracker.js      # 行为埋点批量上报（曝光/播放/完播）
├── components/           # SideNav / TopBar / VideoCard / CommentPanel / ApiSettingsModal
└── views/                # Feed / Search / Publish / Profile / Messages / Login
```

## 与后端调用流程的对应

| 页面 | 后端链路 |
|---|---|
| Feed（推荐/关注/热门） | `GET /feed-items?scene=...` → 网关聚合 video+interaction+user |
| 点赞/收藏/评论 | `interaction.rpc` Redis+Kafka 异步链路 |
| 关注 | `communication.rpc` + 发布扇出 inbox |
| 消息中心 | `GET /messages` 游标分页、未读、批量已读 |
| 发布 | `POST /videos` multipart → OSS → fanout |
| 埋点 | `POST /tracking-events` 批量（曝光≥50%可见 800ms / play / complete） |
| QoS | 卡顿/错误时 `POST /playback-qos-reports`（tracker 扩展点） |
