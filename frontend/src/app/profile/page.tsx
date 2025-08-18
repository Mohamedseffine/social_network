"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "../../context/AuthContext";
import { API_BASE_URL, getImageUrl } from "../../utils/api";
import  PostCard  from "../postcard";

const ProfilePage = () => {
  const { user, isLoading } = useAuth();
  const router = useRouter();
  const [posts, setPosts] = useState<any[]>([]);
  const [followers, setFollowers] = useState<any[]>([]);
  const [following, setFollowing] = useState<any[]>([]);
  const [error, setError] = useState("");
  const [isPageLoading, setIsPageLoading] = useState(true);
  const [activePopover, setActivePopover] = useState<'followers' | 'following' | null>(null);
  const [isPrivate, setIsPrivate] = useState<boolean>()
  
  
  useEffect(() => {
    if (isLoading) return; // Wait until session check is complete

    if (!user) {
      router.push('/'); // Redirect to login if not authenticated
      return;
    }

    const fetchProfileData = async () => {
      setIsPageLoading(true);
      setError("");
      try {
        const [postsRes, followersRes, followingRes] = await Promise.all([
          fetch(`${API_BASE_URL}/users/${user.id}/posts`, { credentials: "include" }),
          fetch(`${API_BASE_URL}/users/${user.id}/followers`, { credentials: "include" }),
          fetch(`${API_BASE_URL}/users/${user.id}/following`, { credentials: "include" }),
        ]);

        if (postsRes.ok) setPosts(await postsRes.json() || []);
        else setError(prev => `${prev}Failed to fetch posts. `);
        if (postsRes.status == 401){
                    router.push("/")
        }

        if (followersRes.ok) setFollowers(await followersRes.json() || []);
        else setError(prev => `${prev}Failed to fetch followers. `);
        if (followersRes.status == 401){
                    router.push("/")

        }

        if (followingRes.ok) setFollowing(await followingRes.json() || []);
        else setError(prev => `${prev}Failed to fetch following. `);
        if (followingRes.status == 401 ){
                    router.push("/")

        }

      } catch (err: any) {
        setError(`An error occurred: ${err.message}`);
      } finally {
        setIsPrivate(user.profile_is_public)
        setIsPageLoading(false);
      }
    };
    fetchProfileData();
  }, [user, isLoading, router]);

  const handleTogglePrivacy = async () => {
    if (!user) return;
    try {
      const res = await fetch(`${API_BASE_URL}/profile/privacy`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ is_public: !user.profile_is_public }),
        credentials: "include",
      });
      if (res.ok) {
        setIsPrivate(!isPrivate)
      } else {
        setError(`Failed to update privacy: ${await res.text()}`);
        if (res.status == 401){
                    router.push("/")

        }
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

  if (isLoading || !user) {
    // Show a loading indicator while the session is being checked, or if redirecting.
    return <div className="loading-container">Loading...</div>;
  }

  return (
    <div className="profile-container">
      <div className="profile-header">
        <img src={getImageUrl(user.avatar)} alt="Avatar" className="profile-avatar" />
        <div className="profile-info">
          <div className="profile-title">
           <h1>
              {user.nickname ? (
                <>
                  {user.nickname}
                  <span style={{ fontWeight: 400, color: 'var(--text-muted)', marginLeft: '0.5rem' }}>
                    ({user.first_name} {user.last_name})
                  </span>
                </>
              ) : (
                `${user.first_name} ${user.last_name}`
              )}
            </h1>            <button onClick={handleTogglePrivacy} className="privacy-toggle-btn">
              Make {isPrivate ? "Private" : "Public"}
            </button>
          </div>
          <div className="profile-stats">
            <div className="stat-item"><strong>{posts.length}</strong> posts</div>
            <div 
              className="stat-item popover-container" 
              onMouseEnter={() => setActivePopover('followers')} 
              onMouseLeave={() => setActivePopover(null)}
            >
              <strong>{followers.length}</strong> followers
              {activePopover === 'followers' && (
                <div className="popover">
                  <div className="popover-user-list">
                    {followers.length > 0 ? followers.map((person: any) => (
                      <div key={person.id} className="user-item">
                        <img src={getImageUrl(person.avatar)} alt="User Avatar" className="user-avatar-small" />
                        <Link href={`/users/${person.id}`}>{person.first_name} {person.last_name}</Link>
                      </div>
                    )) : <p>No followers.</p>}
                  </div>
                </div>
              )}
            </div>
            <div 
              className="stat-item popover-container" 
              onMouseEnter={() => setActivePopover('following')} 
              onMouseLeave={() => setActivePopover(null)}
            >
              <strong>{following.length}</strong> following
              {activePopover === 'following' && (
                <div className="popover">
                   <div className="popover-user-list">
                    {following.length > 0 ? following.map((person: any) => (
                      <div key={person.id} className="user-item">
                        <img src={getImageUrl(person.avatar)} alt="User Avatar" className="user-avatar-small" />
                        <Link href={`/users/${person.id}`}>{person.first_name} {person.last_name}</Link>
                      </div>
                    )) : <p>Not following anyone.</p>}
                  </div>
                </div>
              )}
            </div>
          </div>
          <div className="profile-bio">
            {user.about_me && <p>{user.about_me}</p>}
            <p><strong>Email:</strong> {user.email}</p>
            <p><strong>Profile is:</strong> {user.profile_is_public ? "Public" : "Private"}</p>
          </div>
        </div>
      </div>

      <div className="profile-content">
        <hr className="divider" />
        {isPageLoading ? (
          <div>Loading posts...</div>
        ) : error ? (
          <div className="error">{error}</div>
        ) : (
          <div className="post-grid">
            {posts && posts.length > 0 ? (
              posts.map((post) => (
               <PostCard key={`${post.privacy}-${post.id}`} post={post}  />
              ))
            ) : (
              <p>No posts to display.</p>
            )}
          </div>
        )}
      </div>

    </div>
  );
};

export default ProfilePage;
