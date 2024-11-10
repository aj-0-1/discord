"use client"

import { useAuthStore } from "@/lib/store/auth-store"
import { useRouter, usePathname } from "next/navigation"
import { useEffect } from "react"

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((state) => state.token)
  const router = useRouter()
  const pathname = usePathname()

  useEffect(() => {
    if (!token && !pathname.startsWith("/auth")) {
      router.push("/login")
    }
  }, [token, router, pathname])

  if (!token && !pathname.startsWith("/auth")) {
    return null
  }

  return <>{children}</>
}
