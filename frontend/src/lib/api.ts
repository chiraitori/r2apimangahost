import type {
  Manga,
  Chapter,
  MangaListResponse,
  MangaDetailResponse,
  ChapterDetailResponse,
  HomeResponse,
  StatsResponse,
  APIResponse
} from './types';

// Read API URL from Vite environment variable (e.g. deployed Go backend domain)
// or fallback to localhost:8080/api/v1 for local development
const API_BASE = (import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1').replace(/\/$/, '');

function getAuthHeader(): Record<string, string> {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('mangahost_admin_token');
    if (token) {
      return { Authorization: `Bearer ${token}` };
    }
  }
  return {};
}

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE}${endpoint.startsWith('/') ? '' : '/'}${endpoint}`;
  
  const headers = {
    ...getAuthHeader(),
    ...(options.headers || {})
  };

  const res = await fetch(url, {
    ...options,
    headers
  });

  const json: APIResponse<T> = await res.json().catch(() => ({
    success: false,
    error: `Network error or invalid JSON response from ${url}`
  }));

  if (!res.ok || !json.success) {
    throw new Error(json.error || json.message || `Request failed with status ${res.status}`);
  }

  return json.data as T;
}

export const api = {
  // Base URL
  getBaseUrl: () => API_BASE,

  // Public Endpoints
  getHome: () => request<HomeResponse>('/home'),
  
  getMangas: (params: {
    q?: string;
    genre?: string;
    status?: string;
    sort?: string;
    order?: 'asc' | 'desc';
    page?: number;
    limit?: number;
  } = {}) => {
    const query = new URLSearchParams();
    if (params.q) query.set('q', params.q);
    if (params.genre) query.set('genre', params.genre);
    if (params.status) query.set('status', params.status);
    if (params.sort) query.set('sort', params.sort);
    if (params.order) query.set('order', params.order);
    if (params.page) query.set('page', params.page.toString());
    if (params.limit) query.set('limit', params.limit.toString());
    return request<MangaListResponse>(`/manga?${query.toString()}`);
  },

  getManga: (idOrSlug: string) => request<MangaDetailResponse>(`/manga/${idOrSlug}`),

  getChapter: (chapterId: string) => request<ChapterDetailResponse>(`/chapter/${chapterId}`),

  getGenres: () => request<string[]>('/genres'),

  getStats: () => request<StatsResponse>('/stats'),

  // Auth Endpoints
  login: async (username: string, password: string) => {
    const res = await request<{ token: string; username: string; role: string }>('/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });
    if (typeof window !== 'undefined') {
      localStorage.setItem('mangahost_admin_token', res.token);
      localStorage.setItem('mangahost_admin_user', res.username);
    }
    return res;
  },

  logout: () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('mangahost_admin_token');
      localStorage.removeItem('mangahost_admin_user');
    }
  },

  getMe: () => request<{ userId: string; username: string; role: string }>('/auth/me'),

  // Admin Endpoints
  createManga: (formData: FormData) => {
    return request<Manga>('/manga', {
      method: 'POST',
      body: formData // Note: do not set Content-Type header manually for FormData so browser sets boundary
    });
  },

  updateManga: (idOrSlug: string, formData: FormData) => {
    return request<Manga>(`/manga/${idOrSlug}`, {
      method: 'PUT',
      body: formData
    });
  },

  deleteManga: (idOrSlug: string) => {
    return request<{ id: string; slug: string }>(`/manga/${idOrSlug}`, {
      method: 'DELETE'
    });
  },

  uploadChapter: (mangaIdOrSlug: string, formData: FormData) => {
    return request<Chapter>(`/manga/${mangaIdOrSlug}/chapters`, {
      method: 'POST',
      body: formData
    });
  },

  deleteChapter: (chapterId: string) => {
    return request<{ id: string }>(`/chapter/${chapterId}`, {
      method: 'DELETE'
    });
  }
};
