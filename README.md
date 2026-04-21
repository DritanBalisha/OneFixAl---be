🚀 OneFixAL – Backend

This repository contains the backend API for the OneFixAL platform.

📌 Project Overview

OneFixAL is a platform that connects clients with professional technicians.
Users can book services, manage appointments, and receive real-time notifications.

👉 The project is split into two repositories:

Backend (this repo) – Go (Gin, GORM, PostgreSQL)
Frontend – React + TypeScript (Vercel)
   🔗 https://github.com/DritanBalisha/OneFixAl---fe

🚀 Tech Stack
Backend: Go (Gin, GORM)
Frontend: React + TypeScript + TailwindCSS
Database: PostgreSQL (Railway)
Deployment:
      Frontend: Vercel
      Backend: Railway

📦 Backend Structure 

```
be/
 ├── cmd/
 |     └── server/
 │           └── main.go           # Entry point
 ├── internal/
 │    ├── api/              # Handlers (signup, login, bookings)
 │    ├── db/               # Database connection (Postgres)
 │    ├── middleware/       # Auth middleware (JWT)
 │    ├── models/           # DB models (User, Booking, etc.)
 │    ├── router/           # Route definitions
 │    └── websocket/        # Real-time notifications
 ├── go.mod
 └── go.sum
```
