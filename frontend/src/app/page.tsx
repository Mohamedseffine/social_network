"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { useAuth } from "../context/AuthContext";
import { API_BASE_URL, getImageUrl } from "../utils/api";

const PostCard = ({ post }: { post: any }) => {
  return (
    <div className="post card">
      <div className="post-author">
        <img src={getImageUrl(post.AuthorAvatar)} alt="Author Avatar" className="user-avatar-small" />
        <Link href={`/users/${post.UserID}`}>
          <span>{post.AuthorFirstName} {post.AuthorLastName}</span>
        </Link>
      </div>
      <p>{post.Content}</p>
      {post.Image && (
        <img src={getImageUrl(post.Image)} alt="Post image" className="post-image" />
      )}
      <small>{new Date(post.CreatedAt).toLocaleString()}</small>
    </div>
  );
};

export default function Home() {
  const { user, login } = useAuth();
  const [showLogin, setShowLogin] = useState(true);
  const [message, setMessage] = useState("");
  const [isError, setIsError] = useState(false);

  // State for the feed
  const [posts, setPosts] = useState<any[]>([]);
  const [page, setPage] = useState(1);
  const [loadingFeed, setLoadingFeed] = useState(false);
  const [hasMorePosts, setHasMorePosts] = useState(true);

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
    }
  }, [user]);

  const handleRegister = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const data = Object.fromEntries(formData.entries());
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
          return;
        }
      } catch (err: any) {
        showMessage(`Avatar upload failed: ${err.message}`, true);
        return;
      }
    }
    const registrationData = { ...data, avatar: avatarPath };
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
      }
    } catch (err: any) {
      showMessage(`Login failed: ${err.message}`, true);
    }
  };

  return (
    <main>
      <h1>Social Dilemma</h1>
      <div className="container">
        {user ? (
          <div id="feed-container">
            <h2>Your Feed</h2>
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
        ) : (
          <div className="auth-container">
            {showLogin ? (
              <div id="login-form" className="form-container">
                <h2>Login</h2>
                <form id="login" onSubmit={handleLogin}>
                  <input type="email" name="email" placeholder="Email" required />
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
