<div x-data="{
    notices: [],
    add(notice) {
        notice.id = Date.now()
        this.notices.push(notice)
        // Auto remove after 5 seconds
        setTimeout(() => { this.remove(notice.id) }, 5000)
    },
    remove(id) {
        this.notices = this.notices.filter(n => n.id !== id)
    }
}" @toast.window="add($event.detail)"
    class="fixed top-6 right-6 z-[110] flex flex-col gap-3 pointer-events-none">
    <template x-for="notice in notices" :key="notice.id">
        <div x-show="true" x-transition:enter="transition ease-out duration-300"
            x-transition:enter-start="opacity-0 transform -translate-y-2"
            x-transition:enter-end="opacity-100 transform translate-y-0"
            x-transition:leave="transition ease-in duration-200"
            x-transition:leave-start="opacity-100 transform translate-y-0"
            x-transition:leave-end="opacity-0 transform -translate-y-2"
            class="pointer-events-auto min-w-[300px] max-w-sm rounded-xl p-4 shadow-2xl border flex items-start gap-3 backdrop-blur-md"
            :class="{
                'bg-[#1e1e2e] border-green-500/50 text-green-400': notice.type === 'success',
                'bg-[#1e1e2e] border-red-500/50 text-red-400': notice.type === 'error',
                'bg-[#1e1e2e] border-blue-500/50 text-blue-400': notice.type === 'info',
                'bg-[#1e1e2e] border-yellow-500/50 text-yellow-400': notice.type === 'warning'
            }">
            <span class="material-symbols-outlined text-xl shrink-0"
                x-text="
                notice.type === 'success' ? 'check_circle' : 
                notice.type === 'error' ? 'error' : 
                notice.type === 'info' ? 'info' : 'warning'
            "></span>

            <div class="flex-1 pt-0.5">
                <p class="font-bold text-sm text-white" x-text="notice.message"></p>
                <p x-show="notice.description" class="text-xs text-white/60 mt-1" x-text="notice.description"></p>
            </div>

            <button @click="remove(notice.id)" class="text-white/40 hover:text-white transition-colors">
                <span class="material-symbols-outlined text-lg">close</span>
            </button>
        </div>
    </template>
</div>
