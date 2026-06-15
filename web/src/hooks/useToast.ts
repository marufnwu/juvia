import { useCallback } from 'react'
import { toast } from '../components/ui/Toast'

interface ApiError {
  response?: {
    data?: {
      error?: {
        user_message?: string
        message?: string
      }
    }
    status?: number
  }
  message?: string
}

export function useApiError() {
  return useCallback((err: unknown, fallbackMessage = 'Something went wrong. Please try again.') => {
    const error = err as ApiError
    const status = error.response?.status

    if (status === 401) {
      toast({ variant: 'error', title: 'Session expired', message: 'Please log in again.' })
      return
    }

    if (status === 403) {
      toast({ variant: 'error', title: 'Access denied', message: 'You do not have permission to perform this action.' })
      return
    }

    if (status === 429) {
      toast({ variant: 'warning', title: 'Too many requests', message: 'Please wait a moment and try again.' })
      return
    }

    const message =
      error.response?.data?.error?.user_message ||
      error.response?.data?.error?.message ||
      error.message ||
      fallbackMessage

    toast({ variant: 'error', title: 'Error', message })
  }, [])
}

export function useApiSuccess() {
  return useCallback((title: string, message?: string) => {
    toast({ variant: 'success', title, message })
  }, [])
}
