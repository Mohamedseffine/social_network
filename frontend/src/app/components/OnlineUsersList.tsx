"use client";

import { useAuth } from "../../context/AuthContext";
import Link from "next/link";
import "./OnlineUsersList.css";

const OnlineUsersList = () => {
  const { onlineUsers, user } = useAuth();

  // Filter out the current user from the list
  const otherOnlineUsers = onlineUsers.filter(onlineUser => onlineUser.id !== user?.id);

  return (
    <div className="online-users-container card">
      <h2>Online Users ({otherOnlineUsers.length})</h2>
      {otherOnlineUsers.length > 0 ? (
        <ul>
          {otherOnlineUsers.map((onlineUser) => (
            <li key={onlineUser.id}>
              <Link href={`/users/${onlineUser.id}`}>
                {onlineUser.first_name} {onlineUser.last_name}
              </Link>
            </li>
          ))}
        </ul>
      ) : (
        <p>No other users are currently online.</p>
      )}
    </div>
  );
};

export default OnlineUsersList;
