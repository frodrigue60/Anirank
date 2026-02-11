<div id="requestModal" class="fixed inset-0 z-[60] hidden overflow-y-auto overflow-x-hidden p-4 md:p-8" role="dialog"
    aria-modal="true">
    {{-- Backdrop --}}
    <div class="fixed inset-0 bg-black/60 backdrop-blur-sm transition-opacity" onclick="window.hideModal('requestModal')">
    </div>

    {{-- Modal Content --}}
    <div class="relative mx-auto mt-20 max-w-lg w-full transform transition-all">
        <div class="glass-panel bg-surface-darker/90 rounded-2xl border border-white/10 shadow-2xl overflow-hidden">
            {{-- Header --}}
            <div class="px-6 py-4 border-b border-white/5 flex items-center justify-between">
                <div class="flex items-center gap-2">
                    <span class="material-symbols-outlined text-primary">campaign</span>
                    <h3 class="text-xl font-bold text-white tracking-tight">Send Request</h3>
                </div>
                <button type="button" onclick="window.hideModal('requestModal')"
                    class="text-white/40 hover:text-white transition-colors">
                    <span class="material-symbols-outlined">close</span>
                </button>
            </div>

            {{-- Body --}}
            <div class="p-6">
                <form method="post" action="{{ route('request.store') }}" id="form-request" class="space-y-6">
                    @csrf

                    <div>
                        <label for="select-request-type" class="block text-sm font-medium text-white/60 mb-2">Request
                            Category</label>
                        <div class="relative">
                            <select name="type" id="select-request-type"
                                class="w-full bg-surface-dark border border-white/10 rounded-xl px-4 py-3 text-white focus:border-primary focus:ring-1 focus:ring-primary appearance-none transition-all outline-none">
                                <option value="" disabled selected>Select a category...</option>
                                <option value="1">Add Anime/Song</option>
                                <option value="2">Report Bug/Issue</option>
                                <option value="3">Suggestion/Feedback</option>
                            </select>
                            <span
                                class="material-symbols-outlined absolute right-4 top-3.5 text-white/20 pointer-events-none">expand_more</span>
                        </div>
                    </div>

                    <div>
                        <label for="textarea-request"
                            class="block text-sm font-medium text-white/60 mb-2">Details</label>
                        <textarea name="content" id="textarea-request" rows="4" required
                            class="w-full bg-surface-dark border border-white/10 rounded-xl px-4 py-3 text-white focus:border-primary focus:ring-1 focus:ring-primary transition-all outline-none resize-none placeholder:text-white/20"
                            placeholder="Describe your request in detail..."></textarea>
                    </div>

                    <div class="flex items-center gap-3 pt-2">
                        <button type="button" onclick="window.hideModal('requestModal')"
                            class="flex-1 px-4 py-3 rounded-xl border border-white/10 text-white/60 font-medium hover:bg-white/5 transition-colors">
                            Cancel
                        </button>
                        <button type="submit"
                            class="flex-1 px-4 py-3 rounded-xl bg-primary text-white font-bold shadow-lg shadow-primary/20 hover:scale-[1.02] active:scale-[0.98] transition-all">
                            Send Request
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</div>
