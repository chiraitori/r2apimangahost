export interface Manga {
  id: string;
  title: string;
  altTitles: string[];
  slug: string;
  coverUrl: string;
  description: string;
  author: string;
  artist: string;
  status: 'Ongoing' | 'Completed' | 'Hiatus' | 'Cancelled';
  genres: string[];
  views: number;
  rating: number;
  createdAt: string;
  updatedAt: string;
  lastChapterNumber?: number;
  chapterCount?: number;
}

export interface Chapter {
  id: string;
  mangaId: string;
  mangaSlug: string;
  chapterNumber: number;
  title: string;
  volume?: number;
  language: string;
  pages: string[];
  pageCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface ChapterNav {
  id: string;
  chapterNumber: number;
  title: string;
}

export interface ChapterDetailResponse {
  chapter: Chapter;
  manga: Manga;
  prevChapter?: ChapterNav;
  nextChapter?: ChapterNav;
}

export interface MangaDetailResponse {
  manga: Manga;
  chapters: Chapter[];
}

export interface Pagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface MangaListResponse {
  data: Manga[];
  pagination: Pagination;
}

export interface HomeResponse {
  featured: Manga[];
  latest: Manga[];
  popular: Manga[];
}

export interface StatsResponse {
  totalMangas: number;
  totalChapters: number;
  totalViews: number;
}

export interface APIResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
}
