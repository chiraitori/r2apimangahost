<script lang="ts">
  import { BookOpen, Zap, HardDrive, ShieldCheck } from 'lucide-svelte';
</script>

<footer class="mt-20 border-t border-slate-800/80 bg-slate-950/70 text-slate-400">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <div class="grid grid-cols-1 md:grid-cols-4 gap-8 mb-10">
      <!-- Brand column -->
      <div class="md:col-span-2 space-y-3">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-lg bg-rose-600 flex items-center justify-center text-white">
            <BookOpen class="w-4 h-4" />
          </div>
          <span class="font-bold text-lg text-white">Manga<span class="text-rose-500">Host</span></span>
        </div>
        <p class="text-xs text-slate-400 max-w-sm leading-relaxed">
          Nền tảng manga hosting hiệu năng cao với backend Go siêu nhẹ (tiết kiệm RAM tối đa), lưu trữ ảnh qua Cloudflare R2 CDN không giới hạn băng thông, frontend SvelteKit deploy Vercel và tương thích Paperback iOS.
        </p>
      </div>

      <!-- Tech Stack -->
      <div>
        <h4 class="text-xs font-bold uppercase tracking-wider text-slate-300 mb-3">Công Nghệ</h4>
        <ul class="text-xs space-y-2 text-slate-400">
          <li class="flex items-center gap-1.5"><Zap class="w-3.5 h-3.5 text-amber-400" /> Golang 1.24+ (Chi Router)</li>
          <li class="flex items-center gap-1.5"><HardDrive class="w-3.5 h-3.5 text-orange-400" /> Cloudflare R2 Storage (S3 API)</li>
          <li class="flex items-center gap-1.5"><Zap class="w-3.5 h-3.5 text-rose-400" /> SvelteKit + TailwindCSS (Vercel)</li>
          <li class="flex items-center gap-1.5"><ShieldCheck class="w-3.5 h-3.5 text-emerald-400" /> MongoDB Database (Atlas)</li>
        </ul>
      </div>

      <!-- Paperback Integration -->
      <div>
        <h4 class="text-xs font-bold uppercase tracking-wider text-slate-300 mb-3">Tích Hợp</h4>
        <ul class="text-xs space-y-2 text-slate-400">
          <li><a href="https://github.com/chiraitori/paperback-mangahost-extension" target="_blank" rel="noopener noreferrer" class="hover:text-rose-400 transition-colors">Paperback iOS Extension</a></li>
          <li><a href="/admin/login" class="hover:text-rose-400 transition-colors">Admin Dashboard</a></li>
          <li><a href="https://api-manga.chiraitori.dev/api/v1/home" target="_blank" rel="noopener noreferrer" class="hover:text-rose-400 transition-colors">REST API Docs</a></li>
        </ul>
      </div>
    </div>

    <div class="pt-6 border-t border-slate-900 flex flex-col sm:flex-row items-center justify-between text-xs text-slate-500 gap-4">
      <p>© {new Date().getFullYear()} MangaHost. Tối ưu hóa hiệu năng & lưu trữ Cloudflare R2.</p>
      <div class="flex items-center gap-4">
        <span class="inline-flex items-center gap-1 text-emerald-400">
          <span class="w-2 h-2 rounded-full bg-emerald-500 animate-ping"></span> API Online
        </span>
      </div>
    </div>
  </div>
</footer>
