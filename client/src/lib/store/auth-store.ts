
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { api } from '@/lib/api'
import type { auth_AuthResponse } from '@/api/models/auth_AuthResponse'

interface AuthState {
  user: auth_AuthResponse['user'] | null
  token: string | null
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, username: string) => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      login: async (email: string, password: string) => {
        try {
          const response = await api.auth.postAuthLogin({ email, password })
          if (response.user && response.token) {
            set({ user: response.user, token: response.token })
          } else {
            throw new Error('Login failed, no user or token returned')
          }
        } catch (error) {
          console.error('Login failed:', error)
        }
      },
      register: async (email: string, password: string, username: string) => {
        try {
          const response = await api.auth.postAuthRegister({
            email,
            password,
            username,
          })
          if (response.user && response.token) {
            set({ user: response.user, token: response.token })
          } else {
            throw new Error('Registration failed, no user or token returned')
          }
        } catch (error) {
          console.error('Registration failed:', error)
        }
      },
      logout: () => {
        set({ user: null, token: null })
        // Optionally clear local storage explicitly
        localStorage.removeItem('auth-storage')
      },
    }),
    {
      name: 'auth-storage',
    }
  )
)

