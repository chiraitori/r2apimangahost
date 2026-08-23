<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { api } from '$lib/api';
  import { readerSettings } from '$lib/stores';
  import type { ChapterDetailResponse } from '$lib/types';
  import {
    ArrowLeft,
    ChevronLeft,
    ChevronRight,
    Maximize,
    Minimize,
    Columns,
    Layers,
    Loader2,
    Settings,
    BookOpen
  } from 'lucide-svelte';

  let data: ChapterDetailResponse | null = null;
  let loading = true;
  let errorMsg = '';
  let currentPageIndex = 0;
  let isFullscreen = false;
  let showHeader = true;
  let lastScrollY = 0;
  let mounted = false;
  let loadedChapterId = '';
  let chapterRequestId = 0;

  $: chapterId = $page.params.chapterId;

  $: if (mounted && chapterId && chapterId !== loadedChapterId) {
    loadedChapterId = chapterId;
    void loadChapter(chapterId);
  }

  async function loadChapter(id: string) {
    const requestId = ++chapterRequestId;
    loading = true;
    errorMsg = '';
    currentPageIndex = 0;
    try {
      const response = await api.getChapter(id);
      if (requestId !== chapterRequestId) return;
      data = response;
      window.scrollTo({ top: 0, behavior: 'smooth' });
    } catch (err: any) {
      if (requestId !== chapterRequestId) return;
      errorMsg = err.message || 'Không thể tải nội dung chapter';
    } finally {
      if (requestId === chapterRequestId) loading = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if ($readerSettings.mode === 'single' && data?.chapter?.pages) {
      if (e.key === 'ArrowRight' || e.key === 'd' || e.key === ' ') {
        nextPage();
      } else if (e.key === 'ArrowLeft' || e.key === 'a') {
        prevPage();
      }
    }
    if (e.key === 'f') {
      toggleFullscreen();
    }
  }

  function nextPage() {
    if (!data?.chapter?.pages) return;
    if (currentPageIndex < data.chapter.pages.length - 1) {
      currentPageIndex++;
      window.scrollTo({ top: 0, behavior: 'smooth' });
    } else if (data.nextChapter) {
      goto(`/read/${data.nextChapter.id}`);
    }
  }

  function prevPage() {
    if (currentPageIndex > 0) {
      currentPageIndex--;
      window.scrollTo({ top: 0, behavior: 'smooth' });
    } else if (data?.prevChapter) {
      goto(`/read/${data.prevChapter.id}`);
    }
  }

  function toggleFullscreen() {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().catch(() => {});
      isFullscreen = true;
    } else {
      document.exitFullscreen().catch(() => {});
      isFullscreen = false;
    }
  }

  function handleScroll() {
    const currentScrollY = window.scrollY;
    if (currentScrollY > 150 && currentScrollY > lastScrollY) {
      showHeader = false; // Hide on scroll down
    } else {
      showHeader = true; // Show on scroll up
    }
    lastScrollY = currentScrollY;
  }

  onMount(() => {
    mounted = true;
    if (chapterId) {
      loadedChapterId = chapterId;
      void loadChapter(chapterId);
    }
    window.addEventListener('keydown', handleKeydown);
    window.addEventListener('scroll', handleScroll, { passive: true });
  });

  onDestroy(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('keydown', handleKeydown);
      window.removeEventListener('scroll', handleScroll);
    }
  });
</script>

<svelte:head>
  <title>
    {data ? `${data.manga.title} - Ch. ${data.chapter.chapterNumber}` : 'Manga Reader - MangaHost'}
  </title>
</svelte:head>

<div class="min-h-screen bg-[#08090c] flex flex-col items-center">
  <!-- STICKY READER TOOLBAR -->
  <header
    class="fixed top-0 inset-x-0 z-50 glass border-b border-white/5 transition-transform duration-300 {showHeader ? 'translate-y-0' : '-translate-y-full'}"
  >
    <div class="max-w-7xl mx-auto px-4 h-14 flex items-center justify-between gap-4">
      <!-- Left: Back link & Title -->
      <div class="flex items-center gap-3 overflow-hidden">
        <a
          href={data ? `/manga/${data.manga.slug || data.manga.id}` : '/'}
          class="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-white/5 transition-colors"
          title="Quay lại truyện"
        >
          <ArrowLeft class="w-4 h-4" />
        </a>
        {#if data}
          <div class="flex flex-col truncate">
            <span class="font-bold text-xs text-slate-100 truncate">{data.manga.title}</span>
            <span class="text-[10px] text-rose-400 font-semibold">
              Chapter {data.chapter.chapterNumber} {data.chapter.title ? `- ${data.chapter.title}` : ''}
            </span>
          </div>
        {/if}
      </div>

      <!-- Center: Chapter Navigation -->
      {#if data}
        <div class="flex items-center gap-1.5">
          <button
            disabled={!data.prevChapter}
            on:click={() => data?.prevChapter && goto(`/read/${data.prevChapter.id}`)}
            class="p-1.5 rounded-lg bg-slate-900 border border-slate-800 text-slate-300 hover:text-white hover:bg-slate-800 disabled:opacity-30 disabled:pointer-events-none transition-colors"
            title={data.prevChapter ? `Ch. ${data.prevChapter.chapterNumber}` : 'Không có chapter trước'}
          >
            <ChevronLeft class="w-4 h-4" />
          </button>

          <span class="text-xs font-bold text-slate-300 px-2 py-1 bg-slate-900/80 rounded-lg border border-white/5 whitespace-nowrap">
            Ch. {data.chapter.chapterNumber}
          </span>

          <button
            disabled={!data.nextChapter}
            on:click={() => data?.nextChapter && goto(`/read/${data.nextChapter.id}`)}
            class="p-1.5 rounded-lg bg-slate-900 border border-slate-800 text-slate-300 hover:text-white hover:bg-slate-800 disabled:opacity-30 disabled:pointer-events-none transition-colors"
            title={data.nextChapter ? `Ch. ${data.nextChapter.chapterNumber}` : 'Không có chapter sau'}
          >
            <ChevronRight class="w-4 h-4" />
          </button>
        </div>
      {/if}

      <!-- Right: Reader Mode Controls -->
      <div class="flex items-center gap-2">
        <!-- Mode Switcher -->
        <button
          on:click={() => readerSettings.setMode($readerSettings.mode === 'webtoon' ? 'single' : 'webtoon')}
          class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-slate-900 border border-slate-800 text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800 transition-colors"
          title="Đổi chế độ đọc"
        >
          {#if $readerSettings.mode === 'webtoon'}
            <Layers class="w-3.5 h-3.5 text-rose-400" />
            <span class="hidden sm:inline">Webtoon (Cuộn dọc)</span>
          {:else}
            <Columns class="w-3.5 h-3.5 text-rose-400" />
            <span class="hidden sm:inline">Lật trang</span>
          {/if}
        </button>

        <!-- Fullscreen -->
        <button
          on:click={toggleFullscreen}
          class="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-white/5 transition-colors hidden sm:inline-flex"
          title="Toàn màn hình (F)"
        >
          {#if isFullscreen}
            <Minimize class="w-4 h-4" />
          {:else}
            <Maximize class="w-4 h-4" />
          {/if}
        </button>
      </div>
    </div>
  </header>

  <!-- READER MAIN CONTENT -->
  <main class="w-full flex-1 pt-16 pb-20 flex flex-col items-center">
    {#if loading}
      <div class="min-h-[70vh] flex flex-col items-center justify-center gap-3 text-slate-400">
        <Loader2 class="w-8 h-8 text-rose-500 animate-spin" />
        <span class="text-xs font-medium">Đang tải trang truyện từ Cloudflare R2...</span>
      </div>
    {:else if errorMsg || !data}
      <div class="py-24 text-center glass rounded-2xl p-8 max-w-md mx-auto space-y-3 border border-rose-500/20">
        <p class="text-rose-400 text-sm font-bold">{errorMsg || 'Không thể tải chapter'}</p>
        <a href="/" class="inline-block px-4 py-2 bg-slate-800 rounded-xl text-xs text-white font-semibold">
          Quay Lại Trang Chủ
        </a>
      </div>
    {:else}
      <!-- MODE 1: WEBTOON / CONTINUOUS VERTICAL SCROLL -->
      {#if $readerSettings.mode === 'webtoon'}
        <div class="w-full flex flex-col items-center" style="max-width: {$readerSettings.maxWidth}px;">
          {#each data.chapter.pages as pageUrl, idx}
            <div class="w-full relative min-h-[300px] flex items-center justify-center bg-slate-950/50">
              <img
                src={pageUrl}
                alt={`Trang ${idx + 1}`}
                loading="lazy"
                class="w-full h-auto block select-none"
              />
            </div>
          {/each}
        </div>
      {:else}
        <!-- MODE 2: PAGINATED / SINGLE PAGE -->
        <div
          class="w-full flex flex-col items-center justify-center cursor-pointer select-none"
          style="max-width: {$readerSettings.maxWidth}px;"
          role="button"
          tabindex="0"
          on:click={nextPage}
          on:keydown={(e) => e.key === 'Enter' && nextPage()}
        >
          <div class="relative min-h-[400px] flex items-center justify-center bg-slate-950">
            <img
              src={data.chapter.pages[currentPageIndex]}
              alt={`Trang ${currentPageIndex + 1}`}
              class="max-h-[88vh] w-auto object-contain"
            />
          </div>

          <!-- Bottom Page Indicator -->
          <div class="mt-4 flex items-center gap-4 bg-slate-900/90 px-4 py-2 rounded-xl border border-white/10 text-xs text-slate-300">
            <button
              on:click|stopPropagation={prevPage}
              disabled={currentPageIndex === 0}
              class="px-2 py-1 rounded bg-slate-800 hover:bg-slate-700 disabled:opacity-30"
            >
              Trang trước
            </button>
            <span class="font-bold">
              {currentPageIndex + 1} / {data.chapter.pages.length}
            </span>
            <button
              on:click|stopPropagation={nextPage}
              disabled={currentPageIndex >= data.chapter.pages.length - 1}
              class="px-2 py-1 rounded bg-slate-800 hover:bg-slate-700 disabled:opacity-30"
            >
              Trang sau
            </button>
          </div>
        </div>
      {/if}

      <!-- BOTTOM CHAPTER NAVIGATOR -->
      <div class="w-full max-w-xl mx-auto px-4 mt-12 space-y-4 text-center">
        <div class="p-6 rounded-2xl glass border border-white/10 space-y-4">
          <p class="text-xs text-slate-400">
            Bạn đã đọc hết <span class="text-white font-bold">Chapter {data.chapter.chapterNumber}</span>
          </p>

          <div class="flex items-center justify-center gap-3">
            {#if data.prevChapter}
              <a
                href={`/read/${data.prevChapter.id}`}
                class="px-4 py-2.5 rounded-xl bg-slate-900 hover:bg-slate-800 text-slate-300 text-xs font-bold border border-slate-800 transition-colors"
              >
                ← Chapter {data.prevChapter.chapterNumber}
              </a>
            {/if}

            <a
              href={`/manga/${data.manga.slug || data.manga.id}`}
              class="px-4 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-bold border border-white/10 transition-colors"
            >
              Mục lục truyện
            </a>

            {#if data.nextChapter}
              <a
                href={`/read/${data.nextChapter.id}`}
                class="px-5 py-2.5 rounded-xl bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold shadow-lg shadow-rose-600/30 transition-all"
              >
                Chapter {data.nextChapter.chapterNumber} →
              </a>
            {/if}
          </div>
        </div>
      </div>
    {/if}
  </main>
</div>
