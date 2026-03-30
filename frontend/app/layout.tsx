import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'LinkPulse | Smart URL Shortener',
  description: 'Create, manage, and analyze shortened URLs with real-time analytics.',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
