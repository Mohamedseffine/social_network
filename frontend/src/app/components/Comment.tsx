import Link from 'next/link';
import { getImageUrl } from '@/utils/api';


export const CommentCard = ({ comment }: { comment: Comment }) => {
  return (
    <div className="comment">
      <div className="comment-author">
        <img src={getImageUrl(comment.author_avatar)} alt="Author Avatar" className="user-avatar-small" />
        <Link href={`/users/${comment.user_id}`}>
          <span className="comment-author-name">{comment.author_first_name} {comment.author_last_name}</span>
        </Link>
      </div>
      <small className="comment-date">{new Date(comment.created_at).toLocaleString()}</small>
      <p className="comment-content">{comment.content}</p>
      {comment.image && (
        <img src={getImageUrl(comment.image)} alt="comment image" className="comment-image" />
      )}
    </div>
  );
};
