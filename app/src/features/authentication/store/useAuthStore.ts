import { create } from 'zustand';

interface AuthState {
    user: { id: string; name: string; email: string } | null;
    isAuthenticated: boolean;
    login: (email: string) => Promise<void>;
    logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
    user: null,
    isAuthenticated: false,
    login: async (email) => {
        // Simulate an API call
        await new Promise((resolve) => setTimeout(resolve, 1000));
        set({
            user: { id: '1', name: 'Admin', email },
            isAuthenticated: true
        });
    },
    logout: () => set({ user: null, isAuthenticated: false }),
}));
