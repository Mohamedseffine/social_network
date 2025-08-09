"use client";

import { useEffect, useState } from "react";
import { useAuth } from "../../context/AuthContext";

const Notifications = () => {
  const [notifications, setNotifications] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const { fetchNotifications: fetchUnreadCount } = useAuth();

  const fetchNotificationsData = async () => {
    setLoading(true);
    try {
      const res = await fetch("http://localhost:8080/api/notifications", {
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        setNotifications(data);
        fetchUnreadCount(); // Update the badge count in the navbar
      } else {
        const errorText = await res.text();
        setError(`Failed to fetch notifications: ${errorText}`);
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNotificationsData();
  }, []);

  const handleAction = async (url: string) => {
    try {
      const res = await fetch(url, {
        method: "POST",
        credentials: "include",
      });

      if (res.ok) {
        fetchNotificationsData(); // Refresh notifications and the count
      } else {
        const errorText = await res.text();
        alert(`Action failed: ${errorText}`);
      }
    } catch (err: any) {
      alert(`An error occurred: ${err.message}`);
    }
  };

  const renderActions = (notification: any) => {
    if (notification.is_read) return null;

    let acceptUrl = "";
    let declineUrl = "";

    switch (notification.type) {
      case "follow_request":
        acceptUrl = `http://localhost:8080/api/requests/${notification.related_id}/accept`;
        declineUrl = `http://localhost:8080/api/requests/${notification.related_id}/decline`;
        break;
      case "group_invite":
        acceptUrl = `http://localhost:8080/api/groups/invites/${notification.related_id}/accept`;
        declineUrl = `http://localhost:8080/api/groups/invites/${notification.related_id}/decline`;
        break;
      case "group_join_request":
        acceptUrl = `http://localhost:8080/api/groups/requests/${notification.related_id}/accept`;
        declineUrl = `http://localhost:8080/api/groups/requests/${notification.related_id}/decline`;
        break;
      default:
        return null;
    }

    return (
      <div className="notification-actions">
        <button onClick={() => handleAction(acceptUrl)} className="btn-accept">Accept</button>
        <button onClick={() => handleAction(declineUrl)} className="btn-decline">Decline</button>
      </div>
    );
  };


  if (loading) {
    return <div>Loading notifications...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="notifications-container">
      <h2>Notifications</h2>
      {notifications && notifications.length > 0 ? (
        notifications.map((notification) => (
          <div
            key={notification.id}
            className={`notification-item card ${
              notification.is_read ? "read" : "unread"
            }`}
          >
            <p>{notification.message}</p>
            {renderActions(notification)}
          </div>
        ))
      ) : (
        <p>No notifications.</p>
      )}
    </div>
  );
};

export default Notifications;
