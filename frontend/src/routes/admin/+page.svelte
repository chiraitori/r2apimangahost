<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { api } from '$lib/api';
  import { authStore } from '$lib/stores';
  import type { Manga, StatsResponse, ChapterSummary } from '$lib/types';
  import {
    Plus,
    Upload,
    BookOpen,
    Trash2,
    Edit3,
    Eye,
    HardDrive,
    Layers,
    CheckCircle2,
    AlertCircle,
    X,
    FileArchive,
    Image,
    Loader2,
    RefreshCw,
    Search
  } from 'lucide-svelte';

  let stats: StatsResponse | null = null;
  let mangas: Manga[] = [];
  let loading = true;
  let adminUsername = '';
  let searchQuery = '';

  // Modals state
  let showCreateMangaModal = false;
  let showUploadChapterModal = false;
  let showManageChaptersModal = false;
  let selectedMangaForChapters: Manga | null = null;
  let mangaChaptersList: ChapterSummary[] = [];
  let loadingChapters = false;

  // Form State: Create Manga
  let newManga = {
    title: '',
    slug: '',
    altTitles: '',
    author: '',
    artist: '',
    status: 'Ongoing',
    genres: 'Action, Fantasy, Adventure',
    description: '',
    coverUrl: ''
  };
  let coverFile: File | null = null;
  let coverPreview = '';
  let creatingManga = false;

  // Form State: Upload Chapter
  let targetMangaId = '';
  let chapterNumber = '';
  let chapterTitle = '';
  let chapterVolume = '';
  let chapterLanguage = 'vi';
  let archiveFile: File | null = null;
  let imageFiles: FileList | null = null;
  let uploadingChapter = false;
  let uploadProgressMsg = '';
  let uploadSuccessMsg = '';
  let uploadErrorMsg = '';

  onMount(async () => {
    if (!$authStore.isLoggedIn) {
      goto('/admin/login');
      return;
    }
    await loadAdminData();
  });

  async function loadAdminData() {
    loading = true;
    try {
      const me = await api.getMe();
      adminUsername = me.username;
      const [statsRes, mangasRes] = await Promise.all([
        api.getStats().catch(() => null),
        api.getMangas({ limit: 100 })
      ]);
      stats = statsRes;
      mangas = mangasRes.data;
    } catch (err) {
      console.error('Auth or data fetch error:', err);
      authStore.logout();
      goto('/admin/login');
    } finally {
      loading = false;
    }
  }

  // Cover Image Selection
  function handleCoverSelect(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files && input.files[0]) {
      coverFile = input.files[0];
      coverPreview = URL.createObjectURL(coverFile);
    }
  }

  // Create Manga Submit
  async function handleCreateManga(e: Event) {
    e.preventDefault();
    if (!newManga.title.trim()) return;

    creatingManga = true;
    try {
      const formData = new FormData();
      formData.append('title', newManga.title);
      if (newManga.slug) formData.append('slug', newManga.slug);
      if (newManga.altTitles) formData.append('altTitles', newManga.altTitles);
      if (newManga.author) formData.append('author', newManga.author);
      if (newManga.artist) formData.append('artist', newManga.artist);
      if (newManga.status) formData.append('status', newManga.status);
      if (newManga.genres) formData.append('genres', newManga.genres);
      if (newManga.description) formData.append('description', newManga.description);
      if (newManga.coverUrl) formData.append('coverUrl', newManga.coverUrl);
      if (coverFile) formData.append('cover', coverFile);

      await api.createManga(formData);
      showCreateMangaModal = false;
      // Reset form
      newManga = {
        title: '',
        slug: '',
        altTitles: '',
        author: '',
        artist: '',
        status: 'Ongoing',
        genres: 'Action, Fantasy, Adventure',
        description: '',
        coverUrl: ''
      };
      coverFile = null;
      coverPreview = '';
      await loadAdminData();
    } catch (err: any) {
      alert('Lỗi tạo truyện: ' + (err.message || 'Unknown error'));
    } finally {
      creatingManga = false;
    }
  }

  // Open Upload Chapter Modal for a specific manga
  function openUploadChapter(manga?: Manga) {
    uploadErrorMsg = '';
    uploadSuccessMsg = '';
    archiveFile = null;
    imageFiles = null;
    chapterNumber = '';
    chapterTitle = '';
    chapterVolume = '';
    if (manga) {
      targetMangaId = manga.slug || manga.id;
      // Suggest next chapter number
      if (manga.lastChapterNumber !== undefined && manga.lastChapterNumber !== null) {
        chapterNumber = (manga.lastChapterNumber + 1).toString();
      } else {
        chapterNumber = '1';
      }
    } else if (mangas.length > 0) {
      targetMangaId = mangas[0].slug || mangas[0].id;
      chapterNumber = '1';
    }
    showUploadChapterModal = true;
  }

  // Handle Chapter Upload Submit (.zip/images -> Cloudflare R2)
  async function handleUploadChapter(e: Event) {
    e.preventDefault();
    if (!targetMangaId || !chapterNumber) {
      uploadErrorMsg = 'Vui lòng chọn truyện và nhập số chapter';
      return;
    }
    if (!archiveFile && (!imageFiles || imageFiles.length === 0)) {
      uploadErrorMsg = 'Vui lòng chọn file .zip/.cbz hoặc các file ảnh trang truyện';
      return;
    }

    uploadingChapter = true;
    uploadErrorMsg = '';
    uploadSuccessMsg = '';
    uploadProgressMsg = 'Đang tải file lên và đẩy trực tiếp vào Cloudflare R2...';

    try {
      const formData = new FormData();
      formData.append('chapterNumber', chapterNumber);
      if (chapterTitle) formData.append('title', chapterTitle);
      if (chapterVolume) formData.append('volume', chapterVolume);
      formData.append('language', chapterLanguage);

      if (archiveFile) {
        formData.append('archive', archiveFile);
      } else if (imageFiles) {
        for (let i = 0; i < imageFiles.length; i++) {
          formData.append('pages', imageFiles[i]);
        }
      }

      const res = await api.uploadChapter(targetMangaId, formData);
      uploadSuccessMsg = `Tải lên thành công Chapter ${res.chapterNumber} với ${res.pageCount} trang!`;
      archiveFile = null;
      imageFiles = null;
      await loadAdminData();
      if (selectedMangaForChapters) {
        await openManageChapters(selectedMangaForChapters);
      }
    } catch (err: any) {
      uploadErrorMsg = err.message || 'Lỗi khi upload chapter lên R2';
    } finally {
      uploadingChapter = false;
      uploadProgressMsg = '';
    }
  }

  // Delete Manga
  async function handleDeleteManga(manga: Manga) {
    if (confirm(`Bạn có chắc chắn muốn xóa bộ truyện "${manga.title}" cùng toàn bộ chapter và ảnh trên Cloudflare R2?`)) {
      try {
        await api.deleteManga(manga.slug || manga.id);
        await loadAdminData();
      } catch (err: any) {
        alert('Lỗi xóa truyện: ' + err.message);
      }
    }
  }

  // Open Manage Chapters Modal
  async function openManageChapters(manga: Manga) {
    selectedMangaForChapters = manga;
    showManageChaptersModal = true;
    loadingChapters = true;
    try {
      const res = await api.getManga(manga.slug || manga.id);
      mangaChaptersList = res.chapters;
    } catch (err: any) {
      alert('Không thể tải danh sách chapter: ' + err.message);
    } finally {
      loadingChapters = false;
    }
  }

  // Delete Chapter
  async function handleDeleteChapter(chapter: ChapterSummary) {
    if (confirm(`Xóa Chapter ${chapter.chapterNumber} và toàn bộ ảnh của chapter này khỏi Cloudflare R2?`)) {
      try {
        await api.deleteChapter(chapter.id);
        if (selectedMangaForChapters) {
          await openManageChapters(selectedMangaForChapters);
          await loadAdminData();
        }
      } catch (err: any) {
        alert('Lỗi xóa chapter: ' + err.message);
      }
    }
  }

  $: filteredMangas = mangas.filter((m) =>
    searchQuery ? m.title.toLowerCase().includes(searchQuery.toLowerCase()) : true
  );
</script>

<svelte:head>
  <title>Admin Dashboard - MangaHost</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
  <!-- Top Bar -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-6">
    <div>
      <h1 class="text-2xl font-black text-white tracking-tight flex items-center gap-2">
        Admin Dashboard <span class="text-xs px-2.5 py-0.5 rounded-full bg-rose-500/20 text-rose-400 border border-rose-500/30">Admin: {adminUsername}</span>
      </h1>
      <p class="text-xs text-slate-400 mt-1">Quản lý kho truyện và upload ảnh lưu trữ trên Cloudflare R2</p>
    </div>

    <div class="flex items-center gap-3">
      <button
        on:click={() => openUploadChapter()}
        class="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-gradient-to-r from-rose-600 to-pink-600 text-white font-bold text-xs shadow-lg shadow-rose-600/20 hover:opacity-95 transition-all"
      >
        <Upload class="w-4 h-4" /> Upload Chapter (.zip/ảnh)
      </button>

      <button
        on:click={() => showCreateMangaModal = true}
        class="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-white font-bold text-xs border border-white/10 transition-colors"
      >
        <Plus class="w-4 h-4" /> Thêm Manga Mới
      </button>
    </div>
  </div>

  <!-- STATS OVERVIEW CARDS -->
  <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
    <div class="glass p-5 rounded-2xl border border-white/5 space-y-1">
      <div class="flex items-center justify-between text-slate-400">
        <span class="text-xs font-bold uppercase tracking-wider">Tổng Manga</span>
        <BookOpen class="w-4 h-4 text-rose-500" />
      </div>
      <p class="text-2xl font-black text-white">{stats?.totalMangas ?? mangas.length}</p>
      <span class="text-[11px] text-slate-500">Đã lưu trữ trong MongoDB</span>
    </div>

    <div class="glass p-5 rounded-2xl border border-white/5 space-y-1">
      <div class="flex items-center justify-between text-slate-400">
        <span class="text-xs font-bold uppercase tracking-wider">Tổng Chapter</span>
        <Layers class="w-4 h-4 text-pink-500" />
      </div>
      <p class="text-2xl font-black text-white">{stats?.totalChapters ?? 0}</p>
      <span class="text-[11px] text-slate-500">Lưu ảnh trên Cloudflare R2 CDN</span>
    </div>

    <div class="glass p-5 rounded-2xl border border-white/5 space-y-1">
      <div class="flex items-center justify-between text-slate-400">
        <span class="text-xs font-bold uppercase tracking-wider">Tổng Lượt Xem</span>
        <Eye class="w-4 h-4 text-amber-400" />
      </div>
      <p class="text-2xl font-black text-white">{(stats?.totalViews ?? 0).toLocaleString()}</p>
      <span class="text-[11px] text-slate-500">Lượt đọc truyện trên hệ thống</span>
    </div>
  </div>

  <!-- MANGA LIST TABLE / GRID -->
  <div class="space-y-4">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
      <h2 class="text-lg font-bold text-white tracking-tight">Danh Sách Truyện Đang Quản Lý ({mangas.length})</h2>
      <div class="relative w-full sm:w-64">
        <Search class="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Lọc truyện..."
          class="w-full pl-8 pr-3 py-1.5 bg-slate-900 border border-slate-800 rounded-xl text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-rose-500"
        />
      </div>
    </div>

    {#if loading}
      <div class="py-16 flex flex-col items-center justify-center gap-3 text-slate-400">
        <Loader2 class="w-8 h-8 text-rose-500 animate-spin" />
        <span class="text-xs">Đang tải danh sách...</span>
      </div>
    {:else if filteredMangas.length === 0}
      <div class="py-16 text-center glass rounded-2xl p-8 border border-white/5 space-y-3">
        <BookOpen class="w-10 h-10 text-slate-600 mx-auto" />
        <p class="text-xs text-slate-400">Chưa có truyện nào. Bấm nút "Thêm Manga Mới" ở trên để tạo!</p>
      </div>
    {:else}
      <div class="glass rounded-2xl border border-white/5 overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs text-slate-300">
            <thead class="bg-slate-900/80 uppercase text-[10px] text-slate-400 font-bold tracking-wider border-b border-slate-800">
              <tr>
                <th class="px-4 py-3.5">Ảnh bìa</th>
                <th class="px-4 py-3.5">Tên Manga / Slug</th>
                <th class="px-4 py-3.5">Trạng thái</th>
                <th class="px-4 py-3.5">Chapter mới nhất</th>
                <th class="px-4 py-3.5">Lượt xem</th>
                <th class="px-4 py-3.5 text-right">Thao tác</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60">
              {#each filteredMangas as manga}
                <tr class="hover:bg-slate-800/30 transition-colors">
                  <!-- Cover -->
                  <td class="px-4 py-3">
                    <img
                      src={manga.coverUrl || '/placeholder.jpg'}
                      alt={manga.title}
                      class="w-12 h-16 object-cover rounded-lg bg-slate-900 border border-white/5 shadow"
                    />
                  </td>

                  <!-- Title & Slug -->
                  <td class="px-4 py-3">
                    <a
                      href={`/manga/${manga.slug || manga.id}`}
                      target="_blank"
                      class="font-bold text-white hover:text-rose-400 transition-colors block text-sm"
                    >
                      {manga.title}
                    </a>
                    <span class="text-[10px] text-slate-500 font-mono">slug: {manga.slug}</span>
                  </td>

                  <!-- Status -->
                  <td class="px-4 py-3">
                    <span class="px-2 py-0.5 rounded text-[10px] font-bold {manga.status === 'Completed' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-rose-500/20 text-rose-400'}">
                      {manga.status}
                    </span>
                  </td>

                  <!-- Chapter count -->
                  <td class="px-4 py-3">
                    {#if manga.lastChapterNumber !== undefined && manga.lastChapterNumber !== null}
                      <span class="font-bold text-rose-400">Ch. {manga.lastChapterNumber}</span>
                      <span class="text-slate-500 text-[10px]">({manga.chapterCount || 0} chương)</span>
                    {:else}
                      <span class="text-slate-500 text-[10px]">Chưa có</span>
                    {/if}
                  </td>

                  <!-- Views -->
                  <td class="px-4 py-3 font-medium">
                    {manga.views?.toLocaleString() || 0}
                  </td>

                  <!-- Actions -->
                  <td class="px-4 py-3 text-right">
                    <div class="flex items-center justify-end gap-1.5">
                      <button
                        on:click={() => openUploadChapter(manga)}
                        class="p-1.5 rounded-lg bg-rose-600/20 text-rose-400 hover:bg-rose-600 hover:text-white transition-colors"
                        title="Upload Chapter Mới"
                      >
                        <Upload class="w-3.5 h-3.5" />
                      </button>

                      <button
                        on:click={() => openManageChapters(manga)}
                        class="p-1.5 rounded-lg bg-slate-800 text-slate-300 hover:bg-slate-700 hover:text-white transition-colors"
                        title="Quản lý danh sách chapter"
                      >
                        <Layers class="w-3.5 h-3.5" />
                      </button>

                      <button
                        on:click={() => handleDeleteManga(manga)}
                        class="p-1.5 rounded-lg bg-red-500/10 text-red-400 hover:bg-red-600 hover:text-white transition-colors"
                        title="Xóa truyện"
                      >
                        <Trash2 class="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    {/if}
  </div>
</div>

<!-- MODAL: CREATE MANGA -->
{#if showCreateMangaModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm overflow-y-auto">
    <div class="w-full max-w-2xl glass p-6 sm:p-8 rounded-3xl border border-white/10 space-y-6 max-h-[90vh] overflow-y-auto">
      <div class="flex items-center justify-between border-b border-slate-800 pb-4">
        <h3 class="font-bold text-lg text-white">Thêm Manga Mới</h3>
        <button on:click={() => showCreateMangaModal = false} class="text-slate-400 hover:text-white">
          <X class="w-5 h-5" />
        </button>
      </div>

      <form on:submit={handleCreateManga} class="space-y-4">
        <div>
          <label class="block text-xs font-bold text-slate-300 mb-1">Tên Manga *</label>
          <input
            type="text"
            bind:value={newManga.title}
            placeholder="Ví dụ: Solo Leveling"
            required
            class="w-full px-3.5 py-2 bg-slate-900 border border-slate-800 rounded-xl text-xs text-white focus:outline-none focus:border-rose-500"
          />
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-xs font-bold text-slate-300 mb-1">Tác giả</label>
            <input
              type="text"
              bind:value={newManga.author}
              placeholder="Chugong"
              class="w-full px-3.5 py-2 bg-slate-900 border border-slate-800 rounded-xl text-xs text-white focus:outline-none focus:border-rose-500"
            />
          </div>
          <div>
            <label class="block text-xs font-bold text-slate-300 mb-1">Họa sĩ</label>
            <input
              type="text"
              bind:value={newManga.artist}
              placeholder="DUBU (REDICE Studio)"
              class="w-full px-3.5 py-2 bg-slate-900 border border-slate-800 rounded-xl text-xs text-white focus:outline-none focus:border-rose-500"
            />
          </div>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-xs font-bold text-slate-300 mb-1">Trạng thái</label>
            <select
              bind:value={newManga.status}
              class="w-full px-3.5 py-2 bg-slate-900 border border-slate-800 rounded-xl text-xs text-white focus:outline-none focus:border-rose-500"
            >
              <option value="Ongoing">Đang tiến hành (Ongoing)</option>
              <option value="Completed">Đã hoàn thành (Completed)</option>
              <option value="Hiatus">Tạm ngưng (Hiatus)</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-bold text-slate-300 mb-1">Thể loại (ngăn cách bằng dấu phẩy)</label>
            <input
              type="text"
              bind:value={newManga.genres}
              placeholder="Action, Fantasy, Adventure"
              class="w-full px-3.5 py-2 bg-slate-900 border border-slate-800 rounded-xl text-xs text-white focus:outline-none focus:border-rose-500"
            />
          </div>
        </div>

        <div>
          <label class="block text-xs font-bold text-slate-300 mb-1">Mô tả / Tóm tắt truyện</label>
          <textarea
            bind:value={newManga.description}
            rows="3"
            placeholder="Nội dung tóm tắt cốt truyện..."
            class="w-full px-3.5 py-2 bg-slate-900 border border-slate-800 rounded-xl text-xs text-white focus:outline-none focus:border-rose-500"
          ></textarea>
        </div>

        <!-- Cover Upload -->
        <div>
          <label class="block text-xs font-bold text-slate-300 mb-1">Ảnh bìa (Upload lên Cloudflare R2 hoặc nhập URL)</label>
          <div class="flex items-center gap-4">
            <input
              type="file"
              accept="image/*"
              on:change={handleCoverSelect}
              class="text-xs text-slate-400 file:mr-3 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-xs file:font-semibold file:bg-slate-800 file:text-rose-400 hover:file:bg-slate-700"
            />
            {#if coverPreview}
              <img src={coverPreview} alt="Preview" class="w-12 h-16 object-cover rounded-lg border border-rose-500/50" />
            {/if}
          </div>
        </div>

        <div class="flex items-center justify-end gap-3 pt-4 border-t border-slate-800">
          <button
            type="button"
            on:click={() => showCreateMangaModal = false}
            class="px-4 py-2 rounded-xl text-xs font-semibold bg-slate-800 text-slate-300 hover:bg-slate-700"
          >
            Hủy
          </button>
          <button
            type="submit"
            disabled={creatingManga}
            class="px-5 py-2 rounded-xl text-xs font-bold bg-rose-600 hover:bg-rose-500 text-white flex items-center gap-2 shadow-lg shadow-rose-600/30 disabled:opacity-50"
          >
            {#if creatingManga}
              <Loader2 class="w-3.5 h-3.5 animate-spin" /> Đang tạo...
            {:else}
              Tạo Manga
            {/if}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- MODAL: UPLOAD CHAPTER (.ZIP/IMAGES TO CLOUDFLARE R2) -->
{#if showUploadChapterModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm overflow-y-auto">
    <div class="w-full max-w-xl glass p-6 sm:p-8 rounded-3xl border border-white/10 space-y-6 max-h-[90vh] overflow-y-auto">
      <div class="flex items-center justify-between border-b border-slate-800 pb-4">
        <div>
          <h3 class="font-bold text-lg text-white">Upload Chapter Mới</h3>
          <p class="text-xs text-slate-400">Tự động giải nén và stream toàn bộ ảnh lên Cloudflare R2</p>
        </div>
        <button on:click={() => showUploadChapterModal = false} class="text-slate-400 hover:text-white">
          <X class="w-5 h-5" />
        </button>
      </div>

      {#if uploadErrorMsg}
        <div class="p-3.5 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-400 text-xs font-semibold flex items-center gap-2">
          <AlertCircle class="w-4 h-4 shrink-0" /> {uploadErrorMsg}
        </div>
      {/if}

      {#if uploadSuccessMsg}
        <div class="p-3.5 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-semibold flex items-center gap-2">
          <CheckCircle2 class="w-4 h-4 shrink-0" /> {uploadSuccessMsg}
        </div>
      {/if}

      <form on:submit={handleUploadChapter} class="space-y-4">
        <!-- Target Manga Selection -->
        <div>
          <label class="block text-xs font-bold text-slate-300 mb-1">Chọn Manga *</label>
          <select
            bind:value={targetMangaId}
            required
            class="w-full px-3.5 py-2.5 bg-slate-900 border border-slate-800 rounded-xl text-xs text-white focus:outline-none focus:border-rose-500"
          >
            {#each mangas as m}
              <option value={m.slug || m.id}>{m.title} (slug: {m.slug})</option>
            {/each}
          </select>
        </div>

        <!-- Chapter Number & Title -->
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-xs font-bold text-slate-300 mb-1">Số Chapter (Number) *</label>
            <input
              type="number"
              step="any"
              bind:value={chapterNumber}
              placeholder="1 hoặc 10.5"
              required
              class="w-full px-3.5 py-2 bg-slate-900 border border-slate-800 rounded-xl text-xs text-white focus:outline-none focus:border-rose-500"
            />
          </div>
          <div>
            <label class="block text-xs font-bold text-slate-300 mb-1">Tiêu đề Chapter (tùy chọn)</label>
            <input
              type="text"
              bind:value={chapterTitle}
              placeholder="Khởi đầu mới..."
              class="w-full px-3.5 py-2 bg-slate-900 border border-slate-800 rounded-xl text-xs text-white focus:outline-none focus:border-rose-500"
            />
          </div>
        </div>

        <!-- FILE UPLOAD METHOD (ZIP OR MULTIPLE IMAGES) -->
        <div class="space-y-3 pt-2">
          <label class="block text-xs font-bold text-slate-300">File trang truyện (Chọn 1 trong 2 cách):</label>

          <!-- Option A: ZIP Archive -->
          <div class="p-4 rounded-2xl bg-slate-900/90 border-2 border-dashed border-slate-800 hover:border-rose-500/50 transition-colors">
            <div class="flex items-center gap-3">
              <FileArchive class="w-8 h-8 text-rose-500 shrink-0" />
              <div class="flex-1">
                <span class="block text-xs font-bold text-slate-200">Cách 1: File nén .zip hoặc .cbz (Khuyên dùng)</span>
                <span class="block text-[11px] text-slate-500">Hệ thống Go sẽ tự động giải nén và sắp xếp trang tự nhiên (1, 2, ... 10)</span>
                <input
                  type="file"
                  accept=".zip,.cbz"
                  on:change={(e) => {
                    const files = (e.target as HTMLInputElement).files;
                    if (files && files[0]) {
                      archiveFile = files[0];
                      imageFiles = null;
                    }
                  }}
                  class="mt-2 text-xs text-slate-400 file:mr-2 file:py-1 file:px-3 file:rounded-lg file:border-0 file:text-xs file:bg-slate-800 file:text-rose-400"
                />
              </div>
            </div>
          </div>

          <!-- Option B: Multiple Images -->
          <div class="p-4 rounded-2xl bg-slate-900/90 border-2 border-dashed border-slate-800 hover:border-rose-500/50 transition-colors">
            <div class="flex items-center gap-3">
              <Image class="w-8 h-8 text-pink-500 shrink-0" />
              <div class="flex-1">
                <span class="block text-xs font-bold text-slate-200">Cách 2: Chọn nhiều file ảnh cùng lúc</span>
                <span class="block text-[11px] text-slate-500">Giữ Ctrl / Shift để chọn nhiều ảnh (jpg, png, webp, ...)</span>
                <input
                  type="file"
                  accept="image/*"
                  multiple
                  on:change={(e) => {
                    const files = (e.target as HTMLInputElement).files;
                    if (files && files.length > 0) {
                      imageFiles = files;
                      archiveFile = null;
                    }
                  }}
                  class="mt-2 text-xs text-slate-400 file:mr-2 file:py-1 file:px-3 file:rounded-lg file:border-0 file:text-xs file:bg-slate-800 file:text-pink-400"
                />
              </div>
            </div>
          </div>
        </div>

        {#if uploadingChapter}
          <div class="p-4 rounded-2xl bg-rose-950/40 border border-rose-500/30 text-rose-300 text-xs flex items-center gap-3 animate-pulse">
            <Loader2 class="w-5 h-5 animate-spin text-rose-500 shrink-0" />
            <div>
              <p class="font-bold">{uploadProgressMsg}</p>
              <p class="text-[10px] text-rose-400/80">Vui lòng không đóng trình duyệt cho đến khi hoàn tất.</p>
            </div>
          </div>
        {/if}

        <div class="flex items-center justify-end gap-3 pt-4 border-t border-slate-800">
          <button
            type="button"
            on:click={() => showUploadChapterModal = false}
            class="px-4 py-2 rounded-xl text-xs font-semibold bg-slate-800 text-slate-300 hover:bg-slate-700"
          >
            Đóng
          </button>
          <button
            type="submit"
            disabled={uploadingChapter}
            class="px-6 py-2.5 rounded-xl text-xs font-bold bg-gradient-to-r from-rose-600 to-pink-600 text-white flex items-center gap-2 shadow-lg shadow-rose-600/30 disabled:opacity-50"
          >
            {#if uploadingChapter}
              <Loader2 class="w-4 h-4 animate-spin" /> Đang tải lên R2...
            {:else}
              <Upload class="w-4 h-4" /> Bắt đầu tải lên
            {/if}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- MODAL: MANAGE CHAPTERS OF A SPECIFIC MANGA -->
{#if showManageChaptersModal && selectedMangaForChapters}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm overflow-y-auto">
    <div class="w-full max-w-2xl glass p-6 sm:p-8 rounded-3xl border border-white/10 space-y-6 max-h-[90vh] overflow-y-auto">
      <div class="flex items-center justify-between border-b border-slate-800 pb-4">
        <div>
          <h3 class="font-bold text-lg text-white">Quản lý Chapters: {selectedMangaForChapters.title}</h3>
          <p class="text-xs text-slate-400">Danh sách các chapter đã lưu trong database và Cloudflare R2</p>
        </div>
        <button on:click={() => showManageChaptersModal = false} class="text-slate-400 hover:text-white">
          <X class="w-5 h-5" />
        </button>
      </div>

      {#if loadingChapters}
        <div class="py-12 flex flex-col items-center justify-center gap-2 text-slate-400">
          <Loader2 class="w-6 h-6 animate-spin text-rose-500" />
          <span class="text-xs">Đang tải danh sách chapter...</span>
        </div>
      {:else if mangaChaptersList.length === 0}
        <div class="py-12 text-center text-slate-500 text-xs">
          Truyện này chưa có chapter nào.
        </div>
      {:else}
        <div class="divide-y divide-slate-800/60 max-h-96 overflow-y-auto pr-1">
          {#each mangaChaptersList as chap}
            <div class="py-3 flex items-center justify-between hover:bg-slate-800/20 px-2 rounded-lg transition-colors">
              <div>
                <span class="font-bold text-xs text-white">Chapter {chap.chapterNumber}</span>
                {#if chap.title}
                  <span class="text-slate-400 text-xs ml-1">- {chap.title}</span>
                {/if}
                <span class="block text-[10px] text-slate-500">{chap.pageCount} trang</span>
              </div>

              <div class="flex items-center gap-2">
                <a
                  href={`/read/${chap.id}`}
                  target="_blank"
                  class="px-2.5 py-1 text-xs rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 flex items-center gap-1"
                >
                  <Eye class="w-3.5 h-3.5" /> Xem
                </a>
                <button
                  on:click={() => handleDeleteChapter(chap)}
                  class="p-1.5 rounded-lg bg-red-500/10 hover:bg-red-600 text-red-400 hover:text-white transition-colors"
                  title="Xóa chapter này"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}

      <div class="flex items-center justify-between pt-4 border-t border-slate-800">
        <button
          on:click={() => openUploadChapter(selectedMangaForChapters ?? undefined)}
          class="px-4 py-2 rounded-xl text-xs font-bold bg-rose-600 hover:bg-rose-500 text-white flex items-center gap-2"
        >
          <Upload class="w-3.5 h-3.5" /> Upload Thêm Chapter
        </button>

        <button
          on:click={() => showManageChaptersModal = false}
          class="px-4 py-2 rounded-xl text-xs font-semibold bg-slate-800 text-slate-300 hover:bg-slate-700"
        >
          Đóng
        </button>
      </div>
    </div>
  </div>
{/if}
