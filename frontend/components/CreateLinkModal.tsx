'use client';

import { useState } from 'react';
import { X } from 'lucide-react';
import { createShortLink } from '@/lib/api';
import { getUser } from '@/lib/auth';
import { CreateShortLinkRequest } from '@/types';

interface CreateLinkModalProps {
  readonly isOpen: boolean;
  readonly onClose: () => void;
  readonly onSuccess: (newLink: any) => void;
}

export default function CreateLinkModal({ isOpen, onClose, onSuccess }: Readonly<CreateLinkModalProps>) {
  const [originalUrl, setOriginalUrl] = useState('');
  const [customAlias, setCustomAlias] = useState('');
  const [title, setTitle] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const user = getUser();
      if (!user) {
        setError('Not authenticated. Please log in again.');
        setLoading(false);
        return;
      }

      // Validate URL format
      try {
        new URL(originalUrl);
      } catch {
        setError('Please enter a valid URL (e.g., https://example.com)');
        setLoading(false);
        return;
      }

      const request: CreateShortLinkRequest = {
        original_url: originalUrl,
        workspace_id: user.workspace_id || 'default',
        title: title || undefined,
        custom_alias: customAlias || undefined,
      };

      const result = await createShortLink(request);

      if (result) {
        onSuccess(result);
        onClose();
        setOriginalUrl('');
        setCustomAlias('');
        setTitle('');
      } else {
        setError('Failed to create link. Please try again.');
      }
    } catch (err: any) {
      console.error('Create link error:', err);
      const errorMsg = err.response?.data?.error || err.message || 'Failed to connect to shortener service';
      setError(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 px-4">
      <div className="bg-zinc-900 rounded-2xl border border-zinc-800 p-8 w-full max-w-md shadow-2xl">
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-2xl font-bold">Create New Short Link</h2>
          <button
            onClick={onClose}
            className="text-zinc-400 hover:text-white transition-colors"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {error && (
          <div className="bg-red-900/20 border border-red-800 rounded-lg p-4 mb-4 text-red-400 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="destination-url" className="block text-sm font-medium text-zinc-200 mb-2">
              Destination URL *
            </label>
            <input
              id="destination-url"
              type="url"
              value={originalUrl}
              onChange={(e) => setOriginalUrl(e.target.value)}
              placeholder="https://example.com/very-long-url"
              required
              className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-4 py-3 text-white placeholder-zinc-500 focus:border-emerald-500 focus:outline-none transition-colors"
            />
            <p className="text-xs text-zinc-500 mt-1">The URL you want to shorten</p>
          </div>

          <div>
            <label htmlFor="title" className="block text-sm font-medium text-zinc-200 mb-2">
              Title (optional)
            </label>
            <input
              id="title"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="My awesome link"
              className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-4 py-3 text-white placeholder-zinc-500 focus:border-emerald-500 focus:outline-none transition-colors"
            />
            <p className="text-xs text-zinc-500 mt-1">Helps you identify this link</p>
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
              placeholder="my-promo"
              className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-4 py-3 text-white placeholder-zinc-500 focus:border-emerald-500 focus:outline-none transition-colors"
            />
            <p className="text-xs text-zinc-500 mt-1">Leave blank to auto-generate</p>
          </div>

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="flex-1 bg-zinc-800 hover:bg-zinc-700 disabled:bg-zinc-800/50 text-white font-semibold py-3 rounded-lg transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex-1 bg-emerald-600 hover:bg-emerald-700 disabled:bg-emerald-600/50 text-white font-semibold py-3 rounded-lg transition-colors"
            >
              {loading ? 'Creating...' : 'Create Link'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
