'use client'

import { useEffect, useState, useRef, UIEvent } from 'react'
import axios from 'axios'

type User = { id: number; name: string }
type Message = { from: number; to: number; content: string; created_at: string }

export default function ChatPage() {
  const [users, setUsers] = useState<User[]>([])
  const [selectedUser, setSelectedUser] = useState<number | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [newMessage, setNewMessage] = useState('')
  const [socket, setSocket] = useState<WebSocket | null>(null)
  const [loadingMore, setLoadingMore] = useState(false)
  const [hasMore, setHasMore] = useState(true)

  const currentUserId = 1 // assume logged in user
  const limit = 20

  // ref to chat messages container for scroll event handling
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const offsetRef = useRef(0) // keep track of how many messages are loaded

  // Load users list once
  useEffect(() => {
    axios.get('http://localhost:8080/api/chat_users')
      .then(res => setUsers(res.data))
      .catch(console.error)
  }, [])

  // Setup WebSocket connection once
  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/api/ws')

    ws.onmessage = (event) => {
      const msg: Message = JSON.parse(event.data)
      // Only add message if it belongs to current chat
      if (
        selectedUser !== null &&
        (msg.from === selectedUser || msg.to === selectedUser)
      ) {
        setMessages(prev => [...prev, msg])
        scrollToBottom()
      }
    }

    setSocket(ws)
    return () => ws.close()
  }, [selectedUser])

  // Load initial messages when user is selected
  const loadMessages = async (otherUserId: number, offset = 0) => {
    if (loadingMore) return
    setLoadingMore(true)
    try {
      const res = await axios.get(`http://localhost:8080/api/messages`, {
        params: {
          guest_id: otherUserId,
          offset,
          limit,
        }
      })

      if (res.data.length < limit) setHasMore(false)
      else setHasMore(true)

      if (offset === 0) {
        // first load, replace messages
        setMessages(res.data)
        scrollToBottom()
      } else {
        // prepend older messages on scroll up
        setMessages(prev => [...res.data, ...prev])
      }
      offsetRef.current = offset + res.data.length
    } catch (err) {
      console.error('Failed to load messages:', err)
    }
    setLoadingMore(false)
  }

  // Called when user clicks on a user from the list
  const handleUserSelect = (userId: number) => {
    setSelectedUser(userId)
    offsetRef.current = 0
    setHasMore(true)
    loadMessages(userId, 0)
  }

  // Scroll chat to bottom
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  // Infinite scroll: detect scroll top and load older messages
  const onScroll = (e: UIEvent<HTMLDivElement>) => {
    const el = e.currentTarget
    if (el.scrollTop === 0 && hasMore && !loadingMore && selectedUser !== null) {
      // Load older messages
      loadMessages(selectedUser, offsetRef.current)
    }
  }

  // Send message through WebSocket
  const sendMessage = () => {
    if (!socket || !selectedUser || newMessage.trim() === '') return

    const msg = { to: selectedUser, content: newMessage, type: 'private' }
    socket.send(JSON.stringify(msg))

    // Optimistically update UI
    setMessages(prev => [
      ...prev,
      { from: currentUserId, to: selectedUser, content: newMessage, created_at: new Date().toISOString() }
    ])
    setNewMessage('')
    scrollToBottom()
  }

  return (
    <div style={{ display: 'flex', height: '90vh' }}>
      {/* Users list */}
      <div style={{ width: '250px', borderRight: '1px solid gray', overflowY: 'auto' }}>
        <h2>Users</h2>
        {users.map(u => (
          <div
            key={u.id}
            onClick={() => handleUserSelect(u.id)}
            style={{ padding: '10px', cursor: 'pointer', backgroundColor: selectedUser === u.id ? '#ddd' : 'transparent' }}
          >
            {u.name}
          </div>
        ))}
      </div>

      {/* Chat section */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
        <div
          ref={containerRef}
          onScroll={onScroll}
          style={{ flex: 1, overflowY: 'auto', padding: '10px', borderBottom: '1px solid gray' }}
        >
          {loadingMore && <div>Loading older messages...</div>}

          {messages.map((m, i) => (
            <div
              key={i}
              style={{
                textAlign: m.from === currentUserId ? 'right' : 'left',
                margin: '5px 0',
              }}
            >
              <span
                style={{
                  display: 'inline-block',
                  backgroundColor: m.from === currentUserId ? '#daf8cb' : '#eee',
                  padding: '8px 12px',
                  borderRadius: '15px',
                  maxWidth: '60%',
                  wordWrap: 'break-word',
                }}
              >
                {m.content}
              </span>
              <div style={{ fontSize: '10px', color: '#888' }}>
                {new Date(m.created_at).toLocaleTimeString()}
              </div>
            </div>
          ))}

          {/* Dummy div to scroll into view */}
          <div ref={messagesEndRef} />
        </div>

        {/* Message input */}
        <div style={{ padding: '10px', display: 'flex', borderTop: '1px solid gray' }}>
          <input
            type="text"
            value={newMessage}
            onChange={e => setNewMessage(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') sendMessage() }}
            style={{ flex: 1, padding: '8px', fontSize: '16px' }}
            placeholder={selectedUser ? 'Type your message...' : 'Select a user to chat'}
            disabled={!selectedUser}
          />
          <button onClick={sendMessage} disabled={!selectedUser || newMessage.trim() === ''} style={{ marginLeft: '10px' }}>
            Send
          </button>
        </div>
      </div>
    </div>
  )
}
