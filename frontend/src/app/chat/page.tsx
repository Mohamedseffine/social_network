"use client";

import { useEffect, useState, useRef } from "react";

const ChatPage = () => {
  const [messages, setMessages] = useState<any[]>([]);
  const [message, setMessage] = useState("");
  const [isConnected, setIsConnected] = useState(false);
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    // This is a placeholder for a real authentication token
    const token = "your_session_token_here"; // Replace with actual token retrieval

    // Note: The WebSocket URL needs to be correctly configured.
    // Since the backend is on port 8080, and we're using a secure context (wss),
    // you might need to adjust your backend to handle secure websockets or use 'ws' for local dev.
    // For now, assuming 'ws' for simplicity.
    const wsUrl = `ws://localhost:8080/api/ws`;

    ws.current = new WebSocket(wsUrl);

    ws.current.onopen = () => {
      console.log("WebSocket connected");
      setIsConnected(true);
      // You might need to send an authentication token here
      // ws.current?.send(JSON.stringify({ type: 'auth', token }));
    };

    ws.current.onclose = () => {
      console.log("WebSocket disconnected");
      setIsConnected(false);
    };

    ws.current.onmessage = (event) => {
      const receivedMessage = JSON.parse(event.data);
      setMessages((prevMessages) => [...prevMessages, receivedMessage]);
    };

    ws.current.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    return () => {
      ws.current?.close();
    };
  }, []);

  const handleSendMessage = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (message.trim() && ws.current?.readyState === WebSocket.OPEN) {
      const msgToSend = {
        content: message,
        // You would typically include senderId, receiverId or groupId here
      };
      ws.current.send(JSON.stringify(msgToSend));
      setMessage("");
    }
  };

  return (
    <div className="chat-container">
      <h1>Chat</h1>
      <div className="connection-status">
        Status: {isConnected ? "Connected" : "Disconnected"}
      </div>
      <div className="messages">
        {messages.map((msg, index) => (
          <div key={index} className="message">
            {/* Adjust based on your message structure */}
            <p>{msg.content || JSON.stringify(msg)}</p>
          </div>
        ))}
      </div>
      <form onSubmit={handleSendMessage} className="message-form">
        <input
          type="text"
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="Type a message..."
        />
        <button type="submit" disabled={!isConnected}>
          Send
        </button>
      </form>
    </div>
  );
};

export default ChatPage;
