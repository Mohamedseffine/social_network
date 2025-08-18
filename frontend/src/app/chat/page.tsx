"use client";

import { useEffect, useState, useRef } from "react";
import { API_BASE_URL } from "../../utils/api";
import { useAuth } from "../../context/AuthContext";
import { usePopup } from "../../context/PopupContext";
import Picker from "emoji-picker-react";
import type { EmojiClickData } from "emoji-picker-react";
import "./Chat.css";
import { useRouter } from "next/navigation";

const ChatPage = () => {
  const { user, lastChatMessage, fetchNotifications, fetchUnreadMessageCount, sendMessage } = useAuth();
  const { showPopup } = usePopup();
  const [conversations, setConversations] = useState<any[]>([]);
  const [selectedConversation, setSelectedConversation] = useState<any>(null);
  const [messages, setMessages] = useState<any[]>([]);
  const [newMessage, setNewMessage] = useState("");
  const [showPicker, setShowPicker] = useState(false);
  const router = useRouter()
  const onEmojiClick = (emojiData: EmojiClickData) => {
    setNewMessage((prevInput) => prevInput + emojiData.emoji);
    setShowPicker(false);
  };

  const fetchConversations = async () => {
    if (!user) return;
    try {
      const convRes = await fetch(`${API_BASE_URL}/conversations`, {
        credentials: "include",
      });
      if (convRes.ok) {
        const data = await convRes.json();
        setConversations(data);
      } else {
        if (convRes.status == 401){
          router.push("/")
        }
        console.error("Failed to fetch conversations");
      }
    } catch (error) {
      console.error("Error fetching conversations:", error);
    }
  };

  useEffect(() => {
    fetchConversations();
  }, [user]);

  useEffect(() => {
    if (!lastChatMessage || !user) return;

    let messageBelongsToCurrentConversation = false;
    if (selectedConversation) {
      const [convType, convIdStr] = selectedConversation.id.split('-');
      const convId = parseInt(convIdStr, 10);
      if (lastChatMessage.type === 'private_message' && convType === 'user') {
        if ((lastChatMessage.sender_id === convId && lastChatMessage.target_id === user.id) ||
            (lastChatMessage.sender_id === user.id && lastChatMessage.target_id === convId)) {
          messageBelongsToCurrentConversation = true;
        }
      } else if (lastChatMessage.type === 'group_message' && convType === 'group') {
        if (lastChatMessage.target_id === convId) {
          messageBelongsToCurrentConversation = true;
        }
      }
    }

    if (messageBelongsToCurrentConversation) {
      if (lastChatMessage.sender_id !== user.id) {
          setMessages((prevMessages) => [...prevMessages, lastChatMessage]);
      }
    } else {
      // If message is not for the current conversation, or no conversation is selected,
      // refetch the conversation list to update unread counts.
      fetchConversations();
    }
  }, [lastChatMessage, user, selectedConversation]);

  const handleSelectConversation = async (conversation: any) => {
    setSelectedConversation(conversation);
    setMessages([]);
    const [type, id] = conversation.id.split('-');
    const queryParam = type === 'user' ? `user_id=${id}` : `group_id=${id}`;

    try {
      const res = await fetch(`${API_BASE_URL}/messages?${queryParam}`, {
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        setMessages(data || []);
        // After successfully fetching messages, the backend has marked them as read.
        // Now, we need to re-fetch the counts to update the UI instantly.
        if (fetchUnreadMessageCount) fetchUnreadMessageCount();
        if (fetchNotifications) fetchNotifications();

        // Also update the unread count for this specific conversation in the local state
        setConversations(prev => prev.map(c =>
          c.id === conversation.id ? { ...c, unread_count: 0 } : c
        ));

      } else {
         if (res.status == 401){
          router.push("/")
         }
        console.error("Failed to fetch messages");
      }
    } catch (error) {
      console.error("Error fetching messages:", error);
    }
  };

  const handleSendMessage = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!newMessage.trim() || !selectedConversation) {
      return;
    }

    const [type, idStr] = selectedConversation.id.split('-');
    const id = parseInt(idStr, 10);

    const message = {
      type: type === 'user' ? 'private_message' : 'group_message',
      payload: {
        content: newMessage,
        recipient_id: type === 'user' ? id : 0,
        group_id: type === 'group' ? id : 0,
      },
    };

    sendMessage(message);

    if (user) {
        const optimisticMessage = {
            id: Date.now(),
            sender_id: user.id,
            sender_name: "Me",
            content: newMessage,
            created_at: new Date().toISOString(),
        };
        setMessages([...messages, optimisticMessage]);
    }
    setNewMessage("");
  };

  return (
    <div className="chat-container">
      <div className="conversation-list">
        <h2>Conversations</h2>
        {conversations.map((conv) => (
          <div
            key={conv.id}
            className={`conversation-item ${selectedConversation?.id === conv.id ? "selected" : ""}`}
            onClick={() => handleSelectConversation(conv)}
          >
            {conv.name}
            {conv.unread_count > 0 && (
              <span className="notification-bell-list">🔔</span>
            )}
          </div>
        ))}
      </div>
      <div className="chat-panel">
        {selectedConversation ? (
          <>
            <div className="chat-header"><h3>{selectedConversation.name}</h3></div>
            <div className="messages">
              {messages.map((msg) => (
                <div key={msg.id} className={`message ${msg.sender_id === user?.id ? 'sent' : 'received'}`}>
                  <div className="message-content">
                    <strong>{msg.sender_name}: </strong>{msg.content}
                  </div>
                </div>
              ))}
            </div>
            <form onSubmit={handleSendMessage} className="message-form">
              <input
                type="text"
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
                placeholder="Type a message..."
              />
              <button
                type="button"
                className="emoji-button"
                onClick={() => setShowPicker((val) => !val)}
              >
                😊
              </button>
              <button type="submit">
                Send
              </button>
            </form>
            {showPicker && (
              <div className="picker-container">
                <Picker onEmojiClick={onEmojiClick} />
              </div>
            )}
          </>
        ) : (
          <div className="no-conversation-selected">
            <h2>Select a conversation to start chatting</h2>
          </div>
        )}
      </div>
    </div>
  );
};

export default ChatPage;
