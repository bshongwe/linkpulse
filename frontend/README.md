# LinkPulse Frontend

Modern Next.js 15 dashboard for LinkPulse URL shortener with real-time analytics.

## 🎨 Features

- **Authentication**: JWT-based login/register flow
- **Dashboard**: Beautiful, responsive UI with Tailwind CSS
- **Live Counter**: Real-time WebSocket connection showing click counts
- **Analytics**: Interactive charts with Recharts
- **Responsive Design**: Mobile-first approach with full mobile support
- **Type Safety**: Full TypeScript support with proper typing
- **Dark Mode**: Modern dark theme optimized for developer comfort

## 🚀 Getting Started

### Prerequisites
- Node.js 20+
- npm or yarn

### Installation

```bash
cd frontend
npm install
```

### Development

```bash
npm run dev
```

The frontend will be available at `http://localhost:3000`

### Production Build

```bash
npm run build
npm start
```

## 📁 Project Structure

```
frontend/
├── app/                           # Next.js App Router
│   ├── (auth)/                   # Authentication routes (grouped)
│   │   ├── login/page.tsx        # Login page
│   │   └── register/page.tsx     # Register page
│   ├── (dashboard)/              # Dashboard routes (grouped)
│   │   ├── page.tsx              # Main dashboard
│   │   ├── links/page.tsx        # Links list
│   │   └── analytics/page.tsx    # Analytics details
│   ├── layout.tsx                # Root layout with metadata
│   └── globals.css               # Global Tailwind styles
├── components/
│   ├── ui/                       # Reusable UI components
│   ├── LiveCounter.tsx           # WebSocket-powered click counter
│   ├── AnalyticsChart.tsx        # Recharts analytics visualization
│   └── Navbar.tsx                # Navigation component
├── lib/
│   ├── api.ts                    # Typed API client (Axios)
│   └── auth.ts                   # JWT token management
├── types/
│   └── index.ts                  # TypeScript interfaces
├── public/                       # Static assets
├── package.json
├── tsconfig.json
├── tailwind.config.ts
├── next.config.mjs
└── Dockerfile                    # Production container image
```

## 🔐 Authentication Flow

1. User navigates to `/login`
2. Enters email and password
3. Frontend stores JWT token in localStorage
4. Automatic redirect to dashboard
5. Protected routes check for valid token
6. API requests automatically include Bearer token

## 📡 API Integration

All API calls go through `lib/api.ts` with:
- Automatic JWT token injection
- Error handling and retry logic
- Type-safe request/response

Example:
```ts
import { createShortLink } from '@/lib/api';

const link = await createShortLink({
  original_url: 'https://example.com',
  custom_alias: 'mylink',
});
```

## 🔌 WebSocket Connection

The LiveCounter component establishes a WebSocket connection to:
```
ws://localhost:8083/ws/live/{shortCode}
```

Events:
- `{action: 'click'}` - Increment click counter

## 🎨 Styling

- **Tailwind CSS**: Utility-first CSS framework
- **Dark Theme**: Zinc-950 base with Emerald-600 accents
- **Responsive**: Mobile-first design approach
- **Custom Utilities**: Defined in `globals.css`

## 🔄 Environment Variables

```env
NEXT_PUBLIC_API_BASE=http://localhost:8082
```

## 📦 Dependencies

- **next**: 15.0.0 - React framework
- **react**: 18.3.0 - UI library
- **tailwindcss**: 3.4.0 - CSS framework
- **recharts**: 2.12.0 - Chart library
- **axios**: 1.7.0 - HTTP client
- **lucide-react**: 0.441.0 - Icon library
- **js-cookie**: 3.0.5 - Cookie management

## 🐳 Docker

Build and run in Docker:

```bash
docker build -t linkpulse-frontend .
docker run -p 3000:3000 linkpulse-frontend
```

## 📝 License

MIT - LinkPulse Project
