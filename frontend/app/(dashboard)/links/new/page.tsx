'use client';

import { useRouter } from 'next/navigation';
import { useState, useEffect } from 'react';
import { Link as LinkIcon, LogOut, ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { getUser, removeAuthToken } from '@/lib/auth';
import { createShortLink, getErrorMessage } from '@/lib/api';

export default function CreateLinkPage() {
  const router = useRouter();
  const [user, setUser] = useState<any>(null);
  const [mounted, setMounted] = useState(false);
  const [originalUrl, setOriginalUrl] = useState('');
  const [customAlias, setCustomAlias] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      if (!user) {
        setError('Not authenticated. Please log in again.');
        router.push('/login');
        return;
      }

      // Use BFF gateway instead of calling shortener directly
      // BFF will extract workspace_id from JWT automatically
      await createShortLink({
        original_url: originalUrl,
        custom_alias: customAlias || undefined,
      });

      // Success - redirect to links page
      router.push('/links');
    } catch (err: any) {
      console.error('Create link error:', err);
      const errorMsg = getErrorMessage(err);
      setError(errorMsg || 'Failed to create link');
    } finally {
      setLoading(false);
    }
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

      <main className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <Link
          href="/links"
          className="flex items-center gap-2 text-zinc-400 hover:text-white transition-colors mb-8"
        >
          <ArrowLeft className="w-5 h-5" />
          Back to Links
        </Link>

        <div className="bg-zinc-900 rounded-2xl border border-zinc-800 p-8">
          <h1 className="text-3xl font-bold mb-2">Create New Short Link</h1>
          <p className="text-zinc-400 mb-8">Shorten a long URL and start tracking clicks</p>

          {error && (
            <div className="bg-red-900/20 border border-red-800 rounded-lg p-4 mb-6 text-red-400 text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label htmlFor="original-url" className="block text-sm font-medium text-zinc-200 mb-2">
                Destination URL *
              </label>
              <input
                id="original-url"
                type="url"
                value={originalUrl}
                onChange={(e) => setOriginalUrl(e.target.value)}
                placeholder="https://example.com/very-long-url-here"
                required
                className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-4 py-3 text-white placeholder-zinc-500 focus:border-emerald-500 focus:outline-none transition-colors"
              />
              <p className="text-xs text-zinc-500 mt-2">The URL you want to shorten</p>
            </div>

            <div>
              <label htmlFor="custom-alias" className="block text-sm font-medium text-zinc-200 mb-2">
                Custom Alias (optional)
              </label>
              <input
                id="custom-alias"
                type="text"
                value={customAlias}
                onChange={(e) => setCustomAlias(e.target.value)}
                placeholder="my-awesome-link"
                className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-4 py-3 text-white placeholder-zinc-500 focus:border-emerald-500 focus:outline-none transition-colors"
              />
              <p className="text-xs text-zinc-500 mt-2">Leave blank to auto-generate (e.g., abc123)</p>
            </div>

            <div className="flex gap-4 pt-4">
              <Link
                href="/links"
                className="flex-1 bg-zinc-800 hover:bg-zinc-700 py-3 rounded-lg font-medium text-center transition-colors"
              >
                Cancel
              </Link>
              <button
                type="submit"
                disabled={loading}
                className="flex-1 bg-emerald-600 hover:bg-emerald-700 disabled:bg-emerald-600/50 py-3 rounded-lg font-medium transition-colors"
              >
                {loading ? 'Creating...' : 'Create Link'}
              </button>
            </div>
          </form>
        </div>

        <div className="mt-8 bg-emerald-900/20 border border-emerald-800 rounded-lg p-6">
          <h3 className="font-semibold text-emerald-400 mb-2">💡 Tips</h3>
          <ul className="text-sm text-zinc-300 space-y-2">
            <li>• Use a descriptive custom alias to make links memorable</li>
            <li>• Aliases can only contain letters, numbers, and hyphens</li>
            <li>• Track all clicks and analytics for your links</li>
            <li>• Share your short links anywhere on the web</li>
          </ul>
        </div>
      </main>
    </div>
  );
}
