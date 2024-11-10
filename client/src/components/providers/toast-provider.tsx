"use client"

import { Toaster } from "@/components/ui/toaster"

export function ToastProvider() {
  return <Toaster />
}

// src/app/layout.tsx
import { ToastProvider } from "@/components/providers/toast-provider"

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>
        <AuthGuard>{children}</AuthGuard>
        <ToastProvider />
      </body>
    </html>
  )
}
