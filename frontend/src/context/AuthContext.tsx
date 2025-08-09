"use client";

import { createContext, useContext, useState, useEffect, ReactNode, useCallback } from 'react';

interface AuthContextType {
  user: any | null;
  login: (user: any) => void;
  logout: () => void;
  isLoading: boolean;
  unreadNotifications: number;
  fetchNotifications: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<any | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [unreadNotifications, setUnreadNotifications] = useState(0);

  const fetchNotifications = useCallback(async () => {
    if (!user) return;
    try {
      const res = await fetch("http://localhost:8080/api/notifications", {
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        const unreadCount = data.filter((n: any) => !n.is_read).length;
        setUnreadNotifications(unreadCount);
      }
    } catch (err) {
      console.error("Failed to fetch notifications", err);
    }
  }, [user]);

  useEffect(() => {
    const checkSession = async () => {
      try {
        const res = await fetch("http://localhost:8080/api/session/me", {
          credentials: "include",
        });
        if (res.ok) {
          const data = await res.json();
          setUser(data);
        }
      } catch (err) {
        // User is not logged in
      } finally {
        setIsLoading(false);
      }
    };
    checkSession();
  }, []);

  useEffect(() => {
    if (user) {
      fetchNotifications();
      // Poll for new notifications every 30 seconds
      const interval = setInterval(fetchNotifications, 30000);
      return () => clearInterval(interval);
    }
  }, [user, fetchNotifications]);

  const login = (user: any) => {
    setUser(user);
  };

  const logout = () => {
    setUser(null);
    setUnreadNotifications(0);
  };

  return (
    <AuthContext.Provider value={{ user, login, logout, isLoading, unreadNotifications, fetchNotifications }}>
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
