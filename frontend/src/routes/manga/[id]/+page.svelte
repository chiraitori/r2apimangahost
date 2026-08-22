<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { api } from '$lib/api';
  import type { Manga, Chapter } from '$lib/types';
  import { BookOpen, User, Calendar, Eye, Star, ArrowUpDown, Search, Play, ArrowLeft, Loader2, Sparkles } from 'lucide-svelte';

  let manga: Manga | null = null;
  let chapters: Chapter[] = [];
  let loading = true;
  let errorMsg = '';
  let chapterFilter = '';
  let sortAscending = false;

  $: mangaId = $page.params.id;

  onMount(async () => {
    await fetchMangaDetails();
  });

  async function fetchMangaDetails() {
    loading = true;
    errorMsg = '';
    try {
      const res = await api.getManga(mangaId);
      manga = res.manga;
      chapters = res.chapters;
    } catch (err: any) {
      errorMsg = err.message || 'Không thể tải thông tin truyện';
    } finally {
      loading = false;
    }
  }

  $: firstChap = chapters.length > 0 ? [...chapters].sort((a, b) => a.chapterNumber - b.chapterNumber)[0] : null;
  $: latestChap = chapters.length > 0 ? [...chapters].sort((a, b) => b.chapterNumber - a.chapterNumber)[0] : null;

  $: filteredChapters = chapters
    .filter((c) => {
      if (!chapterFilter.trim()) return true;
      const term = chapterFilter.toLowerCase();
      return (
        c.chapterNumber.toString().includes(term) ||
        (c.title && c.title.toLowerCase().includes(term))
      );
    })
    .sort((a, b) => {
      return sortAscending
        ? a.chapterNumber - b.chapterNumber
        : b.chapterNumber - a.chapterNumber;
    });

  function formatDate(d: string): string {
    if (!d) return '';
    const date = new Date(d);
    return date.toLocaleDateString('vi-VN', { year: 'numeric', month: 'short', day: 'numeric' });
  }
</script>

<svelte:head>
  <title>{manga ? `${manga.title} - MangaHost` : 'Chi Tiết Truyện - MangaHost'}</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-8">
  <a
    href="/"
    class="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-400 hover:text-white transition-colors"
  >
    <ArrowLeft class="w-4 h-4" /> Quay lại trang chủ
  </a>

  {#if loading}
    <div class="py-24 flex flex-col items-center justify-center gap-3 text-slate-400">
      <Loader2 class="w-8 h-8 text-rose-500 animate-spin" />
      <span class="text-sm">Đang tải chi tiết truyện...</span>
    </div>
  {:else if errorMsg || !manga}
    <div class="py-16 text-center glass rounded-2xl p-8 border border-rose-500/20 max-w-md mx-auto space-y-3">
      <p class="text-rose-400 font-bold text-base">{errorMsg || 'Không tìm thấy truyện'}</p>
      <a href="/" class="inline-block px-4 py-2 bg-slate-800 rounded-xl text-xs text-white font-semibold">
        Về Trang Chủ
      </a>
    </div>
  {:else}
    <!-- MANGA HEADER HERO -->
    <div class="relative rounded-3xl overflow-hidden glass p-6 sm:p-8 border border-white/10 shadow-2xl">
      <!-- Glow background -->
      <div class="absolute top-0 right-0 w-96 h-96 bg-rose-600/10 rounded-full blur-3xl pointer-events-none"></div>

      <div class="relative z-10 flex flex-col md:flex-row gap-8">
        <!-- Cover Art -->
        <div class="shrink-0 flex justify-center">
          <div class="relative group">
            <img
              src={manga.coverUrl || '/placeholder.jpg'}
              alt={manga.title}
              class="w-56 sm:w-64 aspect-[3/4] object-cover rounded-2xl shadow-2xl border border-white/10"
            />
          </div>
        </div>

        <!-- Info & Actions -->
        <div class="flex-1 space-y-4">
          <div>
            <div class="flex items-center gap-2 mb-2">
              <span class="px-2.5 py-0.5 text-[11px] font-bold uppercase rounded-md {manga.status === 'Completed' ? 'bg-emerald-500 text-white' : 'bg-rose-600 text-white'}">
                {manga.status || 'Ongoing'}
              </span>
              {#if manga.rating}
                <span class="inline-flex items-center gap-1 text-xs font-bold text-amber-300 bg-slate-900/80 px-2 py-0.5 rounded-md border border-white/5">
                  <Star class="w-3.5 h-3.5 fill-amber-400 text-amber-400" />
                  {manga.rating.toFixed(1)}
                </span>
              {/if}
            </div>

            <h1 class="text-2xl sm:text-4xl font-extrabold text-white tracking-tight">
              {manga.title}
            </h1>
            {#if manga.altTitles && manga.altTitles.length > 0}
              <p class="text-xs text-slate-400 mt-1">
                Tên khác: {manga.altTitles.join(' • ')}
              </p>
            {/if}
          </div>

          <!-- Metadata row -->
          <div class="grid grid-cols-2 sm:grid-cols-3 gap-3 text-xs text-slate-300 bg-slate-900/50 p-4 rounded-xl border border-white/5">
            <div>
              <span class="text-slate-500 block text-[10px] uppercase font-bold">Tác giả</span>
              <span class="font-semibold text-slate-200">{manga.author || 'Đang cập nhật'}</span>
            </div>
            <div>
              <span class="text-slate-500 block text-[10px] uppercase font-bold">Họa sĩ</span>
              <span class="font-semibold text-slate-200">{manga.artist || 'Đang cập nhật'}</span>
            </div>
            <div>
              <span class="text-slate-500 block text-[10px] uppercase font-bold">Lượt xem</span>
              <span class="font-semibold text-slate-200">{manga.views?.toLocaleString() || 0}</span>
            </div>
          </div>

          <!-- Genres -->
          {#if manga.genres && manga.genres.length > 0}
            <div class="flex flex-wrap gap-1.5 pt-1">
              {#each manga.genres as g}
                <a
                  href={`/?genre=${encodeURIComponent(g)}`}
                  class="px-2.5 py-1 text-xs font-medium bg-slate-800 hover:bg-rose-600 hover:text-white rounded-lg text-slate-300 border border-white/5 transition-colors"
                >
                  {g}
                </a>
              {/each}
            </div>
          {/if}

          <!-- Description -->
          <div class="pt-2">
            <h3 class="text-xs font-bold text-slate-400 uppercase tracking-wider mb-1.5">Tóm tắt nội dung</h3>
            <p class="text-sm text-slate-300 leading-relaxed max-w-3xl whitespace-pre-line">
              {manga.description || 'Chưa có tóm tắt cho bộ truyện này.'}
            </p>
          </div>

          <!-- Quick Read Buttons -->
          {#if chapters.length > 0 && firstChap && latestChap}
            <div class="flex flex-wrap items-center gap-3 pt-4 border-t border-white/10">
              <a
                href={`/read/${firstChap.id}`}
                class="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-gradient-to-r from-rose-600 to-pink-600 text-white font-bold text-xs shadow-lg shadow-rose-600/30 hover:scale-105 transition-all"
              >
                <Play class="w-3.5 h-3.5 fill-white" /> Đọc Từ Đầu (Ch. {firstChap.chapterNumber})
              </a>

              <a
                href={`/read/${latestChap.id}`}
                class="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 font-bold text-xs border border-white/10 transition-all"
              >
                <Sparkles class="w-3.5 h-3.5 text-amber-400" /> Đọc Mới Nhất (Ch. {latestChap.chapterNumber})
              </a>
            </div>
          {/if}
        </div>
      </div>
    </div>

    <!-- CHAPTER LIST SECTION -->
    <div class="space-y-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-800 pb-3">
        <div class="flex items-center gap-2">
          <BookOpen class="w-5 h-5 text-rose-500" />
          <h2 class="text-lg font-bold text-white tracking-tight">
            Danh Sách Chương ({chapters.length})
          </h2>
        </div>

        <div class="flex items-center gap-2">
          <!-- Chapter Search -->
          <div class="relative">
            <Search class="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              bind:value={chapterFilter}
              placeholder="Tìm số chapter..."
              class="w-36 sm:w-48 pl-8 pr-3 py-1.5 bg-slate-900 border border-slate-800 rounded-xl text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-rose-500"
            />
          </div>

          <!-- Sort Direction Toggle -->
          <button
            on:click={() => sortAscending = !sortAscending}
            class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-900 border border-slate-800 rounded-xl text-xs font-semibold text-slate-300 hover:bg-slate-800 transition-colors"
          >
            <ArrowUpDown class="w-3.5 h-3.5 text-rose-400" />
            {sortAscending ? 'Cũ nhất' : 'Mới nhất'}
          </button>
        </div>
      </div>

      <!-- Chapters Grid / List -->
      {#if filteredChapters.length === 0}
        <div class="py-12 text-center glass rounded-2xl p-6 text-slate-400 text-xs">
          {chapterFilter ? 'Không tìm thấy chapter phù hợp' : 'Chưa có chapter nào được upload.'}
        </div>
      {:else}
        <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
          {#each filteredChapters as chapter (chapter.id)}
            <a
              href={`/read/${chapter.id}`}
              class="p-3 rounded-xl bg-slate-900/60 hover:bg-slate-800/80 border border-slate-800 hover:border-rose-500/40 flex items-center justify-between group transition-all"
            >
              <div class="flex flex-col">
                <span class="font-bold text-xs text-slate-200 group-hover:text-rose-400 transition-colors">
                  Chapter {chapter.chapterNumber}
                  {#if chapter.title}
                    <span class="font-normal text-slate-400 text-[11px] ml-1">- {chapter.title}</span>
                  {/if}
                </span>
                <span class="text-[10px] text-slate-500 mt-0.5">
                  {chapter.pageCount || chapter.pages.length} trang • {formatDate(chapter.createdAt)}
                </span>
              </div>
              <Play class="w-3.5 h-3.5 text-slate-600 group-hover:text-rose-500 shrink-0 transition-colors" />
            </a>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>
