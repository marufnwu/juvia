export class WSClient {
  private url: string
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10
  private listeners: Map<string, ((data: unknown) => void)[]> = new Map()

  constructor(url: string) {
    this.url = url
  }

  connect() {
    this.ws = new WebSocket(this.url)
    this.ws.onopen = () => {
      this.reconnectAttempts = 0
      const handlers = this.listeners.get('connect') || []
      handlers.forEach((h) => h(null))
    }
    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        const type = msg.type || 'message'
        const handlers = this.listeners.get(type) || []
        handlers.forEach((h) => h(msg.data))
      } catch {
        // ignore parse errors
      }
    }
    this.ws.onclose = () => {
      const handlers = this.listeners.get('close') || []
      handlers.forEach((h) => h(null))
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        const delay = Math.min(1000 * 2 ** this.reconnectAttempts, 30000)
        setTimeout(() => {
          this.reconnectAttempts++
          this.connect()
        }, delay)
      }
    }
    this.ws.onerror = () => {
      this.ws?.close()
    }
  }

  on(type: string, handler: (data: unknown) => void) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, [])
    }
    this.listeners.get(type)!.push(handler)
  }

  off(type: string, handler: (data: unknown) => void) {
    const handlers = this.listeners.get(type) || []
    this.listeners.set(
      type,
      handlers.filter((h) => h !== handler)
    )
  }

  send(data: unknown) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    }
  }

  close() {
    this.ws?.close()
  }
}
