"use client";

import { createContext, useContext, useState, useEffect, ReactNode, useCallback, useRef } from 'react';
import { API_BASE_URL } from '../utils/api';

interface AuthContextType {
  user: any | null;
  login: (user: any) => void;
  logout: () => void;
  isLoading: boolean;
  unreadNotifications: number;
  fetchNotifications: () => void;
  unreadMessages: number;
  ws: React.RefObject<WebSocket | null>;
  lastChatMessage: any | null;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<any | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [unreadNotifications, setUnreadNotifications] = useState(0);
  const [unreadMessages, setUnreadMessages] = useState(0);
  const [lastChatMessage, setLastChatMessage] = useState<any | null>(null);
  const ws = useRef<WebSocket | null>(null);

  const fetchNotifications = useCallback(async () => {
    // ... (implementation unchanged)
  }, [user]);

  const fetchUnreadMessageCount = useCallback(async () => {
    if (!user) return;
    try {
      const res = await fetch(`${API_BASE_URL}/messages/unread-count`, { credentials: "include" });
      if (res.ok) {
        const data = await res.json();
        setUnreadMessages(data.unread_count);
      }
    } catch (err) {
      console.error("Failed to fetch unread message count", err);
    }
  }, [user]);

  useEffect(() => {
    // ... (checkSession implementation unchanged)
  }, []);

  useEffect(() => {
    if (user && !ws.current) {
      const wsUrl = API_BASE_URL.replace(/^http/, 'ws') + '/ws';
      ws.current = new WebSocket(wsUrl);
      ws.current.onopen = () => console.log("Global WebSocket connected");
      ws.current.onclose = () => console.log("Global WebSocket disconnected");
      ws.current.onerror = (error) => console.error("Global WebSocket error:", error);

      ws.current.onmessage = (event) => {
        const message = JSON.parse(event.data);
        if (message.type === 'new_notification') {
          fetchNotifications();
          fetchUnreadMessageCount(); // A notification might relate to a new message
        } else if (message.type === 'chat_message') {
           setLastChatMessage(message.payload);
           fetchUnreadMessageCount();
        }
      };

      fetchNotifications();
      fetchUnreadMessageCount();
    }
    if (!user && ws.current) {
        ws.current.close();
        ws.current = null;
    }
    return () => {
      if (ws.current) ws.current.close();
    };
  }, [user]); // Removed fetchNotifications and fetchUnreadMessageCount to prevent loops

  const login = (user: any) => setUser(user);

  const logout = () => {
    if (ws.current) {
        ws.current.close();
        ws.current = null;
    }
    setUser(null);
    setUnreadNotifications(0);
    setUnreadMessages(0);
  };

  return (
    <AuthContext.Provider value={{ user, login, logout, isLoading, unreadNotifications, fetchNotifications, unreadMessages, ws, lastChatMessage }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
