"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useAuth } from "../../context/AuthContext";
import { API_BASE_URL, getImageUrl } from "../../utils/api";

const UsersPage = () => {
  const [users, setUsers] = useState<any[]>([]);
  const [filteredUsers, setFilteredUsers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const { user: currentUser } = useAuth();

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/users`, {
          credentials: "include",
        });
        if (res.ok) {
          const data = await res.json();
          setUsers(data);
          setFilteredUsers(data);
        } else {
          setError("Failed to fetch users.");
        }
      } catch (err) {
        setError("An error occurred.");
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, []);

  useEffect(() => {
    const results = users.filter(user =>
      `${user.first_name} ${user.last_name}`.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (user.nickname && user.nickname.toLowerCase().includes(searchTerm.toLowerCase()))
    );
    setFilteredUsers(results);
  }, [searchTerm, users]);

  if (loading) {
    return <div>Loading users...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="users-container">
      <h1>Find Users</h1>
      <div className="search-bar">
        <input
          type="text"
          placeholder="Search for users..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
        />
      </div>
      <div className="users-list">
        {filteredUsers.map((user) => {
          if (currentUser && user.id === currentUser.id) {
            return null; // Don't show current user in the list
          }
          return (
            <div key={user.id} className="user-item card">
              <img
                src={getImageUrl(user.avatar)}
                alt="Avatar"
                className="user-avatar-small"
              />
              <div className="user-info">
                <Link href={`/users/${user.id}`}>
                  <h3>{user.first_name} {user.last_name}</h3>
                </Link>
                {user.nickname && <p>@{user.nickname}</p>}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default UsersPage;
