<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { api } from '$lib/api';
  import { authStore } from '$lib/stores';
  import { Shield, Lock, User, Eye, EyeOff, Loader2, BookOpen } from 'lucide-svelte';

  let username = '';
  let password = '';
  let showPassword = false;
  let loading = false;
  let errorMsg = '';

  onMount(async () => {
    if ($authStore.isLoggedIn) {
      goto('/admin');
    }
  });

  async function handleLogin(e: Event) {
    e.preventDefault();
    if (!username || !password) {
      errorMsg = 'Vui lòng nhập đầy đủ tên đăng nhập và mật khẩu';
      return;
    }

    loading = true;
    errorMsg = '';

    try {
      const res = await api.login(username, password);
      authStore.login(res.token, res.username);
      goto('/admin');
    } catch (err: any) {
      errorMsg = err.message || 'Đăng nhập thất bại. Vui lòng kiểm tra lại thông tin.';
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Admin Login - MangaHost</title>
</svelte:head>

<div class="min-h-[80vh] flex items-center justify-center px-4">
  <div class="w-full max-w-md glass p-8 rounded-3xl border border-white/10 shadow-2xl space-y-6 relative overflow-hidden">
    <!-- Glow -->
    <div class="absolute -top-12 -right-12 w-48 h-48 bg-rose-600/20 rounded-full blur-2xl pointer-events-none"></div>

    <div class="text-center space-y-2 relative z-10">
      <div class="w-12 h-12 rounded-2xl bg-rose-600 mx-auto flex items-center justify-center shadow-lg shadow-rose-600/30 text-white">
        <Shield class="w-6 h-6" />
      </div>
      <h1 class="text-2xl font-bold text-white tracking-tight">Admin MangaHost</h1>
      <p class="text-xs text-slate-400">Đăng nhập để quản lý truyện và upload chapter lên Cloudflare R2</p>
    </div>

    {#if errorMsg}
      <div class="p-3.5 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-400 text-xs font-semibold">
        {errorMsg}
      </div>
    {/if}

    <form on:submit={handleLogin} class="space-y-4 relative z-10">
      <div>
        <label for="admin-user-input" class="block text-xs font-bold text-slate-300 uppercase tracking-wider mb-1.5">Tên đăng nhập</label>
        <div class="relative">
          <User class="w-4 h-4 text-slate-500 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            id="admin-user-input"
            type="text"
            bind:value={username}
            placeholder="admin"
            required
            class="w-full pl-10 pr-4 py-2.5 bg-slate-900/90 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-rose-500 transition-colors"
          />
        </div>
      </div>

      <div>
        <label for="admin-pass-input" class="block text-xs font-bold text-slate-300 uppercase tracking-wider mb-1.5">Mật khẩu</label>
        <div class="relative">
          <Lock class="w-4 h-4 text-slate-500 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            id="admin-pass-input"
            type={showPassword ? 'text' : 'password'}
            bind:value={password}
            placeholder="••••••••"
            required
            class="w-full pl-10 pr-10 py-2.5 bg-slate-900/90 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-rose-500 transition-colors"
          />
          <button
            type="button"
            on:click={() => showPassword = !showPassword}
            class="absolute right-3.5 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
          >
            {#if showPassword}
              <EyeOff class="w-4 h-4" />
            {:else}
              <Eye class="w-4 h-4" />
            {/if}
          </button>
        </div>
      </div>

      <button
        type="submit"
        disabled={loading}
        class="w-full py-3 rounded-xl bg-gradient-to-r from-rose-600 to-pink-600 text-white font-bold text-sm shadow-lg shadow-rose-600/30 hover:opacity-95 disabled:opacity-50 flex items-center justify-center gap-2 transition-all"
      >
        {#if loading}
          <Loader2 class="w-4 h-4 animate-spin" /> Đang đăng nhập...
        {:else}
          Đăng Nhập
        {/if}
      </button>
    </form>
  </div>
</div>
