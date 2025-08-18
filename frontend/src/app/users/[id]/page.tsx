"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "../../../context/AuthContext";
import { usePopup } from "../../../context/PopupContext";
import { API_BASE_URL, getImageUrl } from "../../../utils/api";
import  PostCard  from "../../postcard";

const UserProfilePage = ({ params }: { params: { id: string } }) => {
  const { user: currentUser } = useAuth();
  const { showPopup } = usePopup();
  const [profileUser, setProfileUser] = useState<any>(null);
  const [posts, setPosts] = useState<any[]>([]);
  const [followers, setFollowers] = useState<any[]>([]);
  const [following, setFollowing] = useState<any[]>([]);
  const [activePopover, setActivePopover] = useState<'followers' | 'following' | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [followStatus, setFollowStatus] = useState("");
  const router = useRouter();
  const { id } = params;
    const {user} = useAuth()


  const fetchUserData = useCallback(async () => {
    setIsLoading(true);
    setError("");
    try {
      // Fetch user data first to check for privacy
      const userRes = await fetch(`${API_BASE_URL}/users/${id}`, {
        credentials: "include",
      });
      if (!userRes.ok) throw new Error("Failed to fetch user data.");
      if (userRes.status == 401){
                  router.push("/")
      }
      const userData = await userRes.json();
      setProfileUser(userData);
      setFollowStatus(userData.follow_status || "not_following");

      // If the profile is public/accessible, fetch all other data
      if (userData.email) {
        const [postsRes, followersRes, followingRes] = await Promise.all([
          fetch(`${API_BASE_URL}/users/${id}/posts`, { credentials: "include" }),
          fetch(`${API_BASE_URL}/users/${id}/followers`, { credentials: "include" }),
          fetch(`${API_BASE_URL}/users/${id}/following`, { credentials: "include" }),
        ]);

        if (postsRes.ok) setPosts(await postsRes.json() || []);
        else setError(prev => `${prev}Failed to fetch posts. `);
        if (postsRes.status == 401 ){
                    router.push("/")

        }

        if (followersRes.ok) setFollowers(await followersRes.json() || []);
        else setError(prev => `${prev}Failed to fetch followers. `);
        if (followersRes.status == 401){
                    router.push("/")

        }

        if (followingRes.ok) setFollowing(await followingRes.json() || []);
        else setError(prev => `${prev}Failed to fetch following. `);
        if (followingRes.status == 401){
                    router.push("/")

        }
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  }, [id]);

  useEffect(() => {
    if (currentUser && currentUser.id.toString() === id) {
      router.push("/profile");
      return;
    }
    if (id) {
      fetchUserData();
    }
  }, [id, currentUser, router, fetchUserData]);

  const handleFollowAction = async () => {
    const action = (followStatus === 'accepted' || followStatus === 'pending') ? 'unfollow' : 'follow';
    try {
      const res = await fetch(`${API_BASE_URL}/users/${id}/${action}`, {
        method: "POST",
        credentials: "include",
      });
      if (!res.ok) {
        showPopup(`Request failed: ${await res.text()}`, 'error');
      } else {
        // Refetch user data to get the latest follow status
        fetchUserData();
      if (res.status == 401){
                  router.push("/")

      }
      }
    } catch (err: any) {
      showPopup(`An error occurred: ${err.message}`, 'error');
    }
  };

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  if (!profileUser) {
    return <div>User not found.</div>;
  }

  const getButtonText = () => {
    switch (followStatus) {
      case "accepted":
        return "Unfollow";
      case "pending":
        return "Pending";
      default:
        return "Follow";
    }
  };
  return (
  user != null &&( <div className="profile-container">
      <h1>{profileUser.nickname || `${profileUser.first_name} ${profileUser.last_name}`}</h1>
      {followStatus !== "is_self" && (
        <button onClick={handleFollowAction} disabled={isLoading}>
          {getButtonText()}
        </button>
      )}
      <div className="profile-header">
        <img src={getImageUrl(profileUser.avatar)} alt="Avatar" className="profile-avatar" />
        <div className="profile-info">
          {profileUser.email ? (
            <>
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
              <p><strong>Email:</strong> {profileUser.email}</p>
              <p><strong>Date of Birth:</strong> {profileUser.date_of_birth}</p>
              {profileUser.about_me && <p><strong>About Me:</strong> {profileUser.about_me}</p>}
            </>
          ) : (
            <p>This profile is private.</p>
          )}
        </div>
      </div>

      <div className="profile-content">
        <h2>Posts</h2>
        {profileUser.email ? (
            posts.length > 0 ? (
            posts.map((post) => (
               <PostCard key={`${post.privacy}-${post.id}`} post={post} />
            ))
            ) : (
            <p>This user has no posts.</p>
            )
        ) : (
            <p>This profile is private. Follow them to see their posts.</p>
        )}
      </div>
    </div>)
  );
};

export default UserProfilePage;
