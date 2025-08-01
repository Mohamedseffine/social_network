'use client'

import { useEffect, useState, useRef, UIEvent, useCallback } from 'react'
import axios from 'axios'
import useSharedSocket from '../hooks/useSharedSocket'

type User = { id: number; nick_name: string,  is_online: boolean, unread_count: number}// Create a model to ease working on chat

type Message = { from: number; to: number; content: string; created_at: string }

export default function ChatPage() {
    const [users, setUsers] = useState<User[]>([])
    const [selectedUser, setSelectedUser] = useState<number | null>(null)
    const [messages, setMessages] = useState<Message[]>([])
    const [newMessage, setNewMessage] = useState('')
    const [loadingMore, setLoadingMore] = useState(false)
    const [hasMore, setHasMore] = useState(true)

    const currentUserId = 1
    const limit = 20
    const messagesEndRef = useRef<HTMLDivElement>(null)
    const offsetRef = useRef(0)

    useEffect(() => {
        axios.get('http://localhost:8080/api/chat_users', {
            withCredentials: true 
        })
            .then(res => setUsers(res.data))
            .catch(console.error)
    }, [])

    const handleSocketMessage = useCallback((msg: Message) => {
        if (
            selectedUser !== null &&
            (msg.from === selectedUser || msg.to === selectedUser)
        ) {
            setMessages(prev => [...prev, msg])
            scrollToBottom()
        }
    }, [selectedUser])

    const { send } = useSharedSocket(handleSocketMessage)

    const loadMessages = async (otherUserId: number, offset = 0) => {
        if (loadingMore) return
        setLoadingMore(true)
        try {
            const res = await axios.get(`http://localhost:8080/api/messages`, {
                params: { guest_id: otherUserId, offset, limit }
            })

            if (res.data.length < limit) setHasMore(false)
            else setHasMore(true)

            if (offset === 0) {
                setMessages(res.data)
                scrollToBottom()
            } else {
                setMessages(prev => [...res.data, ...prev])
            }

            offsetRef.current = offset + res.data.length
        } catch (err) {
            console.error('Failed to load messages:', err)
        }
        setLoadingMore(false)
    }

    const handleUserSelect = (userId: number) => {
        setSelectedUser(userId)
        offsetRef.current = 0
        setHasMore(true)
        loadMessages(userId, 0)
    }

    const scrollToBottom = () => {
        setTimeout(() => {
            messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
        }, 10)
    }

    const onScroll = (e: UIEvent<HTMLDivElement>) => {
        const el = e.currentTarget
        if (el.scrollTop === 0 && hasMore && !loadingMore && selectedUser !== null) {
            loadMessages(selectedUser, offsetRef.current)
        }
    }

    const sendMessage = () => {
        if (!selectedUser || newMessage.trim() === '') return

        const msg = {
            type: 'private',
            to: selectedUser,
            content: newMessage
        }

        send(msg)

        // Optimistic update
        setMessages(prev => [
            ...prev,
            {
                from: currentUserId,
                to: selectedUser,
                content: newMessage,
                created_at: new Date().toISOString()
            }
        ])
        setNewMessage('')
        scrollToBottom()
    }

    return (
        <div style={{ display: 'flex', height: '90vh' }}>
            {/* Users list */}
            <div style={{ width: '250px', borderRight: '1px solid gray', overflowY: 'auto' }}>
                <h2>Users</h2>
                {users.length > 0 && (
                    users.map(u => (
                        <div
                            key={u.id}
                            onClick={() => handleUserSelect(u.id)}
                            style={{
                                padding: '10px',
                                cursor: 'pointer',
                                backgroundColor: selectedUser === u.id ? '#ddd' : 'transparent'
                            }}
                        >
                            {u.nick_name}
                        </div>
                    ))
                )}
            </div>


            {/* Chat section */}
            <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
                <div
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
