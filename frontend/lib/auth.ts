import Cookies from 'js-cookie';

const TOKEN_KEY = 'access_token';
const USER_KEY = 'user';

/**
 * Decode JWT token to extract claims
 * Note: This does NOT verify the signature - verification happens on the backend
 */
export function decodeJWT(token: string): any {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) {
      throw new Error('Invalid token format');
    }
    const decoded = JSON.parse(atob(parts[1]));
    return decoded;
  } catch (error) {
    console.error('Failed to decode JWT:', error);
    return null;
  }
}

export function setAuthToken(token: string) {
  if (globalThis.window) {
    globalThis.window.localStorage.setItem(TOKEN_KEY, token);
    Cookies.set(TOKEN_KEY, token, { 
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'strict',
    });
  }
}

export function getAuthToken(): string | null {
  if (globalThis.window) {
    return globalThis.window.localStorage.getItem(TOKEN_KEY);
  }
  return null;
}

export function removeAuthToken() {
  if (globalThis.window) {
    globalThis.window.localStorage.removeItem(TOKEN_KEY);
    globalThis.window.localStorage.removeItem(USER_KEY);
    Cookies.remove(TOKEN_KEY);
  }
}

export function isAuthenticated(): boolean {
  return getAuthToken() !== null;
}

export function setUser(user: any) {
  if (globalThis.window) {
    globalThis.window.localStorage.setItem(USER_KEY, JSON.stringify(user));
  }
}

export function getUser() {
  if (globalThis.window) {
    const user = globalThis.window.localStorage.getItem(USER_KEY);
    return user ? JSON.parse(user) : null;
  }
  return null;
}
