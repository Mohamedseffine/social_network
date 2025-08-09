"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

const GroupPage = ({ params }: { params: { id: string } }) => {
  const [group, setGroup] = useState<any>(null);
  const [posts, setPosts] = useState<any[]>([]);
  const [events, setEvents] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const router = useRouter();
  const { id } = params;

  const fetchGroupData = async () => {
    setLoading(true);
    try {
      const groupRes = await fetch(`http://localhost:8080/api/groups/${id}`, {
        credentials: "include",
      });
      if (groupRes.ok) {
        const groupData = await groupRes.json();
        setGroup(groupData);
      } else {
        const errorText = await groupRes.text();
        setError(`Failed to fetch group data: ${errorText}`);
      }

      const postsRes = await fetch(
        `http://localhost:8080/api/groups/${id}/posts`,
        {
          credentials: "include",
        }
      );
      if (postsRes.ok) {
        const postsData = await postsRes.json();
        setPosts(postsData);
      } else {
        const errorText = await postsRes.text();
        setError((prev) => `${prev}\nFailed to fetch group posts: ${errorText}`);
      }

      const eventsRes = await fetch(
        `http://localhost:8080/api/groups/${id}/events`,
        {
          credentials: "include",
        }
      );
      if (eventsRes.ok) {
        const eventsData = await eventsRes.json();
        setEvents(eventsData);
      } else {
        const errorText = await eventsRes.text();
        setError((prev) => `${prev}\nFailed to fetch group events: ${errorText}`);
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (id) {
      fetchGroupData();
    }
  }, [id]);

  const handleCreatePost = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const content = formData.get("content");

    try {
      const res = await fetch(
        `http://localhost:8080/api/groups/${id}/posts`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ content }),
          credentials: "include",
        }
      );

      if (res.ok) {
        form.reset();
        fetchGroupData(); // Refresh data
      } else {
        const errorText = await res.text();
        setError(`Failed to create group post: ${errorText}`);
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

  const handleCreateEvent = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const title = formData.get("title");
    const description = formData.get("description");
    const event_time = formData.get("event_time");

    try {
      const res = await fetch(
        `http://localhost:8080/api/groups/${id}/events`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ title, description, event_time }),
          credentials: "include",
        }
      );

      if (res.ok) {
        form.reset();
        fetchGroupData(); // Refresh data
      } else {
        const errorText = await res.text();
        setError(`Failed to create event: ${errorText}`);
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

  if (!group) {
    return <div>No group data found.</div>;
  }

  return (
    <div className="group-container">
      <h1>{group.title}</h1>
      <p>{group.description}</p>

      <div className="group-content">
        <div className="group-posts">
          <h2>Posts</h2>
          <div className="create-group-post">
            <h3>Create a Post</h3>
            <form onSubmit={handleCreatePost}>
              <textarea
                name="content"
                placeholder="What's on your mind?"
                required
              ></textarea>
              <button type="submit">Post</button>
            </form>
          </div>
          {posts && posts.length > 0 ? (
            posts.map((post) => (
              <div key={post.id} className="post">
                <p>{post.content}</p>
                <small>{new Date(post.created_at).toLocaleString()}</small>
              </div>
            ))
          ) : (
            <p>No posts in this group yet.</p>
          )}
        </div>

        <div className="group-events">
          <h2>Events</h2>
          <div className="create-event">
            <h3>Create an Event</h3>
            <form onSubmit={handleCreateEvent}>
              <input type="text" name="title" placeholder="Event Title" required />
              <textarea
                name="description"
                placeholder="Event Description"
                required
              ></textarea>
              <input type="datetime-local" name="event_time" required />
              <button type="submit">Create Event</button>
            </form>
          </div>
          {events && events.length > 0 ? (
            events.map((event) => (
              <div key={event.id} className="event">
                <h3>{event.title}</h3>
                <p>{event.description}</p>
                <small>
                  When: {new Date(event.event_time).toLocaleString()}
                </small>
              </div>
            ))
          ) : (
            <p>No events in this group yet.</p>
          )}
        </div>
      </div>
    </div>
  );
};

export default GroupPage;
