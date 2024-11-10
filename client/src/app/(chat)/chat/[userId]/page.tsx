import { ChatRoom } from "@/components/chat/chat-room"

interface ChatPageProps {
  params: {
    userId: string
  }
}

export default async function ChatPage({ params }: ChatPageProps) {
  const { userId } = await Promise.resolve(params)  // Properly await params

  return <ChatRoom userId={userId} />
}
