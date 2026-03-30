'use client';

import { useState, useEffect } from 'react';
import { Plus, Link as LinkIcon, LogOut, Trash2, Copy } from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { getUser, removeAuthToken } from '@/lib/auth';
import CreateLinkModal from '@/components/CreateLinkModal';

export default function LinksPage() {
  const router = useRouter();
  const [user, setUser] = useState<any>(null);
  const [mounted, setMounted] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [links, setLinks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setMounted(true);
    const currentUser = getUser();
    if (!currentUser) {
      router.push('/login');
    } else {
      setUser(currentUser);
      // Simulate loading links
      setTimeout(() => {
        setLinks([]);
        setLoading(false);
      }, 500);
    }
  }, [router]);

  const handleLogout = () => {
    removeAuthToken();
    router.push('/login');
  };

  const handleDeleteLink = (id: string) => {
    setLinks(links.filter((link) => link.id !== id));
  };

  const handleCopyLink = (shortCode: string) => {
    const url = `${window.location.origin}/${shortCode}`;
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
        <div className="flex justify-between items-center mb-8">
          <div>
            <h2 className="text-3xl font-bold mb-2">My Links</h2>
            <p className="text-zinc-400">Manage and track all your shortened URLs</p>
          </div>
          <button
            onClick={() => setIsModalOpen(true)}
            className="flex items-center gap-2 bg-emerald-600 hover:bg-emerald-700 px-6 py-3 rounded-lg font-semibold transition-colors"
          >
            <Plus className="w-5 h-5" />
            New Short Link
          </button>
        </div>

        {/* Links Table */}
        <div className="bg-zinc-900 rounded-2xl border border-zinc-800 overflow-hidden">
          {loading ? (
            <div className="p-8 text-center text-zinc-400">
              <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-emerald-500 mx-auto mb-2"></div>
              Loading links...
            </div>
          ) : links.length === 0 ? (
            <div className="p-12 text-center">
              <LinkIcon className="w-16 h-16 text-zinc-700 mx-auto mb-4" />
              <h3 className="text-xl font-semibold mb-2">No links yet</h3>
              <p className="text-zinc-400 mb-6">Create your first short link to get started</p>
              <button
                onClick={() => setIsModalOpen(true)}
                className="inline-flex items-center gap-2 bg-emerald-600 hover:bg-emerald-700 px-6 py-3 rounded-lg font-semibold transition-colors"
              >
                <Plus className="w-5 h-5" />
                Create Link
              </button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-zinc-800 bg-zinc-800/50">
                    <th className="px-6 py-4 text-left text-sm font-semibold text-zinc-300">Short Code</th>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-zinc-300">Original URL</th>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-zinc-300">Clicks</th>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-zinc-300">Created</th>
                    <th className="px-6 py-4 text-right text-sm font-semibold text-zinc-300">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {links.map((link) => (
                    <tr key={link.id} className="border-b border-zinc-800 hover:bg-zinc-800/50 transition-colors">
                      <td className="px-6 py-4 font-mono text-emerald-500">{link.short_code}</td>
                      <td className="px-6 py-4 truncate text-sm text-zinc-300">{link.original_url}</td>
                      <td className="px-6 py-4 text-sm">{link.click_count || 0}</td>
                      <td className="px-6 py-4 text-sm text-zinc-400">{link.created_at}</td>
                      <td className="px-6 py-4 text-right">
                        <div className="flex justify-end gap-2">
                          <button
                            onClick={() => handleCopyLink(link.short_code)}
                            className="p-2 hover:bg-zinc-700 rounded-lg transition-colors text-zinc-400 hover:text-white"
                            title="Copy link"
                          >
                            <Copy className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => handleDeleteLink(link.id)}
                            className="p-2 hover:bg-red-900/20 rounded-lg transition-colors text-zinc-400 hover:text-red-400"
                            title="Delete link"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Create Link Modal */}
        <CreateLinkModal
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
          onSuccess={(newLink) => {
            setLinks([newLink, ...links]);
          }}
        />
      </main>
    </div>
  );
}
