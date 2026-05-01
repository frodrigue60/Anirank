import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import NavbarMaster from '../../lib/components/NavbarMaster.svelte';
import { authState, setUser, logout } from '$lib/state/auth.svelte';
import { goto } from '$app/navigation';

describe('NavbarMaster Component', () => {
    beforeEach(() => {
        // Reset auth state before each test
        setUser(null);
    });

    it('should render User Toggle when not authenticated', async () => {
        render(NavbarMaster);

        const toggle = screen.getByLabelText(/toggle user dropdown menu/i);
        expect(toggle).toBeInTheDocument();

        // Click to see login options
        await fireEvent.click(toggle);
        expect(screen.getByText(/sign in/i)).toBeInTheDocument();
        expect(screen.getByText(/register/i)).toBeInTheDocument();
    });

    it('should render User Profile dropdown when authenticated', async () => {
        const mockUser = {
            id: 1,
            uuid: 'user-uuid',
            name: 'Test User',
            slug: 'test-user',
            email: 'test@example.com'
        };

        setUser(mockUser);
        render(NavbarMaster);

        // Should not see login/register
        expect(screen.queryByText(/sign in/i)).not.toBeInTheDocument();

        // Should see user avatar/toggle (it uses an aria-label "Toggle user dropdown menu")
        const toggle = screen.getByLabelText(/toggle user dropdown menu/i);
        expect(toggle).toBeInTheDocument();

        // Click to open dropdown
        await fireEvent.click(toggle);

        // Should see profile name and logout button
        expect(screen.getByText(/test user/i)).toBeInTheDocument();
        expect(screen.getByText(/logout/i)).toBeInTheDocument();
    });

    it('should handle logout correctly', async () => {
        setUser({
            id: 1,
            uuid: 'user-uuid',
            name: 'Test User',
            slug: 'test-user',
            email: 'test@example.com'
        });

        render(NavbarMaster);

        const toggle = screen.getByLabelText(/toggle user dropdown menu/i);
        await fireEvent.click(toggle);

        const logoutButton = screen.getByText(/logout/i);
        await fireEvent.click(logoutButton);

        // Verify logout function was called (indirectly by checking authState)
        expect(authState.isAuthenticated).toBe(false);
        expect(authState.user).toBe(null);

        // Verify navigation to home
        expect(goto).toHaveBeenCalledWith('/');
        
        // UI should revert to guest state
        // Re-open dropdown to check buttons
        const guestToggle = screen.getByLabelText(/toggle user dropdown menu/i);
        await fireEvent.click(guestToggle);

        await waitFor(() => {
            expect(screen.getByText(/sign in/i)).toBeInTheDocument();
        });
    });
});
