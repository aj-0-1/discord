"use client"

import { useAuthStore } from "@/lib/store/auth-store"
import Link from "next/link"
import { usePathname } from "next/navigation"

type Conversation = {
  userId: string
  username: string
  lastMessage?: string
}

export function ConversationList() {
  const pathname = usePathname()
  // This will be replaced with real data from the API
  const conversations: Conversation[] = [
    // Temporary mock data
    { userId: "1", username: "User 1" },
    { userId: "2", username: "User 2" },
  ]

  return (
    <div className="space-y-1">
      {conversations.map((conversation) => (
        <Link
          key={conversation.userId}
          href={`/chat/${conversation.userId}`}
          className={`block px-4 py-2 hover:bg-accent ${
            pathname === `/chat/${conversation.userId}` ? "bg-accent" : ""
          }`}
        >
          <div className="font-medium">{conversation.username}</div>
          {conversation.lastMessage && (
            <div className="text-sm text-muted-foreground truncate">
              {conversation.lastMessage}
            </div>
          )}
        </Link>
      ))}
    </div>
  )
}
