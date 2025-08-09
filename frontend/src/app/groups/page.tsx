"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";

const GroupsPage = () => {
  const [groups, setGroups] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const router = useRouter();

  const fetchGroups = async () => {
    try {
      const res = await fetch("http://localhost:8080/api/groups", {
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        setGroups(data);
      } else {
        const errorText = await res.text();
        setError(`Failed to fetch groups: ${errorText}`);
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchGroups();
  }, []);

  const handleCreateGroup = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const title = formData.get("title");
    const description = formData.get("description");

    try {
      const res = await fetch(`http://localhost:8080/api/groups`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title, description }),
        credentials: "include",
      });

      if (res.ok) {
        form.reset();
        fetchGroups(); // Refresh groups
      } else {
        const errorText = await res.text();
        setError(`Failed to create group: ${errorText}`);
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="groups-container">
      <h1>Groups</h1>
      <div className="create-group">
        <h2>Create a Group</h2>
        <form onSubmit={handleCreateGroup}>
          <input type="text" name="title" placeholder="Group Title" required />
          <textarea
            name="description"
            placeholder="Group Description"
            required
          ></textarea>
          <button type="submit">Create Group</button>
        </form>
      </div>
      <div className="groups-list">
        <h2>All Groups</h2>
        {groups && groups.length > 0 ? (
          groups.map((group) => (
            <div key={group.id} className="group-item">
              <Link href={`/groups/${group.id}`}>
                <h3>{group.title}</h3>
              </Link>
              <p>{group.description}</p>
            </div>
          ))
        ) : (
          <p>No groups found.</p>
        )}
      </div>
    </div>
  );
};

export default GroupsPage;
