'use client';

import { useEffect, useState } from 'react';
import { TrendingUp, AlertCircle } from 'lucide-react';

export default function LiveCounter({ shortCode }: Readonly<{ shortCode?: string }>) {
  const [count, setCount] = useState(0);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    if (!shortCode) {
      return;
    }

    let ws: WebSocket | null = null;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 5;

    function handleMessage(event: MessageEvent) {
      try {
        const data = JSON.parse(event.data) as { action: string };
        if (data.action === 'click') {
          setCount((prev) => prev + 1);
        }
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e);
      }
    }

    function handleReconnect() {
      if (reconnectAttempts < maxReconnectAttempts) {
        const delay = 1000 * Math.pow(2, reconnectAttempts);
        setTimeout(() => {
          reconnectAttempts += 1;
          connect();
        }, delay);
      }
    }

    function connect() {
      if (!shortCode) return;
      try {
        const protocol = typeof globalThis !== 'undefined' && globalThis.location?.protocol === 'https:' ? 'wss' : 'ws';
        ws = new WebSocket(`${protocol}://localhost:8083/ws/live/${shortCode}`);

        ws.onopen = () => {
          setConnected(true);
          reconnectAttempts = 0;
        };

        ws.onmessage = handleMessage;

        ws.onerror = () => {
          setConnected(false);
        };

        ws.onclose = () => {
          setConnected(false);
          handleReconnect();
        };
      } catch (e) {
        console.error('WebSocket connection error:', e);
        setConnected(false);
      }
    }

    connect();

    return () => {
      if (ws) {
        ws.close();
      }
    };
  }, [shortCode]);

  if (!shortCode) {
    return (
      <div className="bg-gradient-to-br from-zinc-900 to-zinc-800 rounded-3xl p-8 border border-zinc-700 shadow-xl">
        <div className="flex items-center justify-between mb-6">
          <div>
            <p className="text-emerald-400 text-sm uppercase tracking-widest font-semibold">Live Clicks</p>
            <p className="text-zinc-400 text-sm mt-1">Real-time click counter</p>
          </div>
          <div className="w-3 h-3 rounded-full bg-zinc-600"></div>
        </div>
        <div className="flex items-center gap-3 text-zinc-400">
          <AlertCircle className="w-5 h-5" />
          <span className="text-sm">Create a short link to see live stats</span>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-gradient-to-br from-zinc-900 to-zinc-800 rounded-3xl p-8 border border-zinc-700 shadow-xl">
      <div className="flex items-center justify-between mb-6">
        <div>
          <p className="text-emerald-400 text-sm uppercase tracking-widest font-semibold">Live Clicks</p>
          <p className="text-zinc-400 text-sm mt-1">Real-time click counter</p>
        </div>
        <div className={`w-3 h-3 rounded-full ${connected ? 'bg-emerald-500 animate-pulse' : 'bg-zinc-600'}`}></div>
      </div>

      <div className="mb-4">
        <div className="text-7xl font-black text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 to-emerald-600">
          {count.toLocaleString()}
        </div>
      </div>

      <div className="flex items-center gap-2 text-emerald-400">
        <TrendingUp className="w-4 h-4" />
        <span className="text-sm">{connected ? 'Connected' : 'Reconnecting...'}</span>
      </div>
    </div>
  );
}
