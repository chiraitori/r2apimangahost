<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { api } from '$lib/api';
  import type { Manga, HomeResponse, Pagination } from '$lib/types';
  import MangaCard from '$lib/components/MangaCard.svelte';
  import { Sparkles, TrendingUp, Flame, BookOpen, ChevronRight, Filter, Search, Loader2 } from 'lucide-svelte';

  let homeData: HomeResponse | null = null;
  let mangaList: Manga[] = [];
  let pagination: Pagination | null = null;
  let genres: string[] = [];
  let selectedGenre = '';
  let searchQuery = '';
  let selectedStatus = '';
  let sortBy = 'updatedAt';
  let currentPage = 1;
  let loading = true;
  let loadingCatalog = false;

  $: queryParam = $page.url.searchParams.get('q') || '';
  $: genreParam = $page.url.searchParams.get('genre') || '';

  $: if (queryParam !== searchQuery) {
    searchQuery = queryParam;
    fetchCatalog();
  }

  $: if (genreParam !== selectedGenre) {
    selectedGenre = genreParam;
    fetchCatalog();
  }

  onMount(async () => {
    try {
      const [homeRes, genreRes] = await Promise.all([
        api.getHome().catch(() => null),
        api.getGenres().catch(() => [])
      ]);
      homeData = homeRes;
      genres = genreRes;
      await fetchCatalog();
    } catch (err) {
      console.error('Error loading initial home data:', err);
    } finally {
      loading = false;
    }
  });

  async function fetchCatalog(pageNumber = 1) {
    loadingCatalog = true;
    currentPage = pageNumber;
    try {
      const res = await api.getMangas({
        q: searchQuery,
        genre: selectedGenre,
        status: selectedStatus,
        sort: sortBy,
        page: currentPage,
        limit: 18
      });
      mangaList = res.data;
      pagination = res.pagination;
    } catch (err) {
      console.error('Error fetching catalog:', err);
      mangaList = [];
    } finally {
      loadingCatalog = false;
    }
  }

  function handleGenreClick(genre: string) {
    selectedGenre = selectedGenre === genre ? '' : genre;
    currentPage = 1;
    fetchCatalog(1);
  }

  function handleFilterSubmit(e: Event) {
    e.preventDefault();
    fetchCatalog(1);
  }
</script>

<svelte:head>
  <title>MangaHost - Đọc Manga Tốc Độ Cao</title>
</svelte:head>

<div class="space-y-12 pb-16">
  {#if loading}
    <div class="min-h-[50vh] flex flex-col items-center justify-center gap-3 text-slate-400">
      <Loader2 class="w-8 h-8 text-rose-500 animate-spin" />
      <span class="text-sm font-medium">Đang tải dữ liệu truyện...</span>
    </div>
  {:else}
    <!-- HERO / FEATURED SECTION (Only show if no search/filter active) -->
    {#if !searchQuery && !selectedGenre && homeData?.featured && homeData.featured.length > 0}
      <section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-6">
        <div class="relative rounded-3xl overflow-hidden glass p-6 sm:p-10 border border-white/10 shadow-2xl">
          <!-- Background Glow -->
          <div class="absolute -top-24 -right-24 w-96 h-96 bg-rose-600/20 rounded-full blur-3xl pointer-events-none"></div>
          <div class="absolute -bottom-24 -left-24 w-96 h-96 bg-purple-600/15 rounded-full blur-3xl pointer-events-none"></div>

          <div class="relative z-10 grid grid-cols-1 lg:grid-cols-12 gap-8 items-center">
            <!-- Left Info -->
            <div class="lg:col-span-7 space-y-4">
              <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs font-bold">
                <Sparkles class="w-3.5 h-3.5" /> Nổi Bật Hôm Nay
              </div>
              <h1 class="text-3xl sm:text-4xl lg:text-5xl font-black text-white tracking-tight leading-tight">
                {homeData.featured[0].title}
              </h1>
              <p class="text-slate-300 text-sm sm:text-base line-clamp-3 leading-relaxed max-w-xl">
                {homeData.featured[0].description || 'Mô tả truyện đang được cập nhật...'}
              </p>

              <!-- Genres -->
              <div class="flex flex-wrap gap-2 pt-1">
                {#each homeData.featured[0].genres || [] as genre}
                  <span class="px-2.5 py-1 rounded-lg text-xs font-semibold bg-slate-800/80 text-slate-300 border border-white/5">
                    {genre}
                  </span>
                {/each}
              </div>

              <!-- CTA Actions -->
              <div class="flex items-center gap-4 pt-3">
                <a
                  href={`/manga/${homeData.featured[0].slug || homeData.featured[0].id}`}
                  class="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-gradient-to-r from-rose-600 to-pink-600 text-white font-bold text-sm shadow-lg shadow-rose-600/30 hover:scale-105 transition-all"
                >
                  <BookOpen class="w-4 h-4" /> Đọc Ngay
                </a>
              </div>
            </div>

            <!-- Right Cover Preview -->
            <div class="lg:col-span-5 flex justify-center lg:justify-end">
              <div class="relative group">
                <div class="absolute -inset-1 bg-gradient-to-r from-rose-600 to-purple-600 rounded-2xl blur-lg opacity-40 group-hover:opacity-75 transition duration-500"></div>
                <img
                  src={homeData.featured[0].coverUrl || '/placeholder.jpg'}
                  alt={homeData.featured[0].title}
                  class="relative w-52 sm:w-64 aspect-[3/4] object-cover rounded-2xl shadow-2xl border border-white/10"
                />
              </div>
            </div>
          </div>
        </div>
      </section>
    {/if}

    <!-- POPULAR CAROUSEL / LATEST (If available) -->
    {#if !searchQuery && !selectedGenre && homeData?.popular && homeData.popular.length > 0}
      <section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between mb-5">
          <div class="flex items-center gap-2">
            <Flame class="w-5 h-5 text-rose-500" />
            <h2 class="text-xl font-bold text-white tracking-tight">Truyện Hot Được Xem Nhiều</h2>
          </div>
        </div>

        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {#each homeData.popular.slice(0, 6) as manga}
            <MangaCard {manga} />
          {/each}
        </div>
      </section>
    {/if}

    <!-- MAIN CATALOG / BROWSE SECTION -->
    <section id="browse" class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-6">
      <!-- Section Header with Filter Controls -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-5">
        <div class="flex items-center gap-2">
          <TrendingUp class="w-5 h-5 text-rose-500" />
          <h2 class="text-xl font-bold text-white tracking-tight">
            {#if searchQuery}
              Kết quả tìm kiếm cho: <span class="text-rose-400 font-semibold">"{searchQuery}"</span>
            {:else if selectedGenre}
              Thể loại: <span class="text-rose-400 font-semibold">{selectedGenre}</span>
            {:else}
              Tất Cả Manga
            {/if}
          </h2>
        </div>

        <!-- Filter bar -->
        <div class="flex items-center flex-wrap gap-2.5">
          <!-- Sort Dropdown -->
          <select
            bind:value={sortBy}
            on:change={() => fetchCatalog(1)}
            class="bg-slate-900 border border-slate-800 text-slate-300 text-xs rounded-xl px-3 py-2 focus:outline-none focus:border-rose-500 transition-colors"
          >
            <option value="updatedAt">Mới cập nhật</option>
            <option value="views">Lượt xem nhiều</option>
            <option value="rating">Đánh giá cao</option>
            <option value="title">Tên A-Z</option>
            <option value="createdAt">Mới đăng</option>
          </select>

          <!-- Status Dropdown -->
          <select
            bind:value={selectedStatus}
            on:change={() => fetchCatalog(1)}
            class="bg-slate-900 border border-slate-800 text-slate-300 text-xs rounded-xl px-3 py-2 focus:outline-none focus:border-rose-500 transition-colors"
          >
            <option value="">Mọi trạng thái</option>
            <option value="Ongoing">Đang tiến hành</option>
            <option value="Completed">Đã hoàn thành</option>
          </select>
        </div>
      </div>

      <!-- Genre Pills -->
      {#if genres.length > 0}
        <div class="flex items-center gap-2 overflow-x-auto pb-2 scrollbar-none">
          <button
            on:click={() => handleGenreClick('')}
            class="px-3.5 py-1.5 rounded-full text-xs font-semibold whitespace-nowrap transition-all {selectedGenre === '' ? 'bg-rose-600 text-white shadow-md shadow-rose-600/20' : 'bg-slate-900 text-slate-400 hover:text-slate-200 border border-slate-800'}"
          >
            Tất cả
          </button>
          {#each genres as g}
            <button
              on:click={() => handleGenreClick(g)}
              class="px-3.5 py-1.5 rounded-full text-xs font-semibold whitespace-nowrap transition-all {selectedGenre === g ? 'bg-rose-600 text-white shadow-md shadow-rose-600/20' : 'bg-slate-900 text-slate-400 hover:text-slate-200 border border-slate-800'}"
            >
              {g}
            </button>
          {/each}
        </div>
      {/if}

      <!-- Manga Grid -->
      {#if loadingCatalog}
        <div class="py-20 flex flex-col items-center justify-center gap-3 text-slate-400">
          <Loader2 class="w-8 h-8 text-rose-500 animate-spin" />
          <span class="text-xs">Đang tải danh sách...</span>
        </div>
      {:else if mangaList.length === 0}
        <div class="py-20 text-center glass rounded-2xl p-8 border border-white/5 space-y-3">
          <BookOpen class="w-12 h-12 text-slate-600 mx-auto" />
          <h3 class="text-base font-bold text-slate-300">Chưa có truyện nào</h3>
          <p class="text-xs text-slate-500 max-w-sm mx-auto">
            Hiện tại chưa có truyện nào phù hợp với bộ lọc hoặc cơ sở dữ liệu đang trống. Hãy đăng nhập Admin để upload truyện đầu tiên!
          </p>
        </div>
      {:else}
        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4 sm:gap-6">
          {#each mangaList as manga (manga.id)}
            <MangaCard {manga} />
          {/each}
        </div>

        <!-- Pagination -->
        {#if pagination && pagination.totalPages > 1}
          <div class="flex items-center justify-center gap-2 pt-8">
            <button
              disabled={currentPage <= 1}
              on:click={() => fetchCatalog(currentPage - 1)}
              class="px-4 py-2 rounded-xl text-xs font-semibold bg-slate-900 border border-slate-800 text-slate-300 hover:bg-slate-800 disabled:opacity-40 disabled:pointer-events-none transition-colors"
            >
              Trước
            </button>
            <span class="text-xs text-slate-400 px-3">
              Trang <span class="text-white font-bold">{currentPage}</span> / {pagination.totalPages}
            </span>
            <button
              disabled={currentPage >= pagination.totalPages}
              on:click={() => fetchCatalog(currentPage + 1)}
              class="px-4 py-2 rounded-xl text-xs font-semibold bg-slate-900 border border-slate-800 text-slate-300 hover:bg-slate-800 disabled:opacity-40 disabled:pointer-events-none transition-colors"
            >
              Sau
            </button>
          </div>
        {/if}
      {/if}
    </section>
  {/if}
</div>
