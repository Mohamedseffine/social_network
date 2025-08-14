import Link from 'next/link';
import { getImageUrl } from '@/utils/api';

export interface Comment {
  id: number;
  post_id: number;
  user_id: number;
  author_first_name: string;
  author_last_name: string;
  author_avatar: string;
  content: string;
  created_at: string;
}

export const CommentCard = ({ comment }: { comment: Comment }) => {
  return (
    <div className="comment">
      <div className="comment-author">
        <img src={getImageUrl(comment.author_avatar)} alt="Author Avatar" className="user-avatar-small" />
        <Link href={`/users/${comment.user_id}`}>
          <span className="comment-author-name">{comment.author_first_name} {comment.author_last_name}</span>
        </Link>
      </div>
      <p className="comment-content">{comment.content}</p>
      <small className="comment-date">{new Date(comment.created_at).toLocaleString()}</small>
    </div>
  );
};
