let policy: any;

/**
 * Wraps a string in a TrustedHTML object if the browser supports and enforces Trusted Types.
 * This is required by the CSP policy to allow use of {@html} in Svelte components.
 */
export function createTrustedHTML(html: string): any {
    if (typeof window !== 'undefined' && (window as any).trustedTypes) {
        if (!policy) {
            try {
                policy = (window as any).trustedTypes.createPolicy('svelte-trusted-html', {
                    createHTML: (s: string) => s
                });
            } catch (e) {
                // Policy might already exist if this is called after a hot reload or multiple times
                // We just want to use the existing one if possible, but TT doesn't have a getPolicy(name)
                // that is widely supported or consistent. 'allow-duplicates' in CSP helps.
                console.warn('TrustedTypes policy creation failed or already exists:', e);
            }
        }
        if (policy) {
            return policy.createHTML(html);
        }
    }
    return html;
}
