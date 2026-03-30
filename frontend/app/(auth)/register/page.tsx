'use client';

import Link from 'next/link';
import { LinkIcon } from 'lucide-react';

export default function RegisterPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-zinc-950 via-zinc-900 to-zinc-950 flex flex-col items-center justify-center px-4">
      <div className="relative z-10 w-full max-w-md">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-12 h-12 bg-emerald-600 rounded-xl mb-4">
            <LinkIcon className="w-7 h-7 text-white" />
          </div>
          <h1 className="text-3xl font-bold">LinkPulse</h1>
          <p className="text-zinc-400 mt-2">Create your account</p>
        </div>

        <div className="bg-zinc-900 rounded-2xl border border-zinc-800 p-8 shadow-xl text-center">
          <p className="text-zinc-400 mb-4">Registration coming soon!</p>
          <p className="text-zinc-500 text-sm mb-6">For now, use any email to login in the sign in page.</p>
          <Link href="/login" className="text-emerald-500 hover:text-emerald-400">
            Back to login
          </Link>
        </div>
      </div>
    </div>
  );
}
