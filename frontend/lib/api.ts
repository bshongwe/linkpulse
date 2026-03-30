import axios, { AxiosInstance } from 'axios';
import { ShortLink, CreateShortLinkRequest, AnalyticsSummary } from '@/types';

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8082';

const api: AxiosInstance = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add token to requests
api.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

// Shortened links
export async function createShortLink(data: CreateShortLinkRequest): Promise<ShortLink> {
  const response = await api.post('/shorten', data);
  return response.data;
}

export async function getShortLink(shortCode: string): Promise<ShortLink> {
  const response = await api.get(`/link/${shortCode}`);
  return response.data;
}

export async function updateShortLink(linkId: string, data: Partial<ShortLink>): Promise<ShortLink> {
  const response = await api.put(`/link/${linkId}`, data);
  return response.data;
}

export async function deleteShortLink(linkId: string): Promise<void> {
  await api.delete(`/link/${linkId}`);
}

export async function listLinks(workspaceId: string): Promise<ShortLink[]> {
  const response = await api.get(`/links?workspace_id=${workspaceId}`);
  return response.data;
}

// Analytics
export async function getAnalytics(shortCode: string): Promise<AnalyticsSummary> {
  const response = await api.get(`/analytics/${shortCode}`);
  return response.data;
}

export async function getLiveCount(shortCode: string): Promise<number> {
  const response = await api.get(`/analytics/${shortCode}/live`);
  return response.data.count;
}

// Error handling
export function getErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    return error.response?.data?.error || error.message || 'An error occurred';
  }
  return 'An unexpected error occurred';
}

export default api;
