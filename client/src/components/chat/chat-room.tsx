"use client"
import { useState, useEffect, useRef } from 'react'
import { useAuthStore } from '@/lib/store/auth-store'
import { ApiClient } from '@/api'
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { chat_Message } from '@/api/models/chat_Message'

interface ChatRoomProps {
  userId: string
}

export const ChatRoom = ({ userId }: ChatRoomProps) => {
  const [messages, setMessages] = useState<chat_Message[]>([])
  const [loading, setLoading] = useState(false)
  const [newMessage, setNewMessage] = useState("")
  const token = useAuthStore((state) => state.token)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  useEffect(() => {
    const fetchMessages = async () => {
      if (!userId || !token) return
      setLoading(true)
      try {
        const api = new ApiClient()
        const result = await api.chat.getChatMessages(`Bearer ${token}`, userId)
        const sortedMessages = [...result].sort((a, b) => {
          const dateA = a.createdAt ? new Date(a.createdAt).getTime() : 0
          const dateB = b.createdAt ? new Date(b.createdAt).getTime() : 0
          return dateA - dateB
        })
        setMessages(sortedMessages)
      } catch (error) {
        console.error('Failed to fetch messages:', error)
      } finally {
        setLoading(false)
      }
    }
    fetchMessages()
  }, [userId, token])

  const handleSendMessage = async () => {
    if (!newMessage.trim() || !token) return

    try {
      const api = new ApiClient()
      await api.chat.postChatMessages(`Bearer ${token}`, {
        toId: userId,
        content: newMessage.trim()
      })
      // Refresh messages after sending
      const result = await api.chat.getChatMessages(`Bearer ${token}`, userId)
      const sortedMessages = [...result].sort((a, b) => {
        const dateA = a.createdAt ? new Date(a.createdAt).getTime() : 0
        const dateB = b.createdAt ? new Date(b.createdAt).getTime() : 0
        return dateA - dateB
      })
      setMessages(sortedMessages)
      setNewMessage("")
    } catch (error) {
      console.error('Failed to send message:', error)
    }
  }

  if (loading) {
    return <div className="flex-1 flex items-center justify-center">Loading messages...</div>
  }

  return (
    <div className="flex flex-col h-full">
      <div className="border-b p-4">
        <h2 className="font-semibold">Chat</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 ? (
          <div className="text-center text-muted-foreground">
            No messages yet. Start the conversation!
          </div>
        ) : (
          <>
            {messages.map((message) => (
              <div
                key={message.id}
                className={`flex ${message.fromId === userId ? "justify-start" : "justify-end"
                  }`}
              >
                <div
                  className={`rounded-lg px-4 py-2 max-w-sm ${message.fromId === userId
                    ? "bg-muted"
                    : "bg-primary text-primary-foreground"
                    }`}
                >
                  {message.content}
                </div>
              </div>
            ))}
            <div ref={messagesEndRef} />
          </>
        )}
      </div>

      <div className="border-t p-4">
        <form
          onSubmit={(e) => {
            e.preventDefault()
            handleSendMessage()
          }}
          className="flex gap-2"
        >
          <Input
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            placeholder="Type a message..."
            className="flex-1"
          />
          <Button type="submit">Send</Button>
        </form>
      </div>
    </div>
  )
}
