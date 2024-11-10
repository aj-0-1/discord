import { ApiClient } from '@/api'

// Create a singleton instance
export const api = new ApiClient({
  BASE: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'
})
