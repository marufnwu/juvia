import { useState, useCallback, useRef } from 'react'
import { useApiError } from './useToast'

interface UseAsyncActionOptions {
  onSuccess?: () => void
  onError?: (error: any) => void
}

export function useAsyncAction<T = any>(
  asyncFn: (...args: any[]) => Promise<T>,
  options: UseAsyncActionOptions = {}
) {
  const { onSuccess, onError } = options
  const showError = useApiError()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<any>(null)
  const abortControllerRef = useRef<AbortController | null>(null)

  const execute = useCallback(
    async (...args: any[]) => {
      if (loading) return

      abortControllerRef.current?.abort()
      abortControllerRef.current = new AbortController()

      setLoading(true)
      setError(null)

      try {
        const result = await asyncFn(...args)
        onSuccess?.()
        return result
      } catch (err: any) {
        if (err.name === 'AbortError') return
        setError(err)
        showError(err)
        onError?.(err)
      } finally {
        setLoading(false)
      }
    },
    [loading, asyncFn, onSuccess, onError, showError]
  )

  const abort = useCallback(() => {
    abortControllerRef.current?.abort()
    setLoading(false)
  }, [])

  return { execute, loading, error, abort }
}

export function useAsyncActionSimple<T = any>(asyncFn: (...args: any[]) => Promise<T>) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<any>(null)
  const abortControllerRef = useRef<AbortController | null>(null)

  const execute = useCallback(async (...args: any[]) => {
    if (loading) return

    abortControllerRef.current?.abort()
    abortControllerRef.current = new AbortController()

    setLoading(true)
    setError(null)

    try {
      return await asyncFn(...args)
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        setError(err)
      }
      throw err
    } finally {
      setLoading(false)
    }
  }, [loading, asyncFn]) as typeof asyncFn

  const abort = useCallback(() => {
    abortControllerRef.current?.abort()
    setLoading(false)
  }, [])

  return { execute, loading, error, abort }
}
