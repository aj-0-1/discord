"use client"

import { useState } from "react"
import { AuthGuard } from "@/components/auth/auth-guard"
import { Button } from "@/components/ui/button"
import { MessageSquarePlus } from "lucide-react"
import { NewMessageDialog } from "@/components/chat/new-message-dialog"

export default function ChatLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const [newMessageOpen, setNewMessageOpen] = useState(false)

  return (
    <AuthGuard>
      <div className="flex h-screen">
        <aside className="w-64 border-r bg-muted">
          <nav className="flex h-full flex-col">
            <div className="p-4 border-b flex justify-between items-center">
              <h2 className="font-semibold">Messages</h2>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setNewMessageOpen(true)}
              >
                <MessageSquarePlus className="h-5 w-5" />
              </Button>
            </div>
          </nav>
        </aside>

        <main className="flex-1 flex flex-col">
          {children}
        </main>
      </div>

      <NewMessageDialog
        open={newMessageOpen}
        onClose={() => setNewMessageOpen(false)}
      />
    </AuthGuard>
  )
}
