"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { useAuth } from "@/context/AuthContext";
import { API_BASE_URL, getImageUrl } from "@/utils/api";


import { Comment, CommentCard } from "./components/Comment";
import { useRouter } from "next/navigation";
export default function PostCard ({ post }: { post: any }) {
  const [comments, setComments] = useState<Comment[]>([]);
  const [showComments, setShowComments] = useState(false);
  const [loadingComments, setLoadingComments] = useState(false);
  const [commentContent, setCommentContent] = useState("");
  const [message, setMessage] = useState("");
  const [isError, setIsError] = useState(false);
  const router = useRouter()
  const showMessage = (msg: string, error: boolean = false) => {
    setMessage(msg);
    setIsError(error);
    setTimeout(() => {
      setMessage("");
    }, 3000);
  };

  const fetchComments = async () => {
    if (loadingComments) return;
    setLoadingComments(true);
    try {
      const res = await fetch(`${API_BASE_URL}/posts/${post.id}/comments`, {
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        setComments(data || []);
      }else if (res.status == 401){
                  router.push("/")

      }
    } catch (err) {
      console.error("Failed to fetch comments", err);
    } finally {
      setLoadingComments(false);
    }
  };

  const handleToggleComments = () => {
    const newShowState = !showComments;
    setShowComments(newShowState);
    if (newShowState && comments.length === 0) {
      fetchComments();
    }
  };

  const handleCreateComment = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    let form = e.currentTarget
    if (!commentContent.trim()) return;
    let data = new FormData(form)
    const imageFile = data.get("image") as File;

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
      const res = await fetch(`${API_BASE_URL}/posts/${post.id}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content: commentContent, image: imagePath }),
        credentials: "include",
      });

      if (res.ok) {
        setCommentContent("");
        form.reset()
        fetchComments(); // Refresh comments
      } else {
        console.error("Failed to create comment");
        if (res.status ==401){
                    router.push("/")

        }
      }
    } catch (err) {
      console.error("An error occurred while creating the comment", err);
    }
  };

  return (
    <div className="post card">
      <div className="post-author">
        <img src={getImageUrl(post.author_avatar)} alt="Author Avatar" className="user-avatar-small" />
        <Link href={`/users/${post.user_id}`}>
          <span>{post.author_first_name} {post.author_last_name}</span>
        </Link>
      </div>
      <small>{new Date(post.created_at).toLocaleString()}</small>
      <p>{post.content}</p>
      {post.image && (
        <img src={getImageUrl(post.image)} alt="Post image" className="post-image" />
      )}
      <div className="post-actions">
        <button onClick={handleToggleComments} className="toggle-comments-btn">
          {showComments ? "Hide" : "View"} Comments
        </button>
      </div>

      {showComments && (
        <div className="comments-section">
          <form onSubmit={handleCreateComment} className="comment-form">
            <textarea
              value={commentContent}
              onChange={(e) => setCommentContent(e.target.value)}
              placeholder="Write a comment..."
              required
            />
            <input type="file" name="image" accept="image/*" />
            <button type="submit">Comment</button>
          </form>
          {message && (
            <div id="message" className={isError ? "error" : "success"}>
              {message}
            </div>
          )}
          {loadingComments ? (
            <p>Loading comments...</p>
          ) : (
            <div className="comments-list">
              {comments.length > 0 ? (
                comments.map((comment) => (
                  <CommentCard key={comment.id} comment={comment} />
                ))
              ) : (
                <p>No comments yet.</p>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
};