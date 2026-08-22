# 📚 Cloudflare R2 Manga Host & Reader System

Hệ thống Manga Hosting hoàn chỉnh, hiệu năng cao, tối ưu RAM và chi phí:
- **Backend**: Golang 1.24+ (Chi Router, S3 AWS SDK Go cho Cloudflare R2, MongoDB Driver). Siêu nhẹ, chiếm cực ít RAM (< 15MB), hỗ trợ stream giải nén `.zip` / `.cbz` trực tiếp lên R2.
- **Frontend**: SvelteKit 2 + Svelte 5 + TailwindCSS + Lucide Icons. Giao diện đọc truyện Webtoon & Lật trang hiện đại, Admin Dashboard bảo mật, sẵn sàng deploy 1-click lên **Vercel**.
- **Storage**: Cloudflare R2 (Lưu trữ ảnh bìa & hàng triệu trang manga với 0đ phí egress).
- **Paperback iOS Extension**: Source extension TypeScript chuẩn cho app **Paperback iOS**.

---

## 📁 Cấu Trúc Dự Án

```text
r2apimangahost/
├── backend/                  # REST API Server viết bằng Go (Golang)
│   ├── cmd/server/main.go    # Entrypoint máy chủ Go
│   ├── internal/             # Config, Auth, DB (Mongo), R2 Storage, Handlers
│   ├── Dockerfile            # Multi-stage build nhẹ (< 25MB)
│   └── .env.example          # Mẫu cấu hình môi trường
│
├── frontend/                 # Giao diện Web SvelteKit (Deploy Vercel)
│   ├── src/
│   │   ├── routes/           # Trang chủ, Trang đọc truyện, Admin Dashboard
│   │   └── lib/              # API Client, Stores, Components
│   ├── package.json
│   └── .env.example          # VITE_API_URL=https://api.yourdomain.com/api/v1
│
└── paperback-extension/      # Extension TypeScript cho app Paperback iOS
    ├── src/MangaHost/        # Source class, Parser kết nối API
    ├── package.json
    └── README.md
```

---

## 🚀 Hướng Dẫn Cài Đặt & Chạy Local

### 1. Chuẩn Bị Môi Trường

1. **MongoDB**: Cài MongoDB local hoặc tạo tài khoản miễn phí trên [MongoDB Atlas](https://www.mongodb.com/atlas).
2. **Cloudflare R2**:
   - Vào Cloudflare Dashboard > **R2** > Tạo một Bucket (ví dụ: `mangahost`).
   - Tạo **API Tokens** (Quyền *Object Read & Write*).
   - Lấy: `Account ID`, `Access Key ID`, `Secret Access Key`.
   - Bật Public Domain cho Bucket (hoặc gắn Custom Domain).

---

### 2. Chạy Backend (Go)

```bash
cd backend
cp .env.example .env
# Điền thông tin MongoDB và Cloudflare R2 vào file .env

# Chạy server Go:
go run ./cmd/server
```
👉 Server sẽ lắng nghe tại `http://localhost:8080`. Tài khoản Admin mặc định được tự động khởi tạo theo biến `ADMIN_USERNAME` và `ADMIN_PASSWORD` trong file `.env`.

---

### 3. Chạy Frontend (SvelteKit)

```bash
cd frontend
npm install
npm run dev
```
👉 Mở trình duyệt tại `http://localhost:3000`.

---

## 🌐 Hướng Dẫn Deploy Production

### 1. Deploy Backend (Go)
Bạn có thể deploy backend Go lên bất kỳ nền tảng nào (Render, Railway, Fly.io, hoặc VPS với Docker):
- **Với Docker**:
  ```bash
  cd backend
  docker build -t mangahost-backend .
  docker run -d -p 8080:8080 --env-file .env mangahost-backend
  ```
- **Với Render / Railway / Fly.io**: Kết nối Github repo, chọn thư mục `backend`, nạp các biến môi trường từ `.env.example`.

### 2. Deploy Frontend lên Vercel
1. Đẩy code lên GitHub.
2. Đăng nhập [Vercel](https://vercel.com) > **Add New Project** > Chọn repository này.
3. Ở phần **Root Directory**, chọn thư mục `frontend`.
4. Thêm biến môi trường:
   - `VITE_API_URL`: Điền URL backend Go của bạn (ví dụ: `https://api.yourdomain.com/api/v1`).
5. Bấm **Deploy**.

---

## 📱 Cài Đặt Paperback iOS Extension

Xem chi tiết trong [paperback-extension/README.md](file:///d:/Code/r2apimangahost/paperback-extension/README.md).
1. Thay đổi domain API trong `paperback-extension/src/MangaHost/MangaHost.ts`.
2. Build bundle bằng `paperback bundle`.
3. Thêm link repository vào app **Paperback** trên iPhone/iPad để đọc truyện!
