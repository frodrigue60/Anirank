<div
    class="space-y-4 {{ $isReply ? 'ml-8 mt-4 border-l-2 border-white/5 pl-4' : 'bg-surface-darker/30 rounded-3xl p-6 border border-white/5' }} animation-fade-in">
    <div class="flex gap-4">
        <div class="w-10 h-10 rounded-full overflow-hidden shrink-0 border border-white/10">
            <img src="{{ $comment->user->profile_photo_url }}" alt="{{ $comment->user->name }}"
                class="w-full h-full object-cover">
        </div>
        <div class="flex-1 space-y-2">
            <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                    <h4 class="font-bold text-white text-sm">{{ $comment->user->name }}</h4>
                    @if ($comment->parent_id)
                        <span
                            class="text-[10px] text-primary font-bold uppercase tracking-wider bg-primary/10 px-1.5 py-0.5 rounded">Reply</span>
                    @endif
                </div>
                <span
                    class="text-[10px] text-white/30 uppercase tracking-widest">{{ $comment->created_at->diffForHumans() }}</span>
            </div>

            @if ($editingCommentId === $comment->id)
                <div class="space-y-3 mt-2">
                    <div class="bg-surface-darker rounded-2xl p-1 border border-primary/30 shadow-lg shadow-primary/5">
                        <textarea wire:model.defer="editingBody"
                            class="w-full bg-transparent border-none rounded-xl p-3 text-sm text-white placeholder:text-white/20 min-h-[80px] resize-none focus:ring-0"
                            placeholder="Edit your comment..."></textarea>
                        <div class="flex justify-end items-center px-2 pb-2 gap-2">
                            <button wire:click="cancelEditing"
                                class="text-white/40 hover:text-white px-4 py-1.5 rounded-lg font-bold text-xs transition-colors">
                                Cancel
                            </button>
                            <button wire:click="updateComment"
                                class="bg-primary hover:bg-primary/80 text-white px-5 py-1.5 rounded-lg font-bold text-xs transition-all shadow-lg shadow-primary/20">
                                Save Changes
                            </button>
                        </div>
                    </div>
                    @error('editingBody')
                        <span class="text-red-400 text-xs ml-2">{{ $message }}</span>
                    @enderror
                </div>
            @else
                <p class="text-sm text-white/80 leading-relaxed">
                    {{ $comment->content }}
                </p>

                @auth
                    <div class="flex items-center gap-4 pt-1">
                        <button wire:click="setReplyTo({{ $comment->id }})"
                            class="text-[11px] font-bold text-white/40 hover:text-primary transition-colors flex items-center gap-1">
                            <span class="material-symbols-outlined text-[16px]">reply</span> Reply
                        </button>

                        @if ($comment->user_id === Auth::id() || Auth::user()->isAdmin())
                            <button wire:click="startEditing({{ $comment->id }})"
                                class="text-[11px] font-bold text-white/20 hover:text-primary transition-colors flex items-center gap-1">
                                <span class="material-symbols-outlined text-[16px]">edit</span> Edit
                            </button>

                            <button wire:click="deleteComment({{ $comment->id }})"
                                onclick="return confirm('Are you sure you want to delete this comment?')"
                                class="text-[11px] font-bold text-white/20 hover:text-red-400 transition-colors flex items-center gap-1">
                                <span class="material-symbols-outlined text-[16px]">delete</span> Delete
                            </button>
                        @endif
                    </div>
                @endauth
            @endif
        </div>
    </div>

    {{-- Reply Form for this specific comment --}}
    @if ($replyingTo === $comment->id)
        <div class="ml-14 mt-4 space-y-3">
            <div class="bg-surface-darker rounded-2xl p-1 border border-primary/30 shadow-lg shadow-primary/5">
                <textarea wire:model.defer="replyBody"
                    class="w-full bg-transparent border-none rounded-xl p-3 text-sm text-white placeholder:text-white/20 min-h-[80px] resize-none focus:ring-0"
                    placeholder="Write your reply..."></textarea>
                <div class="flex justify-end items-center px-2 pb-2 gap-2">
                    <button wire:click="cancelReply"
                        class="text-white/40 hover:text-white px-4 py-1.5 rounded-lg font-bold text-xs transition-colors">
                        Cancel
                    </button>
                    <button wire:click="postComment"
                        class="bg-primary hover:bg-primary/80 text-white px-5 py-1.5 rounded-lg font-bold text-xs transition-all shadow-lg shadow-primary/20">
                        Post Reply
                    </button>
                </div>
            </div>
            @error('replyBody')
                <span class="text-red-400 text-xs ml-2">{{ $message }}</span>
            @enderror
        </div>
    @endif

    {{-- Recursive Replies --}}
    @if ($comment->replies->count() > 0)
        <div class="space-y-4">
            @foreach ($comment->replies as $reply)
                @include('partials.comments.item', ['comment' => $reply, 'isReply' => true])
            @endforeach
        </div>
    @endif
</div>
