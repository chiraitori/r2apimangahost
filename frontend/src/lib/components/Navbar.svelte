<script lang="ts">
  import { authStore } from '$lib/stores';
  import { Search, Shield, LogOut, Smartphone, Sparkles, Copy, Check, ExternalLink, X } from 'lucide-svelte';
  import { goto } from '$app/navigation';

  let searchQuery = '';
  let showPaperbackModal = false;
  let copied = false;

  const paperbackRepoURL = 'https://chiraitori.github.io/paperback-mangahost-extension/versioning.json';

  function handleSearch(e: Event) {
    e.preventDefault();
    if (searchQuery.trim()) {
      goto(`/?q=${encodeURIComponent(searchQuery.trim())}`);
    }
  }

  function handleLogout() {
    authStore.logout();
    goto('/');
  }

  async function copyRepoURL() {
    try {
      await navigator.clipboard.writeText(paperbackRepoURL);
      copied = true;
      setTimeout(() => (copied = false), 2500);
    } catch (err) {
      console.error('Failed to copy', err);
    }
  }
</script>

<header class="sticky top-0 z-40 w-full glass border-b border-white/5">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between gap-4">
    <!-- Brand Logo -->
    <a href="/" class="flex items-center gap-2.5 group shrink-0">
      <img src="/icon.png" alt="MangaHost" class="w-9 h-9 rounded-xl object-cover shadow-lg shadow-rose-500/20 group-hover:scale-105 transition-transform duration-200 border border-white/10" />
      <div class="flex flex-col">
        <span class="font-extrabold text-lg tracking-tight bg-gradient-to-r from-white via-slate-200 to-rose-400 bg-clip-text text-transparent">
          Manga<span class="text-rose-500">Host</span>
        </span>
        <span class="text-[10px] font-medium text-slate-400 -mt-1 tracking-wider uppercase flex items-center gap-1">
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span> Cloudflare R2
        </span>
      </div>
    </a>

    <!-- Search Bar -->
    <form on:submit={handleSearch} class="flex-1 max-w-md hidden md:block">
      <div class="relative">
        <Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2 pointer-events-none" />
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Tìm tên truyện, tác giả..."
          class="w-full pl-10 pr-4 py-2 bg-slate-900/80 border border-slate-800 rounded-xl text-sm text-slate-200 placeholder-slate-500 focus:outline-none focus:border-rose-500/60 focus:ring-2 focus:ring-rose-500/10 transition-all"
        />
      </div>
    </form>

    <!-- Navigation actions -->
    <div class="flex items-center gap-3">
      <a
        href="/#browse"
        class="hidden sm:inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800/60 rounded-lg transition-colors"
      >
        <Sparkles class="w-3.5 h-3.5 text-amber-400" />
        Khám phá
      </a>

      <!-- Paperback info button -->
      <button
        type="button"
        on:click={() => (showPaperbackModal = true)}
        class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800/60 rounded-lg transition-colors border border-slate-700/50 hover:border-slate-600"
        title="Thêm Extension vào Paperback iOS"
      >
        <Smartphone class="w-3.5 h-3.5 text-blue-400" />
        Paperback iOS
      </button>

      {#if $authStore.isLoggedIn}
        <a
          href="/admin"
          class="inline-flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-bold text-white bg-rose-600 hover:bg-rose-500 shadow-md shadow-rose-600/20 rounded-xl transition-all"
        >
          <Shield class="w-3.5 h-3.5" />
          Admin Panel
        </a>
        <button
          on:click={handleLogout}
          class="p-2 text-slate-400 hover:text-rose-400 hover:bg-slate-800/60 rounded-xl transition-colors"
          title="Đăng xuất"
        >
          <LogOut class="w-4 h-4" />
        </button>
      {:else}
        <a
          href="/admin/login"
          class="inline-flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 rounded-xl transition-colors"
        >
          <Shield class="w-3.5 h-3.5 text-slate-400" />
          Admin
        </a>
      {/if}
    </div>
  </div>
</header>

<!-- Paperback Modal -->
{#if showPaperbackModal}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200"
    role="dialog"
    aria-modal="true"
  >
    <div class="relative w-full max-w-lg bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl p-6 sm:p-8 space-y-6">
      <!-- Close button -->
      <button
        on:click={() => (showPaperbackModal = false)}
        class="absolute top-4 right-4 p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-xl transition-colors"
        aria-label="Đóng"
      >
        <X class="w-5 h-5" />
      </button>

      <!-- Modal Header -->
      <div class="flex items-center gap-3.5">
        <div class="w-12 h-12 rounded-2xl bg-gradient-to-tr from-blue-600 to-indigo-500 flex items-center justify-center shadow-lg shadow-blue-500/20 shrink-0">
          <Smartphone class="w-6 h-6 text-white" />
        </div>
        <div>
          <h3 class="text-lg font-bold text-white">Cài Đặt Paperback (iOS)</h3>
          <p class="text-xs text-slate-400">Đọc toàn bộ kho truyện trên iPhone & iPad qua app Paperback</p>
        </div>
      </div>

      <!-- Repository URL Copy Box -->
      <div class="space-y-2">
        <label for="paperback-url" class="block text-xs font-bold text-slate-300 uppercase tracking-wider">
          External Repository URL:
        </label>
        <div class="flex items-center gap-2 p-1.5 bg-slate-950 border border-slate-800 rounded-xl">
          <input
            id="paperback-url"
            type="text"
            readonly
            value={paperbackRepoURL}
            class="flex-1 bg-transparent px-3 text-xs text-blue-300 font-mono focus:outline-none select-all"
          />
          <button
            on:click={copyRepoURL}
            class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-bold rounded-lg transition-all {copied
              ? 'bg-emerald-600 text-white'
              : 'bg-blue-600 hover:bg-blue-500 text-white shadow-md shadow-blue-600/20'}"
          >
            {#if copied}
              <Check class="w-3.5 h-3.5" />
              Đã sao chép!
            {:else}
              <Copy class="w-3.5 h-3.5" />
              Copy URL
            {/if}
          </button>
        </div>
      </div>

      <!-- 3-Step Guide -->
      <div class="p-4 bg-slate-950/60 border border-slate-800/80 rounded-xl space-y-2.5">
        <h4 class="text-xs font-bold text-slate-300 uppercase tracking-wider">3 Bước Thêm Vào App:</h4>
        <ol class="text-xs text-slate-400 space-y-2 list-decimal list-inside leading-relaxed">
          <li>Mở app <strong class="text-white">Paperback</strong> trên iPhone / iPad.</li>
          <li>Vào <strong class="text-white">Settings</strong> ➔ <strong class="text-white">External Repositories</strong> ➔ bấm <strong class="text-white">Add</strong> (+).</li>
          <li>Dán URL ở trên vào và bấm <strong class="text-white">Save</strong>. Extension <strong>MangaHost R2</strong> sẽ xuất hiện!</li>
        </ol>
      </div>

      <!-- Action buttons -->
      <div class="flex flex-col sm:flex-row items-center justify-between gap-3 pt-2">
        <a
          href="paperback://addRepo?url={encodeURIComponent(paperbackRepoURL)}"
          class="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-4 py-2.5 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white font-bold text-xs rounded-xl shadow-lg shadow-blue-500/25 transition-all"
        >
          <Smartphone class="w-4 h-4" />
          1-Click Thêm Vào Paperback
        </a>

        <a
          href="https://github.com/chiraitori/paperback-mangahost-extension"
          target="_blank"
          rel="noopener noreferrer"
          class="w-full sm:w-auto inline-flex items-center justify-center gap-1.5 px-3 py-2 text-xs font-medium text-slate-400 hover:text-white transition-colors"
        >
          <ExternalLink class="w-3.5 h-3.5" />
          Xem GitHub Repo
        </a>
      </div>
    </div>
  </div>
{/if}
