"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { useAuth } from "@/context/AuthContext";
import { API_BASE_URL, getImageUrl } from "@/utils/api";
import PostCard from "@/app/postcard"
import { Comment, CommentCard } from "./components/Comment";
import { useRouter } from "next/navigation";



export default function Home() {
  const { user, login, isLoading } = useAuth();
  const [showLogin, setShowLogin] = useState(true);
  const [message, setMessage] = useState("");
  const [isError, setIsError] = useState(false);

  // State for the feed
  const [posts, setPosts] = useState<any[]>([]);
  const [page, setPage] = useState(1);
  const [loadingFeed, setLoadingFeed] = useState(false);
  const [hasMorePosts, setHasMorePosts] = useState(true);

  // State for Create Post form
  const [postPrivacy, setPostPrivacy] = useState('public');
  const [followers, setFollowers] = useState<any[]>([]);
  const [selectedFollowers, setSelectedFollowers] = useState<number[]>([]);
  const router = useRouter()
  const showMessage = (msg: string, error: boolean = false) => {
    setMessage(msg);
    setIsError(error);
    setTimeout(() => {
      setMessage("");
    }, 3000);
  };

  const fetchFeedPosts = async (isNewLoad: boolean = false) => {
    if (loadingFeed || (!hasMorePosts && !isNewLoad)) return;

    setLoadingFeed(true);
    const currentPage = isNewLoad ? 1 : page;

    try {
      const res = await fetch(`${API_BASE_URL}/feed?page=${currentPage}`, {
        credentials: "include",
      });
      if (res.ok) {
        const newPosts = await res.json();
        if (newPosts === null || newPosts.length === 0) {
          setHasMorePosts(false);
        } else {
          setPosts(prev => isNewLoad ? newPosts : [...prev, ...newPosts]);
          setPage(currentPage + 1);
          if (res.status == 401){
                      router.push("/")

          }
        }
      }
    } catch (err) {
      console.error("Failed to fetch feed", err);
    } finally {
      setLoadingFeed(false);
    }
  };

  useEffect(() => {
    if (user) {
      fetchFeedPosts(true);
      fetchFollowers();
    }
  }, [user]);

  const fetchFollowers = async () => {
    if (!user) return;
    try {
      const res = await fetch(`${API_BASE_URL}/users/${user.id}/followers`, { credentials: 'include' });
      if (res.ok) {
        const data = await res.json();
        setFollowers(data || []);
      } else {
        if (res.status == 401){
                    router.push("/")

        }
        console.error("Failed to fetch followers");
      }
    } catch (err) {
      console.error("Error fetching followers", err);
    }
  };

  const handleFollowerSelection = (e: React.ChangeEvent<HTMLInputElement>) => {
    const followerId = parseInt(e.target.value, 10);
    setSelectedFollowers(prev =>
      e.target.checked ? [...prev, followerId] : prev.filter(id => id !== followerId)
    );
  };

  const handleCreatePost = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const content = formData.get("content");
    const privacy = formData.get("privacy");
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

    const postData: any = { content, privacy, image: imagePath };
    if (postPrivacy === 'private') {
      postData.viewer_ids = selectedFollowers;
    }

    try {
      const res = await fetch(`${API_BASE_URL}/posts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(postData),
        credentials: "include",
      });
      if (res.ok) {
        form.reset();
        setPostPrivacy('public');
        setSelectedFollowers([]);
        fetchFeedPosts(true); // Refetch feed posts
      } else {
        showMessage(`Failed to create post: ${await res.text()}`, true);
        if (res.status == 401){
                    router.push("/")

        }
      }
    } catch (err: any) {
      showMessage(`An error occurred: ${err.message}`, true);
    }
  };

  const handleRegister = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const avatarFile = formData.get("avatar") as File;

    let avatarPath = "";
    if (avatarFile && avatarFile.size > 0) {
      try {
        const uploadFormData = new FormData();
        uploadFormData.append("image", avatarFile);
        const uploadRes = await fetch(`${API_BASE_URL}/upload`, {
          method: "POST", body: uploadFormData, credentials: "include",
        });
        if (uploadRes.ok) {
          avatarPath = (await uploadRes.json()).path;
        } else {
          showMessage(`Avatar upload failed: ${await uploadRes.text()}`, true);
          if (uploadRes.status == 401){
                      router.push("/")

          }
          return;
        }
      } catch (err: any) {
        showMessage(`Avatar upload failed: ${err.message}`, true);
        return;
      }
    }

    const registrationData = {
      first_name: formData.get("first_name"),
      last_name: formData.get("last_name"),
      email: formData.get("email"),
      password: formData.get("password"),
      date_of_birth: formData.get("date_of_birth"),
      avatar: avatarPath,
      nickname: formData.get("nickname"),
      about_me: formData.get("about_me"),
    };
    
    try {
      const res = await fetch(`${API_BASE_URL}/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(registrationData),
        credentials: "include",
      });
      if (res.ok) {
        showMessage("Registration successful! Please log in.");
        setShowLogin(true);
        form.reset();
      } else {
        showMessage(`Registration failed: ${await res.text()}`, true);
        if (res.status == 401){
                    router.push("/")

        }
      }
    } catch (err: any) {
      showMessage(`Registration failed: ${err.message}`, true);
    }
  };

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const data = Object.fromEntries(formData.entries());
    try {
      const res = await fetch(`${API_BASE_URL}/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
        credentials: "include",
      });
      if (res.ok) {
        login(await res.json());
      } else {
        showMessage(`Login failed: ${await res.text()}`, true);
        if (res.status == 401){
                    router.push("/")

        }
      }
    } catch (err: any) {
      showMessage(`Login failed: ${err.message}`, true);
    }
  };

  if (isLoading) {
    return <div className="loading-container">Loading...</div>;
  }

  return (
    <main>
      <div className="container">
        {user ? (
          <div className="main-content-layout">
            <div className="feed-and-posts">
              <div className="create-post card">
                <h3>Create a Post</h3>
                <form onSubmit={handleCreatePost}>
                  <textarea name="content" placeholder="What's on your mind?" required></textarea>
                  <input type="file" name="image" accept="image/*" />
                  <div className="privacy-options">
                    <label htmlFor="privacy">Privacy:</label>
                    <select name="privacy" value={postPrivacy} onChange={(e) => setPostPrivacy(e.target.value)}>
                      <option value="public">Public</option>
                      <option value="almost private">Followers Only</option>
                      <option value="private">Specific Followers</option>
                    </select>
                  </div>

                  {postPrivacy === 'private' && (
                    <div className="followers-selection">
                      <h4>Select followers who can see this post:</h4>
                      <div className="followers-list-container">
                        {followers.length > 0 ? (
                          followers.map(follower => (
                            <div key={follower.id} className="follower-item">
                              <input
                                type="checkbox"
                                id={`follower-${follower.id}`}
                                value={follower.id}
                                onChange={handleFollowerSelection}
                                checked={selectedFollowers.includes(follower.id)}
                              />
                              <label htmlFor={`follower-${follower.id}`}>{follower.first_name} {follower.last_name}</label>
                            </div>
                          ))
                        ) : (
                          <p>You have no followers to select.</p>
                        )}
                      </div>
                    </div>
                  )}

                  <button type="submit">Post</button>
                </form>
              </div>
              <div className="posts-list">
                {posts.map(post => <PostCard key={`${post.privacy}-${post.id}`} post={post} />)}
              </div>
              {loadingFeed && <div>Loading more posts...</div>}
              {!loadingFeed && hasMorePosts && (
                <button onClick={() => fetchFeedPosts()} className="load-more-btn">Load More</button>
              )}
              {!loadingFeed && !hasMorePosts && posts.length === 0 && <p>No posts in your feed yet. Follow some people or join some groups!</p>}
              {!loadingFeed && !hasMorePosts && posts.length > 0 && <p>You've reached the end of the feed.</p>}
            </div>
          </div>
        ) : (
          <div className="auth-container">
            {showLogin ? (
              <div id="login-form" className="form-container">
                <h2>Login</h2>
                <form id="login" onSubmit={handleLogin}>
                  <input type="text" name="identifier" placeholder="Email or Nickname" required />
                  <input type="password" name="password" placeholder="Password" required />
                  <button type="submit">Login</button>
                </form>
                <p>
                  Don't have an account?{" "}
                  <button onClick={() => setShowLogin(false)} className="toggle-btn">Register</button>
                </p>
              </div>
            ) : (
              <div id="register-form" className="form-container">
                <h2>Register</h2>
                <form id="register" onSubmit={handleRegister}>
                  <input type="text" name="first_name" placeholder="First Name" required />
                  <input type="text" name="last_name" placeholder="Last Name" required />
                  <input type="email" name="email" placeholder="Email" required />
                  <input type="password" name="password" placeholder="Password" required />
                  <input type="date" name="date_of_birth" placeholder="Date of Birth" required />
                  <input type="text" name="nickname" placeholder="Nickname (Optional)" />
                  <textarea name="about_me" placeholder="About Me (Optional)"></textarea>
                  <input type="file" name="avatar" accept="image/*" />
                  <button type="submit">Register</button>
                </form>
                <p>
                  Already have an account?{" "}
                  <button onClick={() => setShowLogin(true)} className="toggle-btn">Login</button>
                </p>
              </div>
            )}
          </div>
        )}
        {message && (
          <div id="message" className={isError ? "error" : "success"}>
            {message}
          </div>
        )}
      </div>
    </main>
  );
}
