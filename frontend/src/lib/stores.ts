import { writable } from 'svelte/store';

// Check if browser environment
const isBrowser = typeof window !== 'undefined';

// Admin Auth Store
function createAuthStore() {
  const initialToken = isBrowser ? localStorage.getItem('mangahost_admin_token') : null;
  const initialUser = isBrowser ? localStorage.getItem('mangahost_admin_user') : null;

  const { subscribe, set, update } = writable({
    isLoggedIn: !!initialToken,
    token: initialToken,
    username: initialUser
  });

  return {
    subscribe,
    login: (token: string, username: string) => {
      if (isBrowser) {
        localStorage.setItem('mangahost_admin_token', token);
        localStorage.setItem('mangahost_admin_user', username);
      }
      set({ isLoggedIn: true, token, username });
    },
    logout: () => {
      if (isBrowser) {
        localStorage.removeItem('mangahost_admin_token');
        localStorage.removeItem('mangahost_admin_user');
      }
      set({ isLoggedIn: false, token: null, username: null });
    }
  };
}

export const authStore = createAuthStore();

// Reader Settings Store
export interface ReaderSettings {
  mode: 'webtoon' | 'single';
  maxWidth: number; // in pixels or percentage
  quality: 'original' | 'compressed';
}

const defaultReaderSettings: ReaderSettings = {
  mode: 'webtoon',
  maxWidth: 900,
  quality: 'original'
};

function createReaderSettingsStore() {
  const saved = isBrowser ? localStorage.getItem('mangahost_reader_settings') : null;
  const initial: ReaderSettings = saved ? JSON.parse(saved) : defaultReaderSettings;

  const { subscribe, set, update } = writable<ReaderSettings>(initial);

  return {
    subscribe,
    setMode: (mode: 'webtoon' | 'single') => {
      update((s) => {
        const next = { ...s, mode };
        if (isBrowser) localStorage.setItem('mangahost_reader_settings', JSON.stringify(next));
        return next;
      });
    },
    setMaxWidth: (maxWidth: number) => {
      update((s) => {
        const next = { ...s, maxWidth };
        if (isBrowser) localStorage.setItem('mangahost_reader_settings', JSON.stringify(next));
        return next;
      });
    }
  };
}

export const readerSettings = createReaderSettingsStore();
