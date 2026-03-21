export interface Toast {
    id: string;
    message: string;
    type: 'success' | 'error' | 'info' | 'warning';
    duration: number;
}

class ToastState {
    toasts = $state<Toast[]>([]);

    addToast(message: string, type: 'success' | 'error' | 'info' | 'warning' = 'info', duration: number = 5000) {
        const id = Math.random().toString(36).substring(2, 9);
        const newToast: Toast = { id, message, type, duration };
        this.toasts.push(newToast);

        if (duration > 0) {
            setTimeout(() => {
                this.removeToast(id);
            }, duration);
        }
    }

    removeToast(id: string) {
        this.toasts = this.toasts.filter(t => t.id !== id);
    }
}

export const toastState = new ToastState();
