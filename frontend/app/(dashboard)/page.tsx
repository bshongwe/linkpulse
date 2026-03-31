'use client';

import { useState, useEffect } from 'react';
import { Plus, Link as LinkIcon, LogOut, BarChart3 } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { getUser, removeAuthToken } from '@/lib/auth';
import LiveCounter from '@/components/LiveCounter';
import AnalyticsChart from '@/components/AnalyticsChart';

export default function Dashboard() {
  const router = useRouter();
  const [user, setUser] = useState<any>(null);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    const currentUser = getUser();
    if (currentUser) {
      setUser(currentUser);
    } else {
      router.push('/login');
    }
  }, [router]);

  const handleLogout = () => {
    removeAuthToken();
    router.push('/login');
  };

  if (!mounted || !user) {
    return (
      <div className="min-h-screen bg-zinc-950 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-emerald-500"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-zinc-950 text-white">
      {/* Navigation */}
      <nav className="border-b border-zinc-800 bg-zinc-900/50 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-emerald-600 rounded-lg flex items-center justify-center">
                <LinkIcon className="w-6 h-6 text-white" />
              </div>
              <h1 className="text-xl font-bold">LinkPulse</h1>
            </div>

            <div className="flex items-center gap-6">
              <Link href="/links" className="text-zinc-400 hover:text-white transition-colors">
                My Links
              </Link>
              <Link href="/analytics" className="text-zinc-400 hover:text-white transition-colors">
                Analytics
              </Link>
              <span className="text-sm text-zinc-500">{user?.email}</span>
              <button
                onClick={handleLogout}
                className="flex items-center gap-2 bg-zinc-800 hover:bg-zinc-700 px-4 py-2 rounded-lg transition-colors"
              >
                <LogOut className="w-4 h-4" />
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Hero Section */}
        <div className="mb-12">
          <div className="flex justify-between items-start mb-8">
            <div>
              <h2 className="text-4xl font-bold mb-2">Welcome back, {user?.name || user?.email.split('@')[0]}</h2>
              <p className="text-zinc-400">Create and manage your short links with real-time analytics</p>
            </div>
            <Link
              href="/links/new"
              className="flex items-center gap-2 bg-emerald-600 hover:bg-emerald-700 px-6 py-3 rounded-lg font-semibold transition-colors"
            >
              <Plus className="w-5 h-5" />
              New Short Link
            </Link>
          </div>
        </div>

        {/* Quick Stats Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-12">
          {/* Live Counter */}
          <div className="lg:col-span-2">
            <LiveCounter />
          </div>

          {/* Stats Card */}
          <div className="bg-zinc-900 rounded-2xl p-6 border border-zinc-800">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 bg-emerald-600/20 rounded-lg flex items-center justify-center">
                <BarChart3 className="w-6 h-6 text-emerald-500" />
              </div>
              <h3 className="text-lg font-semibold">Quick Stats</h3>
            </div>
            <AnalyticsChart />
          </div>
        </div>

        {/* Recent Links Section */}
        <div className="bg-zinc-900 rounded-2xl p-8 border border-zinc-800">
          <h3 className="text-xl font-semibold mb-6">Recent Links</h3>
          <p className="text-zinc-400 text-center py-8">
            No links yet.{' '}
            <Link href="/links/new" className="text-emerald-500 hover:text-emerald-400">
              Create your first short link
            </Link>
          </p>
        </div>
      </main>
    </div>
  );
}
