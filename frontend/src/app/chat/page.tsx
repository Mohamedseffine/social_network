'use client'

import { useEffect, useState } from 'react'
import axios from 'axios'

type User = { id: number; name: string }
type Message = { from: number; to: number; content: string; created_at: string }

export default function ChatPage() {
    const [users, setUsers] = useState<User[]>([])
    const [selectedUser, setSelectedUser] = useState<number | null>(null)
    const [messages, setMessages] = useState<Message[]>([])
    const [newMessage, setNewMessage] = useState('')
    const [socket, setSocket] = useState<WebSocket | null>(null)

    const currentUserId = 1 // assume logged in user

    useEffect(() => {
        axios.get('http://localhost:8080/api/chat_users')
            .then(res => setUsers(res.data))
    }, [])

    useEffect(() => {
        const ws = new WebSocket('ws://localhost:8080/api/ws')
        ws.onmessage = (event) => {
            const msg = JSON.parse(event.data)
            setMessages(prev => [...prev, msg])
        }
        setSocket(ws)
        return () => ws.close()
    }, [])

    const loadMessages = async (otherUserId: number) => {
        const res = await axios.get(`http://localhost:8080/api/messages?guest_id=${otherUserId}`);

        setMessages(res.data)
        setSelectedUser(otherUserId)
    }

    const sendMessage = () => {
        if (!socket || !selectedUser) return
        const msg = { to: selectedUser, content: newMessage, type: 'private' }
        socket.send(JSON.stringify(msg))
        setMessages(prev => [...prev, { from: currentUserId, to: selectedUser, content: newMessage, created_at: new Date().toISOString() }])
        setNewMessage('')
    }

    return (
        <div>
            <h1>Users</h1>
            {users.map(u => (
                <div key={u.id} onClick={() => loadMessages(u.id)}>
                    {u.name}
                </div>
            ))}

            <hr />
            <div>
                {messages.map((m, i) => (
                    <div key={i} style={{ textAlign: m.from === currentUserId ? 'right' : 'left' }}>
                        <span>{m.content}</span>
                    </div>
                ))}
            </div>
            <input value={newMessage} onChange={e => setNewMessage(e.target.value)} />
            <button onClick={sendMessage}>Send</button>
        </div>
    )
}
