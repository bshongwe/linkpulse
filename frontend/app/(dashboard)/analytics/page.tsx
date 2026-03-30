'use client';

import { useState, useEffect } from 'react';
import { Link as LinkIcon, LogOut, ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { getUser, removeAuthToken } from '@/lib/auth';
import AnalyticsChart from '@/components/AnalyticsChart';

export default function AnalyticsPage() {
  const router = useRouter();
  const [user, setUser] = useState<any>(null);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    const currentUser = getUser();
    if (!currentUser) {
      router.push('/login');
    } else {
      setUser(currentUser);
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
            <Link href="/" className="flex items-center gap-3 hover:opacity-80 transition-opacity">
              <div className="w-10 h-10 bg-emerald-600 rounded-lg flex items-center justify-center">
                <LinkIcon className="w-6 h-6 text-white" />
              </div>
              <h1 className="text-xl font-bold">LinkPulse</h1>
            </Link>

            <div className="flex items-center gap-6">
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
        {/* Header */}
        <div className="flex items-center gap-4 mb-8">
          <Link
            href="/"
            className="flex items-center gap-2 text-zinc-400 hover:text-white transition-colors"
          >
            <ArrowLeft className="w-5 h-5" />
            Back to Dashboard
          </Link>
        </div>

        <div className="mb-8">
          <h2 className="text-3xl font-bold mb-2">Analytics</h2>
          <p className="text-zinc-400">View detailed statistics and insights about your links</p>
        </div>

        {/* Analytics Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* Total Clicks */}
          <div className="bg-zinc-900 rounded-2xl p-6 border border-zinc-800">
            <h3 className="text-sm font-semibold text-zinc-400 mb-2">Total Clicks</h3>
            <p className="text-4xl font-bold">0</p>
            <p className="text-xs text-zinc-500 mt-2">All time</p>
          </div>

          {/* Clicks This Month */}
          <div className="bg-zinc-900 rounded-2xl p-6 border border-zinc-800">
            <h3 className="text-sm font-semibold text-zinc-400 mb-2">This Month</h3>
            <p className="text-4xl font-bold">0</p>
            <p className="text-xs text-zinc-500 mt-2">Last 30 days</p>
          </div>

          {/* Clicks This Week */}
          <div className="bg-zinc-900 rounded-2xl p-6 border border-zinc-800">
            <h3 className="text-sm font-semibold text-zinc-400 mb-2">This Week</h3>
            <p className="text-4xl font-bold">0</p>
            <p className="text-xs text-zinc-500 mt-2">Last 7 days</p>
          </div>

          {/* Clicks Today */}
          <div className="bg-zinc-900 rounded-2xl p-6 border border-zinc-800">
            <h3 className="text-sm font-semibold text-zinc-400 mb-2">Today</h3>
            <p className="text-4xl font-bold">0</p>
            <p className="text-xs text-zinc-500 mt-2">Last 24 hours</p>
          </div>
        </div>

        {/* Chart */}
        <div className="bg-zinc-900 rounded-2xl p-8 border border-zinc-800">
          <h3 className="text-xl font-semibold mb-6">Weekly Trends</h3>
          <AnalyticsChart />
        </div>

        {/* Top Links */}
        <div className="mt-8 bg-zinc-900 rounded-2xl p-8 border border-zinc-800">
          <h3 className="text-xl font-semibold mb-6">Top Performing Links</h3>
          <p className="text-zinc-400 text-center py-8">
            No data available yet. Create and share some links to see analytics.
          </p>
        </div>
      </main>
    </div>
  );
}
