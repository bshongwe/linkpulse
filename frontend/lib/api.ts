import axios, { AxiosInstance } from 'axios';
import { ShortLink, CreateShortLinkRequest, AnalyticsSummary } from '@/types';

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080';

console.log('🔍 API_BASE:', API_BASE); // DEBUG: Log the API base URL

const api: AxiosInstance = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add token to requests
api.interceptors.request.use((config) => {
  if (globalThis.window) {
    const token = globalThis.window.localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

// Shortened links (via BFF)
export async function createShortLink(data: CreateShortLinkRequest): Promise<ShortLink> {
  const url = '/api/v1/links';
  
  // Map frontend types to BFF types
  const bffPayload = {
    url: data.original_url,
    custom: data.custom_alias,
    expires_at: data.expires_at,
    tags: data.tags,
  };
  
  console.log('📤 POST', API_BASE + url, 'with data:', bffPayload); // DEBUG
  const response = await api.post(url, bffPayload);
  console.log('✅ Response:', response.data); // DEBUG
  return response.data.data || response.data;
}

export async function getShortLink(shortCode: string): Promise<ShortLink> {
  const response = await api.get(`/api/v1/links/${shortCode}`);
  return response.data.data || response.data;
}

export async function updateShortLink(linkId: string, data: Partial<ShortLink>): Promise<ShortLink> {
  const response = await api.put(`/api/v1/links/id/${linkId}`, data);
  return response.data.data || response.data;
}

export async function deleteShortLink(linkId: string): Promise<void> {
  await api.delete(`/api/v1/links/id/${linkId}`);
}

export async function listLinks(): Promise<ShortLink[]> {
  const response = await api.get(`/api/v1/links`, {
    params: {
      page: 1,
      page_size: 100,
    },
  });
  return response.data.data?.links || response.data.data || response.data || [];
}

// Analytics
export async function getAnalytics(linkId: string): Promise<AnalyticsSummary> {
  const response = await api.get(`/api/v1/links/${linkId}/analytics`);
  return response.data;
}

export async function getLiveCount(linkId: string): Promise<number> {
  const response = await api.get(`/api/v1/links/${linkId}/analytics`);
  return response.data.liveCount || 0;
}

// Error handling
export function getErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    return error.response?.data?.error || error.message || 'An error occurred';
  }
  return 'An unexpected error occurred';
}

export default api;
