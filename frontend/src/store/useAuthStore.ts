import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface Faculty {
  id: string;
  login_id: string;
  name: string;
  first_name: string;
  last_name: string;
  employee_id: string;
  department: string;
  designation: string;
  email: string;
  phone: string;
  address?: string;
  dob?: string;
  gender?: string;
  blood_group?: string;
  marital_status?: string;
  qualification?: string;
  experience?: string;
  joining_date?: string;
  subjects?: string[];
}

interface AuthState {
  user: Faculty | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  setAuth: (user: Faculty) => void;
  clearAuth: () => void;
  setLoading: (status: boolean) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,
      isLoading: true,
      setAuth: (user) => set({ user, isAuthenticated: true, isLoading: false }),
      clearAuth: () => set({ user: null, isAuthenticated: false, isLoading: false }),
      setLoading: (status) => set({ isLoading: status }),
    }),
    {
      name: 'auth-storage',
      onRehydrateStorage: () => (state) => {
        if (state) {
          state.setLoading(false);
        }
      },
    }
  )
);
