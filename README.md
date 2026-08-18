# GCET Faculty Portal - Web Version

A complete production-ready web application replacement for the GCET Faculty Portal Flutter app.

## Project Structure

- `frontend/`: Next.js 16 + React 19 web application (App Router)
  - Features: Zustand state management, TanStack Query, shadcn/ui components, Framer Motion animations.
  - Tailwind CSS configured to exactly match the original Flutter app's organic theme.
- `backend/`: Go (Golang) + Chi Router + Supabase API
  - Features: Headless HTTP web scraping session proxy, robust JWT authentication, concurrent scraping sessions, robust error handling.

## Deployment Setup (Vercel)

The frontend is ready for zero-configuration deployment to **Vercel**.

1. Connect your GitHub repository to Vercel.
2. Select the `web/frontend` directory as the Root Directory in Vercel settings.
3. Add the required Environment Variable:
   - `NEXT_PUBLIC_API_URL` (Points to the deployed Go backend URL, e.g. `https://gcet-backend.example.com/api/v1`)

## Deployment Setup (Go Backend)

The backend can be deployed via Docker or directly on platforms like Render, Fly.io, or AWS.

1. Ensure the `web/backend` has environment variables configured:
   - `PORT`: (default 8080)
   - `GMS_BASE_URL`: https://gms.gcet.ac.in
   - `JWT_SECRET`: <your-secure-secret>
   - `SUPABASE_URL`: <supabase-project-url>
   - `SUPABASE_KEY`: <supabase-service-role-key>

## Running Locally

1. **Backend:**
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

2. **Frontend:**
   ```bash
   cd frontend
   npm run dev
   ```

## Design System
- **Colors**: Primary (`#D35D27`), Background (`#fbf8f1`), Surface (`#ffffff`), Text (`#1a1b41`)
- **Typography**: Inter (Body), Lora (Headings), Lexend (Accents)
