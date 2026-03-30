export interface ShortLink {
  id: string;
  short_code: string;
  original_url: string;
  workspace_id: string;
  title?: string;
  description?: string;
  custom_alias?: string;
  expires_at?: number;
  is_active: boolean;
  qr_code?: string;
  click_count: number;
  created_at: string;
  updated_at: string;
}

export interface CreateShortLinkRequest {
  original_url: string;
  workspace_id: string;
  created_by?: string;
  title?: string;
  description?: string;
  custom_alias?: string;
  expires_at?: number;
  redirect_type?: 'permanent' | 'temporary';
  tags?: string[];
  campaign_id?: string;
}

export interface AnalyticsSummary {
  total_clicks: number;
  clicks_last_24h: number;
  clicks_last_7d: number;
  clicks_last_30d: number;
  top_countries: Record<string, number>;
  top_devices: Record<string, number>;
  top_referrers: Record<string, number>;
  top_utm_sources: Record<string, number>;
}

export interface User {
  id: string;
  email: string;
  name: string;
  workspace_id: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token?: string;
  user: User;
}
