'use client';

import { useState, useEffect } from 'react';
import { Plus, Link as LinkIcon, LogOut, Trash2, Copy, ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { getUser, removeAuthToken } from '@/lib/auth';
import { listLinks, deleteShortLink } from '@/lib/api';

export default function LinksPage() {
  const router = useRouter();
  const [user, setUser] = useState<any>(null);
  const [mounted, setMounted] = useState(false);
  const [links, setLinks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [deleting, setDeleting] = useState<string | null>(null);

  useEffect(() => {
    setMounted(true);
    const currentUser = getUser();
    if (!currentUser) {
      router.push('/login');
      return;
    }
    setUser(currentUser);
    fetchLinks(currentUser.workspace_id || 'default');
  }, [router]);

  const fetchLinks = async (workspaceId: string) => {
    try {
      setLoading(true);
      setError('');
      const result = await listLinks(workspaceId);
      setLinks(result || []);
    } catch (err: any) {
      console.error('Failed to fetch links:', err);
      const errorMsg = err.response?.data?.error || err.message || 'Failed to load links';
      setError(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    removeAuthToken();
    router.push('/login');
  };

  const handleDeleteLink = async (id: string) => {
    if (!confirm('Are you sure you want to delete this link?')) return;
    
    try {
      setDeleting(id);
      await deleteShortLink(id);
      setLinks(links.filter((link) => link.id !== id));
    } catch (err: any) {
      console.error('Failed to delete link:', err);
      const errorMsg = err.response?.data?.error || 'Failed to delete link';
      alert(`Error: ${errorMsg}`);
    } finally {
      setDeleting(null);
    }
  };

  const handleCopyLink = (shortCode: string) => {
    const url = `https://short.url/${shortCode}`;
    navigator.clipboard.writeText(url);
    alert('Link copied to clipboard!');
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
        <div className="flex justify-between items-start mb-8">
          <div>
            <div className="flex items-center gap-3 mb-4">
              <Link
                href="/"
                className="text-zinc-400 hover:text-white transition-colors"
              >
                <ArrowLeft className="w-5 h-5" />
              </Link>
              <h2 className="text-3xl font-bold">My Links</h2>
            </div>
            <p className="text-zinc-400">View and manage all your shortened links</p>
          </div>
          <Link
            href="/links/new"
            className="flex items-center gap-2 bg-emerald-600 hover:bg-emerald-700 px-6 py-3 rounded-lg font-semibold transition-colors"
          >
            <Plus className="w-5 h-5" />
            New Link
          </Link>
        </div>

        {/* Error Message */}
        {error && (
          <div className="bg-red-900/20 border border-red-800 rounded-lg p-4 mb-6 text-red-400 text-sm">
            {error}
          </div>
        )}

        {/* Loading State */}
        {loading && (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-emerald-500"></div>
          </div>
        )}

        {/* Empty State */}
        {!loading && links.length === 0 && (
          <div className="bg-zinc-900 rounded-2xl border border-zinc-800 p-12 text-center">
            <LinkIcon className="w-12 h-12 text-zinc-600 mx-auto mb-4" />
            <h3 className="text-xl font-semibold mb-2">No links yet</h3>
            <p className="text-zinc-400 mb-6">Create your first shortened link to get started</p>
            <Link
              href="/links/new"
              className="inline-flex items-center gap-2 bg-emerald-600 hover:bg-emerald-700 px-6 py-3 rounded-lg font-semibold transition-colors"
            >
              <Plus className="w-5 h-5" />
              Create Link
            </Link>
          </div>
        )}

        {/* Links Table */}
        {!loading && links.length > 0 && (
          <div className="bg-zinc-900 rounded-2xl border border-zinc-800 overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="border-b border-zinc-800 bg-zinc-800/50">
                  <tr>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-zinc-200">Original URL</th>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-zinc-200">Short Code</th>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-zinc-200">Clicks</th>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-zinc-200">Created</th>
                    <th className="px-6 py-4 text-right text-sm font-semibold text-zinc-200">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800">
                  {links.map((link) => (
                    <tr key={link.id} className="hover:bg-zinc-800/50 transition-colors">
                      <td className="px-6 py-4 text-sm text-zinc-300 truncate">
                        <a
                          href={link.original_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="hover:text-emerald-500 transition-colors"
                          title={link.original_url}
                        >
                          {link.original_url}
                        </a>
                      </td>
                      <td className="px-6 py-4 text-sm font-mono text-emerald-400">
                        {link.short_code}
                      </td>
                      <td className="px-6 py-4 text-sm text-zinc-300">
                        {link.click_count || 0}
                      </td>
                      <td className="px-6 py-4 text-sm text-zinc-400">
                        {new Date(link.created_at).toLocaleDateString()}
                      </td>
                      <td className="px-6 py-4 text-right space-x-2 flex justify-end">
                        <button
                          onClick={() => handleCopyLink(link.short_code)}
                          className="p-2 hover:bg-zinc-700 rounded-lg transition-colors text-zinc-400 hover:text-white"
                          title="Copy link"
                        >
                          <Copy className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteLink(link.id)}
                          disabled={deleting === link.id}
                          className="p-2 hover:bg-red-900/20 rounded-lg transition-colors text-zinc-400 hover:text-red-400 disabled:opacity-50"
                          title="Delete link"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
