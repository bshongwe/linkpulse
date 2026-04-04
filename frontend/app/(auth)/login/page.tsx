'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { LinkIcon } from 'lucide-react';
import { setAuthToken, setUser, decodeJWT } from '@/lib/auth';

const AUTH_BASE = process.env.NEXT_PUBLIC_AUTH_BASE || 'http://localhost:8081';

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      // Call real auth service
      const response = await fetch(`${AUTH_BASE}/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      const data = await response.json();

      if (!response.ok) {
        setError(data.error || 'Login failed. Please check your credentials.');
        return;
      }

      // Store token
      setAuthToken(data.access_token);

      // Extract user info from JWT token
      const decoded = decodeJWT(data.access_token);
      if (decoded) {
        const user = {
          id: decoded.user_id || decoded.sub || email.split('@')[0],
          email: decoded.email || email,
          name: decoded.name || email.split('@')[0],
          workspace_id: decoded.workspace_id || 'default',
        };
        setUser(user);
      } else {
        // Fallback user object if JWT decode fails
        setUser({
          id: email.split('@')[0],
          email,
          name: email.split('@')[0],
          workspace_id: 'default',
        });
      }

      router.push('/');
    } catch (err) {
      console.error('Login error:', err);
      setError('Failed to connect to auth service. Is the backend running?');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-zinc-950 via-zinc-900 to-zinc-950 flex flex-col items-center justify-center px-4">
      {/* Background decoration */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute top-0 right-0 w-96 h-96 bg-emerald-600/10 rounded-full blur-3xl"></div>
        <div className="absolute bottom-0 left-0 w-96 h-96 bg-emerald-600/5 rounded-full blur-3xl"></div>
      </div>

      <div className="relative z-10 w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-12 h-12 bg-emerald-600 rounded-xl mb-4">
            <LinkIcon className="w-7 h-7 text-white" />
          </div>
          <h1 className="text-3xl font-bold">LinkPulse</h1>
          <p className="text-zinc-400 mt-2">Smart URL shortener with real-time analytics</p>
        </div>

        {/* Login Form */}
        <div className="bg-zinc-900 rounded-2xl border border-zinc-800 p-8 shadow-xl">
          <h2 className="text-2xl font-bold mb-6">Welcome back</h2>

          {error && (
            <div className="bg-red-900/20 border border-red-800 rounded-lg p-4 mb-6 text-red-400 text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-zinc-200 mb-2">Email</label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                required
                className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-4 py-2 text-white placeholder-zinc-500 focus:border-emerald-500 focus:outline-none transition-colors"
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-zinc-200 mb-2">Password</label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
                className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-4 py-2 text-white placeholder-zinc-500 focus:border-emerald-500 focus:outline-none transition-colors"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-emerald-600 hover:bg-emerald-700 disabled:bg-emerald-600/50 text-white font-semibold py-2 px-4 rounded-lg transition-colors mt-6"
            >
              {loading ? 'Signing in...' : 'Sign in'}
            </button>
          </form>

          <p className="text-center text-zinc-400 text-sm mt-6">
            Don't have an account?{' '}
            <Link href="/register" className="text-emerald-500 hover:text-emerald-400">
              Sign up
            </Link>
          </p>
        </div>

        {/* Login info */}
        <div className="mt-6 p-4 bg-emerald-900/20 border border-emerald-800 rounded-lg">
          <p className="text-emerald-400 text-sm font-semibold mb-2">Login Info</p>
          <p className="text-zinc-400 text-xs">
            Create an account first via the register page, or contact admin for test credentials.
          </p>
        </div>
      </div>
    </div>
  );
}
