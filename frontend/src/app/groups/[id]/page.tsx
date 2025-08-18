"use client";
import "../../style.css"
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { API_BASE_URL, getImageUrl } from "../../../utils/api";
import { usePopup } from "../../../context/PopupContext";
import { Comment, CommentCard } from "../../components/Comment";

const GroupPage = ({ params }: { params: { id: string } }) => {
  const { showPopup } = usePopup();
  const [group, setGroup] = useState<any>(null);
  const [membershipStatus, setMembershipStatus] = useState<string | null>(null);
  const [posts, setPosts] = useState<any[]>([]);
  const [events, setEvents] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const router = useRouter();
  const { id } = params;

  // State for create form toggle
  const [formType, setFormType] = useState('post');

  // State for invite user search
  const [searchTerm, setSearchTerm] = useState("");
  const [searchResults, setSearchResults] = useState<any[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  // Debounce search term
  useEffect(() => {
    if (searchTerm.trim() === "") {
      setSearchResults([]);
      return;
    }

    setIsSearching(true);
    const delayDebounceFn = setTimeout(async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/users?q=${searchTerm}`, {
          credentials: "include",
        });
        if (res.ok) {
          const data = await res.json();
          setSearchResults(data);
        } else {
          if (res.status == 401){
                    router.push("/")
          }
          console.error("Failed to search users");
        }
      } catch (err) {
        console.error("User search error:", err);
      } finally {
        setIsSearching(false);
      }
    }, 500); // 500ms delay

    return () => clearTimeout(delayDebounceFn);
  }, [searchTerm]);

  const fetchMemberContent = async () => {
    try {
      const postsRes = await fetch(`${API_BASE_URL}/groups/${id}/posts`, {
        credentials: "include",
      });
      if (postsRes.ok) setPosts(await postsRes.json());
      if (postsRes.status == 401){
                  router.push("/")
      }
      const eventsRes = await fetch(`${API_BASE_URL}/groups/${id}/events`, {
        credentials: "include",
      });
      if (eventsRes.ok) setEvents(await eventsRes.json());
      if (eventsRes.status == 401){
                  router.push("/")
      }
    } catch (err) {
      console.error("Failed to fetch member content", err);
      setError("Failed to load group content.");
    }
  };

  const [comments, setComments] = useState<{ [key: number]: any[] }>({});
  const [visibleComments, setVisibleComments] = useState<number | null>(null);
  const [message, setMessage] = useState("");
  const [isError, setIsError] = useState(false);
  const showMessage = (msg: string, error: boolean = false) => {
    setMessage(msg);
    setIsError(error);
    setTimeout(() => {
      setMessage("");
    }, 3000);
  };
  const handleCreateComment = async (e: React.FormEvent<HTMLFormElement>, postId: number) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const content = formData.get("content") as string;
    const imageFile = formData.get("image") as File;

    let imagePath = "";
    if (imageFile && imageFile.size > 0) {
      try {
        const uploadFormData = new FormData();
        uploadFormData.append("image", imageFile);
        const uploadRes = await fetch(`${API_BASE_URL}/upload`, {
          method: "POST", body: uploadFormData, credentials: "include",
        });
        if (uploadRes.ok) {
          imagePath = (await uploadRes.json()).path;
        } else {
          showMessage(`Image upload failed: ${await uploadRes.text()}`, true);
          if (uploadRes.status == 401){
                      router.push("/")
          }
          return;
        }
      } catch (err: any) {
        showMessage(`Image upload failed: ${err.message}`, true);
        return;
      }
    }

    try {
      const res = await fetch(`${API_BASE_URL}/group-posts/${postId}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content, image : imagePath }),
        credentials: "include",
      });
      if (res.ok) {
        form.reset();
        fetchComments(postId); // Refetch comments to show the new one
      } else {
        setError(`Failed to create comment: ${await res.text()}`);
        if (res.status == 401){
                    router.push("/")

        }
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

  const fetchComments = async (postId: number) => {
    try {
      const res = await fetch(`${API_BASE_URL}/group-posts/${postId}/comments`, {
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        setComments(prev => ({ ...prev, [postId]: data || [] }));
      } else {
        console.error("Failed to fetch comments");
        if (res.status == 401){
                    router.push("/")

        }
      }
    } catch (err) {
      console.error("Error fetching comments", err);
    }
  };

  const toggleComments = (postId: number) => {
    if (visibleComments === postId) {
      setVisibleComments(null); // Hide if already visible
    } else {
      setVisibleComments(postId);
      if (!comments[postId]) { // Fetch only if not already fetched
        fetchComments(postId);
      }
    }
  };

  const [attendees, setAttendees] = useState<{ [key: number]: any[] }>({});
  const [visibleAttendees, setVisibleAttendees] = useState<number | null>(null);

  const handleRsvp = async (eventId: number, status: 'going' | 'not_going') => {
    try {
      const res = await fetch(`${API_BASE_URL}/events/${eventId}/respond`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status }),
        credentials: "include",
      });
      if (res.ok) {
        showPopup(`You are now marked as ${status}`, 'success');
        // If attendees for this event are visible, refetch them
        if (visibleAttendees === eventId) {
          fetchAttendees(eventId);
        }
      } else {
        setError(`Failed to RSVP: ${await res.text()}`);
        if (res.status == 401){
                    router.push("/")

        }
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

  const fetchAttendees = async (eventId: number) => {
    try {
      const res = await fetch(`${API_BASE_URL}/events/${eventId}/attendees`, {
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        setAttendees(prev => ({ ...prev, [eventId]: data || [] }));
      } else {
        console.error("Failed to fetch attendees");
      }
    } catch (err) {
      console.error("Error fetching attendees", err);
    }
  };

  const toggleAttendees = (eventId: number) => {
    if (visibleAttendees === eventId) {
      setVisibleAttendees(null); // Hide if already visible
    } else {
      setVisibleAttendees(eventId);
      if (!attendees[eventId]) { // Fetch only if not already fetched
        fetchAttendees(eventId);
      }
    }
  };

  useEffect(() => {
    if (!id) return;

    const fetchAllData = async () => {
      setLoading(true);
      setError("");
      try {
        const groupRes = await fetch(`${API_BASE_URL}/groups/${id}`, {
          credentials: "include",
        });
        if (!groupRes.ok) {
          throw new Error("Failed to fetch group data");
        }
        const groupData = await groupRes.json();
        setGroup(groupData);

        const statusRes = await fetch(`${API_BASE_URL}/groups/${id}/membership`, {
          credentials: "include",
        });
        if (!statusRes.ok) {
          throw new Error("Failed to fetch membership status");
        }
        const statusData = await statusRes.json();
        setMembershipStatus(statusData.status);

        if (statusData.status === 'accepted') {
          await fetchMemberContent();
        }
      } catch (err: any) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchAllData();
  }, [id]);

  const handleRequestToJoin = async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/groups/${id}/join`, {
        method: "POST",
        credentials: "include",
      });
      if (res.ok) {
        showPopup("Request to join sent successfully!", 'success');
        setMembershipStatus("pending");
      } else {
        const errorText = await res.text();
        setError(`Failed to send join request: ${errorText}`);
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

  const handleInviteUser = async (userIdToInvite: number) => {
    try {
      const res = await fetch(`${API_BASE_URL}/groups/${id}/invite`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user_id: userIdToInvite }),
        credentials: "include",
      });

      if (res.ok) {
        showPopup("Invitation sent successfully!", 'success');
        setSearchTerm("");
        setSearchResults([]);
      } else {
        const errorText = await res.text();
        showPopup(`Failed to send invitation: ${errorText}`, 'error');
      }
    } catch (err: any) {
      showPopup(`An error occurred: ${err.message}`, 'error');
    }
  };

  const handleCreatePost = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const content = formData.get("content");
    const imageFile = formData.get("image") as File;
    let imagePath = "";
    if (imageFile && imageFile.size > 0) {
      try {
        const uploadFormData = new FormData();
        uploadFormData.append("image", imageFile);
        const uploadRes = await fetch(`${API_BASE_URL}/upload`, {
          method: "POST", body: uploadFormData, credentials: "include",
        });
        if (uploadRes.ok) {
          imagePath = (await uploadRes.json()).path;
          console.log(imagePath);

        } else {
          alert(`Image upload failed: ${await uploadRes.text()}`);
          return;
        }
      } catch (err: any) {
        alert(`Image upload failed: ${err.message}`);
        return;
      }
    }
    try {
      const res = await fetch(`${API_BASE_URL}/groups/${id}/posts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content, image: imagePath }),
        credentials: "include",
      });
      if (res.ok) {
        form.reset();
        fetchMemberContent();
      } else {
        setError(`Failed to create post: ${await res.text()}`);
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
    const event_time_str = formData.get("event_time") as string;

    if (!event_time_str) {
      setError("Event time is required.");
      return;
    }
    function isValidDate(dateString: string) {
      return !isNaN(Date.parse(dateString));
    }
    if (!isValidDate(event_time_str)) {
      alert("invalid date value")
      return
    }


    const event_time = new Date(event_time_str).toISOString();

    try {
      const res = await fetch(`${API_BASE_URL}/groups/${id}/events`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title, description, event_time }),
        credentials: "include",
      });
      if (res.ok) {
        form.reset();
        fetchMemberContent();
      } else {
        let errtxt = await res.text()
        if (errtxt.trim() === "Event time must be in the future".trim()) {
          alert(errtxt)
        } else {
          setError(`Failed to create event: ${errtxt}`);
        }
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

  const renderContent = () => {
    if (membershipStatus === 'accepted') {
      return (
        <>
          <div className="creation-forms card">
            <div className="form-toggle">
              <button onClick={() => setFormType('post')} className={formType === 'post' ? 'active' : ''}>Create Post</button>
              <button onClick={() => setFormType('event')} className={formType === 'event' ? 'active' : ''}>Create Event</button>
            </div>
            {formType === 'post' && (
              <div className="create-group-post">
                <h3>Create a Post in {group.title}</h3>
                <form onSubmit={handleCreatePost}>
                  <textarea name="content" placeholder="What's on your mind?" required />
                  <input type="file" name="image" accept="image/*" />
                  <button type="submit">Post</button>
                </form>
              </div>
            )}
            {formType === 'event' && (
              <div className="create-event">
                <h3>Create an Event in {group.title}</h3>
                <form onSubmit={handleCreateEvent}>
                  <input type="text" name="title" placeholder="Event Title" required />
                  <textarea name="description" placeholder="Event Description" required />
                  <input type="datetime-local" name="event_time" required />
                  <button type="submit">Create Event</button>
                </form>
              </div>
            )}
          </div>
          <div className="group-actions">
            <div className="invite-user">
              <h3>Invite a User</h3>
              <input
                type="text"
                placeholder="Search by name or nickname..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
              {isSearching && <div>Searching...</div>}
              <div className="search-results">
                {searchResults?.map((user) => (
                  <div key={user.id} className="search-result-item">
                    <span>{user.first_name} {user.last_name} (@{user.nickname})</span>
                    <button onClick={() => handleInviteUser(user.id)}>Invite</button>
                  </div>
                ))}
              </div>
            </div>
          </div>
          <div className="group-content">
            <div className="group-posts">
              <h2>Posts</h2>
              {posts && posts.length > 0 ? (
                posts.map((post) => (
                  <div key={post.id} className="post card">
                    <small>{new Date(post.created_at).toLocaleString()}</small>
                    <p>{post.content}</p>
                    {post.image && (
                      <img src={getImageUrl(post.image)} alt="Post image" className="post-image" />
                    )}
                    <div className="post-actions">
                      <button onClick={() => toggleComments(post.id)} className="toggle-comments-btn">
                        {visibleComments === post.id ? 'Hide' : 'View'} Comments
                      </button>
                    </div>
                    {visibleComments === post.id && (
                      <div className="comments-section">
                        <form onSubmit={(e) => handleCreateComment(e, post.id)} className="comment-form">
                          <textarea name="content" placeholder="Write a comment..." required />
                          <input type="file" name="image" accept="image/*" />
                          <button type="submit">Comment</button>
                        </form>
                        {message && (
                          <div id="message" className={isError ? "error" : "success"}>
                            {message}
                          </div>
                        )}
                        <div className="comments-list">
                          {comments[post.id] ? (
                            comments[post.id].map((comment: Comment) => (
                              <CommentCard key={comment.id} comment={comment} />
                            ))
                          ) : (
                            <p>Loading comments...</p>
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                ))
              ) : (
                <p>No posts in this group yet.</p>
              )}
            </div>
            <div className="group-events">
              <h2>Events</h2>
              {events && events.length > 0 ? (
                events.map((event) => (
                  <div key={event.id} className="event card">
                    <h3>{event.title}</h3>
                    <p>{event.description}</p>
                    <small>When: {new Date(event.event_time).toLocaleString()}</small>
                    <div className="event-actions">
                      <button onClick={() => handleRsvp(event.id, 'going')}>Going</button>
                      <button onClick={() => handleRsvp(event.id, 'not_going')}>Not Going</button>
                      <button onClick={() => toggleAttendees(event.id)}>
                        {visibleAttendees === event.id ? 'Hide RSVPs' : 'View RSVPs'}
                      </button>
                    </div>
                    {visibleAttendees === event.id && (
                      <div className="attendees-list">
                        <h4>Going:</h4>
                        <ul>
                          {attendees[event.id]?.filter(a => a.status === 'going').map(a => <li key={a.id}>{a.first_name} {a.last_name}</li>)}
                        </ul>
                        <h4>Not Going:</h4>
                        <ul>
                          {attendees[event.id]?.filter(a => a.status === 'not_going').map(a => <li key={a.id}>{a.first_name} {a.last_name}</li>)}
                        </ul>
                      </div>
                    )}
                  </div>
                ))
              ) : (
                <p>No events in this group yet.</p>
              )}
            </div>
          </div>
        </>
      );
    } else if (membershipStatus === 'pending') {
      return <div className="group-status-info">Your request to join this group is pending approval.</div>;
    } else if (membershipStatus === 'not_member') {
      return (
        <div className="group-actions">
          <div className="join-request">
            <p>You are not a member of this group.</p>
            <button onClick={handleRequestToJoin}>Request to Join Group</button>
          </div>
        </div>
      );
    }
    return null;
  };

  if (loading) return <div>Loading...</div>;
  if (error) return <div className="error">{error}</div>;
  if (!group) return <div>Group not found.</div>;

  return (
    <div className="group-container">
      <h1>{group.title}</h1>
      <p>{group.description}</p>
      {renderContent()}
    </div>
  );
};

export default GroupPage;
