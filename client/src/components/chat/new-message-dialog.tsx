"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { ApiClient } from '@/api'
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { useAuthStore } from "@/lib/store/auth-store"
import { Search } from "lucide-react"
import { useDebounce } from "@/hooks"
import { user_User } from "@/api/models/user_User"

export function NewMessageDialog({
  open,
  onClose
}: {
  open: boolean
  onClose: () => void
}) {
  const [searchQuery, setSearchQuery] = useState("")
  const [users, setUsers] = useState<user_User[]>([])
  const [loading, setLoading] = useState(false)
  const router = useRouter()
  const token = useAuthStore((state) => state.token)
  const debouncedSearch = useDebounce(searchQuery, 300)

  useEffect(() => {
    const searchUsers = async () => {
      if (!debouncedSearch.trim() || !token) return setUsers([]) // Only search if token is available

      try {
        setLoading(true)
        const api = new ApiClient()
        // Use the token as a Bearer token in the headers for the API request

        const results = await api.users.getUsersSearch(`Bearer ${token}`, debouncedSearch)
        setUsers(results)
      } catch (error: any) {
        console.error('Failed to search users:', error);
        if (error.response) {
          console.error('Response error:', error.response.data);
        }
      } finally {
        setLoading(false)
      }
    }

    searchUsers()
  }, [debouncedSearch, token])

  const handleStartChat = (userId: string) => {
    if (userId) {
      router.push(`/chat/${userId}`)
      onClose()
    }
  }

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>New Message</DialogTitle>
        </DialogHeader>
        <div className="relative">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search users..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-8"
          />
        </div>
        <div className="mt-2 space-y-1">
          {loading && (
            <div className="text-sm text-muted-foreground text-center py-4">
              Searching...
            </div>
          )}
          {!loading && users.length === 0 && searchQuery && (
            <div className="text-sm text-muted-foreground text-center py-4">
              No users found
            </div>
          )}
          {users.map((user) => (
            <button
              key={user.id}
              onClick={() => user.id && handleStartChat(user.id)}
              className="w-full p-2 text-left hover:bg-accent rounded-md flex items-center gap-2"
            >
              <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
                <span className="text-sm font-medium">
                  {user.username?.[0].toUpperCase()}
                </span>
              </div>
              <div>
                <div className="font-medium">{user.username}</div>
                <div className="text-sm text-muted-foreground">{user.email}</div>
              </div>
            </button>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  )
}
