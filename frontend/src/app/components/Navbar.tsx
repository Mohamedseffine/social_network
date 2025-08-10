"use client";

import Link from "next/link";
import { useAuth } from "../../context/AuthContext";
import { API_BASE_URL } from "../../utils/api";

const Navbar = () => {
  const { user, logout, unreadNotifications, unreadMessages } = useAuth();

  const handleLogout = async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/logout`, {
        method: "POST",
        credentials: "include",
      });
      if (res.ok) {
        logout();
      } else {
        console.error("Logout failed");
      }
    } catch (err) {
      console.error("Logout failed", err);
    }
  };

  return (
    <nav className="navbar">
      <div className="nav-container">
        <Link href="/" className="nav-logo">
          Social Dilemma
        </Link>
        <ul className="nav-menu">
          {user && (
            <>
              <li className="nav-item">
                <Link href="/profile" className="nav-links">
                  Profile
                </Link>
              </li>
              <li className="nav-item">
                <Link href="/users" className="nav-links">
                  Users
                </Link>
              </li>
              <li className="nav-item">
                <Link href="/groups" className="nav-links">
                  Groups
                </Link>
              </li>
              <li className="nav-item">
                <Link href="/chat" className="nav-links">
                  Chat
                  {unreadMessages > 0 && (
                    <span className="notification-badge">{unreadMessages}</span>
                  )}
                </Link>
              </li>
              <li className="nav-item">
                <Link href="/notifications" className="nav-links">
                  Notifications
                  {unreadNotifications > 0 && (
                    <span className="notification-badge">{unreadNotifications}</span>
                  )}
                </Link>
              </li>
              <li className="nav-item">
                <button onClick={handleLogout} className="nav-links-btn">
                  Logout
                </button>
              </li>
            </>
          )}
        </ul>
      </div>
    </nav>
  );
};

export default Navbar;
