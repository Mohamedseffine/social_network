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
  lastChatMessage: any | null;
  onlineUsers: any[];
  fetchUnreadMessageCount: () => void;
  sendMessage: (message: any) => void; // New function to send messages via worker
  notificationTrigger: number;
  triggerNotificationRefresh: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<any | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [unreadNotifications, setUnreadNotifications] = useState(0);
  const [unreadMessages, setUnreadMessages] = useState(0);
  const [lastChatMessage, setLastChatMessage] = useState<any | null>(null);
  const [onlineUsers, setOnlineUsers] = useState<any[]>([]);
  const worker = useRef<SharedWorker | null>(null);
  const [notificationTrigger, setNotificationTrigger] = useState(0);

  const triggerNotificationRefresh = () => {
    setNotificationTrigger((prev) => prev + 1);
  };

  const fetchNotifications = useCallback(async () => {
    if (!user) return;
    try {
      const res = await fetch(`${API_BASE_URL}/notifications/unread-count`, { credentials: "include" });
      if (res.ok) {
        const data = await res.json();
        setUnreadNotifications(data.unread_count);
      }
    } catch (err) {
      console.error("Failed to fetch unread notifications", err);
    }
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

  const sendMessage = (message: any) => {
    if (worker.current) {
      worker.current.port.postMessage(message);
    } else {
      console.error("Shared Worker is not available.");
    }
  };

  useEffect(() => {
    const checkSession = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/session/me`, { credentials: 'include' });
        if (res.ok) {
          const userData = await res.json();
          setUser(userData);
        } else {
          setUser(null);
        }
      } catch (err) {
        console.error("Session check failed", err);
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };
    checkSession();
  }, []);

  useEffect(() => {
    if (user && !worker.current) {
      // Path to the worker script
      worker.current = new SharedWorker('/socket-worker.js');

      worker.current.port.onmessage = (event) => {
        const message = event.data;
        // Logic to handle messages from the worker
        if (message.type === 'new_notification') {
          fetchNotifications();
          fetchUnreadMessageCount();
        } else if (message.type === 'chat_message') {
           setLastChatMessage(message.payload);
           fetchUnreadMessageCount();
        } else if (message.type === 'online_users') {
          setOnlineUsers(message.payload);
        } else if (message.type === 'WS_OPEN') {
            console.log("AuthContext: WebSocket connection confirmed open by worker.");
            // Fetch initial data now that we know the connection is good
            fetchNotifications();
            fetchUnreadMessageCount();
        }
      };

      // Start the port and initialize the worker with the WebSocket URL
      worker.current.port.start();
      const wsUrl = API_BASE_URL.replace(/^http/, 'ws') + '/ws';
      worker.current.port.postMessage({ type: 'INIT_WS', payload: wsUrl });

    } else if (!user && worker.current) {
        worker.current.port.close();
        worker.current = null;
    }

    // No return/cleanup function needed in the same way, as the worker persists.
    // The port will be garbage collected when the context unmounts.
  }, [user, fetchNotifications, fetchUnreadMessageCount]);

  const login = (user: any) => setUser(user);

  const logout = () => {
    if (worker.current) {
        worker.current.port.close();
        worker.current = null;
    }
    setUser(null);
    setUnreadNotifications(0);
    setUnreadMessages(0);
    setOnlineUsers([]);
  };

  return (
    <AuthContext.Provider value={{ user, login, logout, isLoading, unreadNotifications, fetchNotifications, unreadMessages, fetchUnreadMessageCount, lastChatMessage, onlineUsers, sendMessage, notificationTrigger, triggerNotificationRefresh }}>
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
