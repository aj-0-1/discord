import { useAuthStore } from "@/lib/store/auth-store"

export class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string
  private token: string
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectTimeout = 1000 // Start with 1s, will increase exponentially

  constructor() {
    this.url = `ws://localhost:8080/api/chat/ws`  // Should come from env config
    this.token = useAuthStore.getState().token ?? ''
  }

  connect() {
    try {
      this.ws = new WebSocket(this.url)
      
      this.ws.onopen = () => {
        console.log('WebSocket connected')
        this.reconnectAttempts = 0
        this.reconnectTimeout = 1000
      }

      this.ws.onclose = () => {
        console.log('WebSocket disconnected')
        this.reconnect()
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error)
      }

      return this.ws
    } catch (error) {
      console.error('WebSocket connection error:', error)
      this.reconnect()
      return null
    }
  }

  private reconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached')
      return
    }

    setTimeout(() => {
      this.reconnectAttempts++
      this.reconnectTimeout *= 2 // Exponential backoff
      this.connect()
    }, this.reconnectTimeout)
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }
}
